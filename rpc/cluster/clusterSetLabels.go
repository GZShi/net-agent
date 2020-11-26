package cluster

import (
	"errors"

	"github.com/GZShi/net-agent/tunnel"
)

type setLabelsRequest struct {
	Labels []string `json:"labels"`
}

type setLabelsResponse struct {
	FinnalLabels []string `json:"finalLabels"`
}

func (c *client) SetLabels(labels []string) (finnalLabels []string, err error) {
	return nil, errors.New("not implement")
}

func (s *service) SetLabels(ctx tunnel.Context) {
	// s.cluster.SetLabels()
	ctx.Error(errors.New("method not implement"))
}
