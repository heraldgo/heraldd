package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"gopkg.in/yaml.v2"

	"github.com/sirupsen/logrus"

	"github.com/heraldgo/heraldd/util"
)

var log *logrus.Logger

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
		logDir := filepath.Dir(output)
		if logDir != "" {
			os.MkdirAll(logDir, 0755)
			if err != nil {
				log.Errorf(`Create log directory "%s" failed: %s`, logDir, err)
				return
			}
		}

		f, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf(`Create log file "%s" error: %s`, output, err)
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
		log.Errorf(`[Heraldd] Load config file "%s" error: %s`, *flagConfigFile, err)
		return
	}

	var logFile *os.File
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	setupLog(cfg, &logFile)

	log.Infoln(strings.Repeat("=", 80))
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
	log.Infoln(strings.Repeat("-", 80))
}
