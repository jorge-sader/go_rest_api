package middlewares

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type HPPOptions struct {
	CheckQuery                  bool
	CheckBody                   bool
	CheckBodyOnlyForContentType string
	Whitelist                   []string
}

func Hpp(options HPPOptions) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if options.CheckBody &&
				(r.Method == http.MethodPost ||
					r.Method == http.MethodPut ||
					r.Method == http.MethodPatch) &&
				isValidContentType(r, options.CheckBodyOnlyForContentType) {
				filterBodyParams(r, options.Whitelist)
			}

			if options.CheckQuery && r.URL.Query() != nil {
				filterQueryParams(r, options.Whitelist)
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isValidContentType(r *http.Request, contentType string) bool {
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filterQueryParams(r *http.Request, whitelist []string) {
	query := r.URL.Query()

	for k, v := range query {
		if len(v) > 1 {
			query.Set(k, v[0]) // keep first
			// query.Set(k, v[len(v)-1]) // Keep last
		}
		if !slices.Contains(whitelist, k) {
			query.Del(k)
		}
	}
	r.URL.RawQuery = query.Encode()
}

func filterBodyParams(r *http.Request, whitelist []string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}

	for k, v := range r.Form {
		if len(v) > 1 {
			r.Form.Set(k, v[0]) // keep first
			// query.Set(k, v[len(v)-1]) // Keep last
		}
		if !slices.Contains(whitelist, k) {
			delete(r.Form, k)
		}
	}
}
