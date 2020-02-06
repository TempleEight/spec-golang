package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/TempleEight/spec-golang/user/dao"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/user/{id}", userGetHandler).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe(":80", r))
}

func userGetHandler(w http.ResponseWriter, r *http.Request) {
	userIDStr := mux.Vars(r)["id"]
	if len(userIDStr) == 0 {
		http.Error(w, "No user ID provided", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID provided", http.StatusBadRequest)
		return
	}

	user, err := dao.GetUser(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Something went wrong: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(user)
}
