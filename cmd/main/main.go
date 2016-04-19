// Package main provides main service.
package main

import (
	"github.com/powerman/narada-go/narada/bootstrap"

	"net/http"

	"github.com/powerman/narada-go/narada"

	_ "../../api/eventsvc/"
	_ "../../api/rpcsvc/"
	_ "../../api/websvc/"
	"../../cfg"
)

var log = narada.NewLog("")

func main() {
	bootstrap.Unlock()
	log.NOTICE("Listening on %s", cfg.HTTP.Listen)
	log.Fatal(http.ListenAndServe(cfg.HTTP.Listen, nil))
}
