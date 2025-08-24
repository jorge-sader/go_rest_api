package handlers

import (
	"fmt"
	"net/http"
)

func StudentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	case http.MethodGet:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	case http.MethodPut:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	case http.MethodPatch:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	case http.MethodDelete:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	}
	w.Write([]byte("Howdy Y'all")) // DEBUG
}
