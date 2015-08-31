package main

import (
	"net"
	"time"
)

//TFTPConn TFTP connection
type TFTPConn struct {
	Type        uint16            //Type connection type
	remote      *net.UDPAddr      //remote Remove UDP socket
	block       uint16            //block current block ID
	blockSize   int               //blockSize set blocksize option
	windowSize  int               //windowSize set window size optin
	timeout     int               //timeout set timeout option
	tsize       int               //tsize set tsize option
	filename    string            //filename filename being accessed
	Options     map[string]string //Options TFTP options for connection
	BytesRecv   int               //BytesRecv total bytes received while active
	BytesSent   int               //BytesSent total bytes sent while active
	PacketsSent int               //PacketsSent total packets sent for client
	PacketsRecv int               //PacketsRecv total packets received from client
	ACKsSent    int               //ACKsSent total acks sent to client
	ACKsRecv    int               //ACKsRecv total acks received from client
	OptACKsSent int               //OptACKsSent total option acks sent to client
	OptACKsRecv int               //OptACKsRecv total option acks received from client
	ErrorsSent  int               //ErrorsSent total errors sent to client
	ErrorsRecv  int               //ErrorsRecv total erros received from client
	StartTime   time.Time         //StartTime time connection started
}

//DataSent increment data sent
func (t *TFTPConn) DataSent(count int) {
	t.BytesSent = t.BytesSent + count
	t.PacketSent()
}

//DataRecv increment data received
func (t *TFTPConn) DataRecv(count int) {
	t.BytesRecv = t.BytesRecv + count
	t.PacketRecv()
}

//PacketSent increment packets sent
func (t *TFTPConn) PacketSent() {
	t.PacketsSent = t.PacketsSent + 1
}

//PacketRecv inrement packets received
func (t *TFTPConn) PacketRecv() {
	t.PacketsRecv = t.PacketsRecv + 1
}

//ACKSent increment packets sent
func (t *TFTPConn) ACKSent() {
	t.ACKsSent = t.ACKsSent + 1
	t.PacketSent()
}

//ACKRecv increment acks received
func (t *TFTPConn) ACKRecv() {
	t.ACKsSent = t.ACKsRecv + 1
	t.PacketRecv()
}

//ErrorSent inciment errors sent
func (t *TFTPConn) ErrorSent() {
	t.ErrorsSent = t.ErrorsSent + 1
	t.PacketSent()
}

//ErrorRecv increment errors received
func (t *TFTPConn) ErrorRecv() {
	t.ErrorsRecv = t.ErrorsRecv + 1
	t.PacketRecv()
}

//OptACKSent increment opt acks sent
func (t *TFTPConn) OptACKSent() {
	t.OptACKsSent = t.OptACKsSent + 1
	t.PacketSent()
}

//OptACKRecv increment opt acks received
func (t *TFTPConn) OptACKRecv() {
	t.OptACKsSent = t.OptACKsRecv + 1
	t.PacketRecv()
}
