package main

import "net"

//TFTPConn TFTP connection
type TFTPConn struct {
	Type        uint16
	remote      *net.UDPAddr
	block       uint16
	blockSize   int
	windowSize  int
	timeout     int
	tsize       int
	filename    string
	BytesSent   int
	BytesRecv   int
	Options     map[string]string
	PacketsSent int
	PacketsRecv int
	ACKsSent    int
	ACKsRecv    int
	OptACKsSent int
	OptACKsRecv int
	ErrorsSent  int
	ErrorsRecv  int
}

func (t *TFTPConn) DataSent(count int) {
	t.BytesSent = t.BytesSent + count
	t.PacketSent()
}

func (t *TFTPConn) DataRecv(count int) {
	t.BytesRecv = t.BytesRecv + count
	t.PacketRecv()
}

func (t *TFTPConn) PacketSent() {
	t.PacketsSent = t.PacketsSent + 1
}
func (t *TFTPConn) PacketRecv() {
	t.PacketsRecv = t.PacketsRecv + 1
}

func (t *TFTPConn) ACKSent() {
	t.ACKsSent = t.ACKsSent + 1
	t.PacketSent()
}

func (t *TFTPConn) ACKRecv() {
	t.ACKsSent = t.ACKsRecv + 1
	t.PacketRecv()
}

func (t *TFTPConn) ErrorSent() {
	t.ErrorsSent = t.ErrorsSent + 1
	t.PacketSent()
}

func (t *TFTPConn) ErrorRecv() {
	t.ErrorsRecv = t.ErrorsRecv + 1
	t.PacketRecv()
}

func (t *TFTPConn) OptACKSent() {
	t.OptACKsSent = t.OptACKsSent + 1
	t.PacketSent()
}

func (t *TFTPConn) OptACKRecv() {
	t.OptACKsSent = t.OptACKsRecv + 1
	t.PacketRecv()
}
