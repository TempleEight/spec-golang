package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/TempleEight/spec-golang/user/dao"
	"github.com/TempleEight/spec-golang/user/metric"
	"github.com/TempleEight/spec-golang/user/util"
	valid "github.com/asaskevich/govalidator"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// env defines the environment that requests should be executed within
type env struct {
	dao  dao.Datastore
	hook Hook
}

// createUserRequest contains the client-provided information required to create a single user
type createUserRequest struct {
	Name string `valid:"type(string),required,stringlength(2|255)"`
}

// updateUserRequest contains the client-provided information required to update a single user, excluding ID
type updateUserRequest struct {
	Name string `valid:"type(string),required,stringlength(2|255)"`
}

// createUserResponse contains a newly created user to be returned to the client
type createUserResponse struct {
	ID   uuid.UUID
	Name string
}

// readUserResponse contains a single user to be returned to the client
type readUserResponse struct {
	ID   uuid.UUID
	Name string
}

// updateUserResponse contains a newly updated user to be returned to the client
type updateUserResponse struct {
	ID   uuid.UUID
	Name string
}

// router generates a router for this service
func defaultRouter(env *env) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/user", env.createUserHandler).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", env.readUserHandler).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", env.updateUserHandler).Methods(http.MethodPut)
	r.HandleFunc("/user/{id}", env.deleteUserHandler).Methods(http.MethodDelete)
	r.Use(jsonMiddleware)
	return r
}

func main() {
	// Prometheus metrics
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
	}()

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
	env := env{d, Hook{}}

	// Call into non-generated entry-point
	router := defaultRouter(&env)
	env.setup(router)

	log.Fatal(http.ListenAndServe(":80", router))
}

func jsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All responses are JSON, set header accordingly
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (env *env) createUserHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := util.ExtractAuthIDFromRequest(r.Header)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Could not authorize request: %s", err.Error()))
		http.Error(w, errMsg, http.StatusUnauthorized)
		metric.RequestFailure.WithLabelValues(metric.RequestCreate, strconv.Itoa(http.StatusUnauthorized)).Inc()
		return
	}

	var req createUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		metric.RequestFailure.WithLabelValues(metric.RequestCreate, strconv.Itoa(http.StatusBadRequest)).Inc()
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Invalid request parameters: %s", err.Error()))
		http.Error(w, errMsg, http.StatusBadRequest)
		metric.RequestFailure.WithLabelValues(metric.RequestCreate, strconv.Itoa(http.StatusBadRequest)).Inc()
		return
	}

	input := dao.CreateUserInput{
		ID:   auth.ID,
		Name: req.Name,
	}

	for _, hook := range env.hook.beforeCreateHooks {
		err := (*hook)(env, req, &input)
		if err != nil {
			errMsg := util.CreateErrorJSON(err.Error())
			http.Error(w, errMsg, err.statusCode)
			metric.RequestFailure.WithLabelValues(metric.RequestCreate, strconv.Itoa(err.statusCode)).Inc()
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestCreate))
	user, err := env.dao.CreateUser(input)
	timer.ObserveDuration()

	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		metric.RequestFailure.WithLabelValues(metric.RequestCreate, strconv.Itoa(http.StatusInternalServerError)).Inc()
		return
	}

	for _, hook := range env.hook.afterCreateHooks {
		err := (*hook)(env, user)
		if err != nil {
			errMsg := util.CreateErrorJSON(err.Error())
			http.Error(w, errMsg, err.statusCode)
			metric.RequestFailure.WithLabelValues(metric.RequestCreate, strconv.Itoa(err.statusCode)).Inc()
			return
		}
	}

	json.NewEncoder(w).Encode(createUserResponse{
		ID:   user.ID,
		Name: user.Name,
	})
	metric.RequestSuccess.WithLabelValues(metric.RequestCreate).Inc()
}

func (env *env) readUserHandler(w http.ResponseWriter, r *http.Request) {
	_, err := util.ExtractAuthIDFromRequest(r.Header)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Could not authorize request: %s", err.Error()))
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	input := dao.ReadUserInput{
		ID: userID,
	}

	for _, hook := range env.hook.beforeReadHooks {
		err := (*hook)(env, &input)
		if err != nil {
			errMsg := util.CreateErrorJSON(err.Error())
			http.Error(w, errMsg, err.statusCode)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestRead))
	user, err := env.dao.ReadUser(input)
	timer.ObserveDuration()

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

	for _, hook := range env.hook.afterReadHooks {
		err := (*hook)(env, user)
		if err != nil {
			errMsg := util.CreateErrorJSON(err.Error())
			http.Error(w, errMsg, err.statusCode)
			return
		}
	}

	json.NewEncoder(w).Encode(readUserResponse{
		ID:   user.ID,
		Name: user.Name,
	})
	metric.RequestSuccess.WithLabelValues(metric.RequestRead).Inc()
}

func (env *env) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := util.ExtractAuthIDFromRequest(r.Header)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Could not authorize request: %s", err.Error()))
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	// Only the auth that created the user can update it
	if auth.ID != userID {
		errMsg := util.CreateErrorJSON("Not authorized to make request")
		http.Error(w, errMsg, http.StatusUnauthorized)
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

	input := dao.UpdateUserInput{
		ID:   userID,
		Name: req.Name,
	}

	for _, hook := range env.hook.beforeUpdateHooks {
		err := (*hook)(env, req, &input)
		if err != nil {
			errMsg := util.CreateErrorJSON(err.Error())
			http.Error(w, errMsg, err.statusCode)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestUpdate))
	user, err := env.dao.UpdateUser(input)
	timer.ObserveDuration()

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

	for _, hook := range env.hook.afterUpdateHooks {
		err := (*hook)(env, user)
		if err != nil {
			errMsg := util.CreateErrorJSON(err.Error())
			http.Error(w, errMsg, err.statusCode)
			return
		}
	}

	json.NewEncoder(w).Encode(updateUserResponse{
		ID:   user.ID,
		Name: user.Name,
	})
	metric.RequestSuccess.WithLabelValues(metric.RequestUpdate).Inc()
}

func (env *env) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := util.ExtractAuthIDFromRequest(r.Header)
	if err != nil {
		errMsg := util.CreateErrorJSON(fmt.Sprintf("Could not authorize request: %s", err.Error()))
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		http.Error(w, util.CreateErrorJSON(err.Error()), http.StatusBadRequest)
		return
	}

	// Only the auth that created the user can delete it
	if auth.ID != userID {
		errMsg := util.CreateErrorJSON("Not authorized to make request")
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	input := dao.DeleteUserInput{
		ID: userID,
	}

	for _, hook := range env.hook.beforeDeleteHooks {
		err := (*hook)(env, &input)
		if err != nil {
			errMsg := util.CreateErrorJSON(err.Error())
			http.Error(w, errMsg, err.statusCode)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestDelete))
	err = env.dao.DeleteUser(input)
	timer.ObserveDuration()

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

	for _, hook := range env.hook.afterDeleteHooks {
		err := (*hook)(env)
		if err != nil {
			errMsg := util.CreateErrorJSON(err.Error())
			http.Error(w, errMsg, err.statusCode)
			return
		}
	}

	json.NewEncoder(w).Encode(struct{}{})
	metric.RequestSuccess.WithLabelValues(metric.RequestDelete).Inc()
}
