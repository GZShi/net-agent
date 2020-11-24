package main

import (
	"errors"

	"github.com/GZShi/net-agent/exchanger"
	"github.com/GZShi/net-agent/tunnel"
)

func newJoinClusterHandler(ts exchanger.Cluster) tunnel.OnRequestFunc {
	return func(ctx tunnel.Context) {
		ctx.Error(errors.New("bad handler"))
	}
}
