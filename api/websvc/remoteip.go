package websvc

import (
	"net"
	"net/http"

	"github.com/powerman/narada-go/narada"
)

var realIPHeader = narada.GetConfigLine("real_ip_header")

func remoteIP(r *http.Request) (ip string) {
	if realIPHeader == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	} else {
		ip = r.Header.Get(realIPHeader)
	}
	if ip == "" {
		panic("failed to detect remote IP, check config/real_ip_header")
	}
	return
}
