package main

import (
	"encoding/base64"
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

// createPictureRequest contains the client-provided information required to create a single picture
type createPictureRequest struct {
	Img string `valid:"-"`
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

// createPictureResponse contains a newly created picture to be returned to the client
type createPictureResponse struct {
	ID uuid.UUID
}

// readPictureResponse contains a single picture to be returned to the client
type readPictureResponse struct {
	ID  uuid.UUID
	Img string
}

// router generates a router for this service
func defaultRouter(env *env) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/user", env.createUserHandler).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}", env.readUserHandler).Methods(http.MethodGet)
	r.HandleFunc("/user/{id}", env.updateUserHandler).Methods(http.MethodPut)
	r.HandleFunc("/user/{id}", env.deleteUserHandler).Methods(http.MethodDelete)
	r.HandleFunc("/user/{id}/picture", env.createPictureHandler).Methods(http.MethodPost)
	r.HandleFunc("/user/{id}/picture/{picture_id}", env.readPictureHandler).Methods(http.MethodGet)
	r.Use(jsonMiddleware)
	return r
}

// respondWithError responds to a HTTP request with a JSON error response
func respondWithError(w http.ResponseWriter, err string, statusCode int, requestType string) {
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, util.CreateErrorJSON(err))
	metric.RequestFailure.WithLabelValues(requestType, strconv.Itoa(statusCode)).Inc()
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

	// Prometheus metrics
	promPort, ok := config.Ports["prometheus"]
	if !ok {
		log.Fatal("A port for the key prometheus was not found")
	}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", promPort), nil)
	}()

	d, err := dao.Init(config)
	if err != nil {
		log.Fatal(err)
	}
	env := env{d, Hook{}}

	// Call into non-generated entry-point
	router := defaultRouter(&env)
	env.setup(router)

	servicePort, ok := config.Ports["service"]
	if !ok {
		log.Fatal("A port for the key service was not found")
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", servicePort), router))
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
		respondWithError(w, fmt.Sprintf("Could not authorize request: %s", err.Error()), http.StatusUnauthorized, metric.RequestCreate)
		return
	}

	var req createUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request parameters: %s", err.Error()), http.StatusBadRequest, metric.RequestCreate)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request parameters: %s", err.Error()), http.StatusBadRequest, metric.RequestCreate)
		return
	}

	input := dao.CreateUserInput{
		ID:   auth.ID,
		Name: req.Name,
	}

	for _, hook := range env.hook.beforeCreateHooks {
		err := (*hook)(env, req, &input)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestCreate)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestCreate))
	user, err := env.dao.CreateUser(input)
	timer.ObserveDuration()

	if err != nil {
		respondWithError(w, fmt.Sprintf("Something went wrong: %s", err.Error()), http.StatusInternalServerError, metric.RequestCreate)
		return
	}

	for _, hook := range env.hook.afterCreateHooks {
		err := (*hook)(env, user)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestCreate)
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
		respondWithError(w, fmt.Sprintf("Could not authorize request: %s", err.Error()), http.StatusUnauthorized, metric.RequestRead)
		return
	}

	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest, metric.RequestRead)
		return
	}

	input := dao.ReadUserInput{
		ID: userID,
	}

	for _, hook := range env.hook.beforeReadHooks {
		err := (*hook)(env, &input)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestRead)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestRead))
	user, err := env.dao.ReadUser(input)
	timer.ObserveDuration()

	if err != nil {
		switch err.(type) {
		case dao.ErrUserNotFound:
			respondWithError(w, err.Error(), http.StatusNotFound, metric.RequestRead)
		default:
			respondWithError(w, fmt.Sprintf("Something went wrong: %s", err.Error()), http.StatusInternalServerError, metric.RequestRead)
		}
		return
	}

	for _, hook := range env.hook.afterReadHooks {
		err := (*hook)(env, user)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestRead)
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
		respondWithError(w, fmt.Sprintf("Could not authorize request: %s", err.Error()), http.StatusUnauthorized, metric.RequestUpdate)
		return
	}

	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest, metric.RequestUpdate)
		return
	}

	// Only the auth that created the user can update it
	if auth.ID != userID {
		respondWithError(w, "Not authorized to make request", http.StatusUnauthorized, metric.RequestUpdate)
		return
	}

	var req updateUserRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request parameters: %s", err.Error()), http.StatusBadRequest, metric.RequestUpdate)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request parameters: %s", err.Error()), http.StatusBadRequest, metric.RequestUpdate)
		return
	}

	input := dao.UpdateUserInput{
		ID:   userID,
		Name: req.Name,
	}

	for _, hook := range env.hook.beforeUpdateHooks {
		err := (*hook)(env, req, &input)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestUpdate)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestUpdate))
	user, err := env.dao.UpdateUser(input)
	timer.ObserveDuration()

	if err != nil {
		switch err.(type) {
		case dao.ErrUserNotFound:
			respondWithError(w, err.Error(), http.StatusNotFound, metric.RequestUpdate)
		default:
			respondWithError(w, fmt.Sprintf("Something went wrong: %s", err.Error()), http.StatusInternalServerError, metric.RequestUpdate)
		}
		return
	}

	for _, hook := range env.hook.afterUpdateHooks {
		err := (*hook)(env, user)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestUpdate)
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
		respondWithError(w, fmt.Sprintf("Could not authorize request: %s", err.Error()), http.StatusUnauthorized, metric.RequestDelete)
		return
	}

	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest, metric.RequestDelete)
		return
	}

	// Only the auth that created the user can delete it
	if auth.ID != userID {
		respondWithError(w, "Not authorized to make request", http.StatusUnauthorized, metric.RequestDelete)
		return
	}

	input := dao.DeleteUserInput{
		ID: userID,
	}

	for _, hook := range env.hook.beforeDeleteHooks {
		err := (*hook)(env, &input)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestDelete)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestDelete))
	err = env.dao.DeleteUser(input)
	timer.ObserveDuration()

	if err != nil {
		switch err.(type) {
		case dao.ErrUserNotFound:
			respondWithError(w, err.Error(), http.StatusNotFound, metric.RequestDelete)
		default:
			respondWithError(w, fmt.Sprintf("Something went wrong: %s", err.Error()), http.StatusInternalServerError, metric.RequestDelete)
		}
		return
	}

	for _, hook := range env.hook.afterDeleteHooks {
		err := (*hook)(env)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestDelete)
			return
		}
	}

	json.NewEncoder(w).Encode(struct{}{})
	metric.RequestSuccess.WithLabelValues(metric.RequestDelete).Inc()
}

func (env *env) createPictureHandler(w http.ResponseWriter, r *http.Request) {
	auth, err := util.ExtractAuthIDFromRequest(r.Header)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Could not authorize request: %s", err.Error()), http.StatusUnauthorized, metric.RequestCreatePicture)
		return
	}

	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest, metric.RequestCreatePicture)
		return
	}

	// Only the auth that created the user can create a single picture
	if auth.ID != userID {
		respondWithError(w, "Not authorized to make request", http.StatusUnauthorized, metric.RequestCreatePicture)
		return
	}

	var req createPictureRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request parameters: %s", err.Error()), http.StatusBadRequest, metric.RequestCreatePicture)
		return
	}

	_, err = valid.ValidateStruct(req)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request parameters: %s", err.Error()), http.StatusBadRequest, metric.RequestCreatePicture)
		return
	}

	decoded, err := base64.StdEncoding.DecodeString(req.Img)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Invalid request parameters: %s", err.Error()), http.StatusBadRequest, metric.RequestCreatePicture)
		return
	}

	uuid, err := uuid.NewUUID()
	if err != nil {
		respondWithError(w, fmt.Sprintf("Could not create UUID: %s", err.Error()), http.StatusInternalServerError, metric.RequestCreatePicture)
		return
	}

	input := dao.CreatePictureInput{
		ID:     uuid,
		UserID: auth.ID,
		Img:    decoded,
	}

	for _, hook := range env.hook.beforeCreatePictureHooks {
		err := (*hook)(env, req, &input)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestCreatePicture)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestCreatePicture))
	picture, err := env.dao.CreatePicture(input)
	timer.ObserveDuration()

	if err != nil {
		switch err.(type) {
		case dao.ErrUserNotFound:
			respondWithError(w, err.Error(), http.StatusNotFound, metric.RequestCreatePicture)
		default:
			respondWithError(w, fmt.Sprintf("Something went wrong: %s", err.Error()), http.StatusInternalServerError, metric.RequestCreatePicture)
		}
		return
	}

	for _, hook := range env.hook.afterCreatePictureHooks {
		err := (*hook)(env, picture)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestCreatePicture)
			return
		}
	}

	json.NewEncoder(w).Encode(createPictureResponse{
		ID: picture.ID,
	})
	metric.RequestSuccess.WithLabelValues(metric.RequestCreatePicture).Inc()
}

func (env *env) readPictureHandler(w http.ResponseWriter, r *http.Request) {
	_, err := util.ExtractAuthIDFromRequest(r.Header)
	if err != nil {
		respondWithError(w, fmt.Sprintf("Could not authorize request: %s", err.Error()), http.StatusUnauthorized, metric.RequestReadPicture)
		return
	}

	userID, err := util.ExtractIDFromRequest(mux.Vars(r))
	if err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest, metric.RequestReadPicture)
		return
	}

	pictureID, err := util.ExtractPictureIDFromRequest(mux.Vars(r))
	if err != nil {
		respondWithError(w, err.Error(), http.StatusBadRequest, metric.RequestReadPicture)
		return
	}

	input := dao.ReadPictureInput{
		ID:     pictureID,
		UserID: userID,
	}

	for _, hook := range env.hook.beforeReadPictureHooks {
		err := (*hook)(env, &input)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestReadPicture)
			return
		}
	}

	timer := prometheus.NewTimer(metric.DatabaseRequestDuration.WithLabelValues(metric.RequestReadPicture))
	picture, err := env.dao.ReadPicture(input)
	timer.ObserveDuration()

	if err != nil {
		switch err.(type) {
		case dao.ErrPictureNotFound:
			respondWithError(w, err.Error(), http.StatusNotFound, metric.RequestReadPicture)
		default:
			respondWithError(w, fmt.Sprintf("Something went wrong: %s", err.Error()), http.StatusInternalServerError, metric.RequestReadPicture)
		}
		return
	}

	for _, hook := range env.hook.afterReadPictureHooks {
		err := (*hook)(env, picture)
		if err != nil {
			respondWithError(w, err.Error(), err.statusCode, metric.RequestReadPicture)
			return
		}
	}

	json.NewEncoder(w).Encode(readPictureResponse{
		ID:  picture.ID,
		Img: base64.StdEncoding.EncodeToString(picture.Img),
	})
	metric.RequestSuccess.WithLabelValues(metric.RequestReadPicture).Inc()
}
