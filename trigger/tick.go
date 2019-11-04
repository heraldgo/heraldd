package trigger

import (
	"context"
	"time"

	"github.com/heraldgo/heraldd/util"
)

// Tick is a trigger which will be active periodically
type Tick struct {
	Interval time.Duration
}

// Run the Tick trigger
func (tgr *Tick) Run(ctx context.Context, sendParam func(map[string]interface{})) {
	if tgr.Interval <= 0 {
		tgr.Interval = time.Second
	}

	ticker := time.NewTicker(tgr.Interval)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			counter++
			sendParam(map[string]interface{}{"counter": counter})
		}
	}
}

func newTriggerTick(param map[string]interface{}) interface{} {
	interval, _ := util.GetIntParam(param, "interval")
	return &Tick{
		Interval: time.Duration(interval) * time.Second,
	}
}
