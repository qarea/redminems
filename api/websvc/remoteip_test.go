package websvc

import (
	"net/http"

	. "gopkg.in/check.v1"
)

type TestIP struct {
	origRealIPHeader string
}

var _ = Suite(&TestIP{})

func (s *TestIP) SetUpTest(c *C) {
	s.origRealIPHeader = realIPHeader
}

func (s *TestIP) TearDownTest(c *C) {
	realIPHeader = s.origRealIPHeader
}

func (s *TestIP) Test(c *C) {
	r, err := http.NewRequest("GET", "http://websvc.test/", nil)
	c.Assert(err, IsNil)
	r.RemoteAddr = "1.2.3.4:0"
	r.Header.Set("X-Real-IP", "4.3.2.1")

	realIPHeader = ""
	c.Check(remoteIP(r), Equals, "1.2.3.4")
	realIPHeader = "X-Real-IP"
	c.Check(remoteIP(r), Equals, "4.3.2.1")
	realIPHeader = "X-Real-REMOTE_ADDR"
	c.Check(func() { remoteIP(r) }, PanicMatches, ".*config/real_ip_header.*")
}
