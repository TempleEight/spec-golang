package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"

	authDAO "github.com/TempleEight/spec-golang/auth/dao"
	"github.com/TempleEight/spec-golang/auth/utils"
	valid "github.com/asaskevich/govalidator"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var dao authDAO.DAO

func main() {
	configPtr := flag.String("config", "/etc/auth-service/config.json", "configuration filepath")
	flag.Parse()

	// Require all struct fields by default
	valid.SetFieldsRequiredByDefault(true)

	config, err := utils.GetConfig(*configPtr)
	if err != nil {
		log.Fatal(err)
	}

	dao = authDAO.DAO{}
	err = dao.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/auth", authCreateHandler).Methods(http.MethodPost)
	r.HandleFunc("/auth", authGetHandler).Methods(http.MethodGet)
	log.Fatal(http.ListenAndServe(":82", r))
}

func authCreateHandler(w http.ResponseWriter, r *http.Request) {
	var req authDAO.PlaintextAuth
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

	hashedAuth := authDAO.HashedAuth{
		Email:    req.Email,
		Password: string(hashedPassword),
	}
	err = dao.CreateAuth(hashedAuth)
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Something went wrong: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	accessToken, err := createToken(req.Email, "TODO", "TODO")
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Could not create access token: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	response := authDAO.AuthResponse{
		AccessToken: accessToken,
	}
	json.NewEncoder(w).Encode(response)
}

func authGetHandler(w http.ResponseWriter, r *http.Request) {
	var req authDAO.PlaintextAuth
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

	auth, err := dao.GetAuth(req)
	if err != nil {
		switch err {
		case authDAO.ErrAuthNotFound:
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

	accessToken, err := createToken(req.Email, "TODO", "TODO")
	if err != nil {
		errMsg := utils.CreateErrorJSON(fmt.Sprintf("Could not create access token: %s", err.Error()))
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	response := authDAO.AuthResponse{
		AccessToken: accessToken,
	}
	json.NewEncoder(w).Encode(response)
}

// Create an access token with a 24 hour lifetime
func createToken(email string, issuer string, key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"iss":   issuer,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(key))
}
