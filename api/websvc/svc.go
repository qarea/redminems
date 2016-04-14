// Package websvc provides handlers for Web.
package websvc

import (
	"fmt"
	"net/http"

	"github.com/powerman/narada-go/narada"
)

var log = narada.NewLog("")

func init() {
	http.Handle(basePath+"/web", http.StripPrefix(basePath, http.HandlerFunc(logResponse(web))))
}

func web(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!\n")
}
