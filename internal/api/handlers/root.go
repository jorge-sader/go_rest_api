package handlers

import (
	"fmt"
	"net/http"

	"github.com/jorge-sader/go-rest-api/pkg/utils"
)

func RootHandler(w http.ResponseWriter, r *http.Request) {
	utils.LogRequestDetails(r)
	fmt.Fprintln(w, "Hello Gorgeous")
	w.Write([]byte("You look fantastic today ;)"))
	fmt.Println("Hello Gorgeous!")
}
