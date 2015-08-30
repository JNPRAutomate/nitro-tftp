package main

import (
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"

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
var cpuprofile string
var memprofile string

func init() {
	flag.StringVar(&configFile, "config", "", "Configuration file")
	flag.StringVar(&configFile, "c", "", "Configuration file")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	flag.BoolVar(&debugFlag, "d", false, "Enable debug logging")
	flag.BoolVar(&versionFlag, "version", false, "Display Version")
	flag.BoolVar(&versionFlag, "v", false, "Display Version")
	//flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	//flag.StringVar(&memprofile, "memprofile", "", "write memory profile to file")
}

func main() {
	flag.Parse()

	// add handlers to help us track memory usage - they don't track memory until they're told to
	//profiler.AddMemoryProfilingHandlers()

	// Uncomment if you want to start profiling automatically
	// profiler.StartProfiling()

	// Enable pprof
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		defer f.Close()
	}

	if debugFlag {
		log.SetLevel(log.DebugLevel)
	}

	// listen on port 6060 (pick a port)
	//go http.ListenAndServe(":6060", nil)

	if versionFlag {
		fmt.Printf("Built: %s\nVersion: %s\nGit Commit: %s\n", BuildDate, Version, GitHash)
		return
	}

	s := &TFTPServer{Debug: debugFlag}

	cfg := &Config{}

	if configFile != "" {
		err := cfg.Open(configFile)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}

	s.LoadConfig(cfg)
	ctrlChan := s.Listen()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	sig := <-sigChan
	close(sigChan)
	close(ctrlChan)
	log.Println("Caught Signal", sig)
}
