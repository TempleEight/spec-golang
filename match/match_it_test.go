package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/TempleEight/spec-golang/match/comm"
	"github.com/TempleEight/spec-golang/match/dao"
	"github.com/TempleEight/spec-golang/match/util"
	"github.com/google/uuid"
)

var environment env

func TestMain(m *testing.M) {
	config, err := util.GetConfig("/etc/match-service/config.json")
	if err != nil {
		log.Fatal(err)
	}

	d, err := dao.Init(config)
	if err != nil {
		log.Fatal(err)
	}
	c := comm.Init(config)

	environment = env{d, c}

	// Create two users for the test
	url := config.Services["user"]
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(`{"Name": "Jay"}`)))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", JWT0))
	_, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	req, err = http.NewRequest(http.MethodPost, url, bytes.NewBuffer([]byte(`{"Name": "Lewis"}`)))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", JWT1))
	_, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestIntegrationMatch(t *testing.T) {
	id := testCreateMatch(t)
	testReadMatch(t, id)
	testReadMatchList(t)
	testUpdateMatch(t, id)
	testDeleteMatch(t, id)
}

func testCreateMatch(t *testing.T) uuid.UUID {
	// Create match
	res, err := makeRequest(environment, http.MethodPost, "/match", fmt.Sprintf(`{"UserOne": "%s", "UserTwo": "%s"}`, UUID0, UUID1), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	var createMatchResponse createMatchResponse
	err = json.Unmarshal([]byte(res.Body.String()), &createMatchResponse)
	if err != nil {
		t.Fatalf("Could not decode json: %s", err.Error())
	}

	if createMatchResponse.UserOne.String() != UUID0 {
		t.Fatalf("Wrong value for UserOne, received: %s, expected: %s", createMatchResponse.UserOne.String(), UUID0)
	}

	if createMatchResponse.UserTwo.String() != UUID1 {
		t.Fatalf("Wrong value for UserTwo, received: %s, expected: %s", createMatchResponse.UserTwo.String(), UUID1)
	}

	_, err = time.Parse(time.RFC3339, createMatchResponse.MatchedOn)
	if err != nil {
		t.Fatalf("MatchedOn was in an invalid format: %s", err.Error())
	}

	return createMatchResponse.ID
}

func testReadMatch(t *testing.T, uuid uuid.UUID) {
	res, err := makeRequest(environment, http.MethodGet, fmt.Sprintf("/match/%s", uuid.String()), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	var readMatchResponse readMatchResponse
	err = json.Unmarshal([]byte(res.Body.String()), &readMatchResponse)
	if err != nil {
		t.Fatalf("Could not decode json: %s", err.Error())
	}

	if readMatchResponse.UserOne.String() != UUID0 {
		t.Fatalf("Wrong value for UserOne, received: %s, expected: %s", readMatchResponse.UserOne.String(), UUID0)
	}

	if readMatchResponse.UserTwo.String() != UUID1 {
		t.Fatalf("Wrong value for UserTwo, received: %s, expected: %s", readMatchResponse.UserTwo.String(), UUID1)
	}

	_, err = time.Parse(time.RFC3339, readMatchResponse.MatchedOn)
	if err != nil {
		t.Fatalf("MatchedOn was in an invalid format: %s", err.Error())
	}
}

func testReadMatchList(t *testing.T) {
	// List all matches
	res, err := makeRequest(environment, http.MethodGet, "/match/all", "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	var listMatchResponse listMatchResponse
	err = json.Unmarshal([]byte(res.Body.String()), &listMatchResponse)
	if err != nil {
		t.Fatalf("Could not decode json: %s", err.Error())
	}

	if len(listMatchResponse.MatchList) != 1 {
		t.Fatalf("Incorrect number of matches returned, received: %d, expected: 1", len(listMatchResponse.MatchList))
	}
	match := listMatchResponse.MatchList[0]

	if match.UserOne.String() != UUID0 {
		t.Fatalf("Wrong value for UserOne, received: %s, expected: %s", match.UserOne.String(), UUID0)
	}

	if match.UserTwo.String() != UUID1 {
		t.Fatalf("Wrong value for UserTwo, received: %s, expected: %s", match.UserTwo.String(), UUID1)
	}

	_, err = time.Parse(time.RFC3339, match.MatchedOn)
	if err != nil {
		t.Fatalf("MatchedOn was in an invalid format: %s", err.Error())
	}
}

func testUpdateMatch(t *testing.T, uuid uuid.UUID) {
	// Update that same match by reversing the user order
	res, err := makeRequest(environment, http.MethodPut, fmt.Sprintf("/match/%s", uuid.String()), fmt.Sprintf(`{"UserOne":"%s", "UserTwo":"%s"}`, UUID1, UUID0), JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	var updateMatchResponse updateMatchResponse
	err = json.Unmarshal([]byte(res.Body.String()), &updateMatchResponse)
	if err != nil {
		t.Fatalf("Could not decode json: %s", err.Error())
	}

	if updateMatchResponse.UserOne.String() != UUID1 {
		t.Fatalf("Wrong value for UserOne, received: %s, expected: %s", updateMatchResponse.UserOne.String(), UUID1)
	}

	if updateMatchResponse.UserTwo.String() != UUID0 {
		t.Fatalf("Wrong value for UserTwo, received: %s, expected: %s", updateMatchResponse.UserTwo.String(), UUID0)
	}

	_, err = time.Parse(time.RFC3339, updateMatchResponse.MatchedOn)
	if err != nil {
		t.Fatalf("MatchedOn was in an invalid format: %s", err.Error())
	}
}

func testDeleteMatch(t *testing.T, uuid uuid.UUID) {
	res, err := makeRequest(environment, http.MethodDelete, fmt.Sprintf("/match/%s", uuid.String()), "", JWT0)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Fatalf("Wrong status code: %v", res.Code)
	}

	received := res.Body.String()
	expected := `{}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Fatalf("Handler returned incorrect body, received: %s expected: %s", received, expected)
	}
}
