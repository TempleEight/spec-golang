package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	matchComm "github.com/TempleEight/spec-golang/match/comm"
	matchDAO "github.com/TempleEight/spec-golang/match/dao"
	"github.com/TempleEight/spec-golang/match/util"
	valid "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

var dao matchDAO.DAO
var comm matchComm.Handler

func main() {
	configPtr := flag.String("config", "/etc/match-service/config.json", "configuration filepath")
	flag.Parse()

	// Require all struct fields by default
	valid.SetFieldsRequiredByDefault(true)

	config, err := util.GetConfig(*configPtr)
	if err != nil {
		log.Fatal(err)
	}

	dao = matchDAO.DAO{}
	err = dao.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	comm = matchComm.Handler{}
	comm.Init(config)

	r := mux.NewRouter()
	// Mux directs to first matching route, i.e. the order matters
	r.HandleFunc("/match/all", matchListHandler).Methods(http.MethodGet)
	r.HandleFunc("/match", matchCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/match/{id}", matchReadHandler).Methods(http.MethodGet)
	r.HandleFunc("/match/{id}", matchUpdateHandler).Methods(http.MethodPut)
	r.HandleFunc("/match/{id}", matchDeleteHandler).Methods(http.MethodDelete)
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

func matchListHandler(w http.ResponseWriter, r *http.Request) {
	matchList, err := dao.ListMatch()
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matchList)
}

func matchCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req matchDAO.MatchCreateRequest
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

	userOneValid, err := comm.CheckUser(*req.UserOne)
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

	userTwoValid, err := comm.CheckUser(*req.UserTwo)
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

	resp, err := dao.CreateMatch(req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp)
}

func matchReadHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	match, err := dao.ReadMatch(matchID)
	if err != nil {
		switch err.(type) {
		case matchDAO.ErrMatchNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(match)
}

func matchUpdateHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	var req matchDAO.MatchUpdateRequest
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

	userOneValid, err := comm.CheckUser(*req.UserOne)
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

	userTwoValid, err := comm.CheckUser(*req.UserTwo)
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

	resp, err := dao.UpdateMatch(matchID, req)
	if err != nil {
		switch err.(type) {
		case matchDAO.ErrMatchNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(resp)
}

func matchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	matchID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	err = dao.DeleteMatch(matchID)
	if err != nil {
		switch err.(type) {
		case matchDAO.ErrMatchNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(struct{}{})
}
