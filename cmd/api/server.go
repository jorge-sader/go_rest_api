package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"

	mw "github.com/jorge-sader/go-rest-api/internal/api/middlewares"
	"github.com/jorge-sader/go-rest-api/internal/api/router"
	"github.com/jorge-sader/go-rest-api/pkg/utils"

	"golang.org/x/net/http2"
)

func main() {

	port := 3000

	// Load certificate and key
	cert := "cmd/api/cert.pem"
	key := "cmd/api/key.pem"

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
	secureMux := utils.ApplyMiddlewares(router.Router(),
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
