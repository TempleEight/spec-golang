package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	authDAO "github.com/TempleEight/spec-golang/auth/dao"
	"github.com/TempleEight/spec-golang/auth/utils"
	valid "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
)

var dao authDAO.DAO

func main() {
	configPtr := flag.String("config", "/etc/auth-service/config.json", "configuration filepath")
	flag.Parse()

	// Require all struct fields by default
	valid.SetFieldsRequiredByDefault(true)

	config, err := utils.GetConfig(*configPtr)
	if err != nil {
		log.Fatal(err)
	}

	dao = authDAO.DAO{}
	err = dao.Init(config)
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/auth", authCreateHandler).Methods(http.MethodPost)
	log.Fatal(http.ListenAndServe(":82", r))
}

func authCreateHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(struct{}{})
}
