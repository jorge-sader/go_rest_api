package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	mw "github.com/jorge-sader/go-rest-api/internal/api/middlewares"
	"github.com/jorge-sader/go-rest-api/pkg/utils"
	"golang.org/x/net/http2"
)

type Teacher struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Classroom string `json:"classroom,omitempty"`
	Subject   string `json:"subject,omitempty"`
}

var teachers = make(map[int]Teacher)
var mutex = &sync.Mutex{}
var nextID = 1

// Initialize/mock some data
func init() {
	teachers[nextID] = Teacher{
		ID:        nextID,
		FirstName: "John",
		LastName:  "Doe",
		Classroom: "9A",
		Subject:   "Math",
	}
	nextID++

	teachers[nextID] = Teacher{
		ID:        nextID,
		FirstName: "Jane",
		LastName:  "Doe",
		Classroom: "10A",
		Subject:   "Algebra",
	}
	nextID++

	teachers[nextID] = Teacher{
		ID:        nextID,
		FirstName: "Jane",
		LastName:  "Smith",
		Classroom: "11A",
		Subject:   "Calculus",
	}
	nextID++
}

func getTeachersHandler(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/teachers/")
	idStr := strings.TrimSuffix(path, "/")

	if idStr == "" {
		firstName := r.URL.Query().Get("first_name")
		lastName := r.URL.Query().Get("last_name")

		// teacherList := []Teacher{}
		teacherList := make([]Teacher, 0, len(teachers))

		for _, teacher := range teachers {
			// Apply filters
			if (firstName == "" || teacher.FirstName == firstName) && (lastName == "" || teacher.LastName == lastName) {
				teacherList = append(teacherList, teacher)
			}
		}

		response := struct {
			Status string    `json:"status"`
			Count  int       `json:"count"`
			Data   []Teacher `json:"data"`
		}{
			Status: "success",
			Count:  len(teacherList),
			Data:   teacherList,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}

	// Handle path Parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		return
	}

	teacher, exists := teachers[id]
	if !exists {
		http.Error(w, "Teacher not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teacher)

}

func addTeacherHandler(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	var newTeachers []Teacher
	err := json.NewDecoder(r.Body).Decode(&newTeachers)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	addedTeachers := make([]Teacher, len(newTeachers))
	for i, newTeacher := range newTeachers {
		newTeacher.ID = nextID
		teachers[nextID] = newTeacher
		addedTeachers[i] = newTeacher
		nextID++
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string    `json:"status"`
		Count  int       `json:"count"`
		Data   []Teacher `json:"data"`
	}{
		Status: "success",
		Count:  len(addedTeachers),
		Data:   addedTeachers,
	}
	json.NewEncoder(w).Encode(response)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	logRequestDetails(r)
	fmt.Fprintln(w, "Hello Gorgeous")
	w.Write([]byte("You look fantastic today ;)"))
	fmt.Println("Hello Gorgeous!")
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		addTeacherHandler(w, r)

	case http.MethodGet:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
		getTeachersHandler(w, r)

	case http.MethodPut:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	case http.MethodPatch:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	case http.MethodDelete:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL) // DEBUG
	}

	w.Write([]byte("Welcome brilliant teachers")) // DEBUG
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
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

func execsHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Write([]byte("Good morning Executives!")) // DEBUG
}

func main() {

	port := 3000

	// Load certificate and key
	cert := "cmd/api/cert.pem"
	key := "cmd/api/key.pem"

	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/teachers/", teachersHandler)
	mux.HandleFunc("/students/", studentsHandler)
	mux.HandleFunc("/execs/", execsHandler)

	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		// // Mutual TLS (mTLS) Config
		// // mTLS is more secure but impractical as Client must have same certificate as Server)
		// ClientAuth: tls.RequireAndVerifyClientCert, // enforce mTLS
		// ClientCAs:  loadClientCAs(),
	}

	// TODO: uncomment/reevaluate after routes are done
	// 	// rl := mw.NewRateLimiter(20, time.Minute)
	//
	// 	hppOptions := mw.HPPOptions{
	// 		CheckQuery:                  true,
	// 		CheckBody:                   true,
	// 		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
	// 		Whitelist:                   []string{"sort_by", "sort_order", "first_name", "last_name", "classroom", "subject"},
	// 	}

	// secureMux establishes the middleware chain that secures our server
	// secureMux := mw.Cors(rl.Middleware(mw.ResponseTime(mw.SecurityHeaders(mw.Compression(mw.Hpp(hppOptions)(mux))))))
	secureMux := utils.ApplyMiddlewares(mux,
		// Innermost (runs last, ends first)
		// mw.Hpp(hppOptions), // TODO: uncomment/reevaluate after routes are done
		// mw.Compression,     // TODO: uncomment/reevaluate after routes are done
		mw.SecurityHeaders,
		// mw.ResponseTime, // TODO: uncomment/reevaluate after routes are done
		// rl.Middleware,   // TODO: uncomment/reevaluate after routes are done
		mw.Cors,
		// Outermost (runs first, ends last)
	)

	// Create custom server
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	// Enable http2
	err := http2.ConfigureServer(server, &http2.Server{})
	if err != nil {
		log.Println("Error enabling http2:", err)
	}

	fmt.Println("Server listening on port:", port)
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}

	// HTTP 1.1 Server without TLS
	// err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	// if err != nil {
	// 	log.Fatalln("Error starting server:", err)
	// }
}

func logRequestDetails(r *http.Request) {

	fmt.Println("Received the following", r.Proto, r.Method, "request:")
	fmt.Println("  Body:", r.Body)
	fmt.Println("  ContentLength:", r.ContentLength)
	fmt.Println("  Form:", r.Form)
	fmt.Println("  Header:", r.Header)
	fmt.Println("  Host:", r.Host)
	fmt.Println("  Method:", r.Method)
	fmt.Println("  Proto:", r.Proto)
	fmt.Println("  RemoteAddr:", r.RemoteAddr)
	fmt.Println("  RequestURI:", r.RequestURI)
	fmt.Println("  TLS:", r.TLS)
	fmt.Println("  TLSVersion:", GetTLSVersionName(r.TLS))
	fmt.Println("  Trailer:", r.Trailer)
	fmt.Println("  TransferEncoding:", r.TransferEncoding)
	fmt.Println("  URL:", r.URL)
	fmt.Println("  UserAgent:", r.UserAgent())

	// After Go 1.22 we can also get
	fmt.Println("  UserAgent:", r.URL.Path)

}

func GetTLSVersionName(tlsConnState *tls.ConnectionState) string {
	var tlsVersion string

	if tlsConnState != nil {
		version := tlsConnState.Version

		switch version {
		case tls.VersionTLS10:
			tlsVersion = "TLS 1.0"

		case tls.VersionTLS11:
			tlsVersion = "TLS 1.1"

		case tls.VersionTLS12:
			tlsVersion = "TLS 1.2"

		case tls.VersionTLS13:
			tlsVersion = "TLS 1.3"

		default:
			tlsVersion = "Unknown TLS version"
		}

	} else {
		tlsVersion = "Received request without TLS"
	}

	return tlsVersion
}

// loadClientCAs loads client certificates of authorization required for mutual TLS (mTLS)
func loadClientCAs() *x509.CertPool {
	clientCAs := x509.NewCertPool()

	caCert, err := os.ReadFile("cert.pem")
	if err != nil {
		log.Fatalln("Error loading client CA:", err)
	}

	clientCAs.AppendCertsFromPEM(caCert)
	return clientCAs
}
