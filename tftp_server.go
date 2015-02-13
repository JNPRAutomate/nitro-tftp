package main

import (
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	//ServerNet default network to listen on
	ServerNet = "udp4"
	//DefaultPort default port to listen on
	DefaultPort = 69
	//DefaultIP default IP to listen on
	DefaultIP = "0.0.0.0"
)

//TFTPServer A server to listen for UDP messages
type TFTPServer struct {
	listenAddr  *net.UDPAddr
	sock        *net.UDPConn
	incomingDir string
	outgoingDir string
	protocol    string
}

func (s *TFTPServer) LoadConfig(c *Config) error {
	var err error
	if c == nil {
		//load default
		c = &Config{}
		c.IncomingDir = "./incoming"
		c.OutgoingDir = "./outgoing"
		c.IP = net.ParseIP("0.0.0.0")
		c.Port = 69
		c.Protocol = "udp4"
	}
	//listen addr
	s.listenAddr, err = net.ResolveUDPAddr(c.Protocol, strings.Join([]string{c.IP.String(), strconv.Itoa(c.Port)}, ":"))
	if err != nil {
		panic(err)
	}
	s.protocol = c.Protocol
	return nil
}

//Listen Listen for connections
func (s *TFTPServer) Listen() {
	var err error
	cMgr := &TFTPClientMgr{Connections: make(map[string]*TFTPConn)}
	//s.listenAddr = &net.UDPAddr{IP: net.ParseIP(DefaultIP), Port: DefaultPort}
	bb := make([]byte, 1024000)

	s.sock, err = net.ListenUDP(s.protocol, s.listenAddr)
	if err != nil {
		log.Println(err)
	}
	s.sock.SetReadBuffer(2048000)
	for {
		//handle each packet in a seperate go routine
		msgLen, _, _, addr, err := s.sock.ReadMsgUDP(bb, nil)
		log.Println(msgLen)
		if err != nil {
			switch err := err.(type) {
			case net.Error:
				if err.Timeout() {
					log.Println(err)
				} else if err.Temporary() {
					log.Println(err)
				}
			}
			return
		}

		msg := bb[:msgLen]
		//clear buffer by emptying slice but not reallocating memory
		bb = bb[:cap(bb)]
		log.Println(msg)

		//TODO pull both bytes of message
		if uint16(msg[1]) == OpcodeRead {
			pkt := &TFTPReadWritePkt{}
			pkt.Unpack(msg)
			cMgr.Start(addr, pkt)
		} else if uint16(msg[1]) == OpcodeWrite {
			pkt := &TFTPReadWritePkt{}
			pkt.Unpack(msg)
			cMgr.Start(addr, pkt)
		} else if uint16(msg[1]) == OpcodeErr {
			pkt := &TFTPErrPkt{}
			pkt.Unpack(msg)
		}

	}
}
