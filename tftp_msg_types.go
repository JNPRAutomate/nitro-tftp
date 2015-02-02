package main

import "bytes"

/*
	TFTP v2 RFC http://tools.ietf.org/html/rfc1350
	TFTP Option Extension http://tools.ietf.org/html/rfc2347
	TODO: TFTP Blocksize Option http://tools.ietf.org/html/rfc2348http://tools.ietf.org/html/rfc2348
	TODO: TFTP Timeout Interval and Transfer Size Options http://tools.ietf.org/html/rfc2349http://tools.ietf.org/html/rfc2349
	TODO: TFTP Windowsize Option http://tools.ietf.org/html/rfc7440
*/

const (
	//OpcodeRead Read request (RRQ)
	OpcodeRead uint16 = 1
	//OpcodeWrite Write request (WRQ)
	OpcodeWrite uint16 = 2
	//OpcodeData Data (DATA)
	OpcodeData uint16 = 3
	//OpcodeACK Acknowledgment (ACK)
	OpcodeACK uint16 = 4
	//OpcodeErr Error (ERROR)
	OpcodeErr uint16 = 5
)

const (
	//ErrorNotDefined Not defined, see error message (if any).
	ErrorNotDefined = 0
	//ErrorFileNotFound File not found.
	ErrorFileNotFound = 1
	//ErrorAccessViolation Access violation.
	ErrorAccessViolation = 2
	//ErrorDiskFull Disk full or allocation exceeded.
	ErrorDiskFull = 3
	//ErrorIllegalOp Illegal TFTP operation.
	ErrorIllegalOp = 4
	//ErrorUnknownID Unknown transfer ID.
	ErrorUnknownID = 5
	//ErrorFileAlreadyExists File already exists.
	ErrorFileAlreadyExists = 6
	//ErrorNoSuchUser No such user.
	ErrorNoSuchUser = 7
)

const (
	//ModeNetASCII mode netascii
	ModeNetASCII string = "netascii"
	//ModeOctet mode octet
	ModeOctet string = "octet"
	//ModeMail mode mail
	ModeMail string = "mail"
)

/*
	RRQ/WRQ packet

	2 bytes     string    1 byte     string   1 byte
	------------------------------------------------
	| Opcode |  Filename  |   0  |    Mode    |   0  |
	------------------------------------------------

*/

//TFTPPacket interface to packet types
type TFTPPacket interface {
	Pack() []byte
	Unpack()
}

//TFTPReadWritePkt RRQ/WRQ packet
type TFTPReadWritePkt struct {
	Opcode   uint16
	Filename string
	Mode     string
}

//Pack returns []byte payload
func (p *TFTPReadWritePkt) Pack() []byte {
	var buff bytes.Buffer
	buff.Write([]byte{byte(p.Opcode)})
	buff.Write([]byte(p.Filename))
	buff.Write([]byte{0})
	buff.Write([]byte(p.Mode))
	buff.Write([]byte{0})
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPReadWritePkt) Unpack(data []byte) {
	p.Opcode = uint16(data[1])
	msgParsed := bytes.Split(data[2:len(data)], []byte{00})
	p.Filename = string(msgParsed[0])
	p.Mode = string(msgParsed[1])
}

//TFTPDataPkt TFTP data Packet
type TFTPDataPkt struct {
	Opcode uint16
	Block  int
	Data   []byte
}

//Pack returns []byte payload
func (p *TFTPDataPkt) Pack() []byte {
	var buff bytes.Buffer
	buff.Write([]byte{0})
	buff.Write([]byte{byte(p.Opcode)})
	buff.Write([]byte{0})
	buff.Write([]byte{byte(p.Block)})
	buff.Write([]byte(p.Data))
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPDataPkt) Unpack(data []byte) {
}

//TFTPAckPkt TFTP ACK Packet
type TFTPAckPkt struct {
	Opcode []byte
	Block  []byte
}

//Pack returns []byte payload
func (p *TFTPAckPkt) Pack() []byte {
	var buff bytes.Buffer
	buff.Write(p.Opcode)
	buff.Write([]byte(p.Block))
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPAckPkt) Unpack(data []byte) {
}

//TFTPErrPkt TFTP error Packet
type TFTPErrPkt struct {
	Opcode  []byte
	ErrCode []byte
	ErrMsg  string
}

//Pack returns []byte payload
func (p *TFTPErrPkt) Pack() []byte {
	var buff bytes.Buffer
	buff.Write(p.Opcode)
	buff.Write([]byte(p.ErrCode))
	buff.Write([]byte(p.ErrMsg))
	buff.Write([]byte{0})
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPErrPkt) Unpack(data []byte) {
}

//TFTPOptionPkt TFTP Option packet
type TFTPOptionPkt struct {
	Opcode    []byte
	OptionAck []byte
	Value1    []byte
	OptN      []byte
	ValueN    []byte
}

//Pack returns []byte payload
func (p *TFTPOptionPkt) Pack() []byte {
	var buff bytes.Buffer
	buff.Write(p.Opcode)
	buff.Write(p.OptionAck)
	buff.Write([]byte{0})
	buff.Write(p.Value1)
	buff.Write([]byte{0})
	buff.Write(p.OptN)
	buff.Write([]byte{0})
	buff.Write(p.ValueN)
	buff.Write([]byte{0})
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPOptionPkt) Unpack() {
}
