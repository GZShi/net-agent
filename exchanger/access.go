package exchanger

import "errors"

type accessInfo struct {
	code     string
	secret   string
	tunnelID uint32
	host     string
	port     uint16
}

type accessManager struct {
	infos map[string]*accessInfo
}

func (am *accessManager) SetAccessInfo(info *accessInfo) error {
	return errors.New("not implements")
}

func (am *accessManager) GetAccessInfo(code, challenge, hashstr string) (*accessInfo, error) {
	return nil, errors.New("not found")
}

func (am *accessManager) RemoveAccessInfo(callerID uint32, code string) error {
	return errors.New("not found")
}
