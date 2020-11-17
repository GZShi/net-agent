package main

import (
	"encoding/json"
	"errors"
	"sync"

	log "github.com/GZShi/net-agent/logger"
	"github.com/fsnotify/fsnotify"
)

type blockInfo struct {
	ChannelName string   `json:"channel"`
	Block       []string `json:"block"`
	Allow       []string `json:"allow"`
}

var blockUpdateLock sync.RWMutex
var blockMaps = make(map[string]bool)
var allowMaps = make(map[string]bool)
var errTargetAddrBlocked = errors.New("Target address blocked")
var errTargetAddrBlockedByDefault = errors.New("Target address blocked by default")

func watchBlockList(path string) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	err = w.Add(path)
	if err != nil {
		return err
	}

	log.Get().WithField("path", path).Info("watching file")
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Get().WithField("event.Name", event.Name).Debug("modified file")
				err = initBlockList(path)
				if err != nil {
					log.Get().WithError(err).Error("update blocklist failed")
				} else {
					log.Get().Info("update blocklist success")
				}
			}
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			log.Get().WithError(err).Error("watch file error")
		}
	}
}

func initBlockList(path string) error {
	data, err := readJSON(path)
	if err != nil {
		return err
	}
	var blockInfos []blockInfo
	if err = json.Unmarshal(data, &blockInfos); err != nil {
		return err
	}

	blockUpdateLock.Lock()
	defer blockUpdateLock.Unlock()
	blockMaps = make(map[string]bool)
	allowMaps = make(map[string]bool)
	for _, info := range blockInfos {
		channelName := info.ChannelName
		for _, item := range info.Block {
			blockMaps[item+"@"+channelName] = true
		}
		for _, item := range info.Allow {
			allowMaps[item+"@"+channelName] = true
		}
	}

	return nil
}

func checkBlockList(network, targetAddr, channelName string) error {
	blockUpdateLock.RLock()
	defer blockUpdateLock.RUnlock()
	key := targetAddr + "@" + channelName
	if _, blocked := blockMaps[key]; blocked {
		return errTargetAddrBlocked
	}
	if _, allowAll := allowMaps["*@"+channelName]; allowAll {
		return nil
	}
	if _, allowTarget := allowMaps[key]; allowTarget {
		return nil
	}
	return errTargetAddrBlockedByDefault
}
