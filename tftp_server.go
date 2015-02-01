package main

import (
	"bytes"
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
	s.listenAddr = &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: 69}

	var err error
	buffer := make([]byte, 9600)

	s.sock, err = net.ListenUDP("udp4", s.listenAddr)
	if err != nil {
		log.Println(err)
	}
	s.sock.SetReadBuffer(1024000)
	//TODO: Send ACKStartMsg on control channel
	for {
		//handle each packet in a seperate go routine
		_, _, err := s.sock.ReadFromUDP(buffer)
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
		msg := bytes.Trim(buffer, "\x00")

		if int(msg[0]) == OpcodeRead {
			pkt := &TFTPReadWritePkt{}
			pkt.Unpack(msg)
			log.Printf("%#v", pkt)
		} else if int(msg[0]) == OpcodeWrite {
			pkt := &TFTPReadWritePkt{}
			pkt.Unpack(msg)
			log.Printf("%#v", pkt)
		} else if int(msg[0]) == OpcodeACK {

		} else if int(msg[0]) == OpcodeErr {

		} else if int(msg[0]) == OpcodeData {

		}

	}
}
