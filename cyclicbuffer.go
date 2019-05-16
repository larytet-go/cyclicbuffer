package cyclicbuffer

import (
	"sync"
)

// Thread safe cyclic buffer. I use this one when I need a fast, RAM based log
// of events. The implementation is *not* lockless
type CyclicBuffer struct {
	data  []interface{}
	full  bool
	size  int
	index int
	mutex *sync.Mutex
}

func (cb *CyclicBuffer) Empty() bool {
	return !cb.NotEmpty()
}

func (cb *CyclicBuffer) NotEmpty() bool {
	return (cb.full || (cb.index > 0))
}

func New(size int) *CyclicBuffer {
	return &CyclicBuffer{
		mutex: &sync.Mutex{},
		data:  make([]interface{}, size),
		index: 0,
		full:  false,
		size:  size,
	}
}

func (cb *CyclicBuffer) Append(d interface{}) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	var index = cb.index
	cb.data[index] = d
	index++
	if index >= cb.size {
		index = 0
		cb.full = true
	}
	cb.index = index
}

type Iterator struct {
	index int
	count int
	cb    *CyclicBuffer
}

// I want to call the user supplied callback and thread safeyty
// in the Range() See, for example, sync.Map API
func CreateIterator(cb *CyclicBuffer) *Iterator {
	var it Iterator
	it.cb = cb
	if cb.full {
		it.index = cb.index
		it.count = cb.size
	} else {
		it.index = 0
		it.count = cb.index
	}
	return &it
}

func (it *Iterator) Value() interface{} {
	value := it.cb.data[it.index]
	it.index++
	if it.index >= it.cb.size {
		it.index = 0
	}
	it.count--
	return value
}

func (it *Iterator) Next() bool {
	return (it.count > 0)
}

func (cb *CyclicBuffer) Get() []interface{} {
	var index int
	var count int
	if cb.full {
		index = cb.index
		count = cb.size
	} else {
		index = 0
		count = cb.index
	}
	res := make([]interface{}, 0, count)
	for i := 0; i < count; i++ {
		d := cb.data[index]
		res = append(res, d)
		index++
		if index >= cb.size {
			index = 0
		}
	}
	return res
}
