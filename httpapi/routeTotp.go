package api

import (
	log "github.com/GZShi/net-agent/logger"
	"github.com/GZShi/net-agent/totp"
	"github.com/kataras/iris"
)

// TotpInfo 持久化存储的totp账号信息
type TotpInfo struct {
	Account string `json:"account"`
	Secret  string `json:"secret"`
	URL     string `json:"url"`
}

// SetTotpInfo 初始化TotpInfo
func (s *HTTPServer) SetTotpInfo(infos []TotpInfo) {
	if infos == nil {
		return
	}
	m := make(map[string]*TotpInfo)

	for _, info := range infos {
		m[info.Account] = &info
		log.Get().WithField("account", info.Account).Info("add totp info")
	}

	s.totpMap = m
}

func (s *HTTPServer) enableTotpRoute(ns iris.Party, pathstr string) {
	party := ns.Party(pathstr)

	party.Get("register", func(ctx iris.Context) {
		account := ctx.URLParam("account")
		secret := totp.GenSecret(0)

		response(ctx, nil, TotpInfo{
			Account: account,
			Secret:  secret,
			URL:     totp.GetSecretURL(account, secret),
		})
	})
}
