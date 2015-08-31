package main

import "time"

//StatsMgr Used to manage usage stats
type StatsMgr struct {
	Clients            map[string]*ClientStats //Clients Per client stats
	TotalBytesReceived int                     //TotalBytesReceived total bytes received while active
	TotalBytesSent     int                     //TotalBytesSent total bytes sent while active
	StartTime          time.Time               //StartTime time process was started
	TotalTransferTime  time.Time               //TotalTransferTime total time spent serving files
}

//UpdateClientStats update the cumulative stats for the client
func (s *StatsMgr) UpdateClientStats(c *TFTPConn) {
	if _, ok := s.Clients[c.remote.IP.String()]; !ok {
		s.Clients[c.remote.IP.String()] = &ClientStats{SrcPortsUsed: make(map[int]int), OptionsUsed: make(map[string]int)}
	}
	s.Clients[c.remote.IP.String()].Connections = s.Clients[c.remote.IP.String()].Connections + 1
	s.Clients[c.remote.IP.String()].BytesRecv = s.Clients[c.remote.IP.String()].BytesRecv + c.BytesRecv
	s.Clients[c.remote.IP.String()].BytesSent = s.Clients[c.remote.IP.String()].BytesSent + c.BytesSent
	s.Clients[c.remote.IP.String()].PacketsSent = s.Clients[c.remote.IP.String()].PacketsSent + c.PacketsSent
	s.Clients[c.remote.IP.String()].PacketsRecv = s.Clients[c.remote.IP.String()].PacketsRecv + c.PacketsRecv
	s.Clients[c.remote.IP.String()].ACKsSent = s.Clients[c.remote.IP.String()].ACKsSent + c.ACKsSent
	s.Clients[c.remote.IP.String()].ACKsRecv = s.Clients[c.remote.IP.String()].ACKsRecv + c.ACKsRecv
	s.Clients[c.remote.IP.String()].OptACKsSent = s.Clients[c.remote.IP.String()].OptACKsSent + c.OptACKsSent
	s.Clients[c.remote.IP.String()].OptACKsRecv = s.Clients[c.remote.IP.String()].OptACKsRecv + c.OptACKsRecv
	s.Clients[c.remote.IP.String()].ErrorsSent = s.Clients[c.remote.IP.String()].ErrorsSent + c.ErrorsSent
	s.Clients[c.remote.IP.String()].ErrorsRecv = s.Clients[c.remote.IP.String()].ErrorsRecv + c.ErrorsRecv
	s.Clients[c.remote.IP.String()].TransferTime = s.Clients[c.remote.IP.String()].TransferTime + time.Since(c.StartTime)
	if _, ok := s.Clients[c.remote.IP.String()].SrcPortsUsed[c.remote.Port]; !ok {
		s.Clients[c.remote.IP.String()].SrcPortsUsed[c.remote.Port] = 0
	}
	s.Clients[c.remote.IP.String()].SrcPortsUsed[c.remote.Port] = s.Clients[c.remote.IP.String()].SrcPortsUsed[c.remote.Port] + 1
	if len(c.Options) > 0 {
		for opt := range c.Options {
			if _, ok := s.Clients[c.remote.IP.String()].OptionsUsed[opt]; !ok {
				s.Clients[c.remote.IP.String()].OptionsUsed[opt] = 0
			}
			s.Clients[c.remote.IP.String()].OptionsUsed[opt] = s.Clients[c.remote.IP.String()].OptionsUsed[opt] + 1
		}
	}
}

//ClientStats stats for a single client
type ClientStats struct {
	SrcPortsUsed map[int]int    //SrcPortsUsed all of the source ports
	OptionsUsed  map[string]int //OptionsUsed all of the options used
	Connections  int            //Connections total number of connections
	BytesRecv    int            //BytesRecv total bytes received while active
	BytesSent    int            //BytesSent total bytes sent while active
	PacketsSent  int            //PacketsSent total packets sent for client
	PacketsRecv  int            //PacketsRecv total packets received from client
	ACKsSent     int            //ACKsSent total acks sent to client
	ACKsRecv     int            //ACKsRecv total acks recieved from client
	OptACKsSent  int            //OptACKsSent total option acks sent to client
	OptACKsRecv  int            //OptACKsRecv total option acks recieved from client
	ErrorsSent   int            //ErrorsSent total errors sent to client
	ErrorsRecv   int            //ErrorsRecv total erros received from client
	TransferTime time.Duration  //TransferTime time spent transfering files
}
