package main

import (
	"log"
	"net"
)

//UDPServer A server to listen for UDP messages
type UDPServer struct {
	listenAddr *net.UDPAddr
	sock       *net.UDPConn
}

//Listen Listen to
func (s *UDPServer) Listen() {
	cMgr := &TFTPClientMgr{}
	s.listenAddr = &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 69}

	var err error
	bb := make([]byte, 9600)

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
		log.Println(msg)

		if int(msg[1]) == OpcodeRead {
			pkt := &TFTPReadWritePkt{}
			pkt.Unpack(msg)
			cMgr.Start(addr, pkt)
		} else if int(msg[1]) == OpcodeWrite {
			pkt := &TFTPReadWritePkt{}
			pkt.Unpack(msg)
			cMgr.Start(addr, pkt)
		} else if int(msg[1]) == OpcodeACK {
			pkt := &TFTPAckPkt{}
			pkt.Unpack(msg)
		} else if int(msg[1]) == OpcodeErr {
			pkt := &TFTPErrPkt{}
			pkt.Unpack(msg)
		} else if int(msg[1]) == OpcodeData {
			pkt := &TFTPDataPkt{}
			pkt.Unpack(msg)
		}

	}
}
