package main

import (
	"jfrog-test/src/common"
	"jfrog-test/src/populardownloads"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	pdroutes := &populardownloads.PopularDownloads{}
	pdroutes.Init()

	router := mux.NewRouter()
	router.HandleFunc("/populardownloads", pdroutes.ListTopDownloads).Methods("Post")

	ipAddr := common.GetOutboundIP()
	myHost := ipAddr.String() + ":8080"

	srv := &http.Server{
		Handler: router,
		Addr:    myHost,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Server starts listening @%s ...", myHost)
	log.Fatal(srv.ListenAndServe())
}
