package perf

import (
	"fmt"
)

// UnitValue 有单位的数据对象
type UnitValue interface {
	Value() int64
	String() string

	Add(val UnitValue)
	Sub(val UnitValue)
}

// MemoryUnit 内存单位
type MemoryUnit int64

const (
	// Byte ...
	Byte MemoryUnit = 1
	// KB ...
	KB = 1024 * Byte
	// MB ...
	MB = 1024 * KB
	// GB ...
	GB = 1024 * MB
	// TB ...
	TB = 1024 * GB
	// PB ...
	PB = 1024 * TB
)

type unitValue struct {
	value MemoryUnit
}

func (uv *unitValue) Value() int64 {
	return int64(uv.value)
}

func (uv *unitValue) String() string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	ceilVal := []float64{float64(KB), float64(MB), float64(GB), float64(TB), float64(PB)}
	val := float64(uv.value)

	var index int
	for index = 0; index < len(ceilVal); index++ {
		if val < ceilVal[index] {
			break
		}
		val = val / ceilVal[index]
	}
	return fmt.Sprintf("%v%v", val, units[index])
}
