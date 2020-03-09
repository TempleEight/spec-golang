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
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that a match is not created if UserOne doesn't exist
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

// Test that a match is not created if UserTwo doesn't exist
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

// Test that a match is not created if every reference doesn't exist
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

// Test that a match is not created if the request body is not complete
func TestCreateMatchHandlerFailsOnOnlyProvidingOneUser(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 123456}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that creating a match fails if the request body is malformed
func TestCreateMatchHandlerFailsOnMalformedJSONBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []int64{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"Use}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}
