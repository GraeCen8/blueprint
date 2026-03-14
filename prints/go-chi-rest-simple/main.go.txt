package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var items = []item{
	{ID: 1, Name: "sample"},
}

func main() {
	router := chi.NewRouter()
	router.Get("/items", listItems)
	router.Post("/items", createItem)

	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func listItems(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, items)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var payload item
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid payload"})
		return
	}
	if payload.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	payload.ID = items[len(items)-1].ID + 1
	items = append(items, payload)
	writeJSON(w, http.StatusCreated, payload)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
