package main

import (
	"errors"
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

// Define a picture UUID
const pictureUUID0 = "00000001-1234-5678-9012-000000000000"

type mockDAO struct {
	userList    []dao.User
	pictureList []dao.Picture
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
			return &user, nil
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

func (md *mockDAO) CreatePicture(input dao.CreatePictureInput) (*dao.Picture, error) {
	mockPicture := dao.Picture{
		ID:     uuid.MustParse(pictureUUID0),
		UserID: input.UserID,
		Img:    input.Img,
	}

	// Validate foreign key
	for _, user := range md.userList {
		if user.ID == input.UserID {
			md.pictureList = append(md.pictureList, mockPicture)
			return &mockPicture, nil
		}
	}

	return nil, dao.ErrUserNotFound(input.UserID.String())
}

func (md *mockDAO) ReadPicture(input dao.ReadPictureInput) (*dao.Picture, error) {
	for _, picture := range md.pictureList {
		if picture.ID == input.ID && picture.UserID == input.UserID {
			return &picture, nil
		}
	}
	return nil, dao.ErrPictureNotFound(input.ID.String())
}

func makeRequest(env env, method string, url string, body string, authToken string) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	defaultRouter(&env).ServeHTTP(rec, req)
	return rec, nil
}

func makeMockEnv() env {
	return env{
		&mockDAO{userList: make([]dao.User, 0), pictureList: make([]dao.Picture, 0)},
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

// Test that a before create hook is successfully invoked
func TestCreateUserHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeCreate(func(env *env, req createUserRequest, input *dao.CreateUserInput) *HookError {
		// Set all name fields to Lewis
		input.Name = "Lewis"
		return nil
	})

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
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

// Test that a before create hook is successfully invoked and request is aborted
func TestCreateUserHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeCreate(func(env *env, req createUserRequest, input *dao.CreateUserInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that an after create hook is successfully invoked
func TestCreateUserHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterCreate(func(env *env, user *dao.User) *HookError {
		// Set all response fields to Lewis
		user.Name = "Lewis"
		return nil
	})

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
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

// Test that an after create hook is successfully invoked
func TestCreateUserHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterCreate(func(env *env, user *dao.User) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
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

// Test that a before read hook is successfully invoked
func TestReadUserHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeRead(func(env *env, input *dao.ReadUserInput) *HookError {
		// Set uuid to Nil
		input.ID = uuid.Nil
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read that same user
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a before read hook is successfully invoked and request is aborted
func TestReadUserHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeRead(func(env *env, input *dao.ReadUserInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Read a single user
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that an after read hook is successfully invoked
func TestReadUserHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterRead(func(env *env, user *dao.User) *HookError {
		user.Name = "Lewis"
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read that same user
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
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

// Test that an after read hook is successfully invoked and request is aborted
func TestReadUserHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterRead(func(env *env, user *dao.User) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read a single user
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
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

// Test that a before update hook is successfully invoked
func TestUpdateUserHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeUpdate(func(env *env, req updateUserRequest, input *dao.UpdateUserInput) *HookError {
		input.Name = "Will"
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name":"Lewis"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","Name":"Will"}`, UUID0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that a before update hook is successfully invoked and request is aborted
func TestUpdateUserHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeUpdate(func(env *env, req updateUserRequest, input *dao.UpdateUserInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name":"Lewis"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that an after update hook is successfully invoked
func TestUpdateUserHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterUpdate(func(env *env, user *dao.User) *HookError {
		user.Name = "Will"
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name":"Lewis"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","Name":"Will"}`, UUID0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that an after update hook is successfully invoked and request is aborted
func TestUpdateUserHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterUpdate(func(env *env, user *dao.User) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/user/%s", UUID0), `{"Name": "Lewis"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
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

// Test that a before delete hook is successfully invoked
func TestDeleteUserHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeDelete(func(env *env, input *dao.DeleteUserInput) *HookError {
		input.ID = uuid.Nil
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Delete that same user
	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a before delete hook is successfully invoked and request is aborted
func TestDeleteUserHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeDelete(func(env *env, input *dao.DeleteUserInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Delete that same user
	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that an after delete hook is successfully invoked
func TestDeleteUserHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()
	isHookExecuted := false

	mockEnv.hook.AfterDelete(func(env *env) *HookError {
		isHookExecuted = true
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Delete that same user
	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	if !isHookExecuted {
		t.Errorf("Hook was not executed")
	}
}

// Test that an after delete hook is successfully invoked and request is aborted
func TestDeleteUserHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterDelete(func(env *env) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Delete that same user
	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/user/%s", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single picture can be successfully created
func TestCreatePictureHandlerSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Create a single picture for that user
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s"}`, pictureUUID0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing a malformed JSON body to the create picture endpoint fails
func TestCreatePictureHandlerFailsOnMalformedJSONBody(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single picture
	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), `{"Img"`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the create endpoint fails
func TestCreatePictureHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single picture
	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent ID to the read endpoint fails
func TestCreatePictureHandlerFailsOnNonExistentID(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", uuid.Nil.String()), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no user exists with that given ID
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that creating a single picture without first creating a single user fails
func TestCreatePictureHandlerFailsWithoutUser(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single picture
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no user exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the create endpoint fails
func TestCreatePictureHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single picture
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a before create hook is successfully invoked
func TestCreatePictureHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()
	isExecuted := false

	mockEnv.hook.BeforeCreatePicture(func(env *env, req createPictureRequest, input *dao.CreatePictureInput) *HookError {
		isExecuted = true
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Create a single picture for that user
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	_, err = makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if !isExecuted {
		t.Errorf("Hook was not executed")
	}
}

// Test that a before create hook is successfully invoked and request is aborted
func TestCreatePictureHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeCreatePicture(func(env *env, req createPictureRequest, input *dao.CreatePictureInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single picture
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that an after create hook is successfully invoked
func TestCreatePictureHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterCreatePicture(func(env *env, picture *dao.Picture) *HookError {
		picture.ID = uuid.Nil
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Create a single picture for that user
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s"}`, uuid.Nil.String())
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that an after create hook is successfully invoked
func TestCreatePictureHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterCreatePicture(func(env *env, picture *dao.Picture) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Create a single picture for that user
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	res, err := makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single picture can be successfully created and then read back
func TestReadPictureHandlerSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Create a single picture for that user
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	_, err = makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Read that same picture
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s/picture/%s", UUID0, pictureUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","Img":"c3F1YXRhbmRkYWI="}`, pictureUUID0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing a non-existent user ID to the read picture endpoint fails
func TestReadPictureHandlerFailsOnNonExistentID(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s/picture/%s", uuid.Nil.String(), uuid.Nil.String()), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no picture exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the read endpoint fails
func TestReadPictureHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := makeMockEnv()

	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s/picture/%s", UUID0, pictureUUID0), "", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a valid picture cannot be read from another user
func TestReadPictureHandlerFailsForDifferentUser(t *testing.T) {
	mockEnv := makeMockEnv()

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Create a single picture for that user
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	_, err = makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Read that same picture, but for a different user
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s/picture/%s", UUID1, pictureUUID0), "", JWT1)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a before read hook is successfully invoked
func TestReadPictureHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()
	isExecuted := false

	mockEnv.hook.BeforeReadPicture(func(env *env, input *dao.ReadPictureInput) *HookError {
		isExecuted = true
		return nil
	})

	_, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s/picture/%s", UUID0, pictureUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if !isExecuted {
		t.Errorf("Hook was not executed")
	}
}

// Test that a before read hook is successfully invoked and request is aborted
func TestReadPictureHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.BeforeReadPicture(func(env *env, input *dao.ReadPictureInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Read a single picture
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s/picture/%s", UUID0, pictureUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that an after read hook is successfully invoked
func TestReadPictureHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterReadPicture(func(env *env, picture *dao.Picture) *HookError {
		picture.ID = uuid.Nil
		return nil
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Create a single picture for that user
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	_, err = makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read that same picture
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s/picture/%s", UUID0, pictureUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","Img":"c3F1YXRhbmRkYWI="}`, uuid.Nil.String())
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that an after read hook is successfully invoked
func TestReadPictureHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := makeMockEnv()

	mockEnv.hook.AfterReadPicture(func(env *env, picture *dao.Picture) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Create a single picture for that user
	body := `{"Img": "c3F1YXRhbmRkYWI="}`
	_, err = makeRequest(mockEnv, http.MethodPost, fmt.Sprintf("/user/%s/picture", UUID0), body, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read that same picture
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/user/%s/picture/%s", UUID0, pictureUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}
