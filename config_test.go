package main

import (
	"testing"
	"time"
)

func TestOpenConfig(t *testing.T) {
	cfg := &Config{}
	cfg.Open("/home/rcameron/gopath/src/github.com/jnprautomate/nitrotftp/example_config.cfg")
}

func TestConfigLoad(t *testing.T) {
	cfg := &Config{}
	cfg.Open("/home/rcameron/gopath/src/github.com/jnprautomate/nitrotftp/example_config.cfg")
	s := &TFTPServer{}
	s.LoadConfig(cfg)
}

func TestConfigLoadAndOpen(t *testing.T) {
	cfg := &Config{}
	cfg.Open("/home/rcameron/gopath/src/github.com/jnprautomate/nitrotftp/example_config.cfg")
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
