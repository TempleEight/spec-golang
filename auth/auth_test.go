package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/auth/comm"
	"github.com/TempleEight/spec-golang/auth/dao"
	"github.com/dgrijalva/jwt-go"
)

type MockAuth struct {
	ID       int
	Email    string
	Password string
}

type MockDAO struct {
	AuthList []MockAuth
}

type MockComm struct{}

func (md *MockDAO) CreateAuth(request dao.AuthCreateRequest) (*dao.Auth, error) {
	mockAuth := MockAuth{len(md.AuthList), request.Email, request.Password}
	md.AuthList = append(md.AuthList, mockAuth)
	return &dao.Auth{
		Id:       mockAuth.ID,
		Email:    mockAuth.Email,
		Password: mockAuth.Password,
	}, nil
}

func (md *MockDAO) ReadAuth(request dao.AuthReadRequest) (*dao.Auth, error) {
	for _, auth := range md.AuthList {
		if auth.Email == request.Email {
			return &dao.Auth{
				Id:       auth.ID,
				Email:    auth.Email,
				Password: auth.Password,
			}, nil
		}
	}
	return nil, dao.ErrAuthNotFound
}

func (mc *MockComm) CreateJWTCredential() (*comm.JWTCredential, error) {
	return &comm.JWTCredential{
		Key:    "MyKey",
		Secret: "ShhItsASecret",
	}, nil
}

func makeRequest(env Env, method string, url string, body string) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	Router(env).ServeHTTP(rec, req)
	return rec, nil
}

// Test that an auth can be successfully created
func TestAuthCreateHandlerSucceeds(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := Env{
		&MockDAO{AuthList: make([]MockAuth, 0)},
		&mockComm,
		cred,
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/auth", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

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

	if id.(float64) != 0 {
		t.Fatalf("ID is incorrect: found %+v, wanted: 0", id)
	}

	iss, ok := claims["iss"]
	if !ok {
		t.Fatalf("Claims doesn't contain an iss key")
	}

	if iss.(string) != cred.Key {
		t.Fatalf("iss is incorrect: found %v, wanted: %s", iss, cred.Key)
	}
}

// Test that providing an empty value for the `Email` parameter causes a `StatusBadRequest`
func TestAuthCreateHandlerFailsOnEmptyParameter(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := Env{
		&MockDAO{AuthList: make([]MockAuth, 0)},
		&mockComm,
		cred,
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/auth", `{"email": "", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing no body to the request causes a `StatusBadRequest`
func TestAuthCreateHandlerFailsOnNoBody(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := Env{
		&MockDAO{AuthList: make([]MockAuth, 0)},
		&mockComm,
		cred,
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/auth", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that an auth can be successfully created and then read from
func TestAuthReadHandlerSucceeds(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := Env{
		&MockDAO{AuthList: make([]MockAuth, 0)},
		&mockComm,
		cred,
	}

	// Create an auth
	_, err := makeRequest(mockEnv, http.MethodPost, "/auth", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Access that same auth
	res, err := makeRequest(mockEnv, http.MethodGet, "/auth", `{"email": "jay@test.com", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

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

	if id.(float64) != 0 {
		t.Fatalf("ID is incorrect: found %+v, wanted: 0", id)
	}

	iss, ok := claims["iss"]
	if !ok {
		t.Fatalf("Claims doesn't contain an iss key")
	}

	if iss.(string) != cred.Key {
		t.Fatalf("iss is incorrect: found %v, wanted: %s", iss, cred.Key)
	}
}

// Test that providing an empty parameter to the read endpoint fails
func TestAuthReadHandlerFailsOnEmptyParameter(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := Env{
		&MockDAO{AuthList: make([]MockAuth, 0)},
		&mockComm,
		cred,
	}

	// Access that same auth
	res, err := makeRequest(mockEnv, http.MethodGet, "/auth", `{"email": "", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that looking up an auth that doesn't exist fails
func TestAuthReadHandlerFailsOnNonExistentAuth(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := Env{
		&MockDAO{AuthList: make([]MockAuth, 0)},
		&mockComm,
		cred,
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/auth", `{"email": "idonotexist@test.com", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing an invalid json body to the read endpoint fails
func TestAuthReadHandlerFailsOnMalformedJSON(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := Env{
		&MockDAO{AuthList: make([]MockAuth, 0)},
		&mockComm,
		cred,
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/auth", `{"email`)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}
