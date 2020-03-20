# Net Agent

## Framework

## Install and Run
```bash
go get github.com/GZShi/net-agent/exec
cd GOPATH/src/github.com/GZShi/net-agent/exec
go build -o ../dist/netagent
cd ../dist
./netagent -config "./config.json"
```

### config example
```json
// client mode
{
  "mode": "client",
  "addr": "localhost:1080",
  "privateKey": "secretsecret",
  "clientName": "localuser",
  "channelName": "remotework"
}

// server mode
{
  "mode": "server",
  "addr": "0.0.0.0:1080",
  "privateKey": "secretsecret"
}
```
