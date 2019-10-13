package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
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

// ParamSetter should set param for the instance
type ParamSetter interface {
	SetParam(map[string]interface{})
}

// LoggerSetter should set logger for the instance
type LoggerSetter interface {
	SetLogger(herald.Logger)
}

func loadConfigFile(configFile string) (interface{}, error) {
	buffer, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Errorf("[Heraldd] Config file \"%s\" read error: %s", configFile, err)
		return nil, err
	}

	var cfg interface{}
	err = yaml.Unmarshal(buffer, &cfg)
	if err != nil {
		log.Errorf("[Heraldd] Config file \"%s\" load error: %s", configFile, err)
		return nil, err
	}
	return util.InterfaceToStringMap(cfg), nil
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

func createTrigger(h *herald.Herald, name, triggerType string, param map[string]interface{}) {
	tgr, err := trigger.CreateTrigger(triggerType)
	if err != nil {
		log.Errorf("[Heraldd] Failed to created trigger for type \"%s\": %s", triggerType, err)
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

func loadTrigger(h *herald.Herald, cfg map[string]interface{}) {
	for name, param := range cfg {
		triggerType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf("[Heraldd] Failed to get param for trigger \"%s\": %s", name, err)
			continue
		}

		createTrigger(h, name, triggerType, paramMap)
	}
}

func createExecutor(h *herald.Herald, name, executorType string, param map[string]interface{}) {
	exe, err := executor.CreateExecutor(executorType)
	if err != nil {
		log.Errorf("[Heraldd] Failed to created executor for type \"%s\": %s", executorType, err)
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

func loadExecutor(h *herald.Herald, cfg map[string]interface{}) {
	for name, param := range cfg {
		executorType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf("[Heraldd] Failed to get param for executor \"%s\": %s", name, err)
			continue
		}

		createExecutor(h, name, executorType, paramMap)
	}
}

func createFilter(h *herald.Herald, name, filterType string, param map[string]interface{}) {
	flt, err := filter.CreateFilter(filterType)
	if err != nil {
		log.Errorf("[Heraldd] Failed to created filter for type \"%s\": %s", filterType, err)
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

func loadFilter(h *herald.Herald, cfg map[string]interface{}) {
	for name, param := range cfg {
		filterType, paramMap, err := loadParamAndType(name, param)
		if err != nil {
			log.Warnf("[Heraldd] Failed to get param for filter \"%s\": %s", name, err)
			continue
		}

		createFilter(h, name, filterType, paramMap)
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

func loadRouter(h *herald.Herald, cfg map[string]interface{}) {
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
				createTrigger(h, tgr, tgr, nil)
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
				createFilter(h, filterString, filterString, nil)
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
					createExecutor(h, exe, exe, nil)
				}
			}

			log.Debugf("[Heraldd] Add router job: %s, %v, %s", routerName, jobName, executorsSlice)
			h.AddRouterJob(routerName, jobName, executorsSlice)
		}
	}
}

func newHerald(cfg interface{}) *herald.Herald {
	h := herald.New(log)

	cfgMap, ok := cfg.(map[string]interface{})
	if !ok {
		log.Errorln("[Heraldd] Configuration is not a map")
		return h
	}

	if cfgTrigger, ok := cfgMap["trigger"]; ok {
		if cfgTriggerMap, ok := cfgTrigger.(map[string]interface{}); ok {
			loadTrigger(h, cfgTriggerMap)
		}
	}

	if cfgExecutor, ok := cfgMap["executor"]; ok {
		if cfgExecutorMap, ok := cfgExecutor.(map[string]interface{}); ok {
			loadExecutor(h, cfgExecutorMap)
		}
	}

	if cfgFilter, ok := cfgMap["filter"]; ok {
		if cfgFilterMap, ok := cfgFilter.(map[string]interface{}); ok {
			loadFilter(h, cfgFilterMap)
		}
	}

	if cfgJob, ok := cfgMap["job"]; ok {
		if cfgJobMap, ok := cfgJob.(map[string]interface{}); ok {
			loadJob(h, cfgJobMap)
		}
	}

	if cfgRouter, ok := cfgMap["router"]; ok {
		if cfgRouterMap, ok := cfgRouter.(map[string]interface{}); ok {
			loadRouter(h, cfgRouterMap)
		}
	}

	return h
}

func main() {
	flagConfigFile := flag.String("config", "config.yml", "Configuration file path")
	flag.Parse()

	log = logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&simpleFormatter{
		TimeFormat: "2006-01-02 15:04:05.000 -0700 MST",
	})

	cfg, err := loadConfigFile(*flagConfigFile)
	if err != nil {
		log.Fatalf("[Heraldd] Load config file \"%s\" error: %s", *flagConfigFile, err)
	}

	h := newHerald(cfg)

	go h.Start()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Infoln("[Heraldd] Shutdown...")

	h.Stop()

	log.Infoln("[Heraldd] Exiting...")
}
