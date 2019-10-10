package trigger

import (
	"context"
	"time"
)

// Tick is a trigger which will be active periodically
type Tick struct {
	Interval time.Duration
	counter  int
}

// Run the Tick trigger
func (tgr *Tick) Run(ctx context.Context, param chan map[string]interface{}) {
	if tgr.Interval <= 0 {
		tgr.Interval = time.Second
	}

	ticker := time.NewTicker(tgr.Interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tgr.counter++
			param <- map[string]interface{}{"counter": tgr.counter}
		}
	}
}

// SetParam will set param from a map
func (tgr *Tick) SetParam(param map[string]interface{}) {
	intervalParam, ok := param["interval"]
	if ok {
		interval, ok := intervalParam.(int)
		if ok {
			tgr.Interval = time.Duration(interval) * time.Second
		}
	}
}
