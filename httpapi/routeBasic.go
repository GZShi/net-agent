package api

import (
	"errors"

	"github.com/GZShi/net-agent/socks5"
	"github.com/kataras/iris"
)

func (s *HTTPServer) enableBasicRoute(ns iris.Party, pathstr string) {
	party := ns.Party(pathstr)

	party.Get("socks5-password", func(ctx iris.Context) {
		name := ctx.URLParam("name")
		secret := ctx.URLParam("secret")

		if len(name) <= 3 || len(secret) <= 2 {
			response(ctx, errors.New("参数长度过短"), nil)
			return
		}

		pswd := socks5.CalcSum(name, secret)

		response(ctx, nil, struct {
			Name string `json:"username"`
			Pswd string `json:"password"`
		}{name, pswd})
	})
}
