package comm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TempleEight/spec-golang/auth/utils"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Handler maintains the list of services and their associated hostnames
type Handler struct {
	Services map[string]string
}

// consumerResponse encapsulates the response from Kong after creating a consumer
type consumerResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// JWTCredential store the issuer and secret key that must be used to sign requests
type JWTCredential struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

// Init sets up the Handler object with a list of services from the config
func (coms *Handler) Init(config *utils.Config) {
	coms.Services = config.Services
}

func createKongConsumer(hostname string) (*consumerResponse, error) {
	postData := url.Values{}
	postData.Set("username", "auth-service")

	res, err := http.PostForm(fmt.Sprintf("%s/consumers", hostname), postData)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		// If we have an error code, the message _should_ be in the body
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.New("unable to create JWT consumer")
		}
		return nil, errors.New(string(bodyBytes))
	}

	consumer := consumerResponse{}
	err = json.NewDecoder(res.Body).Decode(&consumer)
	if err != nil {
		return nil, err
	}

	return &consumer, nil
}

func requestCredential(hostname string, consumer *consumerResponse) (*JWTCredential, error) {
	reqUrl := fmt.Sprintf("%s/consumers/%s/jwt", hostname, consumer.Username)
	res, err := http.Post(reqUrl, "application/x-www-form-urlencoded", nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 201 {
		// If we have an error code, the message _should_ be in the body
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, errors.New("unable to create JWT token")
		}
		return nil, errors.New(string(bodyBytes))
	}

	jwt := JWTCredential{}
	err = json.NewDecoder(res.Body).Decode(&jwt)
	if err != nil {
		return nil, err
	}

	return &jwt, nil
}

// CreateJWTCredential provisions a new HS256 JWT credential
func (coms *Handler) CreateJWTCredential() (*JWTCredential, error) {
	hostname, ok := coms.Services["kong-admin"]
	if !ok {
		return nil, fmt.Errorf("service %s's hostname not in config file", "kong-admin")
	}

	// Create a consumer
	consumer, err := createKongConsumer(hostname)
	if err != nil {
		return nil, err
	}

	// Use the consumer to request a credential
	return requestCredential(hostname, consumer)
}
