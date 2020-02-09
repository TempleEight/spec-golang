package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/matches", matchGetHandler).Methods(http.MethodGet)

	log.Fatal(http.ListenAndServe(":80", r))
}

func matchGetHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode("Hello, World")
}
