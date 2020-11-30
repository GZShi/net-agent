package perf

import "sync"

type list struct {
	records []interface{}
	mut     sync.RWMutex
}

func newList() *list {
	return &list{
		records: make([]interface{}, 10),
	}
}

func (l *list) Push(val ...interface{}) {
	l.mut.Lock()
	l.records = append(l.records, val...)
	l.mut.Unlock()
}
