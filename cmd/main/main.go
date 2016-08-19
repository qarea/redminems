// Package main provides main service.
package main

import (
	"github.com/powerman/narada-go/narada/bootstrap"

	"net/http"

	"github.com/powerman/narada-go/narada"

	_ "../../api/eventsvc"
	_ "../../api/rpcsvc"
	_ "../../api/websvc"
	"../../cfg"
)

var log = narada.NewLog("")

func main() {
	if err := bootstrap.Unlock(); err != nil {
		log.Fatal(err)
	}
	log.NOTICE("Listening on %s", cfg.HTTP.Listen+cfg.HTTP.BasePath)
	log.Fatal(http.ListenAndServe(cfg.HTTP.Listen+cfg.HTTP.BasePath, nil))
}
