package trigger

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/heraldgo/heraldd/util"
)

// Cron is a trigger which will be active according to the spec
type Cron struct {
	util.BaseLogger
	Spec        string
	WithSeconds bool
}

// Run the Cron trigger
func (tgr *Cron) Run(ctx context.Context, sendParam func(map[string]interface{})) {
	cronChan := make(chan struct{})

	c := cron.New()
	_, err := c.AddFunc(tgr.Spec, func() {
		select {
		case <-ctx.Done():
		case cronChan <- struct{}{}:
		}
	})
	if err != nil {
		tgr.Errorf(`Cron error "%s": %s`, tgr.Spec, err)
		return
	}

	c.Start()
	defer c.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cronChan:
			sendParam(map[string]interface{}{
				"cron": tgr.Spec,
				"time": time.Now().Format(time.RFC3339),
			})
		}
	}
}

func newTriggerCron(param map[string]interface{}) interface{} {
	spec, _ := util.GetStringParam(param, "cron")
	withSeconds, _ := util.GetBoolParam(param, "with_seconds")
	return &Cron{
		Spec:        spec,
		WithSeconds: withSeconds,
	}
}
