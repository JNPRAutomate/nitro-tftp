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

//UDPServer A server to listen for UDP messages
type UDPServer struct {
	listenAddr *net.UDPAddr
	sock       *net.UDPConn
}

//Listen Listen for connections
func (s *UDPServer) Listen() {
	var err error
	cMgr := &TFTPClientMgr{Connections: make(map[string]*TFTPConn)}
	//s.listenAddr = &net.UDPAddr{IP: net.ParseIP(DefaultIP), Port: DefaultPort}
	s.listenAddr, err = net.ResolveUDPAddr(ServerNet, strings.Join([]string{DefaultIP, strconv.Itoa(DefaultPort)}, ":"))
	if err != nil {
		panic(err)
	}
	bb := make([]byte, 1024000)

	s.sock, err = net.ListenUDP("udp4", s.listenAddr)
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
