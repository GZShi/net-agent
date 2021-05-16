package common

import (
	"io"
	"net/http"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
	"github.com/sirupsen/logrus"
)

const DefaultChatServerPort = uint32(20210516)

// RunChatServer 文件服务
func RunChatServer(t tunnel.Tunnel, cls def.Cluster, param map[string]string, log *logrus.Entry) (io.Closer, error) {

	l, err := t.Listen(DefaultChatServerPort)
	if err != nil {
		return nil, err
	}

	http.HandleFunc("/say-hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("who you are~ ?"))
	})

	httpsvc := &http.Server{}

	err = httpsvc.Serve(l)
	if err != nil {
		return nil, err
	}

	return httpsvc, nil
}
