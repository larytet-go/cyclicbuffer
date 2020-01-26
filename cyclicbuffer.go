package cyclicbuffer

import (
	"sync"
)

// CyclicBuffer is a thread safe cyclic buffer.
// I use this buffer one when I need a fast, RAM based log  of events.
// The implementation is *not* lockless
type CyclicBuffer struct {
	data  []interface{}
	full  bool
	size  int
	index int
	mutex *sync.Mutex
}

// Empty returns true is the buffer is empty
func (cb *CyclicBuffer) Empty() bool {
	return !cb.NotEmpty()
}

// NotEmpty returns true is the buffer is not empty
func (cb *CyclicBuffer) NotEmpty() bool {
	return (cb.full || (cb.index > 0))
}

// New creates a buffer
func New(size int) *CyclicBuffer {
	return &CyclicBuffer{
		mutex: &sync.Mutex{},
		data:  make([]interface{}, size),
		index: 0,
		full:  false,
		size:  size,
	}
}

// Append adds an item to the cyclic buffer
// Returns position of the next entry
func (cb *CyclicBuffer) Append(d interface{}) int {
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
	return index
}

// AppendSafe is a thread safe API
func (cb *CyclicBuffer) AppendSafe(d interface{}) int {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	return cb.Append(d)
}

// Iterator object supporting loops
type Iterator struct {
	index int
	count int
	cb    *CyclicBuffer
}

// CreateIterator returns a new iterator
// I want to call the user supplied callback and thread safety
// in the Range() See, for example, sync.Map API
func (cb *CyclicBuffer) CreateIterator() *Iterator {
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

// CreateIterator is backward compatible API
func CreateIterator(cb *CyclicBuffer) *Iterator {
	return cb.CreateIterator()
}

// Value returns item from the iterator
func (it *Iterator) Value() interface{} {
	value := it.cb.data[it.index]
	it.index++
	if it.index >= it.cb.size {
		it.index = 0
	}
	it.count--
	return value
}

// Next returns true if there anything else
func (it *Iterator) Next() bool {
	return (it.count > 0)
}

// Get returns a copy of the stored data
// This is not a deep copy
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

// GetData returns all items in the buffer
func (cb *CyclicBuffer) GetData() []interface{} {
	return cb.data
}
