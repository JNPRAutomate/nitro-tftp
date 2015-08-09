package main

import "net"

//TFTPConn TFTP connection
type TFTPConn struct {
	Type       uint16
	remote     *net.UDPAddr
	block      uint16
	blockSize  int
	windowSize int
	timeout    int
	tsize      int
	filename   string
	BytesSent  int
	BytesRecv  int
	Options    map[string]string
}
