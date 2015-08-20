package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"

	log "github.com/Sirupsen/logrus"
	"github.com/wblakecaldwell/profiler"
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
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	flag.BoolVar(&versionFlag, "version", false, "Display Version")
	flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&memprofile, "memprofile", "", "write memory profile to file")
}

func main() {
	flag.Parse()

	// add handlers to help us track memory usage - they don't track memory until they're told to
	profiler.AddMemoryProfilingHandlers()

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

	// listen on port 6060 (pick a port)
	go http.ListenAndServe(":6060", nil)

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
