package tunnel

import (
	"errors"
	"fmt"
)

// Service 服务模块接口
type Service interface {
	Prefix() string
	Exec(ctx Context) error
	Hello(t Tunnel) error
}

func (t *tunnel) bindService(s Service) error {
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
	err := s.Hello(t)
	if err != nil {
		delete(t.serviceMap, prefix)
		return err
	}
	return nil
}

func (t *tunnel) BindServices(s ...Service) error {
	for _, service := range s {
		if err := t.bindService(service); err != nil {
			return err
		}
	}
	return nil
}
