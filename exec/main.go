package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"time"

	log "github.com/GZShi/net-agent/logger"
)

type totpInfo struct {
	Account string `json:"account"`
	Secret  string `json:"secret"`
	URL     string `json:"url"`
}

type portProxyInfo struct {
	Listen      string `json:"listen"`
	TargetAddr  string `json:"targetAddr"`
	ChannelName string `json:"channelName"`
}

type config struct {
	Mode      string          `json:"mode"` // server/agent/visitor
	Addr      string          `json:"addr"` // ip:port
	Secret    string          `json:"secret"`
	PortProxy []portProxyInfo `json:"portProxy"`

	// client only
	ClientName  string `json:"clientName"`
	ChannelName string `json:"channelName"`

	// server only
	TotpList []totpInfo `json:"totpList"`
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
	defer func() {
		log.Get().Warn("program closed after 12s")
		<-time.After(time.Second * 12)
	}()

	log.Get().Info("https://github.com/GZShi/net-agent.git")

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

	configDir := path.Dir(configPath)
	switch cfg.Mode {
	case "agent", "ws-agent":
		if len(cfg.ClientName) < 3 {
			log.Get().WithField("clientName", cfg.ClientName).Error("length of config.clientName must >= 3")
			return
		}
		if len(cfg.ChannelName) < 3 {
			log.Get().WithField("channelName", cfg.ChannelName).Error("length of config.channelName must >= 6")
			return
		}
		// 重试间隔
		retryElapse := 12.0
		for {
			startTime := time.Now()
			runAsAgent(&cfg)
			runSeconds := time.Since(startTime).Seconds()
			if runSeconds < retryElapse {
				log.Get().Info(fmt.Sprintf("agent restart after %v seconds", retryElapse-runSeconds))
				<-time.After(time.Second * time.Duration(retryElapse-runSeconds))
				retryElapse *= 1.2
				if retryElapse > 60 {
					retryElapse = 60
				}
			} else {
				retryElapse = 12.0
			}
		}
	case "server":
		runAsServer(&cfg, configDir)
	case "visitor":
		runAsVisitor(&cfg)
	default:
		log.Get().WithField("mode", cfg.Mode).Error("unknown mode")
	}
}
