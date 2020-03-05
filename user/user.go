package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/TempleEight/spec-golang/user/dao"
	"github.com/TempleEight/spec-golang/user/util"
	valid "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

type Env struct {
	dao dao.Datastore
}

// Router generates a router for this service
func Router(env Env) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/user", env.userCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", env.userReadHandler).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", env.userUpdateHandler).Methods(http.MethodPut)
	r.HandleFunc("/user/{id}", env.userDeleteHandler).Methods(http.MethodDelete)
	r.Use(jsonMiddleware)
	return r
}

func main() {
	configPtr := flag.String("config", "/etc/user-service/config.json", "configuration filepath")
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
	env := Env{d}

	r := Router(env)
	log.Fatal(http.ListenAndServe(":80", r))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All responses are JSON, set header accordingly
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (env *Env) userCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req dao.UserCreateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	resp, err := env.dao.CreateUser(req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resp)
}

func (env *Env) userReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	user, err := env.dao.ReadUser(userID)
	if err != nil {
		switch err.(type) {
		case dao.ErrUserNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (env *Env) userUpdateHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	var req dao.UserUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	resp, err := env.dao.UpdateUser(userID, req)
	if err != nil {
		switch err.(type) {
		case dao.ErrUserNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(resp)
}

func (env *Env) userDeleteHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	err = env.dao.DeleteUser(userID)
	if err != nil {
		switch err.(type) {
		case dao.ErrUserNotFound:
			http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusNotFound)
		default:
			errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}
	json.NewEncoder(w).Encode(struct{}{})
}
