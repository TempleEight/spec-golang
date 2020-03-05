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
				Name: user.Name,
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

var mockEnv = Env{
	&MockDAO{Users: make([]MockUser, 0)},
}

func makeRequest(method string, url string, body string, handler http.HandlerFunc) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	handler.ServeHTTP(rec, req)
	return rec, nil
}

// Test that a user can be successfully created
func TestUserCreateHandlerSucceeds(t *testing.T) {
	res, err := makeRequest(http.MethodPost, "/user", `{"Name": "Jay"}`, mockEnv.userCreateHandler)
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
	res, err := makeRequest(http.MethodPost, "/user", `{"Name": ""}`, mockEnv.userCreateHandler)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that providing no body to the request causes a `StatusBadRequest`
func TestUserCreateHandlerFailsOnNoBody(t *testing.T) {
	res, err := makeRequest(http.MethodPost, "/user", "", mockEnv.userCreateHandler)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}
