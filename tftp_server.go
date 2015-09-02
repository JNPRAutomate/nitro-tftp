package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/robwc/tftp"
)

const (
	//ServerNet default network to listen on
	ServerNet = "udp4"
	//DefaultPort default port to listen on
	DefaultPort = 69
	//DefaultIP default IP to listen on
	DefaultIP = "0.0.0.0"
	//DefaultStats enable stats collection by default
	DefaultStats = true
)

//TFTPServer A server to listen for UDP messages
type TFTPServer struct {
	ctrlChan    chan int             //ctrlChan control channel for managing the server
	listenAddr  *net.UDPAddr         //listenAddr the address to listen on
	sock        *net.UDPConn         //sock UDP connection socket
	incomingDir string               //incomingDir incoming directory for files
	outgoingDir string               //outgoingDir outgoingDir for files
	protocol    string               //protocol protocol to listen on: udp, udp4, or udp6
	wg          sync.WaitGroup       //wg wait group for syncing data
	Connections map[string]*TFTPConn //Connections active TFTP connection
	clientwg    sync.WaitGroup       //clientwg wait group to manage client connection
	Debug       bool                 //Debug enable debuging
	StatsMgr    *StatsMgr            //StatsMgr stats collection manager
	Config      *Config              //Config complete config passed to server
}

//LoadConfig load a config from disk
func (s *TFTPServer) LoadConfig(c *Config) error {
	var err error
	if c.IncomingDir == "" && c.OutgoingDir == "" {
		//load default
		c = NewConfig()
	} else {
		s.Config = c
		s.incomingDir = c.IncomingDir
		s.outgoingDir = c.OutgoingDir
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
		return err
	}
	s.protocol = c.Protocol
	return nil
}

//Listen Listen for connections
func (s *TFTPServer) Listen() chan int {
	s.Connections = make(map[string]*TFTPConn)
	s.ctrlChan = make(chan int)
	var err error
	bb := make([]byte, 1024000)

	//Listen on stats API
	if s.Config.Stats {
		s.StatsMgr = NewStatsMgr()
		s.StatsMgr.StatsListener(s.Config.StatsIP, s.Config.StatsPort, s.ctrlChan)
	}

	s.sock, err = net.ListenUDP(s.protocol, s.listenAddr)
	if err != nil {
		log.Println(err)
		return s.ctrlChan
	}
	s.sock.SetReadBuffer(2048000)

	s.wg.Add(1)
	go func(msg <-chan int) {
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

	s.wg.Add(1)
	go func() {
		for {
			//handle each packet in a seperate go routine
			msgLen, _, _, addr, err := s.sock.ReadMsgUDP(bb, nil)
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
			log.Println("New Connection from", addr.String())
			nullBytes := bytes.Count(msg, []byte{'\x00'})

			//Normal TFTP connection
			if nullBytes == 3 {
				if binary.BigEndian.Uint16(msg[:2]) == tftp.OpcodeRead {
					pkt := &tftp.TFTPReadWritePkt{}
					pkt.Unpack(msg)
					s.Start(addr, pkt)
				} else if binary.BigEndian.Uint16(msg[:2]) == tftp.OpcodeWrite {
					pkt := &tftp.TFTPReadWritePkt{}
					pkt.Unpack(msg)
					s.Start(addr, pkt)
				} else if binary.BigEndian.Uint16(msg[:2]) == tftp.OpcodeErr {
					pkt := &tftp.TFTPErrPkt{}
					pkt.Unpack(msg)
				}
				//Option packet sent
			} else if nullBytes > 3 {
				if binary.BigEndian.Uint16(msg[:2]) == tftp.OpcodeRead {
					pkt := &tftp.TFTPOptionPkt{Options: make(map[string]string)}
					pkt.Unpack(msg)
					s.StartOptions(addr, pkt)
				} else if binary.BigEndian.Uint16(msg[:2]) == tftp.OpcodeWrite {
					pkt := &tftp.TFTPOptionPkt{Options: make(map[string]string)}
					pkt.Unpack(msg)
					s.StartOptions(addr, pkt)
				} else if binary.BigEndian.Uint16(msg[:2]) == tftp.OpcodeErr {
					pkt := &tftp.TFTPErrPkt{}
					pkt.Unpack(msg)
				}
			}

		}
	}()
	log.Printf("%s now listening on %s for incoming connections", AppName, s.listenAddr.String())
	return s.ctrlChan
}
