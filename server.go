package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Coaster struct {
	Name         string `json:"name"`
	Manufacturer string `json:"manufacturer"`
	ID           string `json:"id"`
	InPark       string `json:"in_park"`
	Height       int    `json:"height"`
}

type coastersHandler struct {
	sync.Mutex
	store map[string]Coaster
}

func newCoasterHandlers() *coastersHandler {
	return &coastersHandler{
		store: map[string]Coaster{},
	}
}

func (h *coastersHandler) coasters(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.list(w, r)
		return
	case "POST":
		h.post(w, r)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}
}

func (h *coastersHandler) list(w http.ResponseWriter, r *http.Request) {

	// List the coasters
	coasters := make([]Coaster, len(h.store))

	h.Lock()
	i := 0
	for _, coaster := range h.store {
		coasters[i] = coaster
		i++
	}
	h.Unlock()

	// Write body as json to return response
	jsonBytes, err := json.Marshal(coasters)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *coastersHandler) post(w http.ResponseWriter, r *http.Request) {

	// Read body to create the coaster
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("Need content-type 'application/json' but got '%s'", ct)))
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var coaster Coaster
	err = json.Unmarshal(bodyBytes, &coaster)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Save the coaster
	coaster.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	h.Lock()
	h.store[coaster.ID] = coaster
	defer h.Unlock()
}

func (h *coastersHandler) coaster(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if parts[2] == "random" {
		h.getRandom(w, r)
		return
	}

	switch r.Method {
	case "GET":
		h.get(w, r, parts[2])
		return
	case "PUT":
		h.put(w, r, parts[2])
		return
	case "DELETE":
		h.delete(w, r, parts[2])
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
		return
	}
}

func (h *coastersHandler) getRandom(w http.ResponseWriter, r *http.Request) {
	ids := make([]string, len(h.store))

	h.Lock()
	i := 0
	for id := range h.store {
		ids[i] = id
		i++
	}
	defer h.Unlock()

	var target string
	if len(ids) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if len(ids) == 1 {
		target = ids[0]
	} else {
		rand.Seed(time.Now().UnixNano())
		target = ids[rand.Intn(len(ids))]
	}

	w.Header().Add("location", fmt.Sprintf("/coasters/%s", target))
	w.WriteHeader(http.StatusFound)
}

func (h *coastersHandler) get(w http.ResponseWriter, r *http.Request, id string) {

	// Search for existing coaster
	h.Lock()
	coaster, ok := h.store[id]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Write body as json to return response
	jsonBytes, err := json.Marshal(coaster)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *coastersHandler) put(w http.ResponseWriter, r *http.Request, id string) {

	// Search for existing coaster
	h.Lock()
	_, ok := h.store[id]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Read body to update the coaster
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte(fmt.Sprintf("Need content-type 'application/json' but got '%s'", ct)))
		return
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	var newCoaster Coaster
	err = json.Unmarshal(bodyBytes, &newCoaster)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Save the coaster
	h.Lock()
	h.store[id] = newCoaster
	h.Unlock()
}

func (h *coastersHandler) delete(w http.ResponseWriter, r *http.Request, id string) {

	// Search for existing coaster
	h.Lock()
	_, ok := h.store[id]
	h.Unlock()
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Delete the coaster
	h.Lock()
	delete(h.store, id)
	h.Unlock()
}

type adminPortal struct {
	password string
}

func newAdminPortal() *adminPortal {
	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		panic("Required env var ADMIN_PASSWORD not set")
	}

	return &adminPortal{password: password}
}

func (a adminPortal) handler (w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" || pass != a.password {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}

	w.Write([]byte("<html><h1>Super secret admin portal</h1></html>"))
}

func main() {
	admin := newAdminPortal()
	coasterHandlers := newCoasterHandlers()
	http.HandleFunc("/admin", admin.handler)
	http.HandleFunc("/coasters", coasterHandlers.coasters)
	http.HandleFunc("/coasters/", coasterHandlers.coaster)
	http.ListenAndServe(":8081", nil)
}
