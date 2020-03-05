package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/user/dao"
)

type MockUser struct {
	ID   int
	Name string
}

type MockDAO struct {
	Users []MockUser
}

func (md *MockDAO) CreateUser(request dao.UserCreateRequest) (*dao.UserCreateResponse, error) {
	mockUser := MockUser{len(md.Users), request.Name}
	md.Users = append(md.Users, mockUser)
	return &dao.UserCreateResponse{
		ID:   mockUser.ID,
		Name: mockUser.Name,
	}, nil
}

func (md *MockDAO) ReadUser(userID int64) (*dao.UserReadResponse, error) {
	for _, user := range md.Users {
		if int64(user.ID) == userID {
			return &dao.UserReadResponse{
				ID:   user.ID,
				Name: user.Name,
			}, nil
		}
	}
	return nil, dao.ErrUserNotFound(userID)
}

func (md *MockDAO) UpdateUser(userID int64, request dao.UserUpdateRequest) (*dao.UserUpdateResponse, error) {
	for i, user := range md.Users {
		if int64(user.ID) == userID {
			md.Users[i].Name = request.Name
			return &dao.UserUpdateResponse{
				ID:   user.ID,
				Name: request.Name,
			}, nil
		}
	}
	return nil, dao.ErrUserNotFound(userID)
}

func (md *MockDAO) DeleteUser(userID int64) error {
	for i, user := range md.Users {
		if int64(user.ID) == userID {
			md.Users = append(md.Users[:i], md.Users[i+1:]...)
			return nil
		}
	}
	return dao.ErrUserNotFound(userID)
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

// Test that a user can be successfully created
func TestUserCreateHandlerSucceeds(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"Name":"Jay"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing an empty value for the `Name` parameter causes a `StatusBadRequest`
func TestUserCreateHandlerFailsOnEmptyParameter(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": ""}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing no body to the request causes a `StatusBadRequest`
func TestUserCreateHandlerFailsOnNoBody(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/user", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that a user can be successfully created and then read back
func TestUserReadHandlerSucceeds(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	// Make a user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Access that same user
	res, err := makeRequest(mockEnv, http.MethodGet, "/user/0", "")
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"Name":"Jay"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing no ID to the read endpoint fails
func TestUserReadHandlerFailsOnEmptyID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/user/", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// 404 Not Found, since no route is defined for GET at /user
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing an invalid ID to the read endpoint fails
func TestUserReadHandlerFailsOnNonExistentID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/user/123456", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// 404 Not Found, since no user exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing a string as ID to the read endpoint fails
func TestUserReadHandlerFailsOnStringID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/user/abcdef", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that a user can be successfully created and then updated
func TestUserUpdateHandlerSucceeds(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	// Make a user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", `{"Name": "Lewis"}`)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"Name":"Lewis"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing an empty name to the update endpoint fails
func TestUserUpdateHandlerFailsOnEmptyName(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	// Make a user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", `{"Name": ""}`)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing malformed JSON to the update endpoint fails
func TestUserUpdateHandlerFailsOnInvalidRequestBody(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	// Make a user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", `{"Name"}`)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing no ID to the update endpoint fails
func TestUserUpdateHandlerFailsOnEmptyID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/user/", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// 404 Not Found, since no route is defined for PUT at /user
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing an invalid ID to the update endpoint fails
func TestUserUpdateHandlerFailsOnNonExistentID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/user/123456", `{"Name":"Will"}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// 404 Not Found, since no user exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing a string as ID to the update endpoint fails
func TestUserUpdateHandlerFailsOnStringID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/user/abcdef", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that a user can be successfully created and then deleted
func TestUserDeleteHandlerSucceeds(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	// Make a user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/0", "")
	if err != nil {
		t.Fatalf("Could not make DELETE request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

	received := res.Body.String()
	expected := `{}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing no ID to the delete endpoint fails
func TestUserDeleteHandlerFailsOnEmptyID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// 404 Not Found, since no route is defined for DELETE at /user
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing an invalid ID to the delete endpoint fails
func TestUserDeleteHandlerFailsOnNonExistentID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/123456", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// 404 Not Found, since no user exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing a string as ID to the delete endpoint fails
func TestUserDeleteHandlerFailsOnStringID(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Users: make([]MockUser, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/abcdef", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}
