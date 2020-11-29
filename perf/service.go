package perf

type PerfService struct {
}

func NewService() *PerfService {
	return &PerfService{}
}

func (ps *PerfService) NewPerf(topic string) Perf {
	return &perf{}
}
