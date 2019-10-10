package trigger

import (
	"context"
	"time"

	"github.com/robfig/cron"
)

// Cron is a trigger which will be active according to the spec
type Cron struct {
	Spec string
}

// Run the Cron trigger
func (tgr *Cron) Run(ctx context.Context, param chan map[string]interface{}) {
	cronChan := make(chan struct{})

	c := cron.New()
	c.AddFunc(tgr.Spec, func() {
		cronChan <- struct{}{}
	})
	c.Start()
	defer c.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cronChan:
			param <- map[string]interface{}{"time": time.Now().Format(time.RFC3339)}
		}
	}
}

// SetParam will set param from a map
func (tgr *Cron) SetParam(param map[string]interface{}) {
	specParam, ok := param["cron"]
	if ok {
		spec, ok := specParam.(string)
		if ok {
			tgr.Spec = spec
		}
	}
}
