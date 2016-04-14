// Package rpcsvc provides handlers for JSON-RPC 2.0.
package rpcsvc

import (
	"net/http"
	"net/rpc"

	"github.com/powerman/narada-go/narada"
	"github.com/powerman/rpc-codec/jsonrpc2"
)

var log = narada.NewLog("")

func init() {
	http.Handle("/rpc", jsonrpc2.HTTPHandler(nil))
	rpc.Register(RPC{})
}

type RPC struct{}

func (RPC) Version(args *struct{}, res *string) error {
	log.DEBUG("RPC: VERSION")
	*res, _ = narada.Version()
	return nil
}
