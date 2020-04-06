// +build it

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/TempleEight/spec-golang/user/dao"
	"github.com/TempleEight/spec-golang/user/util"
)

var environment env

func TestMain(m *testing.M) {
	config, err := util.GetConfig("/etc/user-service/config.json")
	if err != nil {
		log.Fatal(err)
	}

	d, err := dao.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	environment = env{d, Hook{}}

	os.Exit(m.Run())
}

func TestIntegrationUser(t *testing.T) {
	// Create user
	res, err := makeRequest(environment, http.MethodPost, "/user", `{"Name": "Jay"}`, JWT1)
	if err != nil {
		t.Fatalf("Could not make request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code %v", res.Code)
	}

	received := res.Body.String()
	expected := fmt.Sprintf(`{"ID":"%s","Name":"Jay"}`, UUID1)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}

	// Read that same user
	res, err = makeRequest(environment, http.MethodGet, fmt.Sprintf("/user/%s", UUID1), "", JWT1)
	if err != nil {
		t.Fatalf("Could not make GET request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received = res.Body.String()
	expected = fmt.Sprintf(`{"ID":"%s","Name":"Jay"}`, UUID1)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}

	// Update that same user
	res, err = makeRequest(environment, http.MethodPut, fmt.Sprintf("/user/%s", UUID1), `{"Name": "Lewis"}`, JWT1)
	if err != nil {
		t.Fatalf("Could not make PUT request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received = res.Body.String()
	expected = fmt.Sprintf(`{"ID":"%s","Name":"Lewis"}`, UUID1)
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}

	// Delete that same user
	res, err = makeRequest(environment, http.MethodDelete, fmt.Sprintf("/user/%s", UUID1), "", JWT1)
	if err != nil {
		t.Fatalf("Could not make DELETE request: %s", err.Error())
	}

	if res.Code != http.StatusOK {
		t.Errorf("Wrong status code: %v", res.Code)
	}

	received = res.Body.String()
	expected = `{}`
	if expected != strings.TrimSuffix(received, "\n") {
		t.Errorf("Handler returned incorrect body: got %+v want %+v", received, expected)
	}
}
