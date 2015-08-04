package main

import (
	"bufio"
	"encoding/binary"
	"net"
	"os"
	"path"
	"strconv"
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

//Start starta new TFTP client session
func (c *TFTPServer) StartOptions(addr *net.UDPAddr, msg interface{}) {
	blksize := DefaultBlockSize
	windowsize := DefaultWindowSize
	timeout := DefaultTimeout
	tsize := DefaultTSize

	log.Printf("%#v", msg.(*TFTPOptionPkt).Options)

	if val, ok := msg.(*TFTPOptionPkt).Options["blksize"]; ok {
		var err error
		blksize, err = strconv.Atoi(val)
		if err != nil {
			log.Error(err)
		}
		if blksize < MinBlockSize || blksize > MaxBlockSize {
			log.Error("Block size out of the valid range")
		}
	}

	if val, ok := msg.(*TFTPOptionPkt).Options["windowsize"]; ok {
		var err error
		windowsize, err = strconv.Atoi(val)
		if err != nil {
			log.Error(err)
		}
		if windowsize < MinWindowSize || blksize > MaxWindowSize {
			log.Error("Window size out of the valid range")
		}
	}

	if val, ok := msg.(*TFTPOptionPkt).Options["timeout"]; ok {
		var err error
		timeout, err = strconv.Atoi(val)
		if err != nil {
			log.Error(err)
		}
		if timeout < MinTimeout || timeout > MaxTimeout {
			log.Error("Timeout out of the valid range")
		}
	}

	if val, ok := msg.(*TFTPOptionPkt).Options["tsize"]; ok {
		var err error
		tsize, err = strconv.Atoi(val)
		if err != nil {
			log.Error(err)
		}
	}

	//add connection
	log.Println(blksize)
	if tftpMsg, ok := msg.(*TFTPOptionPkt); ok {
		nc := &TFTPConn{Type: tftpMsg.Opcode, remote: addr, timeout: timeout, tsize: tsize, windowSize: windowsize, blockSize: blksize, filename: msg.(*TFTPOptionPkt).Filename, Options: tftpMsg.Options}
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

func (c *TFTPServer) sendOptAck(conn *net.UDPConn, tid string, opts map[string]string) {
	pkt := &TFTPOptionAckPkt{Opcode: OpcodeOptAck, Options: opts}
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
		if len(c.Connections[tid].Options) > 0 {
			c.sendOptAck(r, tid, c.Connections[tid].Options)
		}

		msgChan := make(chan []byte)
		dataChan := make(chan bool)

		fileName := strings.Join([]string{c.outgoingDir, c.Connections[tid].filename}, "/")

		//listen for message
		c.clientwg.Add(1)
		go func() {
			bb := make([]byte, 1024000)

			for {
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
					close(msgChan)
					c.clientwg.Done()
					return
				}
				//pull message from buffer
				msg := bb[:msgLen]
				//clear buffer
				bb = bb[:cap(bb)]
				msgChan <- msg
			}
		}()

		//process message
		c.clientwg.Add(1)
		go func() {
			for {
				select {
				case msg, open := <-msgChan:
					if !open {
						c.clientwg.Done()
						return
					}
					if binary.BigEndian.Uint16(msg[:2]) == OpcodeACK {
						pkt := &TFTPAckPkt{}
						pkt.Unpack(msg)
						//Write Data
						c.Connections[tid].block = pkt.Block + 1
						dataChan <- true
					} else {
						//TODO: send error
					}
				}
			}
		}()

		//send packet
		c.clientwg.Add(1)
		go func(fileName string) {
			inputFile, err := os.OpenFile(path.Clean(fileName), os.O_RDWR|os.O_CREATE, 0660)
			defer inputFile.Close()
			if err != nil {
				//Unable to open file, send error to client
				log.Error(err)
			}
			buffer := make([]byte, c.Connections[tid].blockSize)

			for {
				select {
				case send := <-dataChan:
					if send {
						dLen, err := inputFile.Read(buffer)
						if err != nil {
							if err.Error() != "EOF" {
								log.Error(err)
							}
							if c.Connections[tid].tsize != 0 {
								if c.Connections[tid].tsize == c.Connections[tid].BytesSent {
									log.Printf("Sending file %s to client %s complete, total size %d matching tsize option", c.Connections[tid].filename, tid, c.Connections[tid].BytesSent)
								} else {
									log.Errorf("Error sending file %s to client %s, total size %d not matching tsize option %d", c.Connections[tid].filename, tid, c.Connections[tid].BytesSent, c.Connections[tid].tsize)
									//TODO: SEND ERROR
									close(dataChan)
									r.Close()
									return
								}
							} else {
								log.Printf("Sending file %s to client %s complete, total size %d", c.Connections[tid].filename, tid, c.Connections[tid].BytesSent)
							}
							close(dataChan)
							r.Close()
							return
						}
						pkt := &TFTPDataPkt{Opcode: OpcodeData, Block: c.Connections[tid].block, Data: buffer[:dLen]}
						r.SetWriteDeadline(time.Now().Add(1 * time.Second))
						if _, err := r.Write(pkt.Pack()); err != nil {
							log.Println(err)
						}
						c.Connections[tid].BytesSent = c.Connections[tid].BytesSent + len(pkt.Data)
						buffer = buffer[:cap(buffer)]
					}
				}
			}

		}(fileName)
	}
}

func (c *TFTPServer) recieveData(tid string) {
	if r, err := net.DialUDP(ServerNet, nil, c.Connections[tid].remote); err != nil {
		log.Println(err)
	} else {
		if len(c.Connections[tid].Options) > 0 {
			c.sendOptAck(r, tid, c.Connections[tid].Options)
		} else {
			c.sendAck(r, tid)
		}

		msgChan := make(chan []byte)
		dataChan := make(chan *TFTPDataPkt)

		fileName := strings.Join([]string{c.incomingDir, c.Connections[tid].filename}, "/")

		//listen for message
		c.clientwg.Add(1)
		go func() {
			bb := make([]byte, 1024000)

			for {
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
					close(msgChan)
					c.clientwg.Done()
					return
				}
				//pull message from buffer
				msg := bb[:msgLen]
				//clear buffer
				bb = bb[:cap(bb)]
				msgChan <- msg
			}
		}()

		//process message
		c.clientwg.Add(1)
		go func() {
			for {
				select {
				case msg, open := <-msgChan:
					if !open {
						c.clientwg.Done()
						return
					}
					if binary.BigEndian.Uint16(msg[:2]) == OpcodeData {
						pkt := &TFTPDataPkt{}
						pkt.Unpack(msg)
						dataChan <- pkt
						//Write Data
					} else if binary.BigEndian.Uint16(msg[:2]) == OpcodeErr {
						//Handle Errors
					} else {
						//TODO: send error
					}
				}
			}
		}()

		//recieve packet
		c.clientwg.Add(1)
		go func(fileName string) {
			outputFile, err := os.OpenFile(path.Clean(fileName), os.O_RDWR|os.O_CREATE, 0660)
			if err != nil {
				//Unable to open file, send error to client
				log.Println(err)
			}
			defer outputFile.Close()
			outputWriter := bufio.NewWriter(outputFile)

			for {
				select {
				case pkt, open := <-dataChan:
					if open {

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
						if len(pkt.Data) < c.Connections[tid].blockSize {
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
							if c.Connections[tid].tsize != 0 {
								if c.Connections[tid].tsize == c.Connections[tid].BytesRecv {
									log.Printf("Writing file %s from client %s complete, total size %d matching tsize option", c.Connections[tid].filename, tid, c.Connections[tid].BytesRecv)
								} else {
									log.Errorf("Error receiving file %s to client %s, total size %d not matching tsize option %d", c.Connections[tid].filename, tid, c.Connections[tid].BytesRecv, c.Connections[tid].tsize)
									//TODO: SEND ERROR
									close(dataChan)
									r.Close()
									return
								}
							} else {
								log.Printf("Writing file %s from client %s complete, total size %d", c.Connections[tid].filename, tid, c.Connections[tid].BytesRecv)
							}
							return
						}
						//continue to read data
						c.sendAck(r, tid)
						err = outputWriter.Flush()
						if err != nil {
							//Unable to write to file
							log.Errorln(err)
						}
					} else {
						//channel closed
					}
				}
			}

		}(fileName)

	}
}

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
