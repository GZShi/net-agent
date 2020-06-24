package main

import (
	"encoding/json"
	"errors"
)

type blockInfo struct {
	ChannelName string   `json:"channel"`
	Block       []string `json:"block"`
	Allow       []string `json:"allow"`
}

var blockMaps = make(map[string]bool)
var allowMaps = make(map[string]bool)
var errTargetAddrBlocked = errors.New("Target address blocked")
var errTargetAddrBlockedByDefault = errors.New("Target address blocked by default")

func initBlockList() error {
	data, err := readJSON("./blocklist.json")
	if err != nil {
		return err
	}
	var blockInfos []blockInfo
	if err = json.Unmarshal(data, &blockInfos); err != nil {
		return err
	}

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
