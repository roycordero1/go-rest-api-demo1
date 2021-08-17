package main

import (
	"net/http"
	"rest-api-tutorial2/admin"
	"rest-api-tutorial2/coasters"
)

func main() {
	admin := admin.NewAdminHandler()
	coasterHandlers := coasters.NewCoasterHandlers()
	http.HandleFunc("/admin", admin.Handler)
	http.HandleFunc("/coasters", coasterHandlers.Coasters)
	http.HandleFunc("/coasters/", coasterHandlers.Coaster)
	http.ListenAndServe(":8081", nil)
}
