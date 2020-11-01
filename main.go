package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	Elasticsearch "github.com/elastic/go-elasticsearch/v7"
	"github.com/gorilla/mux"
)

// Geolocation refers to the identification of the geographic location of a user
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
	var userGeolocation Geolocation

	err := json.NewDecoder(r.Body).Decode(&userGeolocation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cfg := Elasticsearch.Config{
		Addresses: []string{
			"localhost:8080",
		},
	}

	es, err := Elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	parkingSpotGeolocation := searchNearestUnoccupiedSpot(userGeolocation, es)

	bytes, _ := json.Marshal(parkingSpotGeolocation)

	fmt.Fprintf(w, string(bytes))
}

func searchNearestUnoccupiedSpot(userGeolocation Geolocation, es *Elasticsearch.Client) Geolocation {
	var requestBuff bytes.Buffer

	getESQuery(userGeolocation, &requestBuff)

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("parking-sensor"),
		es.Search.WithBody(&requestBuff),
		es.Search.WithTrackTotalHits(true),
		es.Search.WithPretty(),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	var esResponseBody map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&esResponseBody); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	var lat float64
	var lon float64

	for _, hit := range esResponseBody["hits"].(map[string]interface{})["hits"].([]interface{}) {
		location := hit.(map[string]interface{})["_source"].(map[string]interface{})["location"].(map[string]interface{})
		lat = location["lat"].(float64)
		lon = location["lon"].(float64)
	}

	return Geolocation{Lat: lat, Lon: lon}
}

func getESQuery(userGeolocation Geolocation, requestBuff *bytes.Buffer) {
	query := map[string]interface{}{
		"from": 0,
		"size": 1,
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"unoccupied": true,
			},
		},
		"sort": []map[string]interface{}{
			map[string]interface{}{
				"_geo_distance": map[string]interface{}{
					"location": map[string]interface{}{
						"lat": userGeolocation.Lat,
						"lon": userGeolocation.Lon,
					},
					"order":         "asc",
					"unit":          "km",
					"distance_type": "arc",
				},
			},
		},
	}

	if err := json.NewEncoder(requestBuff).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}
}
