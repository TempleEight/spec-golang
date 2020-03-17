package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/TempleEight/spec-golang/match/dao"
	"github.com/google/uuid"
)

// Define 2 UUIDs
const UUID0 = "00000000-1234-5678-9012-000000000000"
const UUID1 = "00000000-1234-5678-9012-000000000001"

// Define 2 JWTs corresponding to UUIDs
const JWT0 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODMyNTc5NzcsImlkIjoiMDAwMDAwMDAtMTIzNC01Njc4LTkwMTItMDAwMDAwMDAwMDAwIiwiaXNzIjoiZkZTOEttVll1S0FDeUYzd2RwUEtIU1FxbVpWVndqRHEifQ.jMpelsEJUwONtRCQnQCo2v5Ph7cZHloc5R1OvKkU2Ck"
const JWT1 = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1ODMyNTc5NzcsImlkIjoiMDAwMDAwMDAtMTIzNC01Njc4LTkwMTItMDAwMDAwMDAwMDAxIiwiaXNzIjoiZkZTOEttVll1S0FDeUYzd2RwUEtIU1FxbVpWVndqRHEifQ.lMmkaK9L2kD2ZnbblSlXdz93cz6jZCALR0KoGlzQKpc"

// Define 3 Match UUIDs
const matchUUID0 = "00000001-1234-5678-9012-000000000000"
const matchUUID1 = "00000001-1234-5678-9012-000000000001"
const matchUUID2 = "00000001-1234-5678-9012-000000000002"

// Define 3 User UUIDs
const userUUID0 = "00000002-1234-5678-9012-000000000000"
const userUUID1 = "00000002-1234-5678-9012-000000000001"
const userUUID2 = "00000002-1234-5678-9012-000000000002"

const time0 = "2020-01-01T12:00:00Z"
const time1 = "2020-12-31T12:00:00Z"

type mockDAO struct {
	matchList []dao.Match
}

type mockComm struct {
	userIDs []uuid.UUID
}

func (md *mockDAO) ListMatch(input dao.ListMatchInput) (*[]dao.Match, error) {
	mockMatchList := make([]dao.Match, 0)
	for _, match := range md.matchList {
		if match.CreatedBy == input.AuthID {
			mockMatchList = append(mockMatchList, match)
		}
	}

	return &mockMatchList, nil
}

func (md *mockDAO) CreateMatch(input dao.CreateMatchInput) (*dao.Match, error) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		return nil, err
	}

	mockMatch := dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: input.AuthID,
		UserOne:   input.UserOne,
		UserTwo:   input.UserTwo,
		MatchedOn: time,
	}
	md.matchList = append(md.matchList, mockMatch)
	return &dao.Match{
		ID:        mockMatch.ID,
		CreatedBy: mockMatch.CreatedBy,
		UserOne:   mockMatch.UserOne,
		UserTwo:   mockMatch.UserTwo,
		MatchedOn: mockMatch.MatchedOn,
	}, nil
}

func (md *mockDAO) ReadMatch(input dao.ReadMatchInput) (*dao.Match, error) {
	for _, match := range md.matchList {
		if match.ID == input.ID {
			return &match, nil
		}
	}
	return nil, dao.ErrMatchNotFound(input.ID.String())
}

func (md *mockDAO) UpdateMatch(input dao.UpdateMatchInput) (*dao.Match, error) {
	time, err := time.Parse(time.RFC3339, time1)
	if err != nil {
		return nil, err
	}

	for i, match := range md.matchList {
		if match.ID == input.ID {
			md.matchList[i].UserOne = input.UserOne
			md.matchList[i].UserTwo = input.UserTwo
			md.matchList[i].MatchedOn = time
			return &dao.Match{
				ID:        md.matchList[i].ID,
				CreatedBy: md.matchList[i].CreatedBy,
				UserOne:   md.matchList[i].UserOne,
				UserTwo:   md.matchList[i].UserTwo,
				MatchedOn: md.matchList[i].MatchedOn,
			}, nil
		}
	}
	return nil, dao.ErrMatchNotFound(input.ID.String())
}

func (md *mockDAO) DeleteMatch(input dao.DeleteMatchInput) error {
	for i, match := range md.matchList {
		if match.ID == input.ID {
			md.matchList = append(md.matchList[:i], md.matchList[i+1:]...)
			return nil
		}
	}
	return dao.ErrMatchNotFound(input.ID.String())
}

func (mc *mockComm) CheckUser(userID uuid.UUID, token string) (bool, error) {
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
	defaultRouter(&env).ServeHTTP(rec, req)
	return rec, nil
}

// Test that a match list can be read successfully for a given ID
func TestListMatchHandlerSucceeds(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{
		dao.Match{
			ID:        uuid.MustParse(matchUUID0),
			CreatedBy: uuid.MustParse(UUID0),
			UserOne:   uuid.MustParse(userUUID0),
			UserTwo:   uuid.MustParse(userUUID1),
			MatchedOn: time,
		},
		dao.Match{
			ID:        uuid.MustParse(matchUUID1),
			CreatedBy: uuid.MustParse(UUID0),
			UserOne:   uuid.MustParse(userUUID0),
			UserTwo:   uuid.MustParse(userUUID2),
			MatchedOn: time,
		},
		dao.Match{
			ID:        uuid.MustParse(matchUUID2),
			CreatedBy: uuid.MustParse(UUID1),
			UserOne:   uuid.MustParse(userUUID1),
			UserTwo:   uuid.MustParse(userUUID2),
			MatchedOn: time,
		},
	}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
			uuid.MustParse(userUUID2),
		}},
		Hook{},
	}

	// Read the match list for UUID0
	res, err := makeRequest(mockEnv, http.MethodGet, "/match/all", "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"MatchList":[{"ID":"%s","UserOne":"%s","UserTwo":"%s","MatchedOn":"%s"},{"ID":"%s","UserOne":"%s","UserTwo":"%s","MatchedOn":"%s"}]}`,
		matchUUID0, userUUID0, userUUID1, time0, matchUUID1, userUUID0, userUUID2, time0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that a single match can be created successfully
func TestCreateMatchHandlerSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0,
		userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","UserOne":"%s","UserTwo":"%s","MatchedOn":"%s"}`, matchUUID0,
		userUUID0, userUUID1, time0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that providing an incomplete body to the create endpoint fails
func TestCreateMatchHandlerFailsOnIncompleteBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s"}`, userUUID0), JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", `{"UserOne"`, JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	// Create a single match
	res, err := makeRequest(mockEnv, http.MethodPost, "/match", "", JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`,
		uuid.Nil.String(), userUUID0), JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`,
		userUUID0, uuid.Nil.String()), JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, uuid.Nil.String(), uuid.Nil.String()), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a before create hook is successfully invoked
func TestCreateMatchHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	mockEnv.hook.BeforeCreate(func(env *env, req createMatchRequest, input *dao.CreateMatchInput) *HookError {
		input.UserOne = uuid.Nil
		return nil
	})

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, UUID0, UUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Fatalf("Wrong status code: %v", res.Code)
	}
}

// Test that a before create hook is successfully invoked and request is aborted
func TestCreateMatchHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	mockEnv.hook.BeforeCreate(func(env *env, req createMatchRequest, input *dao.CreateMatchInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Fatalf("Wrong status code: %v", res.Code)
	}
}

// Test that an after create hook is successfully invoked
func TestCreateMatchHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	mockEnv.hook.AfterCreate(func(env *env, match *dao.Match) *HookError {
		match.ID = uuid.Nil
		return nil
	})

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","UserOne":"%s","UserTwo":"%s","MatchedOn":"%s"}`, uuid.Nil, userUUID0, userUUID1, time0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that an after create hook is successfully invoked
func TestCreateMatchHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	mockEnv.hook.AfterCreate(func(env *env, match *dao.Match) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Fatalf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be read successfully
func TestReadMatchHandlerSucceeds(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: uuid.MustParse(UUID0),
		UserOne:   uuid.MustParse(userUUID0),
		UserTwo:   uuid.MustParse(userUUID1),
		MatchedOn: time,
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/match/%s", matchUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","UserOne":"%s","UserTwo":"%s","MatchedOn":"%s"}`, matchUUID0,
		userUUID0, userUUID1, time0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that providing no ID to the read endpoint fails
func TestReadMatchHandlerFailsOnEmptyID(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, "/match/", "", JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/match/%s", uuid.Nil.String()), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no match exists with the given ID
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a before read hook is successfully invoked
func TestReadMatchHandlerBeforeHookSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	mockEnv.hook.BeforeRead(func(env *env, input *dao.ReadMatchInput) *HookError {
		input.ID = uuid.Nil
		return nil
	})

	// Create a single match
	_, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read that same match
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/match/%s", matchUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusNotFound {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a before read hook is successfully invoked
func TestReadMatchHandlerBeforeHookAbortsRequest(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	mockEnv.hook.BeforeRead(func(env *env, input *dao.ReadMatchInput) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single match
	_, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read a single match
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/match/%s", matchUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that an after read hook is successfully invoked
func TestReadMatchHandlerAfterHookSucceeds(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	mockEnv.hook.AfterRead(func(env *env, match *dao.Match) *HookError {
		match.ID = uuid.Nil
		return nil
	})

	// Create a single match
	_, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read a single match
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/match/%s", matchUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","UserOne":"%s","UserTwo":"%s","MatchedOn":"%s"}`, uuid.Nil, userUUID0, userUUID1, time0)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that an after read hook is successfully invoked and request is aborted
func TestReadUserHandlerAfterHookAbortsRequest(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	mockEnv.hook.AfterRead(func(env *env, user *dao.Match) *HookError {
		return &HookError{http.StatusTeapot, errors.New("Example")}
	})

	// Create a single match
	_, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Read a single match
	res, err := makeRequest(mockEnv, http.MethodGet, fmt.Sprintf("/match/%s", matchUUID0), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusTeapot {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be updated successfully
func TestUpdateUserHandlerSucceeds(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: uuid.MustParse(UUID0),
		UserOne:   uuid.MustParse(userUUID0),
		UserTwo:   uuid.MustParse(userUUID1),
		MatchedOn: time,
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
			uuid.MustParse(userUUID2),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/match/%s", matchUUID0),
		fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID2), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","UserOne":"%s","UserTwo":"%s","MatchedOn":"%s"}`,
		matchUUID0, userUUID0, userUUID2, time1)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: received %+v, expected %+v", received, expected)
	}
}

// Test that providing an incomplete to the update endpoint fails
func TestUpdateMatchHandlerFailsOnIncompleteBody(t *testing.T) {
	mockEnv := env{
		&mockDAO{matchList: make([]dao.Match, 0)},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s"}`, userUUID0), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing a malformed JSON body to the update endpoint fails
func TestUpdateMatchHandlerFailsOnMalformedJSONBody(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: uuid.MustParse(UUID0),
		UserOne:   uuid.MustParse(userUUID0),
		UserTwo:   uuid.MustParse(userUUID1),
		MatchedOn: time,
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/match/%s", matchUUID0), `{"UserOne"`, JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing no body to the update endpoint fails
func TestUpdateMatchHandlerFailsOnNoBody(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: uuid.MustParse(UUID0),
		UserOne:   uuid.MustParse(userUUID0),
		UserTwo:   uuid.MustParse(userUUID1),
		MatchedOn: time,
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/match/%s", matchUUID0), "", JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, "/match/", "", JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/match/%s", uuid.Nil.String()),
		fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0, userUUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no match exists with the given ID
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an invalid UserOne reference to the update endpoint fails
func TestUpdateMatchHandlerFailsOnInvalidUserOne(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: uuid.MustParse(UUID0),
		UserOne:   uuid.MustParse(userUUID0),
		UserTwo:   uuid.MustParse(userUUID1),
		MatchedOn: time,
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/match/%s", matchUUID0),
		fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, uuid.Nil.String(), userUUID0), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing an invalid UserTwo reference to the update endpoint fails
func TestUpdateMatchHandlerFailsOnInvalidUserTwo(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: uuid.MustParse(UUID0),
		UserOne:   uuid.MustParse(userUUID0),
		UserTwo:   uuid.MustParse(userUUID1),
		MatchedOn: time,
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/match/%s", matchUUID0),
		fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, userUUID0,
			uuid.Nil.String()), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that providing all invalid references to the update endpoint fails
func TestUpdateMatchHandlerFailsOnAllInvalidReferences(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: uuid.MustParse(UUID0),
		UserOne:   uuid.MustParse(userUUID0),
		UserTwo:   uuid.MustParse(userUUID1),
		MatchedOn: time,
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodPut, fmt.Sprintf("/match/%s", matchUUID0),
		fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, uuid.Nil.String(),
			uuid.Nil.String()), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}

// Test that a single match can be deleted successfully
func TestDeleteMatchHandlerSucceeds(t *testing.T) {
	time, err := time.Parse(time.RFC3339, time0)
	if err != nil {
		t.Fatalf("Could not parse time: %s", err.Error())
	}

	// Populate mock datastore
	matchList := []dao.Match{dao.Match{
		ID:        uuid.MustParse(matchUUID0),
		CreatedBy: uuid.MustParse(UUID0),
		UserOne:   uuid.MustParse(userUUID0),
		UserTwo:   uuid.MustParse(userUUID1),
		MatchedOn: time,
	}}

	mockEnv := env{
		&mockDAO{matchList},
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/match/%s", matchUUID0), "", JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, "/match/", "", JWT0)
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
		&mockComm{userIDs: []uuid.UUID{
			uuid.MustParse(userUUID0),
			uuid.MustParse(userUUID1),
		}},
		Hook{},
	}

	res, err := makeRequest(mockEnv, http.MethodDelete, fmt.Sprintf("/match/%s", uuid.Nil.String()), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	// Unauthorized, since no match exists with the given ID
	if res.Code != http.StatusUnauthorized {
		t.Errorf("Wrong status code: %v", res.Code)
	}
}
