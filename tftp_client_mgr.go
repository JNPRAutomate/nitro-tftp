package main

import (
	"log"
	"net"
	"sync"
	"time"
)

//TFTPClientMgr manages TFTP clients
type TFTPClientMgr struct {
	Connections map[string]*TFTPConn
	wg          sync.WaitGroup
}

//Start starta new TFTP client session
func (c *TFTPClientMgr) Start(addr *net.UDPAddr, msg interface{}) {
	//add connection
	if tftpMsg, ok := msg.(*TFTPReadWritePkt); ok {
		nc := &TFTPConn{Type: tftpMsg.Opcode, remote: addr}
		c.Connections[addr.String()] = nc
		if tftpMsg.Opcode == OpcodeRead {
			//Setting block to min of 1
			c.Connections[addr.String()].block = 1
			c.sendData(addr.String())
		} else if tftpMsg.Opcode == OpcodeWrite {
			//Setting block to min of 0
			c.Connections[addr.String()].block = 0
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

func (c *TFTPClientMgr) sendAck(conn *net.UDPConn, tid string) {
	pkt := &TFTPAckPkt{Opcode: OpcodeACK, Block: c.Connections[tid].block}
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	if _, err := conn.Write(pkt.Pack()); err != nil {
		log.Println(err)
	}
}

func (c *TFTPClientMgr) sendError(opcode int, tid string) {

}

func (c *TFTPClientMgr) sendData(tid string) {
	//TODO: Implement reverse of recieve data
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
	if r, err := net.DialUDP(ServerNet, nil, c.Connections[tid].remote); err != nil {
		log.Println(err)
	} else {
		c.sendAck(r, tid)
		bb := make([]byte, 1024000)
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			for {
				//handle each packet in a seperate go routine
				msgLen, _, _, _, err := r.ReadMsgUDP(bb, nil)
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
				//pull message from buffer
				msg := bb[:msgLen]
				//clear buffer
				bb = bb[:cap(bb)]
				if uint16(msg[1]) == OpcodeData {
					pkt := &TFTPDataPkt{}
					pkt.Unpack(msg)
					//	log.Printf("%#v", pkt)
					c.Connections[tid].block = pkt.Block
					if len(pkt.Data) < DefaultBlockSize {
						//last packet
						c.sendAck(r, tid)
						err := r.Close()
						if err != nil {
							panic(err)
						}
						return
					}
					//continue to read data
					c.sendAck(r, tid)
					//TODO: Write data
				} else {
					//TODO: send error
				}
			}
		}()

	}
}

//TFTPConn TFTP connection
type TFTPConn struct {
	Type      uint16
	remote    *net.UDPAddr
	block     uint16
	blockSize int
}
