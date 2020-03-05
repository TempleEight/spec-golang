package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/match/dao"
)

type MockMatch struct {
	ID        int
	UserOne   int
	UserTwo   int
	MatchedOn string
}

type MockDAO struct {
	Matches []MockMatch
}

type MockComm struct {
	UserIds []int
}

func (md *MockDAO) ListMatch() (*dao.MatchListResponse, error) {
	matches := make([]dao.MatchReadResponse, 0)
	for _, match := range md.Matches {
		matches = append(matches, dao.MatchReadResponse{
			ID:        match.ID,
			UserOne:   match.UserOne,
			UserTwo:   match.UserTwo,
			MatchedOn: match.MatchedOn,
		})
	}

	return &dao.MatchListResponse{
		MatchList: matches,
	}, nil
}

func (md *MockDAO) CreateMatch(request dao.MatchCreateRequest) (*dao.MatchCreateResponse, error) {
	mockMatch := MockMatch{len(md.Matches), *request.UserOne, *request.UserTwo, "2020-01-01T12:00:00.000000Z"}
	md.Matches = append(md.Matches, mockMatch)
	return &dao.MatchCreateResponse{
		ID:        mockMatch.ID,
		UserOne:   mockMatch.UserOne,
		UserTwo:   mockMatch.UserTwo,
		MatchedOn: mockMatch.MatchedOn,
	}, nil
}

func (md *MockDAO) ReadMatch(matchID int64) (*dao.MatchReadResponse, error) {
	for _, match := range md.Matches {
		if int64(match.ID) == matchID {
			return &dao.MatchReadResponse{
				ID:        match.ID,
				UserOne:   match.UserOne,
				UserTwo:   match.UserTwo,
				MatchedOn: match.MatchedOn,
			}, nil
		}
	}
	return nil, dao.ErrMatchNotFound(matchID)
}

func (md *MockDAO) UpdateMatch(matchID int64, request dao.MatchUpdateRequest) (*dao.MatchUpdateResponse, error) {
	for i, match := range md.Matches {
		if int64(match.ID) == matchID {
			md.Matches[i].UserOne = *request.UserOne
			md.Matches[i].UserTwo = *request.UserTwo
			md.Matches[i].MatchedOn = "2020-12-31T12:00:00.000000Z"
			return &dao.MatchUpdateResponse{
				ID:        md.Matches[i].ID,
				UserOne:   md.Matches[i].UserOne,
				UserTwo:   md.Matches[i].UserTwo,
				MatchedOn: md.Matches[i].MatchedOn,
			}, nil
		}
	}
	return nil, dao.ErrMatchNotFound(matchID)
}

func (md *MockDAO) DeleteMatch(matchID int64) error {
	for i, match := range md.Matches {
		if int64(match.ID) == matchID {
			md.Matches = append(md.Matches[:i], md.Matches[i+1:]...)
			return nil
		}
	}
	return dao.ErrMatchNotFound(matchID)
}

func (mc *MockComm) CheckUser(userID int) (bool, error) {
	for _, id := range mc.UserIds {
		if id == userID {
			return true, nil
		}
	}
	return false, nil
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

// Test that a match can be created successfully, assuming the users exist
func TestMatchCreateHandlerSucceeds(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Matches: make([]MockMatch, 0)},
		&MockComm{UserIds: []int{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 1}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

	received := res.Body.String()
	expected := `{"ID":0,"UserOne":0,"UserTwo":1,"MatchedOn":"2020-01-01T12:00:00.000000Z"}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}

// Test that a match is not created if one of the users doesn't exist
func TestMatchCreateHandlerFailsOnOneInvalidUser(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Matches: make([]MockMatch, 0)},
		&MockComm{UserIds: []int{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 0, "UserTwo": 123456}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that a match is not created if both of the users don't exist
func TestMatchCreateHandlerFailsOnBothInvalidUsers(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Matches: make([]MockMatch, 0)},
		&MockComm{UserIds: []int{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 123456, "UserTwo": 234567}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

func TestMatchCreateHandlerFailsOnOnlyProvidingOneUser(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Matches: make([]MockMatch, 0)},
		&MockComm{UserIds: []int{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne": 123456}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}

// Test that creating a match fails if the request body is malformed
func TestMatchCreateHandlerFailsOnMalformedJSONBody(t *testing.T) {
	var mockEnv = Env{
		&MockDAO{Matches: make([]MockMatch, 0)},
		&MockComm{UserIds: []int{0, 1}},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"Use}`)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code %v", res.Code)
	}
}
