package main

import (
	"bytes"
	"encoding/binary"
	"log"
)

/*
	TFTP v2 RFC http://tools.ietf.org/html/rfc1350
	TFTP Option Extension http://tools.ietf.org/html/rfc2347
	TODO: TFTP Blocksize Option http://tools.ietf.org/html/rfc2348
	TODO: TFTP Timeout Interval and Transfer Size Options http://tools.ietf.org/html/rfc2349
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
	//OpcodeOptAck Options Acknowledgment
	OpcodeOptAck uint16 = 6
)

const (
	//ErrorNotDefined Not defined, see error message (if any).
	ErrorNotDefined uint16 = 0
	//ErrorFileNotFound File not found.
	ErrorFileNotFound uint16 = 1
	//ErrorAccessViolation Access violation.
	ErrorAccessViolation uint16 = 2
	//ErrorDiskFull Disk full or allocation exceeded.
	ErrorDiskFull uint16 = 3
	//ErrorIllegalOp Illegal TFTP operation.
	ErrorIllegalOp uint16 = 4
	//ErrorUnknownID Unknown transfer ID.
	ErrorUnknownID uint16 = 5
	//ErrorFileAlreadyExists File already exists.
	ErrorFileAlreadyExists uint16 = 6
	//ErrorNoSuchUser No such user.
	ErrorNoSuchUser uint16 = 7
)

const (
	//ModeNetASCII mode netascii
	ModeNetASCII string = "netascii"
	//ModeOctet mode octet
	ModeOctet string = "octet"
	//ModeMail mode mail
	ModeMail string = "mail"
)

const (
	//DefaultBlockSize the default block size of a connection
	MinBlockSize     int = 8
	MaxBlockSize     int = 65464
	DefaultBlockSize int = 512
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
	Unpack([]byte)
}

//TFTPReadWritePkt RRQ/WRQ packet
type TFTPReadWritePkt struct {
	Opcode   uint16
	Filename string
	Mode     string
}

//Pack returns []byte payload
func (p *TFTPReadWritePkt) Pack() []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, p.Opcode)
	if err != nil {
		panic(err)
	}
	buff.Write([]byte(p.Filename))
	buff.Write([]byte{0})
	buff.Write([]byte(p.Mode))
	buff.Write([]byte{0})
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPReadWritePkt) Unpack(data []byte) {
	p.Opcode = binary.BigEndian.Uint16(data[:2])
	msgParsed := bytes.Split(data[2:len(data)], []byte{00})
	p.Filename = string(msgParsed[0])
	p.Mode = string(msgParsed[1])
}

//TFTPDataPkt TFTP data Packet
type TFTPDataPkt struct {
	Opcode uint16
	Block  uint16
	Data   []byte
}

//Pack returns []byte payload
func (p *TFTPDataPkt) Pack() []byte {
	var err error
	buff := new(bytes.Buffer)
	err = binary.Write(buff, binary.BigEndian, p.Opcode)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buff, binary.BigEndian, p.Block)
	if err != nil {
		panic(err)
	}
	buff.Write([]byte(p.Data))
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPDataPkt) Unpack(data []byte) {
	p.Opcode = binary.BigEndian.Uint16(data[:2])
	p.Block = binary.BigEndian.Uint16(data[2:4])
	p.Data = data[4:]
}

//TFTPAckPkt TFTP ACK Packet
type TFTPAckPkt struct {
	Opcode uint16
	Block  uint16
}

//Pack returns []byte payload
func (p *TFTPAckPkt) Pack() []byte {
	var err error
	buff := new(bytes.Buffer)
	err = binary.Write(buff, binary.BigEndian, p.Opcode)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buff, binary.BigEndian, p.Block)
	if err != nil {
		panic(err)
	}
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPAckPkt) Unpack(data []byte) {
	p.Opcode = binary.BigEndian.Uint16(data[:2])
	p.Block = binary.BigEndian.Uint16(data[2:4])
}

//TFTPAckPkt TFTP ACK Packet
type TFTPOptionAckPkt struct {
	Opcode  uint16
	Options map[string]string
}

//Pack returns []byte payload
func (p *TFTPOptionAckPkt) Pack() []byte {
	var err error
	buff := new(bytes.Buffer)
	err = binary.Write(buff, binary.BigEndian, p.Opcode)
	if err != nil {
		panic(err)
	}
	for k, v := range p.Options {
		buff.Write([]byte(k))
		buff.Write([]byte{0})
		buff.Write([]byte(v))
		buff.Write([]byte{0})
	}
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPOptionAckPkt) Unpack(data []byte) {
	p.Opcode = binary.BigEndian.Uint16(data[:2])
	msgParsed := bytes.Split(data[2:len(data)], []byte{00})
	parsedLen := len(msgParsed)
	log.Println("PL", parsedLen)
	log.Println(msgParsed)
	k := 0
	v := 1
	for parsedLen > v {
		p.Options[string(msgParsed[k])] = string(msgParsed[v])
		k = k + 2
		v = v + 2
	}
}

//TFTPErrPkt TFTP error Packet
type TFTPErrPkt struct {
	Opcode  uint16
	ErrCode uint16
	ErrMsg  string
}

//Pack returns []byte payload
func (p *TFTPErrPkt) Pack() []byte {
	var err error
	buff := new(bytes.Buffer)
	err = binary.Write(buff, binary.BigEndian, p.Opcode)
	if err != nil {
		panic(err)
	}
	err = binary.Write(buff, binary.BigEndian, p.ErrCode)
	if err != nil {
		panic(err)
	}
	buff.Write([]byte(p.ErrMsg))
	buff.Write([]byte{0})
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPErrPkt) Unpack(data []byte) {
}

//TFTPOptionPkt TFTP Option packet
type TFTPOptionPkt struct {
	Opcode   uint16
	Filename string
	Mode     string
	Options  map[string]string
}

//TFTPOptionPktMaxSize Maximum Packet Size of an Option Packet
const TFTPOptionPktMaxSize = 512

//Pack returns []byte payload
func (p *TFTPOptionPkt) Pack() []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, p.Opcode)
	if err != nil {
		panic(err)
	}
	buff.Write([]byte(p.Filename))
	buff.Write([]byte{0})
	buff.Write([]byte(p.Mode))
	buff.Write([]byte{0})
	//Wrote Options
	for k, v := range p.Options {
		buff.Write([]byte(k))
		buff.Write([]byte{0})
		buff.Write([]byte(v))
		buff.Write([]byte{0})
	}
	return buff.Bytes()
}

//Unpack loads []byte payload
func (p *TFTPOptionPkt) Unpack(data []byte) {
	p.Opcode = binary.BigEndian.Uint16(data[:2])
	msgParsed := bytes.Split(data[2:len(data)], []byte{00})
	parsedLen := len(msgParsed)
	p.Filename = string(msgParsed[0])
	p.Mode = string(msgParsed[1])
	log.Println("PL", parsedLen)
	log.Println(msgParsed)
	k := 2
	v := 3
	if parsedLen > 2 {
		for parsedLen > v {
			p.Options[string(msgParsed[k])] = string(msgParsed[v])
			k = k + 2
			v = v + 2
		}
	}
}
