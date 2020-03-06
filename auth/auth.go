package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"

	"github.com/TempleEight/spec-golang/auth/comm"
	"github.com/TempleEight/spec-golang/auth/dao"
	"github.com/TempleEight/spec-golang/auth/utils"
	valid "github.com/asaskevich/govalidator"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

type Env struct {
	dao           dao.Datastore
	comm          comm.Comm
	jwtCredential *comm.JWTCredential
}

func Router(env Env) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/auth", env.authCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/auth", env.authReadHandler).Methods(http.MethodGet)
	return r
}

func main() {
	configPtr := flag.String("config", "/etc/auth-service/config.json", "configuration filepath")
	flag.Parse()

	// Require all struct fields by default
	valid.SetFieldsRequiredByDefault(true)

	config, err := utils.GetConfig(*configPtr)
	if err != nil {
		log.Fatal(err)
	}

	d, err := dao.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	c := comm.Init(config)
	jwtCredential, err := c.CreateJWTCredential()
	if err != nil {
		log.Fatal(err)
	}

	env := Env{d, c, jwtCredential}

	log.Fatal(http.ListenAndServe(":82", Router(env)))
}

func (env *Env) authCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req dao.AuthCreateRequest
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

	// Hash and salt the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Could not hash password: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	hashedAuth := dao.AuthCreateRequest{
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	auth, err := env.dao.CreateAuth(hashedAuth)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	accessToken, err := createToken(auth.Id, env.jwtCredential.Key, env.jwtCredential.Secret)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Could not create access token: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	response := dao.AuthCreateResponse{
		AccessToken: accessToken,
	}
	json.NewEncoder(w).Encode(response)
}

func (env *Env) authReadHandler(w http.ResponseWriter, r *http.Request) {
	var req dao.AuthReadRequest
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

	auth, err := env.dao.ReadAuth(req)
	if err != nil {
		switch err {
		case dao.ErrAuthNotFound:
			errMsg := utils.CreateErrorJSON(fmt.Sprintf("Invalid email or password"))
			http.Error(w, errMsg, http.StatusUnauthorized)
		default:
			errMsg := utils.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
			http.Error(w, errMsg, http.StatusInternalServerError)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(auth.Password), []byte(req.Password))
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Invalid email or password"))
		http.Error(w, errMsg, http.StatusUnauthorized)
		return
	}

	accessToken, err := createToken(auth.Id, env.jwtCredential.Key, env.jwtCredential.Secret)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Could not create access token: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	response := dao.AuthReadResponse{
		AccessToken: accessToken,
	}
	json.NewEncoder(w).Encode(response)
}

// Create an access token with a 24 hour lifetime
func createToken(id int, issuer string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"iss": issuer,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(secret))
}
