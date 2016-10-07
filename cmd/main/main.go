// Package main provides main service.
package main

import (
	"github.com/powerman/narada-go/narada"
	"github.com/powerman/narada-go/narada/bootstrap"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/qarea/ctxtg"
	"github.com/qarea/redminems/api/rpcsvc"
	"github.com/qarea/redminems/cfg"
	"github.com/qarea/redminems/redmine"

	"net/http"
)

var log = narada.NewLog("")

func main() {
	p, err := ctxtg.NewRSATokenParser(cfg.RSAPublicKey)
	if err != nil {
		log.Fatal(err)
	}

	r := redmine.NewClient(cfg.HTTP.Timeout)

	rpcsvc.Init(r, p)

	if err := bootstrap.Unlock(); err != nil {
		log.Fatal(err)
	}
	http.Handle(cfg.HTTP.BasePath+"/metrics", prometheus.Handler())
	log.NOTICE("Listening on %s", cfg.HTTP.Listen+cfg.HTTP.BasePath)
	log.Fatal(http.ListenAndServe(cfg.HTTP.Listen, nil))
}
