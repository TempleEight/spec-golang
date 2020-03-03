package comm

import (
	"fmt"
	"net/http"

	"github.com/TempleEight/spec-golang/match/util"
)

// Comm provides the interface adopted by Handler, allowing for mocking
type Comm interface {
	CheckUser(userID int) (bool, error)
}

// Handler maintains the list of services and their associated hostnames
type Handler struct {
	Services map[string]string
}

// Init sets up the Handler object with a list of services from the config
func Init(config *util.Config) *Handler {
	return &Handler{config.Services}
}

// CheckUser makes a request to the target service to check if a user ID exists
func (comm *Handler) CheckUser(userID int) (bool, error) {
	hostname, ok := comm.Services["user"]
	if !ok {
		return false, fmt.Errorf("service %s's hostname not in config file", "user")
	}

	resp, err := http.Get(fmt.Sprintf("%s/%d", hostname, userID))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}
