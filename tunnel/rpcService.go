package tunnel

import (
	"errors"
	"fmt"
)

// Service 服务模块接口
type Service interface {
	Prefix() string
	Exec(ctx Context) error
}

func (t *tunnel) BindService(s Service) error {
	if s == nil {
		return errors.New("service is nil")
	}
	if t.serviceMap == nil {
		t.serviceMap = make(map[string]Service)
	}

	prefix := s.Prefix()
	_, found := t.serviceMap[prefix]
	if found {
		return fmt.Errorf("bind failed: service '%v' exists", prefix)
	}

	t.serviceMap[prefix] = s
	return nil
}
