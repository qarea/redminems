package cfg

import (
	"strings"

	"github.com/powerman/narada-go/narada"
)

var log = narada.NewLog("")

var (
	Debug bool
	HTTP  struct {
		Listen       string
		RealIPHeader string
		BasePath     string
	}
)

func init() {
	if err := load(); err != nil {
		log.Fatal(err)
	}
}

func load() error {
	Debug = narada.GetConfigLine("log/level") == "DEBUG"

	HTTP.Listen = narada.GetConfigLine("listen")
	if strings.Index(HTTP.Listen, ":") == -1 {
		log.Fatal("please setup config/listen")
	}

	HTTP.RealIPHeader = narada.GetConfigLine("real_ip_header")

	HTTP.BasePath = narada.GetConfigLine("basepath")
	if HTTP.BasePath != "" && (HTTP.BasePath[0] != '/' || HTTP.BasePath[len(HTTP.BasePath)-1] == '/') {
		log.Fatal("config/basepath should begin with / and should not end with /")
	}

	return nil
}
