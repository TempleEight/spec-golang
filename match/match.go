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

type Env struct {
	dao  dao.Datastore
	comm comm.Comm
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

	env := Env{d, c}

	r := mux.NewRouter()
	// Mux directs to first matching route, i.e. the order matters
	r.HandleFunc("/match/all", env.matchListHandler).Methods(http.MethodGet)
	r.HandleFunc("/match", env.matchCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/match/{id}", env.matchReadHandler).Methods(http.MethodGet)
	r.HandleFunc("/match/{id}", env.matchUpdateHandler).Methods(http.MethodPut)
	r.HandleFunc("/match/{id}", env.matchDeleteHandler).Methods(http.MethodDelete)
	r.Use(jsonMiddleware)

	log.Fatal(http.ListenAndServe(":81", r))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All responses are JSON, set header accordingly
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (env *Env) matchListHandler(w http.ResponseWriter, r *http.Request) {
	matchList, err := env.dao.ListMatch()
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matchList)
}

func (env *Env) matchCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req dao.MatchCreateRequest
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

	resp, err := env.dao.CreateMatch(req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp)
}

func (env *Env) matchReadHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	match, err := env.dao.ReadMatch(matchID)
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

	json.NewEncoder(w).Encode(match)
}

func (env *Env) matchUpdateHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	var req dao.MatchUpdateRequest
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

	resp, err := env.dao.UpdateMatch(matchID, req)
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
	json.NewEncoder(w).Encode(resp)
}

func (env *Env) matchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	err = env.dao.DeleteMatch(matchID)
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
