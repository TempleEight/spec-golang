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

// env defines the environment that requests should be executed within
type env struct {
	dao dao.Datastore
}

// createUserRequest contains the information required to create a new user
type createUserRequest struct {
	Name string `valid:"type(string),required,stringlength(2|255)"`
}

// updateUserRequest contains all the information about an existing user
type updateUserRequest struct {
	Name string `valid:"type(string),required,stringlength(2|255)"`
}

// createUserResponse contains the information about the newly created user
type createUserResponse struct {
	ID   int64
	Name string
}

// readUserResponse returns all the information stored about a single user
type readUserResponse struct {
	ID   int64
	Name string
}

// updateUserResponse contains the information about the newly updated user
type updateUserResponse struct {
	ID   int64
	Name string
}

// router generates a router for this service
func (env *env) router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/user", env.createUserHandler).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", env.readUserHandler).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", env.updateUserHandler).Methods(http.MethodPut)
	r.HandleFunc("/user/{id}", env.deleteUserHandler).Methods(http.MethodDelete)
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
	env := env{d}

	log.Fatal(http.ListenAndServe(":80", env.router()))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All responses are JSON, set header accordingly
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (env *env) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
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

	resp, err := env.dao.CreateUser(dao.CreateUserInput{
		Name: req.Name,
	})
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(createUserResponse{
		ID:   resp.ID,
		Name: resp.Name,
	})
}

func (env *env) readUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	user, err := env.dao.ReadUser(dao.ReadUserInput{
		ID: userID,
	})
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
	json.NewEncoder(w).Encode(readUserResponse{
		ID:   user.ID,
		Name: user.Name,
	})
}

func (env *env) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	var req updateUserRequest
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

	resp, err := env.dao.UpdateUser(dao.UpdateUserInput{
		ID:   userID,
		Name: req.Name,
	})
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

	json.NewEncoder(w).Encode(updateUserResponse{
		ID:   resp.ID,
		Name: resp.Name,
	})
}

func (env *env) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	err = env.dao.DeleteUser(dao.DeleteUserInput{
		ID: userID,
	})
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
