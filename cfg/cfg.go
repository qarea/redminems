package cfg

import (
	"strings"
	"time"

	"github.com/powerman/narada-go/narada"
)

var log = narada.NewLog("")

var (
	// Debug enable debug logs
	Debug bool

	// LockTimeout for narada.SharedLock
	LockTimeout time.Duration

	// RSAPublicKey for JWT token verification
	RSAPublicKey []byte

	// HTTP configuration for application http server
	HTTP struct {
		Listen   string
		BasePath string

		// Default timeout for http requests to adapters
		Timeout      time.Duration
		RealIPHeader string
	}
)

func init() {
	if err := load(); err != nil {
		log.Fatal(err)
	}
}

func load() error {
	Debug = narada.GetConfigLine("log/level") == "DEBUG"

	HTTP.Listen = narada.GetConfigLine("http/listen")
	if strings.Index(HTTP.Listen, ":") == -1 {
		log.Fatal("please setup config/http/listen")
	}

	HTTP.BasePath = narada.GetConfigLine("http/basepath")
	if HTTP.BasePath != "" && (HTTP.BasePath[0] != '/' || HTTP.BasePath[len(HTTP.BasePath)-1] == '/') {
		log.Fatal("config/http/basepath should begin with / and should not end with /")
	}

	HTTP.RealIPHeader = narada.GetConfigLine("http/real_ip_header")

	var err error
	RSAPublicKey, err = narada.GetConfig("rsa_public_key")
	if err != nil {
		return err
	}

	HTTP.Timeout = narada.GetConfigDuration("http/timeout")

	LockTimeout = narada.GetConfigDuration("lock_timeout")
	return nil
}
