package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Auth contains the unique identifier for a given auth
type Auth struct {
	ID int64
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

// ExtractIDFromRequest extracts the parameter provided under parameter ID and converts it into an integer
func ExtractIDFromRequest(requestParams map[string]string) (int64, error) {
	idStr := requestParams["id"]
	if len(idStr) == 0 {
		return 0, errors.New("No ID provided")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, errors.New("Invalid ID provided")
	}

	return id, nil
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

	// Convert to an integer
	intID, err := id.(json.Number).Int64()
	if err != nil {
		return nil, err
	}

	return &Auth{intID}, nil
}
