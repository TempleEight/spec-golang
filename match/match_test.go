package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/match/dao"
)

type mockDAO struct {
	matchList []dao.Match
}

type mockComm struct {
	userIDs []int64
}

func (md *mockDAO) ListMatch() (*[]dao.Match, error) {
	matchList := make([]dao.Match, 0)
	for _, match := range md.matchList {
		matchList = append(matchList, dao.Match{
			ID:        match.ID,
			UserOne:   match.UserOne,
			UserTwo:   match.UserTwo,
			MatchedOn: match.MatchedOn,
		})
	}

	return &matchList, nil
}

func (md *mockDAO) CreateMatch(input dao.CreateMatchInput) (*dao.Match, error) {
	mockMatch := dao.Match{
		ID:        int64(len(md.matchList)),
		UserOne:   input.UserOne,
		UserTwo:   input.UserTwo,
		MatchedOn: "2020-01-01T12:00:00.000000Z",
	}
	md.matchList = append(md.matchList, mockMatch)
	return &dao.Match{
		ID:        mockMatch.ID,
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

func makeRequest(env env, method string, url string, body string) (*httptest.ResponseRecorder, error) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	env.router().ServeHTTP(rec, req)
	return rec, nil
}

// Test that a match list can be read successfully
func TestListMatchHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1, 2, 3}},
	}

	// Create a first match
	_, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 1}`)
	if err != nil {
		t.Fatalf("Could not make first POST request: %s", err.Error())
	}

	// Create a second match
	_, err = makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 2, "UserTwo": 3}`)
	if err != nil {
		t.Fatalf("Could not make second POST request: %s", err.Error())
	}

	// Read the match list
	res, err := makeRequest(mockEnv, http.MethodGet, "/match/all", "")
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"MatchList":[{"ID":0,"UserOne":0,"UserTwo":1,"MatchedOn":"2020-01-01T12:00:00.000000Z"},{"ID":1,"UserOne":2,"UserTwo":3,"MatchedOn":"2020-01-01T12:00:00.000000Z"}]}`
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

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 1}`)
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

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0}`)
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

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne"`)
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
	res, err := makeRequest(mockEnv, http.MethodPost, "/match", "")
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

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 123456, "UserTwo": 0}`)
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

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 123456}`)
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

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 123456, "UserTwo": 234567}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be read successfully
func TestReadMatchHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	// Create a single match
	_, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 1}`)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Read the match
	res, err := makeRequest(mockEnv, http.MethodGet, "/match/0", "")
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
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

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/", "")
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

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/123456", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no match exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the read endpoint fails
func TestReadUserHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/abcdef", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad Request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be updated successfully
func TestUpdateUserHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	// Create a single match
	_, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 1}`)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Update the match
	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne": 1, "UserTwo": 0}`)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"UserOne":1,"UserTwo":0,"MatchedOn":"2020-12-31T12:00:00.000000Z"}`
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

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the update endpoint fails
func TestUpdateMatchHandlerFailsOnMalformedJSONBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne"`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the update endpoint fails
func TestUpdateMatchHandlerFailsOnNoBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", "")
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

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/", "")
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

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/123456", `{"UserOne": 0, "UserTwo": 1}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no match exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the update endpoint fails
func TestUpdateMatchHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/abcdef", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad Request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an invalid UserOne reference to the update endpoint fails
func TestUpdateMatchHandlerFailsOnInvalidUserOne(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne": 123456, "UserTwo": 0}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an invalid UserTwo reference to the update endpoint fails
func TestUpdateMatchHandlerFailsOnInvalidUserTwo(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne": 0, "UserTwo": 123456}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing all invalid references to the update endpoint fails
func TestUpdateMatchHandlerFailsOnAllInvalidReferences(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/0", `{"UserOne": 123456, "UserTwo": 234567}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be deleted successfully
func TestDeleteMatchHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	// Create a single match
	_, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 1}`)
	if err != nil {
		t.Fatalf("Could not make POST request: %s", err.Error())
	}

	// Delete the match
	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/0", "")
	if err != nil {
		t.Fatalf("Could not make DELETE request: %s", err.Error())
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

	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/", "")
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

	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/123456", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Not Found, since no match exists with that given ID
	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a string ID to the delete endpoint fails
func TestDeleteMatchHandlerFailsOnStringID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/abcdef", "")
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Bad Request, since we require an integer
	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}
