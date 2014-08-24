// counter implements a very simple thread safe Counter type.
package utils

import "sync"

type Counter struct {
	value int
	lock  sync.Mutex
}

// NewCounter creates a new Counter
func NewCounter() *Counter {
	c := Counter{}
	return &c
}

// Inc increments the counter by one and returns the value
func (c *Counter) Inc() int {
	c.lock.Lock()
	c.value++
	c.lock.Unlock()
	return c.value
}
