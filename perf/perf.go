package perf

import "sync"

// Perf 记录器
type Perf interface {
	Record(topic string, value interface{}) Perf
	Sum(key string, value UnitValue) Perf
	SetState(stateName string, state int) Perf
}

type perf struct {
	topicRecords sync.Map
	values       sync.Map
	states       sync.Map
}

func (p *perf) Record(topic string, value interface{}) Perf {
	val, _ := p.topicRecords.LoadOrStore(topic, newList())
	records := val.(*list)

	records.Push(value)

	return p
}

func (p *perf) Sum(key string, value UnitValue) Perf {
	val, _ := p.values.LoadOrStore(key, 0)

	uval := val.(UnitValue)
	uval.Add(value)

	return p
}

func (p *perf) SetState(stateName string, state int) Perf {
	p.states.Store(stateName, state)
	return p
}
