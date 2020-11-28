package dial

type Dial interface {
	Dial(dialSessionID uint32, network, address string) (connSessionID uint32, err error)
}
