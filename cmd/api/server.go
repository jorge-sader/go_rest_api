package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/jorge-sader/go-rest-api/internal/api/middlewares"
	"github.com/jorge-sader/go-rest-api/internal/api/repositories/sqlconnect"
	"github.com/jorge-sader/go-rest-api/internal/api/router"
	"github.com/jorge-sader/go-rest-api/pkg/utils"

	"golang.org/x/net/http2"
)

func init() {
	err := godotenv.Load("cmd/api/.env")
	if err != nil {
		log.Fatalln("error loading environment variables:", err)
	}
}

func main() {
	// Load certificate and key
	cert := "cmd/api/cert.pem"
	key := "cmd/api/key.pem"

	// Connect to database
	_, err := sqlconnect.ConnectDB()
	if err != nil {
		fmt.Println("error connecting to DB: ", err)
		return
	}

	// Configure TLS
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		// // Mutual TLS (mTLS) Config
		// // mTLS is more secure but impractical as Client must have same certificate as Server)
		// ClientAuth: tls.RequireAndVerifyClientCert, // enforce mTLS
		// ClientCAs:  loadClientCAs(),
	}

	// TODO: uncomment/reevaluate after routes are done
	// 	// rl := middlewares.NewRateLimiter(20, time.Minute)
	//
	// 	hppOptions := middlewares.HPPOptions{
	// 		CheckQuery:                  true,
	// 		CheckBody:                   true,
	// 		CheckBodyOnlyForContentType: "application/x-www-form-urlencoded",
	// 		Whitelist:                   []string{"sort_by", "sort_order", "first_name", "last_name", "classroom", "subject"},
	// 	}

	// secureMux establishes the middleware chain that secures our server
	// secureMux := middlewares.Cors(rl.Middleware(middlewares.ResponseTime(middlewares.SecurityHeaders(middlewares.Compression(middlewares.Hpp(hppOptions)(mux))))))
	secureMux := utils.ApplyMiddlewares(router.Router(),
		// Innermost (runs last, ends first)
		// middlewares.Hpp(hppOptions), // TODO: uncomment/reevaluate after routes are done
		// middlewares.Compression,     // TODO: uncomment/reevaluate after routes are done
		middlewares.SecurityHeaders,
		// middlewares.ResponseTime, // TODO: uncomment/reevaluate after routes are done
		// rl.Middleware,   // TODO: uncomment/reevaluate after routes are done
		middlewares.Cors,
		// Outermost (runs first, ends last)
	)

	port, err := strconv.Atoi(os.Getenv("SERVER_PORT"))
	if err != nil {

	}

	// Create custom server
	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	// Enable http2
	err = http2.ConfigureServer(server, &http2.Server{})
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
