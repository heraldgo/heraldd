package main

import (
	"errors"
	"fmt"
	"plugin"

	"github.com/heraldgo/herald"

	"github.com/heraldgo/heraldd/executor"
	"github.com/heraldgo/heraldd/filter"
	"github.com/heraldgo/heraldd/trigger"
	"github.com/heraldgo/heraldd/util"
)

var pluginComponents = [3]string{"trigger", "executor", "filter"}
var pluginFuncs = [3]string{"CreateTrigger", "CreateExecutor", "CreateFilter"}

type mapParam map[string]interface{}

type mapCreator map[string]func(string, map[string]interface{}) (interface{}, error)
type mapPlugin map[string]mapCreator

// LoggerSetter should set logger for the instance
type LoggerSetter interface {
	SetLogger(interface{})
}

// LoggerPrefixSetter should set logger for the instance
type LoggerPrefixSetter interface {
	SetLoggerPrefix(string)
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
		lgr.SetLogger(log)
	}

	lgrp, ok := ifc.(LoggerPrefixSetter)
	if ok {
		lgrp.SetLoggerPrefix(prefix)
	}
}

func loadCreator(plugins []string) []mapPlugin {
	creators := make([]mapPlugin, 0, len(plugins)+1)

	for _, p := range plugins {
		pln, err := plugin.Open(p)
		if err != nil {
			log.Errorf(`[Heraldd] Failed to open plugin "%s": %s`, p, err)
			continue
		}

		creator := make(mapPlugin)
		creator[p] = make(mapCreator)

		for i := range pluginComponents {
			m, err := pln.Lookup(pluginFuncs[i])
			if err != nil {
				log.Debugf(`[Heraldd] Function "%s" not found in plugin "%s"`, pluginFuncs[i], p)
				continue
			}
			f, ok := m.(func(string, map[string]interface{}) (interface{}, error))
			if !ok {
				log.Warnf(`[Heraldd] Invalid function "%s" in plugin "%s"`, pluginFuncs[i], p)
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
	creator["heraldd"]["filter"] = filter.CreateFilter
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
				log.Debugf(`[Heraldd] Component "%s" type "%s" not in plugin "%s"`, component, instanceType, p)
				continue
			}

			if !validateFunc(ifc) {
				log.Warnf(`[Heraldd] "%s" in plugin "%s" is not a "%s"`, instanceType, p, component)
				continue
			}

			return ifc
		}
	}

	log.Errorf(`[Heraldd] Failed to created "%s" with type "%s"`, component, instanceType)
	return nil
}

func createTrigger(h *herald.Herald, name, triggerType string, param map[string]interface{}, creators []mapPlugin) {
	tgrI := createInstance("trigger", triggerType, param, creators, func(ifc interface{}) bool {
		_, ok := ifc.(herald.Trigger)
		return ok
	})

	if tgrI == nil {
		return
	}

	tgr := tgrI.(herald.Trigger)

	loggerPrefix := fmt.Sprintf("[Trigger:%s(%s)]", triggerType, name)
	setLogger(tgr, loggerPrefix)

	h.AddTrigger(name, tgr)
}

func loadTrigger(h *herald.Herald, cfg map[string]interface{}, creators []mapPlugin) {
	for name, param := range cfg {
		triggerType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf(`[Heraldd] Failed to get param for trigger "%s": %s`, name, err)
			continue
		}

		createTrigger(h, name, triggerType, paramMap, creators)
	}
}

func createExecutor(h *herald.Herald, name, executorType string, param map[string]interface{}, creators []mapPlugin) {
	exeI := createInstance("executor", executorType, param, creators, func(ifc interface{}) bool {
		_, ok := ifc.(herald.Executor)
		return ok
	})

	if exeI == nil {
		return
	}

	exe := exeI.(herald.Executor)

	loggerPrefix := fmt.Sprintf("[Executor:%s(%s)]", executorType, name)
	setLogger(exe, loggerPrefix)

	h.AddExecutor(name, exe)
}

func loadExecutor(h *herald.Herald, cfg map[string]interface{}, creators []mapPlugin) {
	for name, param := range cfg {
		executorType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf(`[Heraldd] Failed to get param for executor "%s": %s`, name, err)
			continue
		}

		createExecutor(h, name, executorType, paramMap, creators)
	}
}

func createFilter(h *herald.Herald, name, filterType string, param map[string]interface{}, creators []mapPlugin) {
	fltI := createInstance("filter", filterType, param, creators, func(ifc interface{}) bool {
		_, ok := ifc.(herald.Filter)
		return ok
	})

	if fltI == nil {
		return
	}

	flt := fltI.(herald.Filter)

	loggerPrefix := fmt.Sprintf("[Filter:%s(%s)]", filterType, name)
	setLogger(flt, loggerPrefix)

	h.AddFilter(name, flt)
}

func loadFilter(h *herald.Herald, cfg map[string]interface{}, creators []mapPlugin) {
	for name, param := range cfg {
		filterType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf(`[Heraldd] Failed to get param for filter "%s": %s`, name, err)
			continue
		}

		createFilter(h, name, filterType, paramMap, creators)
	}
}

func loadJob(h *herald.Herald, cfg map[string]interface{}) {
	for name, param := range cfg {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			log.Warnf("[Heraldd] Param is not a map for job: %s", name)
			continue
		}

		h.SetJobParam(name, paramMap)
	}
}

func loadRouter(h *herald.Herald, cfg map[string]interface{}, creators []mapPlugin) {
	for routerName, param := range cfg {
		paramMap, ok := param.(map[string]interface{})
		if !ok {
			log.Warnf("[Heraldd] Param is not a map for job: %s", routerName)
			continue
		}

		// Load Trigger
		triggersSlice, err := util.GetStringSliceParam(paramMap, "trigger")
		if err != nil {
			log.Warnf("[Heraldd] Invalid trigger value: %s", err)
		}
		for _, tgr := range triggersSlice {
			_, ok := h.GetTrigger(tgr)
			if !ok {
				createTrigger(h, tgr, tgr, nil, creators)
			}
		}

		// Load Filter
		var filterString string
		filter, ok := paramMap["filter"]
		if ok {
			filterString, ok = filter.(string)
			if !ok {
				log.Warnf(`[Heraldd] Filter name "%v" is not a string`, filter)
			}
			_, ok = h.GetFilter(filterString)
			if !ok {
				createFilter(h, filterString, filterString, nil, creators)
			}
		}

		// Load routerParam
		newParam := make(map[string]interface{})
		for k, v := range paramMap {
			if k != "trigger" && k != "filter" && k != "job" {
				newParam[k] = v
			}
		}

		log.Debugf(`[Heraldd] Add router "%s", trigger(%v), filter(%s)`, routerName, triggersSlice, filterString)
		h.AddRouter(routerName, triggersSlice, filterString, newParam)

		// Load job
		job, ok := paramMap["job"]
		if !ok {
			continue
		}
		jobMap, ok := job.(map[string]interface{})
		if !ok {
			log.Warnf(`[Heraldd] Job in router "%s" is not a map`, routerName)
			continue
		}

		// Load job Executors
		for jobName := range jobMap {
			executorsSlice, err := util.GetStringSliceParam(jobMap, jobName)
			if err != nil {
				log.Warnf("[Heraldd] Invalid executor value: %s", err)
			}
			for _, exe := range executorsSlice {
				_, ok := h.GetExecutor(exe)
				if !ok {
					createExecutor(h, exe, exe, nil, creators)
				}
			}

			log.Debugf(`[Heraldd] Add job for router "%s", job(%s), executor(%v)`, routerName, jobName, executorsSlice)
			h.AddRouterJob(routerName, jobName, executorsSlice)
		}
	}
}

func newHerald(cfg map[string]interface{}) *herald.Herald {
	h := herald.New(log)

	plugins, _ := util.GetStringSliceParam(cfg, "plugin")
	creators := loadCreator(plugins)

	cfgTrigger, _ := util.GetMapParam(cfg, "trigger")
	loadTrigger(h, cfgTrigger, creators)

	cfgExecutor, _ := util.GetMapParam(cfg, "executor")
	loadExecutor(h, cfgExecutor, creators)

	cfgFilter, _ := util.GetMapParam(cfg, "filter")
	loadFilter(h, cfgFilter, creators)

	cfgJob, _ := util.GetMapParam(cfg, "job")
	loadJob(h, cfgJob)

	cfgRouter, _ := util.GetMapParam(cfg, "router")
	loadRouter(h, cfgRouter, creators)

	return h
}
