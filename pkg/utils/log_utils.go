package utils

import (
	"crypto/tls"
	"fmt"
	"net/http"
)

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

func LogRequestDetails(r *http.Request) {

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
