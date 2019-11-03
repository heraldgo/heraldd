package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/heraldgo/heraldd/util"
)

var log *logrus.Logger

var cfg struct {
	LogLevel      string `yaml:"log_level"`
	LogTimeFormat string `yaml:"time_format"`
	LogOutput     string `yaml:"log_output"`
	WorkDir       string `yaml:"work_dir"`
	Secret        string `yaml:"secret"`
	UnixSocket    string `yaml:"unix_socket"`
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
}

func loadConfigFile(configFile string) error {
	buffer, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buffer, &cfg)
	if err != nil {
		return err
	}

	return nil
}

func setupLog(logFile **os.File) {
	level := logrus.InfoLevel
	timeFormat := "2006-01-02 15:04:05.000 -0700 MST"

	levelLogrusTemp, err := logrus.ParseLevel(cfg.LogLevel)
	if err == nil {
		level = levelLogrusTemp
	}

	if cfg.LogTimeFormat != "" {
		timeFormat = cfg.LogTimeFormat
	}

	log.SetLevel(level)
	log.SetFormatter(&util.SimpleFormatter{
		TimeFormat: timeFormat,
	})

	if cfg.LogOutput != "" {
		f, err := os.OpenFile(cfg.LogOutput, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf(`[HeraldExeServer] Create log file "%s" error: %s`, cfg.LogOutput, err)
		} else {
			log.SetOutput(f)
			*logFile = f
		}
	}
}

func newExeServer() *exeServer {
	s := &exeServer{}

	s.UnixSocket = cfg.UnixSocket
	s.Host = cfg.Host
	s.Port = cfg.Port
	s.secret = cfg.Secret
	s.exeGit.WorkDir = cfg.WorkDir

	s.SetLogger(log)
	s.SetLoggerPrefix("[HeraldExeServer]")

	return s
}

func main() {
	flagConfigFile := flag.String("config", "config.yml", "Configuration file path")
	flag.Parse()

	log = logrus.New()

	err := loadConfigFile(*flagConfigFile)
	if err != nil {
		log.Errorf(`[HeraldExeServer] Load config file "%s" error: %s`, *flagConfigFile, err)
		return
	}

	var logFile *os.File
	defer func() {
		if logFile != nil {
			logFile.Close()
		}
	}()

	setupLog(&logFile)

	log.Infoln("[HeraldExeServer] Initialize...")

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	s := newExeServer()

	log.Infoln("[HeraldExeServer] Start...")

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Run(ctx)
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Infoln("[HeraldExeServer] Shutdown...")

	cancel()

	wg.Wait()

	log.Infoln("[HeraldExeServer] Exiting...")
}
