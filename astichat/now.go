package astichat

import (
	"sync"
	"time"
)

// Now represents a time that increments by itself
type Now struct {
	mutex *sync.Mutex
	time  time.Time
}

// NewNow creates a new Now
func NewNow(t time.Time) (n *Now) {
	n = &Now{mutex: &sync.Mutex{}, time: t}
	go n.update()
	return
}

// update updates now
func (n *Now) update() {
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			n.mutex.Lock()
			n.time = n.time.Add(time.Second)
			n.mutex.Unlock()
		}
	}
}

// Time returns the time
func (n *Now) Time() time.Time {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.time
}
