package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
)

//Version info
var GitHash = ""
var Version = ""
var BuildDate = ""
var AppName = "Nitro TFTP"

var debugFlag bool
var configFile string
var versionFlag bool

func init() {
	flag.StringVar(&configFile, "config", "", "Configuration file")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	flag.BoolVar(&versionFlag, "version", false, "Display Version")
}

func main() {
	flag.Parse()

	if versionFlag {
		fmt.Printf("Built: %s\nVersion: %s\nGit Commit: %s\n", BuildDate, Version, GitHash)
		return
	}

	s := &TFTPServer{}
	s.Debug = debugFlag
	s.LoadConfig(&Config{})
	ctrlChan := s.Listen()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	close(c)
	close(ctrlChan)
	log.Println("Caught Signal", sig)
}
