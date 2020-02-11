package main

import (
	"encoding/json"
	"errors"
	"strconv"
)

// CreateErrorJSON returns a JSON string containing the key error associated with provided value
func CreateErrorJSON(message string) string {
	payload := map[string]string{"error": message}
	json, err := json.Marshal(payload)
	if err != nil {
		return err.Error()
	}
	return string(json)
}

// ExtractUserIDFromRequest extracts the parameter provided under parameter ID and converts it into an integer
func ExtractUserIDFromRequest(requestParams map[string]string) (int64, error) {
	userIDStr := requestParams["id"]
	if len(userIDStr) == 0 {
		return 0, errors.New("No user ID provided")
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, errors.New("Invalid user ID provided")
	}

	return userID, nil
}
