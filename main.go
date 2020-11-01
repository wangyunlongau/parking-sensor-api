package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Geolocation struct {
	Lat float64
	Lon float64
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", handleRequest).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var geolocation Geolocation

	err := json.NewDecoder(r.Body).Decode(&geolocation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Geolocation: %+v", geolocation)
}

func searchNearestUnoccupiedSpot(userGeolocation Geolocation) {

}
