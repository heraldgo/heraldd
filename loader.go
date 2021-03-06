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
			newParam[k] = util.DeepCopyParam(v)
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

func loadParamWithPreset(cfg interface{}, cfgPreset map[string]interface{}) map[string]interface{} {
	param := make(map[string]interface{})

	var cfgMap map[string]interface{}
	var presetNames []string

	cfgMap, ok := cfg.(map[string]interface{})
	if ok {
		presetNames, _ = util.GetStringSliceParam(cfgMap, "preset")
	} else {
		presetName, ok := cfg.(string)
		if ok {
			presetNames = append(presetNames, presetName)
		} else {
			presetNameSlice, _ := cfg.([]interface{})
			for _, value := range presetNameSlice {
				valueString, ok := value.(string)
				if !ok {
					continue
				}
				presetNames = append(presetNames, valueString)
			}
		}
	}

	// Reverse iteration
	for i := len(presetNames) - 1; i >= 0; i-- {
		presetParam, err := util.GetMapParam(cfgPreset, presetNames[i])
		if err != nil {
			log.Warnf(`Preset "%s" not loaded: %s`, presetNames[i], err)
			continue
		}
		util.MergeMapParam(param, presetParam)
	}

	for k, v := range cfgMap {
		if k != "preset" {
			param[k] = util.DeepCopyParam(v)
		}
	}

	return param
}

func loadRouterTrigger(h *herald.Herald, paramMap map[string]interface{}, creators []mapPlugin) string {
	trigger, _ := util.GetStringParam(paramMap, "trigger")
	if trigger == "" {
		log.Errorf("Invalid trigger value in router")
		return ""
	}
	if h.GetTrigger(trigger) == nil {
		err := createTrigger(h, trigger, trigger, nil, creators)
		if err != nil {
			log.Errorf(`Auto create trigger "%s" failed`, trigger)
			return ""
		}
	}
	return trigger
}

func loadRouterSelector(h *herald.Herald, paramMap map[string]interface{}, creators []mapPlugin) string {
	selector, _ := util.GetStringParam(paramMap, "selector")
	if selector == "" {
		return ""
	}
	if h.GetSelector(selector) == nil {
		err := createSelector(h, selector, selector, nil, creators)
		if err != nil {
			log.Errorf(`Auto create selector "%s" failed`, selector)
			return ""
		}
	}
	return selector
}

func loadRouterExecutor(h *herald.Herald, cfgTask interface{}, cfgPreset map[string]interface{}, creators []mapPlugin) (string, map[string]interface{}, map[string]interface{}) {
	cfgTaskMap := make(map[string]interface{})

	executor, ok := cfgTask.(string)
	if !ok {
		cfgTaskMap, _ = cfgTask.(map[string]interface{})
		executor, _ = util.GetStringParam(cfgTaskMap, "executor")
	}
	if executor == "" {
		log.Errorf("Invalid executor for task")
		return "", nil, nil
	}

	if h.GetExecutor(executor) == nil {
		err := createExecutor(h, executor, executor, nil, creators)
		if err != nil {
			log.Errorf(`Auto create executor "%s" failed for task`, executor)
			return "", nil, nil
		}
	}

	cfgSelectParam := cfgTaskMap["select_param"]
	selectParam := loadParamWithPreset(cfgSelectParam, cfgPreset)

	cfgJobParam := cfgTaskMap["job_param"]
	jobParam := loadParamWithPreset(cfgJobParam, cfgPreset)

	return executor, selectParam, jobParam
}

func loadRouter(h *herald.Herald, cfg, cfgPreset map[string]interface{}, creators []mapPlugin) {
	for router, param := range cfg {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			log.Errorf("Param is not a map for router: %s", router)
			continue
		}

		trigger := loadRouterTrigger(h, paramMap, creators)
		if trigger == "" {
			log.Errorf(`Get trigger error in router "%s"`, router)
			continue
		}

		selector := loadRouterSelector(h, paramMap, creators)

		log.Debugf(`Register router "%s": trigger(%s), selector(%s)`, router, trigger, selector)
		err := h.RegisterRouter(router, trigger, selector)
		if err != nil {
			log.Errorf(`Register router error for router "%s": %s`, router, err)
			continue
		}

		// Load router param
		cfgRouterSelectParam, _ := util.GetMapParam(paramMap, "select_param")
		routerSelectParam := loadParamWithPreset(cfgRouterSelectParam, cfgPreset)

		cfgRouterJobParam, _ := util.GetMapParam(paramMap, "job_param")
		routerJobParam := loadParamWithPreset(cfgRouterJobParam, cfgPreset)

		// Load tasks in router
		tasks, err := util.GetMapParam(paramMap, "task")
		if err != nil {
			log.Errorf(`Get tasks error for router "%s"`, router)
			continue
		}

		for task, cfgTask := range tasks {
			executor, taskSelectParam, taskJobParam := loadRouterExecutor(h, cfgTask, cfgPreset, creators)
			if executor == "" {
				log.Errorf(`Get executor error for task "%s" in router "%s"`, task, router)
				continue
			}

			selectParam := make(map[string]interface{})
			util.MergeMapParam(selectParam, routerSelectParam)
			util.MergeMapParam(selectParam, taskSelectParam)

			jobParam := make(map[string]interface{})
			util.MergeMapParam(jobParam, routerJobParam)
			util.MergeMapParam(jobParam, taskJobParam)

			log.Debugf(`Add task for router "%s", task(%s), executor(%v)`, router, task, executor)
			err = h.AddRouterTask(router, task, executor, selectParam, jobParam)
			if err != nil {
				log.Errorf(`Add router task failed: %s`, err)
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

	cfgPreset, _ := util.GetMapParam(cfg, "preset")

	cfgRouter, _ := util.GetMapParam(cfg, "router")
	loadRouter(h, cfgRouter, cfgPreset, creators)

	return h
}
