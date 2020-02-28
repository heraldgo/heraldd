package main

import (
	"errors"
	"fmt"
	"plugin"

	"github.com/heraldgo/herald"

	"github.com/heraldgo/heraldd/executor"
	"github.com/heraldgo/heraldd/selector"
	"github.com/heraldgo/heraldd/trigger"
	"github.com/heraldgo/heraldd/util"
)

var pluginComponents = [3]string{"trigger", "executor", "selector"}
var pluginFuncs = [3]string{"CreateTrigger", "CreateExecutor", "CreateSelector"}

type mapParam map[string]interface{}

type mapCreator map[string]func(string, map[string]interface{}) (interface{}, error)
type mapPlugin map[string]mapCreator

// LoggerSetter should set logger for the instance
type LoggerSetter interface {
	SetLogger(interface{})
}

func loadParamAndType(name string, param interface{}) (string, map[string]interface{}, error) {
	paramMap, ok := param.(map[string]interface{})
	if !ok {
		return "", nil, errors.New("Param is not a map")
	}

	var typeName string

	typeParam, ok := paramMap["type"]
	if !ok {
		typeName = name
	} else {
		typeName, ok = typeParam.(string)
		if !ok {
			return "", nil, errors.New(`"type" is not a string`)
		}
	}

	newParam := make(map[string]interface{})
	for k, v := range paramMap {
		if k != "type" {
			newParam[k] = v
		}
	}

	return typeName, newParam, nil
}

func setLogger(ifc interface{}, prefix string) {
	lgr, ok := ifc.(LoggerSetter)
	if ok {
		lgr.SetLogger(&util.PrefixLogger{
			Logger: logger,
			Prefix: prefix,
		})
	}
}

func loadCreator(plugins []string) []mapPlugin {
	creators := make([]mapPlugin, 0, len(plugins)+1)

	for _, p := range plugins {
		pln, err := plugin.Open(p)
		if err != nil {
			log.Errorf(`Failed to open plugin "%s": %s`, p, err)
			continue
		}

		creator := make(mapPlugin)
		creator[p] = make(mapCreator)

		for i := range pluginComponents {
			m, err := pln.Lookup(pluginFuncs[i])
			if err != nil {
				log.Debugf(`Function "%s" not found in plugin "%s"`, pluginFuncs[i], p)
				continue
			}
			f, ok := m.(func(string, map[string]interface{}) (interface{}, error))
			if !ok {
				log.Warnf(`Invalid function "%s" in plugin "%s"`, pluginFuncs[i], p)
				continue
			}
			creator[p][pluginComponents[i]] = f
		}

		creators = append(creators, creator)
	}

	creator := make(mapPlugin)
	creator["heraldd"] = make(mapCreator)
	creator["heraldd"]["trigger"] = trigger.CreateTrigger
	creator["heraldd"]["executor"] = executor.CreateExecutor
	creator["heraldd"]["selector"] = selector.CreateSelector
	creators = append(creators, creator)

	return creators
}

func createInstance(component, instanceType string, param map[string]interface{}, creators []mapPlugin, validateFunc func(interface{}) bool) interface{} {
	for _, pluginMap := range creators {
		for p, creatorMap := range pluginMap {
			createFunc, ok := creatorMap[component]
			if !ok {
				continue
			}

			ifc, err := createFunc(instanceType, param)
			if err != nil {
				log.Debugf(`Component "%s" type "%s" not in plugin "%s": %s`, component, instanceType, p, err)
				continue
			}

			if !validateFunc(ifc) {
				log.Warnf(`"%s" in plugin "%s" is not a "%s"`, instanceType, p, component)
				continue
			}

			return ifc
		}
	}

	log.Errorf(`Failed to created "%s" with type "%s"`, component, instanceType)
	return nil
}

func createTrigger(h *herald.Herald, name, triggerType string, param map[string]interface{}, creators []mapPlugin) error {
	tgrI := createInstance("trigger", triggerType, param, creators, func(ifc interface{}) bool {
		_, ok := ifc.(herald.Trigger)
		return ok
	})

	if tgrI == nil {
		return errors.New("Failed to create trigger")
	}

	tgr := tgrI.(herald.Trigger)

	loggerPrefix := fmt.Sprintf("[Trigger:%s(%s)]", triggerType, name)
	setLogger(tgr, loggerPrefix)

	err := h.RegisterTrigger(name, tgr)
	if err != nil {
		return err
	}

	return nil
}

func loadTrigger(h *herald.Herald, cfg map[string]interface{}, creators []mapPlugin) {
	for name, param := range cfg {
		triggerType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Errorf(`Failed to get param for trigger "%s": %s`, name, err)
			continue
		}

		createTrigger(h, name, triggerType, paramMap, creators)
	}
}

func createExecutor(h *herald.Herald, name, executorType string, param map[string]interface{}, creators []mapPlugin) error {
	exeI := createInstance("executor", executorType, param, creators, func(ifc interface{}) bool {
		_, ok := ifc.(herald.Executor)
		return ok
	})

	if exeI == nil {
		return errors.New("Failed to create executor")
	}

	exe := exeI.(herald.Executor)

	loggerPrefix := fmt.Sprintf("[Executor:%s(%s)]", executorType, name)
	setLogger(exe, loggerPrefix)

	err := h.RegisterExecutor(name, exe)
	if err != nil {
		return err
	}

	return nil
}

func loadExecutor(h *herald.Herald, cfg map[string]interface{}, creators []mapPlugin) {
	for name, param := range cfg {
		executorType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Errorf(`Failed to get param for executor "%s": %s`, name, err)
			continue
		}

		createExecutor(h, name, executorType, paramMap, creators)
	}
}

func createSelector(h *herald.Herald, name, selectorType string, param map[string]interface{}, creators []mapPlugin) error {
	sltI := createInstance("selector", selectorType, param, creators, func(ifc interface{}) bool {
		_, ok := ifc.(herald.Selector)
		return ok
	})

	if sltI == nil {
		return errors.New("Failed to create selector")
	}

	slt := sltI.(herald.Selector)

	loggerPrefix := fmt.Sprintf("[Selector:%s(%s)]", selectorType, name)
	setLogger(slt, loggerPrefix)

	err := h.RegisterSelector(name, slt)
	if err != nil {
		return err
	}

	return nil
}

func loadSelector(h *herald.Herald, cfg map[string]interface{}, creators []mapPlugin) {
	for name, param := range cfg {
		selectorType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Errorf(`Failed to get param for selector "%s": %s`, name, err)
			continue
		}

		createSelector(h, name, selectorType, paramMap, creators)
	}
}

func loadJob(h *herald.Herald, cfg map[string]interface{}) {
	for name, param := range cfg {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			log.Errorf("Param is not a map for job: %s", name)
			continue
		}

		err := h.SetJobParam(name, paramMap)
		if err != nil {
			log.Errorf(`Set job param error for job "%s": %s`, name, err)
		}
	}
}

func loadRouter(h *herald.Herald, cfg map[string]interface{}, creators []mapPlugin) {
	for router, param := range cfg {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			log.Errorf("Param is not a map for job: %s", router)
			continue
		}

		// Load Trigger
		trigger, _ := util.GetStringParam(paramMap, "trigger")
		if trigger == "" {
			log.Errorf(`Invalid trigger value in router "%s"`, router)
			continue
		}
		if h.GetTrigger(trigger) == nil {
			err := createTrigger(h, trigger, trigger, nil, creators)
			if err != nil {
				log.Errorf(`Auto create trigger "%s" failed for router "%s"`, trigger, router)
				continue
			}
		}

		// Load Selector
		selector, err := util.GetStringParam(paramMap, "selector")
		if selector != "" {
			if h.GetSelector(selector) == nil {
				err := createSelector(h, selector, selector, nil, creators)
				if err != nil {
					log.Errorf(`Auto create selector "%s" failed for router "%s"`, selector, router)
					continue
				}
			}
		}

		// Load routerParam
		newParam := make(map[string]interface{})
		for k, v := range paramMap {
			if k != "trigger" && k != "selector" && k != "job" {
				newParam[k] = v
			}
		}

		log.Debugf(`Register router "%s": trigger(%s), selector(%s)`, router, trigger, selector)
		err = h.RegisterRouter(router, trigger, selector, newParam)
		if err != nil {
			log.Errorf(`Register router error for router "%s": %s`, router, err)
			continue
		}

		// Load jobs in router
		jobs, err := util.GetMapParam(paramMap, "job")
		if err != nil {
			log.Errorf(`Get jobs error for router "%s"`, router)
			continue
		}

		// Load job Executors
		for job := range jobs {
			executor, err := util.GetStringParam(jobs, job)
			if err != nil {
				log.Errorf(`Invalid executor value for job "%s" in router "%s": %s`, job, router, err)
				continue
			}

			if h.GetExecutor(executor) == nil {
				err := createExecutor(h, executor, executor, nil, creators)
				if err != nil {
					log.Errorf(`Auto create executor "%s" failed for job "%s" in router "%s"`, executor, job, router)
					continue
				}
			}

			log.Debugf(`Add job for router "%s", job(%s), executor(%v)`, router, job, executor)
			err = h.AddRouterJob(router, job, executor)
			if err != nil {
				log.Errorf(`Add router job failed: %s`, err)
				continue
			}
		}
	}
}

func newHerald(cfg map[string]interface{}) *herald.Herald {
	h := herald.New(logger)

	plugins, _ := util.GetStringSliceParam(cfg, "plugin")
	creators := loadCreator(plugins)

	cfgTrigger, _ := util.GetMapParam(cfg, "trigger")
	loadTrigger(h, cfgTrigger, creators)

	cfgExecutor, _ := util.GetMapParam(cfg, "executor")
	loadExecutor(h, cfgExecutor, creators)

	cfgSelector, _ := util.GetMapParam(cfg, "selector")
	loadSelector(h, cfgSelector, creators)

	cfgJob, _ := util.GetMapParam(cfg, "job")
	loadJob(h, cfgJob)

	cfgRouter, _ := util.GetMapParam(cfg, "router")
	loadRouter(h, cfgRouter, creators)

	return h
}
