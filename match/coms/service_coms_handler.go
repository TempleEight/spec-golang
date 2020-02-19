package coms

import (
	"github.com/TempleEight/spec-golang/match/utils"
	"fmt"
	"net/http"
)

// Handler maintaints the list of services and their associated hostnames
type Handler struct {
	Services map[string]string
}

// Initialise sets up the Handler object with a list of services from the config
func (coms *Handler) Initialise(config *utils.Config) {
	coms.Services = config.Services
}

// CheckUser makes requests to the target service to check if a userId exists.
func (coms *Handler) CheckUser(service string, userID int) (bool, error) {
	hostname := coms.Services[service]

	resp, err := http.Get(fmt.Sprintf("%s/%d", hostname, userID))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}