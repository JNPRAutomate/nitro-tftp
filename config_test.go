package main

import (
	"encoding/json"
	"log"
	"testing"
	"time"
)

func TestOpenConfig(t *testing.T) {
	cfg := &Config{}
	cfg.Open("/home/rcameron/gopath/src/github.com/robwc/nitro-tftp/example_config.cfg")
}

func TestConfigLoad(t *testing.T) {
	cfg := &Config{}
	cfg.Open("/home/rcameron/gopath/src/github.com/robwc/nitro-tftp/example_config.cfg")
	s := &TFTPServer{}
	s.LoadConfig(cfg)
}

func TestConfigMarshal(t *testing.T) {
	cfg := &Config{}
	cfg.Open("/home/rcameron/gopath/src/github.com/robwc/nitro-tftp/example_config.cfg")
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Error(err)
	}
	log.Println(string(data))
}

func TestConfigLoadAndOpen(t *testing.T) {
	cfg := &Config{}
	cfg.Open("/home/rcameron/gopath/src/github.com/robwc/nitro-tftp/example_config.cfg")
	s := &TFTPServer{}
	s.LoadConfig(cfg)
	ctrlChan := s.Listen()
	timer := time.NewTimer(time.Second * 5)
	<-timer.C
	close(ctrlChan)
}

func TestConfigLoadDefaultAndOpen(t *testing.T) {
	s := &TFTPServer{}
	s.LoadConfig(&Config{})
	ctrlChan := s.Listen()
	timer := time.NewTimer(time.Second * 5)
	<-timer.C
	close(ctrlChan)
}
