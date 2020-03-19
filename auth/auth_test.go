package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/auth/comm"
	"github.com/TempleEight/spec-golang/auth/dao"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type mockDAO struct {
	authList []dao.Auth
}

type mockComm struct{}

func (md *mockDAO) CreateAuth(input dao.CreateAuthInput) (*dao.Auth, error) {
	// Check if auth already exists
	for _, auth := range md.authList {
		if auth.Email == input.Email {
			return nil, dao.ErrDuplicateAuth
		}
	}

	mockAuth := dao.Auth{
		ID:       input.ID,
		Email:    input.Email,
		Password: input.Password,
	}
	md.authList = append(md.authList, mockAuth)
	return &mockAuth, nil
}

func (md *mockDAO) ReadAuth(input dao.ReadAuthInput) (*dao.Auth, error) {
	for _, auth := range md.authList {
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

func (mc *mockComm) CreateJWTCredential() (*comm.JWTCredential, error) {
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

	defaultRouter(&env).ServeHTTP(rec, req)
	return rec, nil
}

func makeMockEnv() env {
	mockComm := mockComm{}
	cred, _ := mockComm.CreateJWTCredential()
	return env{
		&mockDAO{authList: make([]dao.Auth, 0)},
		&mockComm,
		cred,
		Hook{},
	}
}

// Test that a single auth can be successfully created
func TestCreateAuthHandlerSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
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

	_, err = uuid.Parse(id.(string))
	if err != nil {
		t.Fatalf("ID is not a valid UUID")
	}

	iss, ok := claims["iss"]
	if !ok {
		t.Fatalf("Claims doesn't contain an iss key")
	}

	if iss.(string) != mockEnv.jwtCredential.Key {
		t.Fatalf("iss is incorrect: found %v, wanted %s", iss, mockEnv.jwtCredential.Key)
	}
}

// Test that providing an empty parameter to the create endpoint fails
func TestCreateAuthHandlerFailsOnEmptyParameter(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the create endpoint fails
func TestCreateAuthHandlerFailsOnMalformedJSON(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to create endpoint fails
func TestCreateAuthHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that repeating the same request to the create endpoint fails
func TestCreateAuthHandlerFailsOnDuplicate(t *testing.T) {
	mockEnv := makeMockEnv()

	// Make first request
	_, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Make repeat first request
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusForbidden {
		t.Fatalf("Invalid response code: %d", res.Code)
	}
}

// Test that a before create hook is successfully invoked
func TestCreateAuthHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()
	isHookExecuted := false

	mockEnv.hook.BeforeCreate(func(env *env, req registerAuthRequest, input *dao.CreateAuthInput) *HookError {
		isHookExecuted = true
		return nil
	})

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	if !isHookExecuted {
		t.Fatalf("Hook was not executed")
	}
}

// Test that a before create hook is successfully invoked and request is aborted
func TestCreateUserHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeCreate(func(env *env, req registerAuthRequest, input *dao.CreateAuthInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Fatalf("Wrong status code: %v", res.Code)
	}
}

// Test that an after create hook is successfully invoked
func TestCreateUserHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()
	isHookExecuted := false

	mockEnv.hook.AfterCreate(func(env *env, auth *dao.Auth, accessToken string) *HookError {
		isHookExecuted = true
		return nil
	})

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	if !isHookExecuted {
		t.Fatalf("Hook was not executed")
	}
}

// Test that an after create hook is successfully invoked
func TestCreateUserHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterCreate(func(env *env, auth *dao.Auth, accessToken string) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Fatalf("Wrong status code: %v", res.Code)
	}
}

// Test that a single auth can be successfully created and then read back
func TestReadAuthHandlerSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create an auth
	_, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Access that same auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", `{"email": "jay@test.com", "password":"BlackcurrantCrush123"}`)
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

	_, err = uuid.Parse(id.(string))
	if err != nil {
		t.Fatalf("ID is not a valid UUID")
	}

	iss, ok := claims["iss"]
	if !ok {
		t.Fatalf("Claims doesn't contain an iss key")
	}

	if iss.(string) != mockEnv.jwtCredential.Key {
		t.Fatalf("iss is incorrect: found %v, wanted %s", iss, mockEnv.jwtCredential.Key)
	}
}

// Test that providing an empty parameter to the read endpoint fails
func TestReadAuthHandlerFailsOnEmptyParameter(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", `{"email": "", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the read endpoint fails
func TestReadAuthHandlerFailsOnMalformedJSON(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", `{"email`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing no body to the read endpoint fails
func TestReadAuthHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent auth to the read endpoint fails
func TestReadAuthHandlerFailsOnNonExistentAuth(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", `{"email": "idonotexist@test.com", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that a before read hook is successfully invoked
func TestReadAuthHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()
	isHookExecuted := false

	mockEnv.hook.BeforeRead(func(env *env, req loginAuthRequest, input *dao.ReadAuthInput) *HookError {
		isHookExecuted = true
		return nil
	})

	// Create a single auth
	_, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Access that same auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", `{"email": "jay@test.com", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	if !isHookExecuted {
		t.Fatalf("Hook was not executed")
	}
}

// Test that a before read hook is successfully invoked and request is aborted
func TestReadAuthHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeRead(func(env *env, req loginAuthRequest, input *dao.ReadAuthInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single auth
	_, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Access that same auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", `{"email": "jay@test.com", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Fatalf("Wrong status code: %v", res.Code)
	}
}

// Test that an after create hook is successfully invoked
func TestReadAuthHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()
	isHookExecuted := false

	mockEnv.hook.AfterRead(func(env *env, auth *dao.Auth, accessToken string) *HookError {
		isHookExecuted = true
		return nil
	})

	// Create a single auth
	_, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Access that same auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", `{"email": "jay@test.com", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	if !isHookExecuted {
		t.Fatalf("Hook was not executed")
	}
}

// Test that an after read hook is successfully invoked and request aborted
func TestReadAuthHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterRead(func(env *env, auth *dao.Auth, accessToken string) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single auth
	_, err := makeRequest(mockEnv, http.MethodPost, "/auth/register", `{"email": "jay@test.com", "password": "BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Access that same auth
	res, err := makeRequest(mockEnv, http.MethodPost, "/auth/login", `{"email": "jay@test.com", "password":"BlackcurrantCrush123"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Fatalf("Wrong status code: %v", res.Code)
	}
}
