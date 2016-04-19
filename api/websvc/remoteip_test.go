package websvc

import (
	"net/http"

	. "gopkg.in/check.v1"

	"../../cfg"
)

type TestIP struct {
	origRealIPHeader string
}

var _ = Suite(&TestIP{})

func (s *TestIP) SetUpTest(c *C) {
	s.origRealIPHeader = cfg.HTTP.RealIPHeader
}

func (s *TestIP) TearDownTest(c *C) {
	cfg.HTTP.RealIPHeader = s.origRealIPHeader
}

func (s *TestIP) Test(c *C) {
	r, err := http.NewRequest("GET", "http://websvc.test/", nil)
	c.Assert(err, IsNil)
	r.RemoteAddr = "1.2.3.4:0"
	r.Header.Set("X-Real-IP", "4.3.2.1")

	cfg.HTTP.RealIPHeader = ""
	c.Check(remoteIP(r), Equals, "1.2.3.4")
	cfg.HTTP.RealIPHeader = "X-Real-IP"
	c.Check(remoteIP(r), Equals, "4.3.2.1")
	cfg.HTTP.RealIPHeader = "X-Real-REMOTE_ADDR"
	// TODO test log.Fatal using subprocess
	// c.Check(func() { remoteIP(r) }, PanicMatches, ".*config/real_ip_header.*")
}
