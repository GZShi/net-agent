package common

import (
	"errors"
	"net"
	"strconv"
	"strings"

	"github.com/GZShi/net-agent/rpc/cluster/def"
	"github.com/GZShi/net-agent/tunnel"
)

func splitHostPort(addr string) (host, port string, isTunnelAddr bool, err error) {
	host, port, err = net.SplitHostPort(addr)
	if err != nil {
		return "", "", false, err
	}

	isTunnelAddr = host == "tunnel" || strings.HasSuffix(host, ".tunnel")

	return
}

// Listen 监听地址，提供服务
func Listen(t tunnel.Tunnel, network, address string) (net.Listener, error) {
	_, port, isTunneAddr, err := splitHostPort(address)
	if err != nil {
		return nil, err
	}
	if t != nil && isTunneAddr {
		vport, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
		return t.Listen(uint32(vport))
	}
	return net.Listen(network, address)
}

// Dial 连接地址，使用服务
func Dial(t tunnel.Tunnel, cls def.Cluster, network, address string) (net.Conn, error) {
	host, port, isTunneAddr, err := splitHostPort(address)
	if err != nil {
		return nil, err
	}
	if isTunneAddr {
		vport, err := strconv.Atoi(port)
		if err != nil {
			return nil, err
		}
		if host == "tunnel" {
			if t == nil {
				return nil, errors.New("tunnel is nil")
			}
			return t.Dial(uint32(vport))
		}
		if cls == nil {
			return nil, errors.New("cluster is nil")
		}
		return cls.Dial(host, uint32(vport))
	}
	return net.Dial(network, address)
}
