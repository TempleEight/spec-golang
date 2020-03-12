package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/match/dao"
)

// Define 2 JWTs with ID 0 and 1
const auth0JWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODMyNTc5NzcsImlkIjowLCJpc3MiOiJmRlM4S21WWXVLQUN5RjN3ZHBQS0hTUXFtWlZWd2pEcSJ9.KzUa-OpHEjFQlsSy7YZI1Kppu4eIU5nyivLvivWcpRc"
const auth1JWT = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODMyNTc5NzcsImlkIjoxLCJpc3MiOiJmRlM4S21WWXVLQUN5RjN3ZHBQS0hTUXFtWlZWd2pEcSJ9.kXaTT0Yl3-zeWreKOl5Zd6dG1gJG49JSS0zfdBRG_oU"

type mockDAO struct {
	matchList []dao.Match
}

type mockComm struct {
	userIDs []int64
}

func (md *mockDAO) ListMatch(input dao.ListMatchInput) (*[]dao.Match, error) {
	mockMatchList := make([]dao.Match, 0)
	for _, match := range md.matchList {
		if match.AuthID == input.AuthID {
			mockMatchList = append(mockMatchList, match)
		}
	}

	return &mockMatchList, nil
}

func (md *mockDAO) CreateMatch(input dao.CreateMatchInput) (*dao.Match, error) {
	mockMatch := dao.Match{
		ID:        int64(len(md.matchList)),
		AuthID:    input.AuthID,
		UserOne:   input.UserOne,
		UserTwo:   input.UserTwo,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}
	md.matchList = append(md.matchList, mockMatch)
	return &dao.Match{
		ID:        mockMatch.ID,
		AuthID:    mockMatch.AuthID,
		UserOne:   mockMatch.UserOne,
		UserTwo:   mockMatch.UserTwo,
		MatchedOn: mockMatch.MatchedOn,
	}, nil
}

func (md *mockDAO) ReadMatch(input dao.ReadMatchInput) (*dao.Match, error) {
	for _, match := range md.matchList {
		if match.ID == input.ID {
			return &dao.Match{
				ID:        match.ID,
				AuthID:    match.AuthID,
				UserOne:   match.UserOne,
				UserTwo:   match.UserTwo,
				MatchedOn: match.MatchedOn,
			}, nil
		}
	}
	return nil, dao.ErrMatchNotFound(input.ID)
}

func (md *mockDAO) UpdateMatch(input dao.UpdateMatchInput) (*dao.Match, error) {
	for i, match := range md.matchList {
		if match.ID == input.ID {
			md.matchList[i].UserOne = input.UserOne
			md.matchList[i].UserTwo = input.UserTwo
			md.matchList[i].MatchedOn = "2020-12-31T12:00:00.000000Z"
			return &dao.Match{
				ID:        md.matchList[i].ID,
				AuthID:    md.matchList[i].AuthID,
				UserOne:   md.matchList[i].UserOne,
				UserTwo:   md.matchList[i].UserTwo,
				MatchedOn: md.matchList[i].MatchedOn,
			}, nil
		}
	}
	return nil, dao.ErrMatchNotFound(input.ID)
}

func (md *mockDAO) DeleteMatch(input dao.DeleteMatchInput) error {
	for i, match := range md.matchList {
		if match.ID == input.ID {
			md.matchList = append(md.matchList[:i], md.matchList[i+1:]...)
			return nil
		}
	}
	return dao.ErrMatchNotFound(input.ID)
}

func (mc *mockComm) CheckUser(userID int64) (bool, error) {
	for _, id := range mc.userIDs {
		if id == userID {
			return true, nil
		}
	}
	return false, nil
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

// Test that a match list can be read successfully for a given ID
func TestListMatchHandlerSucceeds(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{
		dao.Match{
			ID:        0,
			AuthID:    0,
			UserOne:   0,
			UserTwo:   1,
			MatchedOn: "2020-01-01T12:00:00.000000Z",
		},
		dao.Match{
			ID:        1,
			AuthID:    0,
			UserOne:   0,
			UserTwo:   2,
			MatchedOn: "2020-01-01T12:00:00.000000Z",
		},
		dao.Match{
			ID:        2,
			AuthID:    1,
			UserOne:   1,
			UserTwo:   2,
			MatchedOn: "2020-01-01T12:00:00.000000Z",
		},
	}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1, 2}},
	}

	// Read the match list for AuthID 0
	res, err := makeRequest(mockEnv, http.MethodGet, "/match/all", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"MatchList":[{"ID":0,"UserOne":0,"UserTwo":1,"MatchedOn":"2020-01-01T12:00:00.000000Z"},{"ID":1,"UserOne":0,"UserTwo":2,"MatchedOn":"2020-01-01T12:00:00.000000Z"}]}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that a single match can be created successfully
func TestCreateMatchHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 1}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"UserOne":0,"UserTwo":1,"MatchedOn":"2020-01-01T12:00:00.000000Z"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that providing an incomplete body to the create endpoint fails
func TestCreateMatchHandlerFailsOnIncompleteBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the create endpoint fails
func TestCreateMatchHandlerFailsOnMalformedJSONBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne"`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the create endpoint fails
func TestCreateMatchHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	// Create a single match
	res, err := makeRequest(mockEnv, http.MethodPost, "/match", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an invalid UserOne reference to the create endpoint fails
func TestCreateMatchHandlerFailsOnInvalidUserOne(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 123456, "UserTwo": 0}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an invalid UserTwo reference to the create endpoint fails
func TestCreateMatchHandlerFailsOnInvalidUserTwo(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 123456}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing all invalid references to the create endpoint fails
func TestCreateMatchHandlerFailsOnAllInvalidReferences(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 123456, "UserTwo": 234567}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be read successfully
func TestReadMatchHandlerSucceeds(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        0,
		AuthID:    0,
		UserOne:   0,
		UserTwo:   1,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/0", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"UserOne":0,"UserTwo":1,"MatchedOn":"2020-01-01T12:00:00.000000Z"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that providing no ID to the read endpoint fails
func TestReadMatchHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no route is defined for GET at /match
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent ID to the read endpoint fails
func TestReadMatchHandlerFailsOnNonExistentID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/123456", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no match exists with the given ID
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the read endpoint fails
func TestReadUserHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/abcdef", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad Request, since an integer ID is required
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be updated successfully
func TestUpdateUserHandlerSucceeds(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        0,
		AuthID:    0,
		UserOne:   0,
		UserTwo:   1,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1, 2}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne": 0, "UserTwo": 2}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"UserOne":0,"UserTwo":2,"MatchedOn":"2020-12-31T12:00:00.000000Z"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that providing an incomplete to the update endpoint fails
func TestUpdateMatchHandlerFailsOnIncompleteBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the update endpoint fails
func TestUpdateMatchHandlerFailsOnMalformedJSONBody(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        0,
		AuthID:    0,
		UserOne:   0,
		UserTwo:   1,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne"`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the update endpoint fails
func TestUpdateMatchHandlerFailsOnNoBody(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        0,
		AuthID:    0,
		UserOne:   0,
		UserTwo:   1,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no ID to the update endpoint fails
func TestUpdateMatchHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no route is defined for PUT at /match
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent ID to the update endpoint fails
func TestUpdateMatchHandlerFailsOnNonExistentID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/123456", `{"UserOne": 0, "UserTwo": 1}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no match exists with the given ID
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the update endpoint fails
func TestUpdateMatchHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/abcdef", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad Request, since an integer ID is required
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an invalid UserOne reference to the update endpoint fails
func TestUpdateMatchHandlerFailsOnInvalidUserOne(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        0,
		AuthID:    0,
		UserOne:   0,
		UserTwo:   1,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne": 123456, "UserTwo": 0}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an invalid UserTwo reference to the update endpoint fails
func TestUpdateMatchHandlerFailsOnInvalidUserTwo(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        0,
		AuthID:    0,
		UserOne:   0,
		UserTwo:   1,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne": 0, "UserTwo": 123456}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing all invalid references to the update endpoint fails
func TestUpdateMatchHandlerFailsOnAllInvalidReferences(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        0,
		AuthID:    0,
		UserOne:   0,
		UserTwo:   1,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne": 123456, "UserTwo": 234567}`, auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be deleted successfully
func TestDeleteMatchHandlerSucceeds(t *testing.T) {
	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        0,
		AuthID:    0,
		UserOne:   0,
		UserTwo:   1,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/0", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v expected %+v", received, expected)
	}
}

// Test that providing no ID to the delete endpoint fails
func TestDeleteMatchHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no route is defined for DELETE at /match
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a non-existent ID to the delete endpoint fails
func TestDeleteMatchHandlerFailsOnNonExistentID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/123456", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no match exists with the given ID
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the delete endpoint fails
func TestDeleteMatchHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/abcdef", "", auth0JWT)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad Request, since an integer ID is required
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}
