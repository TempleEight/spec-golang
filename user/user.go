package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	userDAO "github.com/TempleEight/spec-golang/user/dao"
	"github.com/TempleEight/spec-golang/user/utils"
	valid "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

var dao userDAO.DAO

func main() {
	configPtr := flag.String("config", "/etc/user-service/config.json", "configuration filepath")
	flag.Parse()

	// Require all struct fields by default
	valid.SetFieldsRequiredByDefault(true)

	dao = userDAO.DAO{}
	err := dao.Initialise(*configPtr)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/user", userCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", userGetHandler).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", userUpdateHandler).Methods(http.MethodPut)
	r.HandleFunc("/user/{id}", userDeleteHandler).Methods(http.MethodDelete)
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
	userID, err := utils.ExtractUserIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, utils.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	user, err := dao.GetUser(userID)
	if err != nil {
		switch err.(type) {
		case userDAO.ErrUserNotFound:
			http.Error(w, utils.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := utils.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(user)
}

func userCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req userDAO.UserCreateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	response, err := dao.CreateUser(req)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(response)
}

func userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ExtractUserIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, utils.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	var req userDAO.UserUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	err = dao.UpdateUser(userID, req)
	if err != nil {
		switch err.(type) {
		case userDAO.ErrUserNotFound:
			http.Error(w, utils.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := utils.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(struct{}{})
}

func userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ExtractUserIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, utils.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	err = dao.DeleteUser(userID)
	if err != nil {
		switch err.(type) {
		case userDAO.ErrUserNotFound:
			http.Error(w, utils.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := utils.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(struct{}{})
}
