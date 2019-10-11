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
		log.Errorf("Config file \"%s\" read error: %s\n", configFile, err)
		return nil, err
	}

	var cfg interface{}
	err = yaml.Unmarshal(buffer, &cfg)
	if err != nil {
		log.Errorf("Config file \"%s\" load error: %s\n", configFile, err)
		return nil, err
	}
	return cfg, nil
}

func loadParamAndType(name string, param interface{}) (string, map[string]interface{}, error) {
	paramMap, ok := param.(map[interface{}]interface{})
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
		key, ok := k.(string)
		if !ok {
			log.Warnf("Key \"%v\" is not a string\n", k)
			continue
		}
		if key != "type" {
			newParam[key] = v
		}
	}

	return typeName, newParam, nil
}

func createTrigger(h *herald.Herald, name, triggerType string, param map[string]interface{}) {
	tgr, err := trigger.CreateTrigger(triggerType)
	if err != nil {
		log.Errorf("Failed to created trigger for type \"%s\": %s\n", triggerType, err)
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

func loadTrigger(h *herald.Herald, cfg map[interface{}]interface{}) {
	for name, param := range cfg {
		nameString, ok := name.(string)
		if !ok {
			log.Warnf("Trigger name \"%v\" is not a string", name)
			continue
		}

		triggerType, paramMap, err := loadParamAndType(nameString, param)
		if err != nil {
			log.Warnf("Failed to get param for trigger \"%s\": %s\n", nameString, err)
			continue
		}

		createTrigger(h, nameString, triggerType, paramMap)
	}
}

func createExecutor(h *herald.Herald, name, executorType string, param map[string]interface{}) {
	exe, err := executor.CreateExecutor(executorType)
	if err != nil {
		log.Errorf("Failed to created executor for type \"%s\": %s\n", executorType, err)
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

func loadExecutor(h *herald.Herald, cfg map[interface{}]interface{}) {
	for name, param := range cfg {
		nameString, ok := name.(string)
		if !ok {
			log.Warnf("Executor name \"%v\" is not a string", name)
			continue
		}

		executorType, paramMap, err := loadParamAndType(nameString, param)
		if err != nil {
			log.Warnf("Failed to get param for executor \"%s\": %s\n", nameString, err)
			continue
		}

		createExecutor(h, nameString, executorType, paramMap)
	}
}

func createFilter(h *herald.Herald, name, filterType string, param map[string]interface{}) {
	flt, err := filter.CreateFilter(filterType)
	if err != nil {
		log.Errorf("Failed to created filter for type \"%s\": %s\n", filterType, err)
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

func loadFilter(h *herald.Herald, cfg map[interface{}]interface{}) {
	for name, param := range cfg {
		nameString, ok := name.(string)
		if !ok {
			log.Warnf("Filter name \"%v\" is not a string", name)
			continue
		}

		filterType, paramMap, err := loadParamAndType(nameString, param)
		if err != nil {
			log.Warnf("Failed to get param for filter \"%s\": %s\n", nameString, err)
			continue
		}

		createFilter(h, nameString, filterType, paramMap)
	}
}

func loadJob(h *herald.Herald, cfg map[interface{}]interface{}) {
	for name, param := range cfg {
		nameString, ok := name.(string)
		if !ok {
			log.Warnf("Job name \"%v\" is not a string", name)
			continue
		}

		paramMap, ok := param.(map[interface{}]interface{})
		if !ok {
			log.Warnf("Param is not a map for job: %s\n", nameString)
			continue
		}

		newParam := make(map[string]interface{})
		for k, v := range paramMap {
			key, ok := k.(string)
			if !ok {
				log.Warnf("Key \"%v\" is not a string\n", k)
				continue
			}
			newParam[key] = v
		}

		h.SetJobParam(nameString, newParam)
	}
}

func loadRouter(h *herald.Herald, cfg map[interface{}]interface{}) {
	for name, param := range cfg {
		nameString, ok := name.(string)
		if !ok {
			log.Warnf("Router name \"%v\" is not a string", name)
			continue
		}

		paramMap, ok := param.(map[interface{}]interface{})
		if !ok {
			log.Warnf("Param is not a map for job: %s\n", nameString)
			continue
		}

		// Load Trigger
		var triggersSlice []string
		triggers, ok := paramMap["trigger"]
		if ok {
			value, ok := triggers.(string)
			if ok {
				triggersSlice = []string{value}
			} else {
				triggersSlice, _ = triggers.([]string)
			}
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
				log.Warnf("Filter name \"%v\" is not a string", filter)
			}
		}
		_, ok = h.GetFilter(filterString)
		if !ok {
			createFilter(h, filterString, filterString, nil)
		}

		// Load routeParam
		newParam := make(map[string]interface{})
		for k, v := range paramMap {
			key, ok := k.(string)
			if !ok {
				log.Warnf("Key \"%v\" is not a string\n", k)
				continue
			}
			if key != "trigger" && key != "filter" && key != "job" {
				newParam[key] = v
			}
		}

		h.AddRouter(nameString, triggersSlice, filterString, newParam)

		// Load job
		job, ok := paramMap["job"]
		if !ok {
			continue
		}
		jobMap, ok := job.(map[interface{}]interface{})
		if !ok {
			log.Warnf("Job in router \"%s\" is not a map: \n", nameString)
			continue
		}

		// Load job Executors
		for jobName, executors := range jobMap {
			jobNameString, ok := jobName.(string)
			if !ok {
				log.Warnf("Job name \"%v\" is not a string", name)
				continue
			}

			var executorsSlice []string
			value, ok := executors.(string)
			if ok {
				executorsSlice = []string{value}
			} else {
				executorsSlice, _ = executors.([]string)
			}
			for _, exe := range executorsSlice {
				_, ok := h.GetExecutor(exe)
				if !ok {
					createExecutor(h, exe, exe, nil)
				}
			}

			h.AddRouterJob(nameString, jobNameString, executorsSlice)
		}
	}
}

func newHerald(cfg interface{}) *herald.Herald {
	h := herald.New()
	h.Log = log

	cfgMap, ok := cfg.(map[interface{}]interface{})
	if !ok {
		log.Println("Configuration is not a map")
		return h
	}

	if cfgTrigger, ok := cfgMap["trigger"]; ok {
		if cfgTriggerMap, ok := cfgTrigger.(map[interface{}]interface{}); ok {
			loadTrigger(h, cfgTriggerMap)
		}
	}

	if cfgExecutor, ok := cfgMap["executor"]; ok {
		if cfgExecutorMap, ok := cfgExecutor.(map[interface{}]interface{}); ok {
			loadExecutor(h, cfgExecutorMap)
		}
	}

	if cfgFilter, ok := cfgMap["filter"]; ok {
		if cfgFilterMap, ok := cfgFilter.(map[interface{}]interface{}); ok {
			loadFilter(h, cfgFilterMap)
		}
	}

	if cfgJob, ok := cfgMap["job"]; ok {
		if cfgJobMap, ok := cfgJob.(map[interface{}]interface{}); ok {
			loadJob(h, cfgJobMap)
		}
	}

	if cfgRouter, ok := cfgMap["router"]; ok {
		if cfgRouterMap, ok := cfgRouter.(map[interface{}]interface{}); ok {
			loadRouter(h, cfgRouterMap)
		}
	}

	return h
}

func main() {
	flagConfigFile := flag.String("config", "config.yml", "Configuration file path")
	flag.Parse()

	log = logrus.New()
	log.SetFormatter(&simpleFormatter{
		TimeFormat: "2006-01-02 15:04:05.000 -0700 MST",
	})

	cfg, err := loadConfigFile(*flagConfigFile)
	if err != nil {
		log.Fatalln(err)
	}

	h := newHerald(cfg)

	go h.Start()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown...")

	h.Stop()

	log.Println("Exiting...")
}
