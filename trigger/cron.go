package trigger

import (
	"context"
	"time"

	"github.com/robfig/cron"

	"github.com/heraldgo/heraldd/util"
)

// Cron is a trigger which will be active according to the spec
type Cron struct {
	Spec string
}

// Run the Cron trigger
func (tgr *Cron) Run(ctx context.Context, sendParam func(map[string]interface{})) {
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
			sendParam(map[string]interface{}{"time": time.Now().Format(time.RFC3339)})
		}
	}
}

func newTriggerCron(param map[string]interface{}) interface{} {
	spec, _ := util.GetStringParam(param, "cron")
	return &Cron{
		Spec: spec,
	}
}
