package main

import "net"

//Config server configration file
type Config struct {
	IncomingDir string //IncomingDir incoming directory
	OutgoingDir string //OutgoingDir outgoing directory, can be the same as incoming
	IP          net.IP //IP IP address to listen on
	Port        int    //Port port to listen on 69 is the default, requires root or administrator privledges
	Protocol    string //Protocol protocol to listen on can be udp,udp4,udp6
}

//NewConfig creates a new config struct and returns it
//If not loaded with Parse then the method returns the defaults
// Default configuration
// IncomingDir "./incoming"
// OutgoingDir "./outgong"
// IP "0.0.0.0"
// Port 69
// Protocol "udp4"
func NewConfig() *Config {
	return &Config{IncomingDir: "./incoming", OutgoingDir: "./outgoing", IP: net.ParseIP("0.0.0.0"), Port: 69, Protocol: "udp4"}
}

//Parse parse config file
func (c *Config) Parse() {

}
