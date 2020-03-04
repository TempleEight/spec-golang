package main

import (
	"encoding/json"
	"github.com/TempleEight/spec-golang/user/dao"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockUser struct {
	Id   int
	Name string
}

type MockDAO struct {
	Users []MockUser
}

func (md *MockDAO) CreateUser(request dao.UserCreateRequest) (*dao.UserCreateResponse, error) {
	mockUser := MockUser{len(md.Users), request.Name}
	md.Users = append(md.Users, mockUser)
	return &dao.UserCreateResponse{
		ID:   mockUser.Id,
		Name: mockUser.Name,
	}, nil
}

func (md *MockDAO) ReadUser(userID int64) (*dao.UserReadResponse, error) {
	for _, user := range md.Users {
		if int64(user.Id) == userID {
			return &dao.UserReadResponse{
				ID:   user.Id,
				Name: user.Name,
			}, nil
		}
	}
	return nil, dao.ErrUserNotFound(userID)
}

func (md *MockDAO) UpdateUser(userID int64, request dao.UserUpdateRequest) (*dao.UserUpdateResponse, error) {
	for i, user := range md.Users {
		if int64(user.Id) == userID {
			md.Users[i].Name = request.Name
			return &dao.UserUpdateResponse{
				ID:   user.Id,
				Name: user.Name,
			}, nil
		}
	}
	return nil, dao.ErrUserNotFound(userID)
}

func (md *MockDAO) DeleteUser(userID int64) error {
	for i, user := range md.Users {
		if int64(user.Id) == userID {
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

func TestUserCreateHandlerSucceeds(t *testing.T) {
	res, err := makeRequest(http.MethodPost, "/user", `{"Name": "Jay"}`, mockEnv.userCreateHandler)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

	var received dao.UserUpdateResponse
	err = json.Unmarshal([]byte(res.Body.String()), &received)
	if err != nil {
		t.Errorf("Could not decode body: %s", res.Body.String())
	}
	expected := dao.UserUpdateResponse{
		ID:   0,
		Name: "Jay",
	}
	if expected != received {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

func TestUserCreateHandlerFailsOnEmptyParameter(t *testing.T) {
	res, err := makeRequest(http.MethodPost, "/user", `{"Name": ""}`, mockEnv.userCreateHandler)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

func TestUserCreateHandlerFailsOnNoBody(t *testing.T) {
	res, err := makeRequest(http.MethodPost, "/user", "", mockEnv.userCreateHandler)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}
