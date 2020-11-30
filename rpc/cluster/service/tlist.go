package service

import (
	"errors"
	"sync"

	"github.com/GZShi/net-agent/rpc/cluster/def"
)

// InvalidTID 错误的tid
const InvalidTID = def.TID(0)

type tlist struct {
	container []def.TID
	length    int

	mut         sync.RWMutex
	selectIndex int
}

func (l *tlist) Append(tid def.TID) {
	if tid == InvalidTID {
		return
	}
	l.mut.Lock()
	defer l.mut.Unlock()

	if len(l.container) > l.length {
		l.container[l.length] = tid
	} else {
		l.container = append(l.container, tid)
	}

	l.length++
}

func (l *tlist) Remove(tid def.TID) {
	if tid == InvalidTID {
		return
	}
	l.mut.Lock()
	defer l.mut.Unlock()

	for i := 0; i < l.length; i++ {
		if l.container[i] == tid {
			l.container[i] = l.container[l.length-1]
			l.container[l.length-1] = InvalidTID
			l.length--
		}
	}
}

func (l *tlist) Select() (def.TID, error) {
	l.mut.RLock()
	defer l.mut.RUnlock()

	if l.length <= 0 {
		return InvalidTID, errors.New("list is empty")
	}

	if l.selectIndex > l.length {
		l.selectIndex = 0
	}

	tid := l.container[l.selectIndex]
	l.selectIndex++

	return tid, nil
}
