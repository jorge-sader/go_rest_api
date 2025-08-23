package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"

	mw "github.com/jorge-sader/go-rest-api/internal/api/middlewares"
	"golang.org/x/net/http2"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	logRequestDetails(r)
	fmt.Fprintln(w, "Hello Gorgeous")
	w.Write([]byte("You look fantastic today ;)"))
	fmt.Println("Hello Gorgeous!")
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodGet:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodPut:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodPatch:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodDelete:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	}

	w.Write([]byte("Welcome brilliant teachers"))
	fmt.Println("Welcome brilliant teachers")
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodGet:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodPut:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodPatch:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodDelete:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	}
	w.Write([]byte("Howdy Y'all"))
	fmt.Println("Howdy Y'all")
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodGet:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodPut:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodPatch:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	case http.MethodDelete:
		fmt.Printf("Received %s request on '%s' route\n", r.Method, r.URL)
		fmt.Fprintf(w, "Received %s request on '%s' route\n", r.Method, r.URL)
	}
	w.Write([]byte("Good morning Executives!"))
	fmt.Println("Good morning Executives!")
}

func main() {

	port := 3000

	// Load certificate and key
	cert := "cmd/api/cert.pem"
	key := "cmd/api/key.pem"

	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/teachers", teachersHandler)
	mux.HandleFunc("/students", studentsHandler)
	mux.HandleFunc("/execs", execsHandler)

	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		// // Mutual TLS (mTLS) Config
		// // mTLS is more secure but impractical as Client must have same certificate as Server)
		// ClientAuth: tls.RequireAndVerifyClientCert, // enforce mTLS
		// ClientCAs:  loadClientCAs(),
	}

	// Create custom server
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   mw.Compression(mw.ResponseTime(mw.SecurityHeaders(mw.Cors(mux)))),
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
