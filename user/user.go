package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/TempleEight/spec-golang/user/dao"
	valid "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

func main() {
	// Require all struct fields by default
	valid.SetFieldsRequiredByDefault(true)

	r := mux.NewRouter()
	r.HandleFunc("/user", userCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", userGetHandler).Methods(http.MethodGet)
	r.Use(jsonMiddleware)

	log.Fatal(http.ListenAndServe(":80", r))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All responses are JSON, set header accordingly
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func userGetHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["id"]
	if len(userIDStr) == 0 {
		http.Error(w, CreateErrorJSON("No user ID provided"), http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, CreateErrorJSON("Invalid user ID provided"), http.StatusBadRequest)
		return
	}

	user, err := dao.GetUser(userID)
	if err != nil {
		switch err.(type) {
		case dao.ErrUserNotFound:
			http.Error(w, CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(user)
}

func userCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req dao.UserCreateRequest
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

	response, err := dao.CreateUser(req)
	if err != nil {
		errMsg := CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}
