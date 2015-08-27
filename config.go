package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

//Config server configration file
type Config struct {
	IncomingDir string `json:"incomingdir"` //IncomingDir incoming directory
	OutgoingDir string `json:"outgoingdir"` //OutgoingDir outgoing directory, can be the same as incoming
	IP          net.IP `json:"listenip"`    //IP IP address to listen on
	Port        int    `json:"port"`        //Port port to listen on 69 is the default, requires root or administrator privledges
	Protocol    string `json:"protocol"`    //Protocol protocol to listen on can be udp,udp4,udp6
	Stats       bool   `json:"stats"`       //Stats determines if stats are to be collected or not
}

//NewConfig creates a new config struct and returns it
//If not loaded with Parse then the method returns the defaults
// Default configuration
// IncomingDir "./incoming"
// OutgoingDir "./outgong"
// IP "0.0.0.0"
// Port 69
// Protocol "udp4"
// Stats true
func NewConfig() *Config {
	return &Config{IncomingDir: "./incoming", OutgoingDir: "./outgoing", IP: net.ParseIP("0.0.0.0"), Port: 69, Protocol: "udp4", Stats: true}
}

//Open open a new config file
func (c *Config) Open(config string) {
	file, e := ioutil.ReadFile(config)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	newConfig := &Config{}
	if err := json.Unmarshal(file, &newConfig); err != nil {
		panic(err)
	}
	c.IncomingDir = newConfig.IncomingDir
	c.OutgoingDir = newConfig.OutgoingDir
	c.IP = newConfig.IP
	c.Port = newConfig.Port
	c.Protocol = newConfig.Protocol
}

//MarshalJSON Return
/*func (c *Config) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("{\"incomingdir\" : %s, \"outgoingdir\" : %s,\"listenip\":\"%s\",\"port\":%d,\"protocol\":\"%s\"}", c.IncomingDir, c.OutgoingDir, c.IP.String(), c.Port, c.Protocol)), nil
}*/

func (c *Config) UnmarshalJSON(data []byte) error {
	var tmpConfig struct {
		IncomingDir string
		OutgoingDir string
		IP          string `json:"listenip"`
		Port        int
		Protocol    string
	}
	err := json.Unmarshal(data, &tmpConfig)
	if err != nil {
		return err
	}
	c.IncomingDir = tmpConfig.IncomingDir
	c.OutgoingDir = tmpConfig.OutgoingDir
	c.IP = net.ParseIP(tmpConfig.IP)
	c.Port = tmpConfig.Port
	c.Protocol = tmpConfig.Protocol

	return nil
}
