package main

import "testing"

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
	s.Listen()
	//stop after 5 seconds
}

func TestConfigLoadDefaultAndOpen(t *testing.T) {
	s := &TFTPServer{}
	s.LoadConfig(nil)
	s.Listen()
}
