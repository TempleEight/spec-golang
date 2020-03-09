package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/TempleEight/spec-golang/match/comm"
	"github.com/TempleEight/spec-golang/match/dao"
	"github.com/TempleEight/spec-golang/match/util"
	valid "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

// env defines the environment that requests should be executed within
type env struct {
	dao  dao.Datastore
	comm comm.Comm
}

// createMatchRequest contains the client-provided information required to create a single match
type createMatchRequest struct {
	UserOne *int64 `valid:"-"`
	UserTwo *int64 `valid:"-"`
}

// updateMatchRequest contains the client-provided information required to update a single match, excluding ID
type updateMatchRequest struct {
	UserOne *int64 `valid:"-"`
	UserTwo *int64 `valid:"-"`
}

// listMatchResponse contains a single match list to be returned to the client
type listMatchResponse struct {
	MatchList []readMatchResponse
}

// createMatchResponse contains a newly created match to be returned to the client
type createMatchResponse struct {
	ID        int64
	UserOne   int64
	UserTwo   int64
	MatchedOn string
}

// readMatchResponse contains a single match to be returned to the client
type readMatchResponse struct {
	ID        int64
	UserOne   int64
	UserTwo   int64
	MatchedOn string
}

// updateMatchResponse contains a newly updated match to be returned to the client
type updateMatchResponse struct {
	ID        int64
	UserOne   int64
	UserTwo   int64
	MatchedOn string
}

// router generates a router for this service
func (env *env) router() *mux.Router {
	r := mux.NewRouter()
	// Mux directs to first matching route, i.e. the order matters
	r.HandleFunc("/match/all", env.listMatchHandler).Methods(http.MethodGet)
	r.HandleFunc("/match", env.createMatchHandler).Methods(http.MethodPost)
	r.HandleFunc("/match/{id}", env.readMatchHandler).Methods(http.MethodGet)
	r.HandleFunc("/match/{id}", env.updateMatchHandler).Methods(http.MethodPut)
	r.HandleFunc("/match/{id}", env.deleteMatchHandler).Methods(http.MethodDelete)
	r.Use(jsonMiddleware)
	return r
}

func main() {
	configPtr := flag.String("config", "/etc/match-service/config.json", "configuration filepath")
	flag.Parse()

	// Require all struct fields by default
	valid.SetFieldsRequiredByDefault(true)

	config, err := util.GetConfig(*configPtr)
	if err != nil {
		log.Fatal(err)
	}

	d, err := dao.Init(config)
	if err != nil {
		log.Fatal(err)
	}
	c := comm.Init(config)

	env := env{d, c}

	log.Fatal(http.ListenAndServe(":81", env.router()))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All responses are JSON, set header accordingly
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (env *env) listMatchHandler(w http.ResponseWriter, r *http.Request) {
	matchList, err := env.dao.ListMatch()
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	matchListResp := listMatchResponse{
		MatchList: make([]readMatchResponse, 0),
	}
	for _, match := range *matchList {
		matchListResp.MatchList = append(matchListResp.MatchList, readMatchResponse{
			ID:        match.ID,
			UserOne:   match.UserOne,
			UserTwo:   match.UserTwo,
			MatchedOn: match.MatchedOn,
		})
	}

	json.NewEncoder(w).Encode(matchListResp)
}

func (env *env) createMatchHandler(w http.ResponseWriter, r *http.Request) {
	var req createMatchRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if req.UserOne == nil || req.UserTwo == nil {
		errMsg := util.CreateErrorJSON("Missing request parameter(s)")
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	userOneValid, err := env.comm.CheckUser(*req.UserOne)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Unable to reach user service: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if !userOneValid {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Unknown User: %d", *req.UserOne))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	userTwoValid, err := env.comm.CheckUser(*req.UserTwo)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Unable to reach user service: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if !userTwoValid {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Unknown User: %d", *req.UserTwo))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	match, err := env.dao.CreateMatch(dao.CreateMatchInput{
		UserOne: *req.UserOne,
		UserTwo: *req.UserTwo,
	})
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(createMatchResponse{
		ID:        match.ID,
		UserOne:   match.UserOne,
		UserTwo:   match.UserTwo,
		MatchedOn: match.MatchedOn,
	})
}

func (env *env) readMatchHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	match, err := env.dao.ReadMatch(dao.ReadMatchInput{
		ID: matchID,
	})
	if err != nil {
		switch err.(type) {
		case dao.ErrMatchNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(readMatchResponse{
		ID:        match.ID,
		UserOne:   match.UserOne,
		UserTwo:   match.UserTwo,
		MatchedOn: match.MatchedOn,
	})
}

func (env *env) updateMatchHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	var req updateMatchRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	if req.UserOne == nil || req.UserTwo == nil {
		errMsg := util.CreateErrorJSON("Missing request parameter")
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	userOneValid, err := env.comm.CheckUser(*req.UserOne)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Unable to reach %s service: %s", "user", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if !userOneValid {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Unknown User: %d", *req.UserOne))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	userTwoValid, err := env.comm.CheckUser(*req.UserTwo)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Unable to reach %s service: %s", "user", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if !userTwoValid {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Unknown User: %d", *req.UserTwo))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	match, err := env.dao.UpdateMatch(dao.UpdateMatchInput{
		ID:      matchID,
		UserOne: *req.UserOne,
		UserTwo: *req.UserTwo,
	})
	if err != nil {
		switch err.(type) {
		case dao.ErrMatchNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(updateMatchResponse{
		ID:        match.ID,
		UserOne:   match.UserOne,
		UserTwo:   match.UserTwo,
		MatchedOn: match.MatchedOn,
	})
}

func (env *env) deleteMatchHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	err = env.dao.DeleteMatch(dao.DeleteMatchInput{
		ID: matchID,
	})
	if err != nil {
		switch err.(type) {
		case dao.ErrMatchNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(struct{}{})
}
