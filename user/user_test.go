package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/user/dao"
	"github.com/google/uuid"
)

// Define 2 UUIDs
const UUID0 = "00000000-1234-5678-9012-000000000000"
const UUID1 = "00000000-1234-5678-9012-000000000001"

// Define 2 JWTs corresponding to UUIDs
const JWT0 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODMyNTc5NzcsImlkIjoiMDAwMDAwMDAtMTIzNC01Njc4LTkwMTItMDAwMDAwMDAwMDAwIiwiaXNzIjoiZkZTOEttVll1S0FDeUYzd2RwUEtIU1FxbVpWVndqRHEifQ.jMpelsEJUwONtRCQnQCo2v5Ph7cZHloc5R1OvKkU2Ck"
const JWT1 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODMyNTc5NzcsImlkIjoiMDAwMDAwMDAtMTIzNC01Njc4LTkwMTItMDAwMDAwMDAwMDAxIiwiaXNzIjoiZkZTOEttVll1S0FDeUYzd2RwUEtIU1FxbVpWVndqRHEifQ.lMmkaK9L2kD2ZnbblSlXdz93cz6jZCALR0KoGlzQKpc"

type mockDAO struct {
	userList []dao.User
}

func (md *mockDAO) CreateUser(input dao.CreateUserInput) (*dao.User, error) {
	mockUser := dao.User{
		ID:   input.ID,
		Name: input.Name,
	}
	md.userList = append(md.userList, mockUser)
	return &mockUser, nil
}

func (md *mockDAO) ReadUser(input dao.ReadUserInput) (*dao.User, error) {
	for _, user := range md.userList {
		if user.ID == input.ID {
			return &dao.User{
				ID:   user.ID,
				Name: user.Name,
			}, nil
		}
	}
	return nil, dao.ErrUserNotFound(input.ID.String())
}

func (md *mockDAO) UpdateUser(input dao.UpdateUserInput) (*dao.User, error) {
	for i, user := range md.userList {
		if user.ID == input.ID {
			md.userList[i].Name = user.Name
			return &dao.User{
				ID:   user.ID,
				Name: input.Name,
			}, nil
		}
	}
	return nil, dao.ErrUserNotFound(input.ID.String())
}

func (md *mockDAO) DeleteUser(input dao.DeleteUserInput) error {
	for i, user := range md.userList {
		if user.ID == input.ID {
			md.userList = append(md.userList[:i], md.userList[i+1:]...)
			return nil
		}
	}
	return dao.ErrUserNotFound(input.ID.String())
}

func makeRequest(env env, method string, url string, body string, authToken string) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	env.router().ServeHTTP(rec, req)
	return rec, nil
}

func makeMockEnv() env {
	return env{
		&mockDAO{userList: make([]dao.User, 0)},
		Hook{},
	}
}

// Test that a single user can be successfully created
func TestCreateUserHandlerSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","Name":"Jay"}`, UUID0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing an empty parameter to the create endpoint fails
func TestCreateUserHandlerFailsOnEmptyParameter(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": ""}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the create endpoint fails
func TestCreateUserHandlerFailsOnMalformedJSONBody(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name"`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the create endpoint fails
func TestCreateUserHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the create endpoint fails
func TestCreateUserHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single user can be successfully created and then read back
func TestReadUserHandlerSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Read that same user
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","Name":"Jay"}`, UUID0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing no ID to the read endpoint fails
func TestReadUserHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodGet, "/user/", "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no route is defined for GET at /user
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent ID to the read endpoint fails
func TestReadUserHandlerFailsOnNonExistentID(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s", uuid.Nil.String()), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no user exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the read endpoint fails
func TestReadUserHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s", UUID0), `{"Name": "Jay"}`, "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single user can be successfully created and then updated
func TestUpdateUserHandlerSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name": "Lewis"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","Name":"Lewis"}`, UUID0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing an empty parameter to the update endpoint fails
func TestUpdateUserHandlerFailsOnEmptyParameter(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name": ""}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the update endpoint fails
func TestUpdateUserHandlerFailsOnMalformedJSONBody(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the update endpoint fails
func TestUpdateUserHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no ID to the update endpoint fails
func TestUpdateUserHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodPut, "/user/", "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no route is defined for PUT at /user
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent ID to the update endpoint fails
func TestUpdateUserHandlerFailsOnNonExistentID(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", uuid.Nil.String()), `{"Name":"Will"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no user can update information about another user
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the update endpoint fails
func TestUpdateUserHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name": "Jay"}`, "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a different JWT to the update endpoint fails
func TestUpdateUserHandlerFailsOnDifferentJWT(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user with JWT0
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Update a single user with JWT1
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name": "Lewis"}`, JWT1)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single user can be successfully created and then deleted
func TestDeleteUserHandlerSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Delete that same user
	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make DELETE request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing no ID to the delete endpoint fails
func TestDeleteUserHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/", "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no route is defined for DELETE at /user
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent ID to the delete endpoint fails
func TestDeleteUserHandlerFailsOnNonExistentID(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/user/%s", uuid.Nil.String()), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no user can delete information about another user
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the delete endpoint fails
func TestDeleteUserHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/user/%s", UUID0), "", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a different JWT to the delete endpoint fails
func TestDeleteUserHandlerFailsOnDifferentJWT(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user with JWT0
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Delete a single user with JWT1
	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/user/%s", UUID0), "", JWT1)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}
