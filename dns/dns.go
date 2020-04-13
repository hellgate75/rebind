// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package dns

import (
	"encoding/json"
	errs "errors"
	"fmt"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	pnet "github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/store"
	"github.com/hellgate75/rebind/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DNS server default port
	udpPort int = 53
	// DNS model.Packet max length
	packetLen int = 512
)

// DNSService is an implementation of DNSServer interface.
type dnsService struct {
	Conn       *net.UDPConn
	Store      registry.Store
	Bucket     store.RequestsCacheStore
	Answers    store.AnswersCacheStore
	Forwarders []net.UDPAddr
	log        log.Logger
	pipe       pnet.NetPipe
	_started   bool
}

func (s *dnsService) pipeHandler(message string) {
	tokens := strings.Split(message, " ")
	if "reload" == tokens[0] {
		s.Store.GetGroupBucket().ReLoad()
		s.pipe.Write([]byte(fmt.Sprintf("reponse %s %s", "ok", tokens[1])))
	} else if "load" == tokens[0] {
		groupId := strings.TrimSpace(tokens[1])
		g, _ := s.Store.GetGroupBucket().GetGroupById(groupId)
		s.Store.GetGroupBucket().GetGroupStore(g)
		s.pipe.Write([]byte(fmt.Sprintf("reponse %s %s", "ok", tokens[2])))
	} else if "shutdown" == tokens[0] {
		s.pipe.Write([]byte(fmt.Sprintf("reponse %s %s", "ok", tokens[1])))
		os.Exit(0)
	}
}

// Listen starts a DNS server on port 53
func (s *dnsService) Listen(ipAddress string, port int, pipeAddress string, pipePort int, pipeResponsePort int) error {
	var err error
	ipTokens := strings.Split(ipAddress, ".")
	if len(ipTokens) < 4 {
		return errs.New(fmt.Sprintf("DNSServer: Invalid ip address: %s", ipAddress))
	}
	s.pipe, err = pnet.NewInputOutputPipeWith(pipeAddress, pipePort, pipeAddress, pipeResponsePort, pnet.PipeHandler(s.pipeHandler), s.log)
	if err == nil {
		err := s.pipe.Start()
		if err != nil {
			s.log.Error("Unable to start net pipe on %s:%v", pipeAddress, pipePort)
			os.Exit(1)
		}
	} else {
		s.log.Error("Unable to create net pipe on %s:%v", pipeAddress, pipePort)
		os.Exit(1)
	}
	defer s.pipe.Stop()
	var token1, token2, token3, token4 byte
	intV, _ := strconv.Atoi(ipTokens[0])
	token1 = byte(intV)
	intV, _ = strconv.Atoi(ipTokens[1])
	token2 = byte(intV)
	intV, _ = strconv.Atoi(ipTokens[2])
	token3 = byte(intV)
	intV, _ = strconv.Atoi(ipTokens[3])
	token4 = byte(intV)
	var ip net.IP = net.IPv4(token1, token2, token3, token4)
	s.log.Debugf("Ip address: %s, Port: %v", ip.String(), port)
	s.Conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: ip, Port: port})
	if err != nil {
		s.log.Fatal(err)
	}
	if s.Conn == nil {
		return errs.New(fmt.Sprintf("Nil Connection from dialing %s:%v", ip.String(), port))
	}
	defer s.Conn.Close()
	s._started = true
	for s._started {
		buf := make([]byte, packetLen)
		s.log.Debug("Reading Network packets ...")
		_, addr, err := s.Conn.ReadFromUDP(buf)
		if err != nil {
			s.log.Errorf("DNSServer: Error reading the request -> Error: %v", err)
			continue
		}
		var m dnsmessage.Message
		err = m.Unpack(buf)
		if err != nil {
			s.log.Errorf("DNSServer: Error unpacking the request -> Error: %v", err)
			continue
		}
		s.log.Debugf("DNSServer: Questions: %v", len(m.Questions))
		if len(m.Questions) == 0 {
			continue
		}
		go s.Query(model.Packet{*addr, m})
	}
	return err
}

// Query lookup answers for DNS Message.
func (s *dnsService) Query(p model.Packet) {
	// got response from forwarder, send it back to client
	// got response from forwarder, send it back to client
	if p.Message.Header.Response {
		pKey := utils.PToString(p)
		if addresses, err := s.Bucket.Get(pKey); err != nil {
			for _, address := range addresses {
				go sendPacket(s.Conn, p.Message, address)
			}
			s.Bucket.Remove(pKey)
			qRep, _ := utils.ReplaceQuestionUnrelated(utils.QToString(p.Message.Questions[0]))
			go s.SaveBulk(qRep, p.Message.Answers)
			s.Answers.Set(qRep, p.Message.Answers...)
		}
		return
	}
	s.log.Debugf("Ip: %v", p.Addr.IP.String())
	s.log.Debugf("Port: %v", p.Addr.Port)
	zone := p.Addr.Zone
	if zone == "" {
		zone = "home"
	}
	s.log.Debugf("_zone: %v (req:%s)", zone, p.Addr.Zone)
	s.log.Debugf("Response: %v", p.Message.Header.Response)
	jsonText, _ := json.Marshal(p)
	s.log.Debugf("PACKET=%s", jsonText)
	s.log.Debugf("pKey: %s", utils.PToString(p))
	s.log.Debugf("Questions: %v", len(p.Message.Questions))
	for idx, q := range p.Message.Questions {
		qRep, _ := utils.ReplaceQuestionUnrelated(utils.QToString(q))
		s.log.Debugf("Question n. %v: %s", idx, qRep)
	}
	q1 := p.Message.Questions[0]
	qRep, _ := utils.ReplaceQuestionUnrelated(utils.QToString(q1))
	s.log.Debugf("qKey: %s", qRep)
	questionType := q1.Type.String()[4:]
	s.log.Debugf("qType: %s", questionType)
	s.log.Debugf("UDPADDR->_zone: %s", p.Addr.Zone)
	s.log.Debugf("UDPADDR->IP: %s", p.Addr.IP.String())
	s.log.Debugf("UDPADDR->Port: %v", p.Addr.Port)

	//Recover by zone, the by question
	// was checked before entering this routine
	if dnsRes, err := s.Answers.Get(qRep); err == nil {
		//Taking from cache
		p.Message.Answers = append(p.Message.Answers, dnsRes...)
		go sendPacket(s.Conn, p.Message, p.Addr)
	} else {
		// answer the question
		// seeking into the store
		val, fwds, ok := s.Store.Get(qRep)
		if ok && len(val) > 0 {
			s.Answers.Set(qRep, val...)
			p.Message.Answers = append(p.Message.Answers, val...)
			go sendPacket(s.Conn, p.Message, p.Addr)
		} else if len(fwds) > 0 {
			// forwarding
			for i := 0; i < len(fwds); i++ {
				s.Bucket.Set(utils.PToString(p), p.Addr)
				go sendPacket(s.Conn, p.Message, s.Forwarders[i])
			}
		} else {
			s.log.Warnf("Request: %s has 0 records")
		}
	}
}

func (s *dnsService) Wait() {
	for s._started {
		time.Sleep(1 * time.Second)
	}
	s.log.Info("DNSServer: Exit wait mode, server stopped")
}

// Query lookup answers for DNS Message.
func (s *dnsService) GetService() model.DNSService {
	return s
}

func sendPacket(conn *net.UDPConn, message dnsmessage.Message, addr net.UDPAddr) {
	packed, err := message.Pack()
	if err != nil {
		//log.Println(err)
		return
	}

	_, err = conn.WriteToUDP(packed, &addr)
	if err != nil {
		//log.Println(err)
	}
}

// New setups a DNSService, rwDirPath is read-writable directory path for storing dns records.
func New(rwDirPath string, logger log.Logger, forwarders []net.UDPAddr) model.DNSServer {
	return &dnsService{
		Store:      registry.NewStore(logger, rwDirPath, forwarders),
		Bucket:     store.NewRequestsCacheStore(),
		Answers:    store.NewAnswersCacheStore(),
		Forwarders: forwarders,
		log:        logger,
	}
}

// Start conveniently init every parts of DNS service.
func Start(rwDirPath string, ip string, port int, pipeIP string, pipePort int, pipeResponsePort int, logger log.Logger, forwarders []net.UDPAddr) model.DNSServer {
	s := New(rwDirPath, logger, forwarders)
	s.(*dnsService).Store.Load()
	go s.Listen(ip, port, pipeIP, pipePort, pipeResponsePort)
	return s
}

func (s *dnsService) Save(key string, resource dnsmessage.Resource, addr net.IPAddr, recordData string, old *dnsmessage.Resource) bool {
	ok := s.Store.Set(key, resource, addr.IP, recordData, old)
	go s.Store.Save()
	return ok
}

func (s *dnsService) SaveBulk(key string, resources []dnsmessage.Resource) {
	s.Store.Override(key, resources)
	go s.Store.Save()
}

func (s *dnsService) All() []model.Get {
	store := s.Store.Clone()
	var recs []model.Get
	for _, r := range store {
		for _, v := range r.Keys() {
			//TODO: Finish algorhitm
			r.Get(v)
			//body := v.Body.GoString()
			//i := strings.Index(body, "{")
			//recs = append(recs, model.Get{
			//	Host: v.Header.Name.String(),
			//	TTL:  v.Header.TTL,
			//	Type: v.Header.Type.String()[4:],
			//	Data: body[i : len(body)-1], // get content within "{" and "}"
			//})
		}
	}
	return recs
}

func (s *dnsService) Remove(key string, r *dnsmessage.Resource) bool {
	ok := s.Store.Remove(key, r)
	if ok {
		go s.Store.Save()
	}
	return ok
}
