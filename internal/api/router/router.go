package router

import (
	"net/http"

	"github.com/jorge-sader/go-rest-api/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/teachers/", handlers.TeachersHandler)
	mux.HandleFunc("/students/", handlers.StudentsHandler)
	mux.HandleFunc("/execs/", handlers.ExecsHandler)

	return mux
}
