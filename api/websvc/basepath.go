package websvc

import (
	_ "log" // work around goimports bug

	"github.com/powerman/narada-go/narada"
)

var basePath = narada.GetConfigLine("basepath")

func init() {
	if basePath != "" && (basePath[0] != '/' || basePath[len(basePath)-1] == '/') {
		log.Fatal("config/basepath should begin with / and should not end with /")
	}
}
