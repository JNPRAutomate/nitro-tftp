package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"path"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

//Start starta new TFTP client session
func (c *TFTPServer) Start(addr *net.UDPAddr, msg interface{}) {
	//add connection
	if tftpMsg, ok := msg.(*TFTPReadWritePkt); ok {
		nc := &TFTPConn{Type: tftpMsg.Opcode, remote: addr, blockSize: DefaultBlockSize, filename: msg.(*TFTPReadWritePkt).Filename}
		c.Connections[addr.String()] = nc
		if tftpMsg.Opcode == OpcodeRead {
			//Setting block to min of 1
			log.Printf("Sending file %s to client %s", nc.filename, addr.String())
			c.Connections[addr.String()].block = 1
			c.sendData(addr.String())
		} else if tftpMsg.Opcode == OpcodeWrite {
			//Setting block to min of 0
			log.Printf("Receiving file %s from client %s", nc.filename, addr.String())
			c.Connections[addr.String()].block = 0
			c.recieveData(addr.String())
		}
	}
	//send error message
}

//ACK handle ack packet
func (c *TFTPServer) ACK(addr *net.UDPAddr, msg interface{}) {
	if tftpMsg, ok := msg.(*TFTPAckPkt); ok {
		log.Printf("%#v", tftpMsg)
	}
}

func (c *TFTPServer) sendAck(conn *net.UDPConn, tid string) {
	pkt := &TFTPAckPkt{Opcode: OpcodeACK, Block: c.Connections[tid].block}
	conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	if _, err := conn.Write(pkt.Pack()); err != nil {
		log.Println(err)
	}
}

func (c *TFTPServer) sendError(opcode int, tid string) {

}

func (c *TFTPServer) sendData(tid string) {
	//TODO: Implement reverse of recieve data
	//read from file send to destination, update block
	if r, err := net.DialUDP(ServerNet, nil, c.Connections[tid].remote); err != nil {
		log.Println(err)
	} else {
		c.clientwg.Add(1)
		go func() {
			defer c.clientwg.Done()
			buffer := make([]byte, c.Connections[tid].blockSize)
			bb := make([]byte, 1024000)
			fileName := strings.Join([]string{c.outgoingDir, c.Connections[tid].filename}, "/")
			inputFile, err := os.OpenFile(path.Clean(fileName), os.O_RDWR|os.O_CREATE, 0660)
			defer inputFile.Close()
			if err != nil {
				//Unable to open file, send error to client
				log.Println(err)
			}
			inputReader := bufio.NewReader(inputFile)
			for {
				dLen, err := inputReader.Read(buffer)

				if err != nil {
					//unable to read from file
					log.Println(err)
				}
				pkt := &TFTPDataPkt{Opcode: OpcodeData, Block: c.Connections[tid].block, Data: bytes.Trim(buffer, "\x00")}
				r.SetWriteDeadline(time.Now().Add(1 * time.Second))
				if _, err := r.Write(pkt.Pack()); err != nil {
					log.Println(err)
				}
				c.Connections[tid].BytesSent = c.Connections[tid].BytesSent + len(pkt.Data)
				buffer = buffer[:cap(buffer)]
				//TODO: send next packet once block is sent
				msgLen, _, _, _, err := r.ReadMsgUDP(bb, nil)
				if err != nil {
					switch err := err.(type) {
					case net.Error:
						if err.Timeout() {
							log.Error(err)
						} else if err.Temporary() {
							log.Error(err)
						}
					}
					return
				}
				//pull message from buffer
				msg := bb[:msgLen]
				//clear buffer
				bb = bb[:cap(bb)]

				//TODO: Process ACK
				if uint16(msg[1]) == OpcodeACK {
					pkt := &TFTPAckPkt{}
					pkt.Unpack(msg)
					//Write Data
					c.Connections[tid].block = c.Connections[tid].block + 1
				} else {
					//TODO: send error
				}
				if c.Connections[tid].blockSize > dLen {
					log.Printf("Sending file %s to client %s complete, total size %d", c.Connections[tid].filename, tid, c.Connections[tid].BytesSent)
					return
				}
			}
		}()
	}
}

func (c *TFTPServer) recieveData(tid string) {
	if r, err := net.DialUDP(ServerNet, nil, c.Connections[tid].remote); err != nil {
		log.Println(err)
	} else {
		c.sendAck(r, tid)
		c.clientwg.Add(1)
		go func() {
			defer c.clientwg.Done()
			bb := make([]byte, 1024000)
			fileName := strings.Join([]string{c.outgoingDir, c.Connections[tid].filename}, "/")
			outputFile, err := os.OpenFile(path.Clean(fileName), os.O_RDWR|os.O_CREATE, 0660)
			if err != nil {
				//Unable to open file, send error to client
				log.Println(err)
			}
			outputWriter := bufio.NewWriter(outputFile)
			for {
				//handle each packet in a seperate go routine
				msgLen, _, _, _, err := r.ReadMsgUDP(bb, nil)
				if err != nil {
					switch err := err.(type) {
					case net.Error:
						if err.Timeout() {
							log.Error(err)
						} else if err.Temporary() {
							log.Error(err)
						}
					}
					return
				}
				//pull message from buffer
				msg := bb[:msgLen]
				//clear buffer
				bb = bb[:cap(bb)]
				opCode := binary.BigEndian.Uint16(msg[:2])
				if opCode == OpcodeData {
					pkt := &TFTPDataPkt{}
					pkt.Unpack(msg)
					//Write Data
					ofb, err := outputWriter.Write(pkt.Data)
					if err != nil {
						//Unable to write to file
						log.Println(err)
					}
					//add bytes received
					c.Connections[tid].BytesRecv = c.Connections[tid].BytesRecv + ofb
					if c.Debug {
						log.Debug("Wrote %d bytes to file %s", ofb, c.Connections[tid].filename)
					}
					c.Connections[tid].block = pkt.Block
					if len(pkt.Data) < DefaultBlockSize {
						//last packet
						c.sendAck(r, tid)
						err := r.Close()
						if err != nil {
							panic(err)
						}
						err = outputWriter.Flush()
						if err != nil {
							log.Println(err)
						}
						log.Printf("Writing file %s from client %s complete, total size %d", c.Connections[tid].filename, tid, c.Connections[tid].BytesRecv)
						return
					}
					//continue to read data
					c.sendAck(r, tid)
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
	filename  string
	BytesSent int
	BytesRecv int
}
