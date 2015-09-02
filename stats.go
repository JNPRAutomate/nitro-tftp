package main

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

//StatsMgr Used to manage usage stats
type StatsMgr struct {
	Clients           map[string]*ClientStats //Clients Per client stats
	TotalBytesRecv    int                     //TotalBytesRecv total bytes received while active
	TotalBytesSent    int                     //TotalBytesSent total bytes sent while active
	StartTime         time.Time               //StartTime time process was started
	TotalTransferTime time.Duration           //TotalTransferTime total time spent serving files
	wg                sync.WaitGroup
}

func NewStatsMgr() *StatsMgr {
	return &StatsMgr{Clients: make(map[string]*ClientStats), StartTime: time.Now()}
}

//UpdateClientStats update the cumulative stats for the client
func (s *StatsMgr) UpdateClientStats(c *TFTPConn) {
	if _, ok := s.Clients[c.remote.IP.String()]; !ok {
		s.Clients[c.remote.IP.String()] = &ClientStats{SrcPortsUsed: make(map[string]int), OptionsUsed: make(map[string]int)}
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
	ttime := time.Since(c.StartTime)
	s.Clients[c.remote.IP.String()].TransferTime = s.Clients[c.remote.IP.String()].TransferTime + ttime
	s.TotalTransferTime = s.TotalTransferTime + ttime
	s.TotalBytesRecv = s.TotalBytesRecv + s.Clients[c.remote.IP.String()].BytesRecv
	s.TotalBytesSent = s.TotalBytesSent + s.Clients[c.remote.IP.String()].BytesSent
	if _, ok := s.Clients[c.remote.IP.String()].SrcPortsUsed[strconv.Itoa(c.remote.Port)]; !ok {
		s.Clients[c.remote.IP.String()].SrcPortsUsed[strconv.Itoa(c.remote.Port)] = 0
	}
	s.Clients[c.remote.IP.String()].SrcPortsUsed[strconv.Itoa(c.remote.Port)] = s.Clients[c.remote.IP.String()].SrcPortsUsed[strconv.Itoa(c.remote.Port)] + 1
	if len(c.Options) > 0 {
		for opt := range c.Options {
			if _, ok := s.Clients[c.remote.IP.String()].OptionsUsed[opt]; !ok {
				s.Clients[c.remote.IP.String()].OptionsUsed[opt] = 0
			}
			s.Clients[c.remote.IP.String()].OptionsUsed[opt] = s.Clients[c.remote.IP.String()].OptionsUsed[opt] + 1
		}
	}
}

func (s *StatsMgr) StatsAllJSON(w http.ResponseWriter, r *http.Request) {
	statsAll := struct {
		ClientStats       map[string]*ClientStats `json:"clientstats"`
		TotalBytesRecv    int                     `json:"totalbytesrecv"`
		TotalBytesSent    int                     `json:"totalbytessent"`
		StartTime         time.Time               `json:"starttime"`
		TotalTransferTime time.Duration           `json:"totaltransfertime"`
	}{s.Clients, s.TotalBytesRecv, s.TotalBytesSent, s.StartTime, s.TotalTransferTime}
	statsJSON, err := json.Marshal(statsAll)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(statsJSON)
}

//StatsListener Listen for requests to serve server stats
func (s *StatsMgr) StatsListener(ip net.IP, port int, ctrlChan chan int) {
	s.wg.Add(1)

	go func() {
		r := mux.NewRouter()
		r.HandleFunc("/api/v1/stats/all", s.StatsAllJSON)
		svr := &http.Server{}
		svr.Handler = r
		tcpaddr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(ip.String(), strconv.Itoa(port)))
		if err != nil {
			log.Error("Stats Listener ", err)
			return
		}
		l, err := net.ListenTCP("tcp", tcpaddr)
		log.Printf("%s now listening on %s for incoming connections", strings.Join([]string{AppName, "Stats Listener"}, " "), tcpaddr.String())
		if err != nil {
			log.Error("Stats Listener ", err)
			return
		}
		if err := svr.Serve(l); err != nil {
			log.Errorln(err)
		}
		for item := range ctrlChan {
			if item == -1 {
				err = l.Close()
				if err != nil {
					log.Println(err)
				}
				s.wg.Done()
				log.Printf("%s shutting down", strings.Join([]string{AppName, "Stats Listener"}, " "))
				return
			}
		}
	}()
}

//ClientStats stats for a single client
type ClientStats struct {
	SrcPortsUsed map[string]int `json:"srcportsused"`  //SrcPortsUsed all of the source ports
	OptionsUsed  map[string]int `json:"optionsused"`   //OptionsUsed all of the options used
	Connections  int            `json:"connections"`   //Connections total number of connections
	BytesRecv    int            `json:"bytesrecieved"` //BytesRecv total bytes received while active
	BytesSent    int            `json:"bytessent"`     //BytesSent total bytes sent while active
	PacketsSent  int            `json:"packetssent"`   //PacketsSent total packets sent for client
	PacketsRecv  int            `json:"packetsrecv"`   //PacketsRecv total packets received from client
	ACKsSent     int            `json:"ackssent"`      //ACKsSent total acks sent to client
	ACKsRecv     int            `json:"acksrecv"`      //ACKsRecv total acks recieved from client
	OptACKsSent  int            `json:"optackssent"`   //OptACKsSent total option acks sent to client
	OptACKsRecv  int            `json:"optacksrecv"`   //OptACKsRecv total option acks recieved from client
	ErrorsSent   int            `json:"errorssent"`    //ErrorsSent total errors sent to client
	ErrorsRecv   int            `json:"errorsrecv"`    //ErrorsRecv total erros received from client
	TransferTime time.Duration  `json:"transfertime"`  //TransferTime time spent transfering files
}
