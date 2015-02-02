package main

import (
	"log"
	"net"
	"time"
)

type TFTPClientMgr struct {
	Connections map[string]*TFTPConn
}

func (c *TFTPClientMgr) Start(addr *net.UDPAddr, msg interface{}) {
	//add connection
	if tftpMsg, ok := msg.(*TFTPReadWritePkt); ok {
		nc := &TFTPConn{Type: tftpMsg.Opcode, remote: addr, block: 1}
		c.Connections[addr.String()] = nc
		//send data
		log.Printf("%#v", tftpMsg)
		if tftpMsg.Opcode == OpcodeRead {
			c.sendData(addr.String())
		} else if tftpMsg.Opcode == OpcodeWrite {
			c.recieveData(addr.String())
		}
	}
	//send error message
}

//ACK handle ack packet
func (c *TFTPClientMgr) ACK(addr *net.UDPAddr, msg interface{}) {
	if tftpMsg, ok := msg.(*TFTPAckPkt); ok {
		log.Printf("%#v", tftpMsg)
	}
}

func (c *TFTPClientMgr) sendAck(tid string) {

}

func (c *TFTPClientMgr) sendError(opcode int, tid string) {

}

func (c *TFTPClientMgr) sendData(tid string) {
	//read from file send to destination, update block
	if r, err := net.DialUDP(ServerNet, nil, c.Connections[tid].remote); err != nil {
		log.Println(err)
	} else {
		pkt := &TFTPDataPkt{Opcode: OpcodeData, Block: c.Connections[tid].block, Data: []byte("food")}
		log.Printf("%#v %#v", r.RemoteAddr(), r.LocalAddr())
		log.Printf("%b", pkt.Pack())
		r.SetWriteDeadline(time.Now().Add(1 * time.Second))
		if _, err := r.Write(pkt.Pack()); err != nil {
			log.Println(err)
		}
	}

}

func (c *TFTPClientMgr) recieveData(tid string) {

}

type TFTPConn struct {
	Type   uint16
	remote *net.UDPAddr
	block  int
}
