package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"regexp"
	"time"

	log "github.com/GZShi/net-agent/logger"
)

type totpInfo struct {
	Account string `json:"account"`
	Secret  string `json:"secret"`
	URL     string `json:"url"`
}

type config struct {
	Mode      string `json:"mode"`
	Addr      string `json:"addr"`
	Secret    string `json:"secret"`
	WhiteList string `json:"whiteList"` // TODO:可访问白名单

	// client only
	ClientName  string `json:"clientName"`
	ChannelName string `json:"channelName"`

	// server only
	PortProxy string     `json:"portProxy"`
	TotpList  []totpInfo `json:"totpList"`
}

func readJSON(path string) ([]byte, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`(^|\n)\s*\/\/.*`)
	text := re.ReplaceAllString(string(bytes), "")

	return []byte(text), nil
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./config.json", "the path of config file")
	flag.Parse()

	data, err := readJSON(configPath)
	if err != nil {
		log.Get().WithField("path", configPath).WithError(err).Error("read config file failed")
		return
	}
	log.Get().WithField("path", configPath).Info("read config file success")

	var cfg config
	if err = json.Unmarshal(data, &cfg); err != nil {
		log.Get().WithError(err).Error("parse json config failed")
		return
	}

	log.Get().WithField("mode", cfg.Mode).Info("will run as config.mode")

	switch cfg.Mode {
	case "client":
		if len(cfg.ClientName) < 3 {
			log.Get().WithField("clientName", cfg.ClientName).Error("length of config.clientName must >= 3")
			<-time.After(time.Second * 12)
			return
		}
		if len(cfg.ChannelName) < 3 {
			log.Get().WithField("channelName", cfg.ChannelName).Error("length of config.channelName must >= 6")
			<-time.After(time.Second * 12)
			return
		}
		for {
			runAsClient(&cfg)
			<-time.After(time.Second * 12)
		}
	case "server":
		runAsServer(&cfg)
	default:
		log.Get().WithField("mode", cfg.Mode).Error("unknown mode")
	}
}
