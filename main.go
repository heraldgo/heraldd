package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"plugin"
	"syscall"

	"gopkg.in/yaml.v2"

	"github.com/sirupsen/logrus"

	"github.com/xianghuzhao/herald"

	"github.com/xianghuzhao/heraldd/executor"
	"github.com/xianghuzhao/heraldd/filter"
	"github.com/xianghuzhao/heraldd/trigger"
	"github.com/xianghuzhao/heraldd/util"
)

var log *logrus.Logger

type mapParam map[string]interface{}

type mapCreater map[string]func(string) (interface{}, error)
type mapPlugin map[string]mapCreater

// ParamSetter should set param for the instance
type ParamSetter interface {
	SetParam(map[string]interface{})
}

// LoggerSetter should set logger for the instance
type LoggerSetter interface {
	SetLogger(interface{})
}

func loadConfigFile(configFile string) (map[string]interface{}, error) {
	buffer, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg interface{}
	err = yaml.Unmarshal(buffer, &cfg)
	if err != nil {
		return nil, err
	}
	cfg = util.InterfaceMapToStringMap(cfg)

	cfgMap, ok := cfg.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Configuration is not a map")
	}

	return cfgMap, nil
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
			return "", nil, errors.New("\"type\" is not a string")
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

func loadCreater(plugins []string) []mapPlugin {
	var creaters []mapPlugin

	for _, p := range plugins {
		pln, err := plugin.Open(p)
		if err != nil {
			log.Errorf("[Heraldd] Failed to open plugin \"%s\": %s", p, err)
			continue
		}

		creater := make(mapPlugin)
		creater[p] = make(mapCreater)

		m, err := pln.Lookup("CreateTrigger")
		if err == nil {
			f, ok := m.(func(string) (interface{}, error))
			if ok {
				creater[p]["trigger"] = f
			} else {
				log.Warnf("[Heraldd] Invalid function \"CreateTrigger\" in plugin \"%s\"", p)
			}
		} else {
			log.Debugf("[Heraldd] Function \"CreateTrigger\" not found in plugin \"%s\"", p)
		}

		m, err = pln.Lookup("CreateExecutor")
		if err == nil {
			f, ok := m.(func(string) (interface{}, error))
			if ok {
				creater[p]["executor"] = f
			} else {
				log.Warnf("[Heraldd] Invalid function \"CreateExecutor\" in plugin \"%s\"", p)
			}
		}

		m, err = pln.Lookup("CreateFilter")
		if err == nil {
			f, ok := m.(func(string) (interface{}, error))
			if ok {
				creater[p]["filter"] = f
			} else {
				log.Warnf("[Heraldd] Invalid function \"CreateFilter\" in plugin \"%s\"", p)
			}
		}

		creaters = append(creaters, creater)
	}

	creater := make(mapPlugin)
	creater["heraldd"] = make(mapCreater)
	creater["heraldd"]["trigger"] = trigger.CreateTrigger
	creater["heraldd"]["executor"] = executor.CreateExecutor
	creater["heraldd"]["filter"] = filter.CreateFilter
	creaters = append(creaters, creater)

	return creaters
}

func createTrigger(h *herald.Herald, name, triggerType string, param map[string]interface{}, creaters []mapPlugin) {
	var tgr herald.Trigger

	for _, pluginMap := range creaters {
		for p, createrMap := range pluginMap {
			createTriggerFunc, ok := createrMap["trigger"]
			if !ok {
				continue
			}

			tgrI, err := createTriggerFunc(triggerType)
			if err != nil {
				continue
			}

			tgrTemp, ok := tgrI.(herald.Trigger)
			if !ok {
				log.Warnf("[Heraldd] \"%s\" in plugin \"%s\" is not a trigger", name, p)
				continue
			}

			tgr = tgrTemp
		}
	}

	if tgr == nil {
		log.Errorf("[Heraldd] Failed to created trigger for type \"%s\"", triggerType)
		return
	}

	prm, ok := tgr.(ParamSetter)
	if ok {
		prm.SetParam(param)
	}

	lgr, ok := tgr.(LoggerSetter)
	if ok {
		lgr.SetLogger(log)
	}

	h.AddTrigger(name, tgr)
}

func loadTrigger(h *herald.Herald, cfg map[string]interface{}, creaters []mapPlugin) {
	for name, param := range cfg {
		triggerType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf("[Heraldd] Failed to get param for trigger \"%s\": %s", name, err)
			continue
		}

		createTrigger(h, name, triggerType, paramMap, creaters)
	}
}

func createExecutor(h *herald.Herald, name, executorType string, param map[string]interface{}, creaters []mapPlugin) {
	var exe herald.Executor

	for _, pluginMap := range creaters {
		for p, createrMap := range pluginMap {
			createExecutorFunc, ok := createrMap["executor"]
			if !ok {
				continue
			}

			exeI, err := createExecutorFunc(executorType)
			if err != nil {
				continue
			}

			exeTemp, ok := exeI.(herald.Executor)
			if !ok {
				log.Warnf("[Heraldd] \"%s\" in plugin \"%s\" is not a executor", name, p)
				continue
			}

			exe = exeTemp
		}
	}

	if exe == nil {
		log.Errorf("[Heraldd] Failed to created executor for type \"%s\"", executorType)
		return
	}

	prm, ok := exe.(ParamSetter)
	if ok {
		prm.SetParam(param)
	}

	lgr, ok := exe.(LoggerSetter)
	if ok {
		lgr.SetLogger(log)
	}

	h.AddExecutor(name, exe)
}

func loadExecutor(h *herald.Herald, cfg map[string]interface{}, creaters []mapPlugin) {
	for name, param := range cfg {
		executorType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf("[Heraldd] Failed to get param for executor \"%s\": %s", name, err)
			continue
		}

		createExecutor(h, name, executorType, paramMap, creaters)
	}
}

func createFilter(h *herald.Herald, name, filterType string, param map[string]interface{}, creaters []mapPlugin) {
	var flt herald.Filter

	for _, pluginMap := range creaters {
		for p, createrMap := range pluginMap {
			createFilterFunc, ok := createrMap["filter"]
			if !ok {
				continue
			}

			fltI, err := createFilterFunc(filterType)
			if err != nil {
				continue
			}

			fltTemp, ok := fltI.(herald.Filter)
			if !ok {
				log.Warnf("[Heraldd] \"%s\" in plugin \"%s\" is not a filter", name, p)
				continue
			}

			flt = fltTemp
		}
	}

	if flt == nil {
		log.Errorf("[Heraldd] Failed to created filter for type \"%s\"", filterType)
		return
	}

	prm, ok := flt.(ParamSetter)
	if ok {
		prm.SetParam(param)
	}

	lgr, ok := flt.(LoggerSetter)
	if ok {
		lgr.SetLogger(log)
	}

	h.AddFilter(name, flt)
}

func loadFilter(h *herald.Herald, cfg map[string]interface{}, creaters []mapPlugin) {
	for name, param := range cfg {
		filterType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf("[Heraldd] Failed to get param for filter \"%s\": %s", name, err)
			continue
		}

		createFilter(h, name, filterType, paramMap, creaters)
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

func loadRouter(h *herald.Herald, cfg map[string]interface{}, creaters []mapPlugin) {
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
				createTrigger(h, tgr, tgr, nil, creaters)
			}
		}

		// Load Filter
		var filterString string
		filter, ok := paramMap["filter"]
		if ok {
			filterString, ok = filter.(string)
			if !ok {
				log.Warnf("[Heraldd] Filter name \"%v\" is not a string", filter)
			}
			_, ok = h.GetFilter(filterString)
			if !ok {
				createFilter(h, filterString, filterString, nil, creaters)
			}
		}

		// Load routerParam
		newParam := make(map[string]interface{})
		for k, v := range paramMap {
			if k != "trigger" && k != "filter" && k != "job" {
				newParam[k] = v
			}
		}

		log.Debugf("[Heraldd] Add router: %s, %v, %s", routerName, triggersSlice, filterString)
		h.AddRouter(routerName, triggersSlice, filterString, newParam)

		// Load job
		job, ok := paramMap["job"]
		if !ok {
			continue
		}
		jobMap, ok := job.(map[string]interface{})
		if !ok {
			log.Warnf("[Heraldd] Job in router \"%s\" is not a map", routerName)
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
					createExecutor(h, exe, exe, nil, creaters)
				}
			}

			log.Debugf("[Heraldd] Add router job: %s, %v, %s", routerName, jobName, executorsSlice)
			h.AddRouterJob(routerName, jobName, executorsSlice)
		}
	}
}

func newHerald(cfg map[string]interface{}) *herald.Herald {
	h := herald.New(log)

	plugins, _ := util.GetStringSliceParam(cfg, "plugin")
	creaters := loadCreater(plugins)

	cfgTrigger, _ := util.GetMapParam(cfg, "trigger")
	loadTrigger(h, cfgTrigger, creaters)

	cfgExecutor, _ := util.GetMapParam(cfg, "executor")
	loadExecutor(h, cfgExecutor, creaters)

	cfgFilter, _ := util.GetMapParam(cfg, "filter")
	loadFilter(h, cfgFilter, creaters)

	cfgJob, _ := util.GetMapParam(cfg, "job")
	loadJob(h, cfgJob)

	cfgRouter, _ := util.GetMapParam(cfg, "router")
	loadRouter(h, cfgRouter, creaters)

	return h
}

func setupLog(cfg map[string]interface{}, logFile **os.File) {
	level := logrus.InfoLevel
	timeFormat := "2006-01-02 15:04:05.000 -0700 MST"
	var output string

	cfgLog, err := util.GetMapParam(cfg, "log")
	if err == nil {
		levelTemp, err := util.GetStringParam(cfgLog, "level")
		if err == nil {
			levelLogrusTemp, err := logrus.ParseLevel(levelTemp)
			if err == nil {
				level = levelLogrusTemp
			}
		}

		timeFormatTemp, err := util.GetStringParam(cfgLog, "time_format")
		if err == nil {
			timeFormat = timeFormatTemp
		}

		outputTemp, err := util.GetStringParam(cfgLog, "output")
		if err == nil {
			output = outputTemp
		}
	}

	log.SetLevel(level)
	log.SetFormatter(&util.SimpleFormatter{
		TimeFormat: timeFormat,
	})

	if output != "" {
		f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Create log file \"%s\" error: %s", output, err)
		} else {
			log.SetOutput(f)
			*logFile = f
		}
	}
}

func main() {
	flagConfigFile := flag.String("config", "config.yml", "Configuration file path")
	flag.Parse()

	log = logrus.New()

	cfg, err := loadConfigFile(*flagConfigFile)
	if err != nil {
		log.Errorf("[Heraldd] Load config file \"%s\" error: %s", *flagConfigFile, err)
		return
	}

	var logFile *os.File
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	setupLog(cfg, &logFile)

	log.Infoln("[Heraldd] Initialize...")

	h := newHerald(cfg)

	log.Infoln("[Heraldd] Start...")

	h.Start()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Infoln("[Heraldd] Shutdown...")

	h.Stop()

	log.Infoln("[Heraldd] Exit...")
}
