package main

import (
	"log"
	"net"
)

type TFTPClientMgr struct {
	Connections map[string]*TFTPConn
}

func (c *TFTPClientMgr) Start(addr *net.UDPAddr, msg interface{}) {
	//add connection
	log.Println(addr.String())
}

type TFTPConn struct {
	Type   int
	remote net.UDPAddr
	block  int
}
