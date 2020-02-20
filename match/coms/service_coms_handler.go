package coms

import (
	"fmt"
	"github.com/TempleEight/spec-golang/match/utils"
	"net/http"
)

// Handler maintaints the list of services and their associated hostnames
type Handler struct {
	Services map[string]string
}

// Init sets up the Handler object with a list of services from the config
func (coms *Handler) Init(config *utils.Config) {
	coms.Services = config.Services
}

// CheckUser makes requests to the target service to check if a userId exists.
func (coms *Handler) CheckUser(userID int) (bool, error) {
	hostname, ok := coms.Services["user"]
	if !ok {
		return false, fmt.Errorf("service %s's hostname not in config file", "user")
	}

	resp, err := http.Get(fmt.Sprintf("%s/%d", hostname, userID))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}
