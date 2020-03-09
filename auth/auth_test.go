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

type MockDAO struct {
	AuthList []dao.Auth
}

type MockComm struct{}

func (md *MockDAO) CreateAuth(input dao.CreateAuthInput) (*dao.Auth, error) {
	// Check if auth already exists
	for _, auth := range md.AuthList {
		if auth.Email == input.Email {
			return nil, dao.ErrDuplicateAuth
		}
	}

	mockAuth := dao.Auth{
		ID:       len(md.AuthList),
		Email:    input.Email,
		Password: input.Password,
	}
	md.AuthList = append(md.AuthList, mockAuth)
	return &dao.Auth{
		ID:       mockAuth.ID,
		Email:    mockAuth.Email,
		Password: mockAuth.Password,
	}, nil
}

func (md *MockDAO) ReadAuth(input dao.ReadAuthInput) (*dao.Auth, error) {
	for _, auth := range md.AuthList {
		if auth.Email == input.Email {
			return &dao.Auth{
				ID:       auth.ID,
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

func makeRequest(env env, method string, url string, body string) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	env.router().ServeHTTP(rec, req)
	return rec, nil
}

// Test that a single auth can be successfully created
func TestCreateAuthHandlerSucceeds(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
		&mockComm,
		cred,
	}

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
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
		t.Fatalf("ID is incorrect, found: %+v, wanted: 0", id)
	}

	iss, ok := claims["iss"]
	if !ok {
		t.Fatalf("Claims doesn't contain an iss key")
	}

	if iss.(string) != cred.Key {
		t.Fatalf("iss is incorrect: found %v, wanted %s", iss, cred.Key)
	}
}

// Test that providing an empty parameter to the create endpoint fails
func TestCreateAuthHandlerFailsOnEmptyParameter(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
		&mockComm,
		cred,
	}

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth", `{"email": "", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the create endpoint fails
func TestCreateAuthHandlerFailsOnMalformedJSON(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
		&mockComm,
		cred,
	}

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodGet, "/auth", `{"email`)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to create endpoint fails
func TestCreateAuthHandlerFailsOnNoBody(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
		&mockComm,
		cred,
	}

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that repeating the same request to the create endpoint fails
func TestCreateAuthHandlerFailsOnDuplicate(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
		&mockComm,
		cred,
	}

	// Make first request
	_, err := makeRequest(mockEnv, http.MethodPost, "/auth", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Make repeat first request
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusForbidden {
		t.Fatalf("Invalid response code: %d", res.Code)
	}
}

// Test that a single auth can be successfully created and then read back
func TestReadAuthHandlerSucceeds(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
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
		t.Errorf("Wrong status code: %v", res.Code)
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
		t.Fatalf("ID is incorrect, found: %+v, wanted: 0", id)
	}

	iss, ok := claims["iss"]
	if !ok {
		t.Fatalf("Claims doesn't contain an iss key")
	}

	if iss.(string) != cred.Key {
		t.Fatalf("iss is incorrect: found %v, wanted %s", iss, cred.Key)
	}
}

// Test that providing an empty parameter to the read endpoint fails
func TestReadAuthHandlerFailsOnEmptyParameter(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
		&mockComm,
		cred,
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/auth", `{"email": "", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the read endpoint fails
func TestReadAuthHandlerFailsOnMalformedJSON(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
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

// Test that providing no body to the read endpoint fails
func TestReadAuthHandlerFailsOnNoBody(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
		&mockComm,
		cred,
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/auth", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent auth to the read endpoint fails
func TestReadAuthHandlerFailsOnNonExistentAuth(t *testing.T) {
	mockComm := MockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	mockEnv := env{
		&MockDAO{AuthList: make([]dao.Auth, 0)},
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
