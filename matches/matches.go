package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/TempleEight/spec-golang/matches/dao"
	valid "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/matches/{id1}", matchListHandler).Methods(http.MethodGet)
	r.HandleFunc("/match/{id}", matchGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/match", matchCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/match", matchDeleteHandler).Methods(http.MethodDelete)
	r.HandleFunc("/match", matchUpdateHandler).Methods(http.MethodPatch)

	log.Fatal(http.ListenAndServe(":80", r))
}

func matchGetHandler(w http.ResponseWriter, r *http.Request) {
	matchIDStr := mux.Vars(r)["id"]
	if len(matchIDStr) == 0 {
		http.Error(w, CreateErrorJSON("No match ID provided"), http.StatusBadRequest)
		return
	}

	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		http.Error(w, CreateErrorJSON("Invalid match ID provided"), http.StatusBadRequest)
		return
	}

	match, err := dao.GetMatch(matchID)
	if err != nil {
		switch err.(type) {
		case dao.ErrMatchNotFound:
			http.Error(w, CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(match)
}

func matchListHandler(w http.ResponseWriter, r *http.Request) {
	userOneStr := mux.Vars(r)["id1"]
	if len(userOneStr) == 0 {
		http.Error(w, CreateErrorJSON("No match ID provided"), http.StatusBadRequest)
		return
	}

	userOne, err := strconv.ParseInt(userOneStr, 10, 64)
	if err != nil {
		http.Error(w, CreateErrorJSON("Invalid User ID provided"), http.StatusBadRequest)
		return
	}

	matches, err := dao.ListMatch(userOne)
	if err != nil {
		switch err.(type) {
		case dao.ErrMatchNotFound:
			http.Error(w, CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(matches)
}

func matchCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req dao.MatchCreateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	response, err := dao.CreateMatch(req)
	if err != nil {
		errMsg := CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}

func matchDeleteHandler(w http.ResponseWriter, r *http.Request) {
	matchIDStr := mux.Vars(r)["id"]
	if len(matchIDStr) == 0 {
		http.Error(w, CreateErrorJSON("No match ID provided"), http.StatusBadRequest)
		return
	}

	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		http.Error(w, CreateErrorJSON("Invalid match ID provided"), http.StatusBadRequest)
		return
	}

	err = dao.DeleteMatch(matchID)
	if err != nil {
		switch err.(type) {
		case dao.ErrMatchNotFound:
			http.Error(w, CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(struct{}{})
}

func matchUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var req dao.MatchUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	err = dao.UpdateMatch(req)
	if err != nil {
		switch err.(type) {
		case dao.ErrMatchNotFound:
			http.Error(w, CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(struct{}{})
}
