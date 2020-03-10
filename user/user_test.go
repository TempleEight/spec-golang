package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/user/dao"
)

// Define 2 JWTs with ID 0 and 1
const user0JWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODMyNTc5NzcsImlkIjowLCJpc3MiOiJmRlM4S21WWXVLQUN5RjN3ZHBQS0hTUXFtWlZWd2pEcSJ9.KzUa-OpHEjFQlsSy7YZI1Kppu4eIU5nyivLvivWcpRc"
const user1JWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODMyNTc5NzcsImlkIjoxLCJpc3MiOiJmRlM4S21WWXVLQUN5RjN3ZHBQS0hTUXFtWlZWd2pEcSJ9.kXaTT0Yl3-zeWreKOl5Zd6dG1gJG49JSS0zfdBRG_oU"

type mockDAO struct {
	userList []dao.User
	authList []dao.Auth
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
	return nil, dao.ErrUserNotFound(input.ID)
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
	return nil, dao.ErrUserNotFound(input.ID)
}

func (md *mockDAO) DeleteUser(input dao.DeleteUserInput) error {
	for i, user := range md.userList {
		if user.ID == input.ID {
			md.userList = append(md.userList[:i], md.userList[i+1:]...)
			return nil
		}
	}
	return dao.ErrUserNotFound(input.ID)
}

func (md *mockDAO) CreateUserAuth(input dao.CreateUserAuthInput) (*dao.Auth, error) {
	// Check if auth already exists
	for _, auth := range md.authList {
		if auth.Email == input.Email {
			return nil, dao.ErrDuplicateAuth
		}
	}

	mockAuth := dao.Auth{
		ID:       int64(len(md.authList)),
		Email:    input.Email,
		Password: input.Password,
	}
	md.authList = append(md.authList, mockAuth)
	return &mockAuth, nil
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

// Test that a single user can be successfully created
func TestCreateUserHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"Name":"Jay"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing an empty parameter to the create endpoint fails
func TestCreateUserHandlerFailsOnEmptyParameter(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the create endpoint fails
func TestCreateUserHandlerFailsOnMalformedJSONBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name"`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the create endpoint fails
func TestCreateUserHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", "", user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the create endpoint fails
func TestCreateUserHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing the same request to the create endpoint fails
func TestCreateAuthHandlerFailsOnDuplicate(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Create the same single user
	res, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusForbidden {
		t.Fatalf("Invalid response code: %d", res.Code)
	}
}

// Test that a single user can be successfully created and then read back
func TestReadUserHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Read that same user
	res, err := makeRequest(mockEnv, http.MethodGet, "/user/0", "", user0JWT)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"Name":"Jay"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing no ID to the read endpoint fails
func TestReadUserHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/user/", "", user0JWT)
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
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/user/123456", "", user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no user exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the read endpoint fails
func TestReadUserHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/user/abcdef", "", user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the read endpoint fails
func TestReadUserHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/user/0", "", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single user can be successfully created and then updated
func TestUpdateUserHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", `{"Name": "Lewis"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"Name":"Lewis"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that providing an empty parameter to the update endpoint fails
func TestUpdateUserHandlerFailsOnEmptyParameter(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123" }`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", `{"Name": ""}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the update endpoint fails
func TestUpdateUserHandlerFailsOnMalformedJSONBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", `{"Name"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the update endpoint fails
func TestUpdateUserHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update that same user
	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", "", user0JWT)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no ID to the update endpoint fails
func TestUpdateUserHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/user/", "", user0JWT)
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
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/user/123456", `{"Name":"Lewis"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no user can update information about another user
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the update endpoint fails
func TestUpdateUserHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/user/abcdef", `{"Name":"Lewis"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the update endpoint fails
func TestUpdateUserHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", `{"Name": "Lewis"}`, "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a different JWT to the update endpoint fails
func TestUpdateUserHandlerFailsOnDifferentJWT(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user with user0JWT
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Update a single user with user1JWT
	res, err := makeRequest(mockEnv, http.MethodPut, "/user/0", `{"Name": "Lewis"}`, user1JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single user can be successfully created and then deleted
func TestDeleteUserHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Delete that same user
	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/0", "", user0JWT)
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
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/", "", user0JWT)
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
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/123456", "", user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no user can delete information about another user
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the delete endpoint fails
func TestDeleteUserHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/abcdef", "", user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an empty JWT to the delete endpoint fails
func TestDeleteUserHandlerFailsOnEmptyJWT(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/0", "", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a different JWT to the delete endpoint fails
func TestDeleteUserHandlerFailsOnDifferentJWT(t *testing.T) {
	mockEnv := env{
		&mockDAO{userList: make([]dao.User, 0)},
	}

	// Create a single user with user0JWT
	_, err := makeRequest(mockEnv, http.MethodPost, "/user", `{"Name": "Jay", "Email": "jay@test.com", "Password": "BlackcurrantCrush123"}`, user0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Delete a single user with user1JWT
	res, err := makeRequest(mockEnv, http.MethodDelete, "/user/0", "", user1JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}
