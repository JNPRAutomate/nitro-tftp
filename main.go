package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

//GitHash set by ld flags at build time
var GitHash = ""

//Version set by git tag
var Version = ""

//BuildDate set by ld flags at build time
var BuildDate = ""

//AppName Application name
var AppName = "Nitro TFTP"

var debugFlag bool
var configFile string
var configString string
var consulConfig string
var versionFlag bool
var genconfigFlag bool
var cpuprofile string
var memprofile string

func init() {
	cfgUsage := "Configuration file"
	flag.StringVar(&configFile, "config", "", cfgUsage)
	flag.StringVar(&configFile, "c", "", cfgUsage+" (shorthand)")
	cfgstrUsage := "Configuration string in JSON format"
	flag.StringVar(&configString, "config-string", "", cfgstrUsage)
	flag.StringVar(&configString, "cs", "", cfgstrUsage+" (shorthand)")
	debugUsage := "Enable debug logging"
	flag.BoolVar(&debugFlag, "debug", false, debugUsage)
	flag.BoolVar(&debugFlag, "d", false, debugUsage+" (shorthand)")
	verUsage := "Display Version"
	flag.BoolVar(&versionFlag, "version", false, verUsage)
	flag.BoolVar(&versionFlag, "v", false, verUsage+" (shorthand)")
	gencfgUsage := "Generate example config"
	flag.BoolVar(&genconfigFlag, "gencfg", false, gencfgUsage)
	flag.BoolVar(&genconfigFlag, "g", false, gencfgUsage+" (shorthand)")
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
		os.Exit(0)
	}

	if genconfigFlag {
		cfg := NewConfig()
		c, err := json.MarshalIndent(cfg, "", "\t")
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		fmt.Println(string(c))
		os.Exit(0)
	}

	s := &TFTPServer{Debug: debugFlag}

	cfg := NewConfig()

	if configFile != "" && configString == "" {
		err := cfg.Open(configFile)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	} else if configFile == "" && configString != "" {
		err := cfg.StringParse(configString)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	}

	s.LoadConfig(cfg)
	ctrlChan := s.Listen()

	//register as a consul service
	serviceid := strings.Join([]string{AppName, strconv.Itoa(int(time.Now().Unix() * rand.Int63()))}, "_")
	if consulConfig != "" {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		agent := client.Agent()
		host, port, err := net.SplitHostPort(s.sock.LocalAddr().String())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		iport, err := strconv.Atoi(port)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		service := &api.AgentServiceRegistration{ID: serviceid, Name: "tftp", Address: host, Port: iport}
		agent.ServiceRegister(service)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	<-sigChan
	close(sigChan)
	close(ctrlChan)

	if consulConfig != "" {
		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		agent := client.Agent()
		agent.ServiceDeregister(serviceid)
	}
	log.Println("Shutting down $s", AppName)
	os.Exit(0)
}
