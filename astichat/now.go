package astichat

import "time"

// Now represents a time that increments by itself
type Now struct {
	time.Time
}

// NewNow creates a new Now
func NewNow(t time.Time) *Now {
	return &Now{Time: t}
}

// Update updates now
func (n *Now) Update() {
	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			n.Time = n.Time.Add(time.Second)
		}
	}
}
