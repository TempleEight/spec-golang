// +build it

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/TempleEight/spec-golang/auth/comm"
	"github.com/TempleEight/spec-golang/auth/dao"
	"github.com/TempleEight/spec-golang/auth/util"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

var environment env

func TestMain(m *testing.M) {
	config, err := util.GetConfig("/etc/auth-service/config.json")
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

	environment = env{d, c, jwtCredential}

	os.Exit(m.Run())
}

func TestIntegrationAuth(t *testing.T) {
	// Create a single auth
	res, err := makeRequest(environment, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	// Validate the JWT returned
	var decoded map[string]string
	err = json.Unmarshal([]byte(res.Body.String()), &decoded)
	if err != nil {
		t.Fatalf("Could not decode json: %s", err.Error())
	}

	rawToken, ok := decoded["AccessToken"]
	if !ok {
		t.Fatalf("Token doesn't contain an access token: %s", err.Error())
	}

	token, _, err := new(jwt.Parser).ParseUnverified(rawToken, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("Could not decode JWT: %s", err.Error())
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Could not decode claims")
	}

	id, ok := claims["id"]
	if !ok {
		t.Fatalf("Claims doesn't contain an ID key")
	}

	_, err = uuid.Parse(id.(string))
	if err != nil {
		t.Fatalf("ID is not a valid UUID")
	}

	iss, ok := claims["iss"]
	if !ok {
		t.Fatalf("Claims doesn't contain an iss key")
	}

	if iss.(string) != environment.jwtCredential.Key {
		t.Fatalf("iss is incorrect: found %v, wanted %s", iss, environment.jwtCredential.Key)
	}

	// Access that same auth
	res, err = makeRequest(environment, http.MethodPost, "/auth/login", `{"email": "jay@test.com", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	// Validate the JWT returned
	err = json.Unmarshal([]byte(res.Body.String()), &decoded)
	if err != nil {
		t.Fatalf("Could not decode json: %s", err.Error())
	}

	rawToken, ok = decoded["AccessToken"]
	if !ok {
		t.Fatalf("Token doesn't contain an access token: %s", err.Error())
	}

	token, _, err = new(jwt.Parser).ParseUnverified(rawToken, jwt.MapClaims{})
	if err != nil {
		t.Fatalf("Could not decode JWT: %s", err.Error())
	}

	claims, ok = token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("Could not decode claims")
	}

	id, ok = claims["id"]
	if !ok {
		t.Fatalf("Claims doesn't contain an ID key")
	}

	_, err = uuid.Parse(id.(string))
	if err != nil {
		t.Fatalf("ID is not a valid UUID")
	}

	iss, ok = claims["iss"]
	if !ok {
		t.Fatalf("Claims doesn't contain an iss key")
	}

	if iss.(string) != environment.jwtCredential.Key {
		t.Fatalf("iss is incorrect: found %v, wanted %s", iss, environment.jwtCredential.Key)
	}
}
