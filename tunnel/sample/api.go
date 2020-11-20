package main

type dialReqeust struct {
	Network   string `json:"network"`
	Address   string `json:"address"`
	SessionID uint32 `json:"sid"`
}

type dialResponse struct {
	SessionID uint32 `json:"sid"`
}
