package main

import (
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
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
	ctrlChan    chan int
	listenAddr  *net.UDPAddr
	sock        *net.UDPConn
	incomingDir string
	outgoingDir string
	protocol    string
	wg          sync.WaitGroup
}

func (s *TFTPServer) LoadConfig(c *Config) error {
	var err error
	if c.IncomingDir == "" || c.OutgoingDir == "" {
		//load default
		c = &Config{}
		c.IncomingDir = "./incoming"
		c.OutgoingDir = "./outgoing"
		c.IP = net.ParseIP("0.0.0.0")
		c.Port = 69
		c.Protocol = "udp4"
	}
	//check inc dir exists and can be written to
	iStat, err := os.Stat(c.IncomingDir)
	if err != nil {
		//Directory does not exist, create dir
		err := os.Mkdir(c.IncomingDir, 0777)
		if err != nil {

		}
		iStat, err = os.Stat(c.IncomingDir)
		if err != nil {

		}
	}
	if !iStat.IsDir() {
		//is not a dir
	}
	//check for rw

	//check outgoing dir exists and can be written to
	oStat, err := os.Stat(c.OutgoingDir)
	if err != nil {
		//Directory does not exist, create dir
		err := os.Mkdir(c.OutgoingDir, 0777)
		if err != nil {

		}
		oStat, err = os.Stat(c.OutgoingDir)
		if err != nil {

		}
	}
	if !oStat.IsDir() {
		//is not a dir
	}
	//check for rw

	//listen addr
	s.listenAddr, err = net.ResolveUDPAddr(c.Protocol, strings.Join([]string{c.IP.String(), strconv.Itoa(c.Port)}, ":"))
	if err != nil {
		panic(err)
	}
	s.protocol = c.Protocol
	return nil
}

//Listen Listen for connections
func (s *TFTPServer) Listen() chan int {
	s.ctrlChan = make(chan int)
	var err error
	cMgr := &TFTPClientMgr{Connections: make(map[string]*TFTPConn)}
	//s.listenAddr = &net.UDPAddr{IP: net.ParseIP(DefaultIP), Port: DefaultPort}
	bb := make([]byte, 1024000)

	s.sock, err = net.ListenUDP(s.protocol, s.listenAddr)
	if err != nil {
		log.Println(err)
		return s.ctrlChan
	}
	s.sock.SetReadBuffer(2048000)

	go func(msg <-chan int) {
		s.wg.Add(1)
		for item := range msg {
			if item == -1 {
				break
			}
		}
		err := s.sock.Close()
		if err != nil {
			log.Println(err)
		}
		s.wg.Done()
	}(s.ctrlChan)

	go func() {
		s.wg.Add(1)
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
				s.wg.Done()
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
	}()
	return s.ctrlChan
}
