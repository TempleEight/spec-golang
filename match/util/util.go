package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// Auth contains the unique identifier for a given auth
type Auth struct {
	ID uuid.UUID
}

// GetConfig returns a configuration object from decoding the given configuration file
func GetConfig(filePath string) (*Config, error) {
	config := Config{}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateErrorJSON returns a JSON string containing the key error associated with provided value
func CreateErrorJSON(message string) string {
	payload := map[string]string{"error": message}
	json, err := json.Marshal(payload)
	if err != nil {
		return err.Error()
	}
	return string(json)
}

// ExtractIDFromRequest extracts the parameter provided under parameter ID and converts it into a UUID
func ExtractIDFromRequest(requestParams map[string]string) (uuid.UUID, error) {
	id := requestParams["id"]
	if len(id) == 0 {
		return uuid.Nil, errors.New("No ID provided")
	}

	return uuid.Parse(id)
}

// ExtractAuthIDFromRequest extracts a token from a header of the form `Authorization: Bearer <token>`
func ExtractAuthIDFromRequest(headers http.Header) (*Auth, error) {
	authHeader := headers.Get("Authorization")
	if len(authHeader) == 0 {
		return nil, errors.New("Authorization header not provided")
	}

	// Extract and parse JWT
	rawToken := strings.Replace(authHeader, "Bearer ", "", 1)
	jwtParser := jwt.Parser{UseJSONNumber: true}
	token, _, err := jwtParser.ParseUnverified(rawToken, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	// Extract claims from JWT
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("JWT claims are invalid")
	}

	// Extract ID from JWT claims
	id, ok := claims["id"]
	if !ok {
		return nil, errors.New("JWT does not contain an id")
	}

	// Convert to a UUID
	uuid, err := uuid.Parse(id.(string))
	if err != nil {
		return nil, err
	}

	return &Auth{uuid}, nil
}
