// Package main provides main service.
package main

import (
	"github.com/powerman/narada-go/narada/bootstrap"
	"gitlab.qarea.org/tgms/ctxtg"
	"gitlab.qarea.org/tgms/redminems/api/rpcsvc"

	"net/http"

	"github.com/powerman/narada-go/narada"
	"github.com/prometheus/client_golang/prometheus"

	_ "../../api/rpcsvc"
	"../../cfg"
)

var log = narada.NewLog("")

func main() {
	p, err := ctxtg.NewRSATokenParser(cfg.RSAPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	// Implement your own trackerClient and pass it instead of "nil" to rpcsvc.Init
	// trackerClient = NewTrackerClient()
	rpcsvc.Init(nil, p)

	if err := bootstrap.Unlock(); err != nil {
		log.Fatal(err)
	}
	http.Handle(cfg.HTTP.BasePath+"/metrics", prometheus.Handler())
	log.NOTICE("Listening on %s", cfg.HTTP.Listen+cfg.HTTP.BasePath)
	log.Fatal(http.ListenAndServe(cfg.HTTP.Listen, nil))
}
