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
	"github.com/hellgate75/rebind/store"
	"github.com/hellgate75/rebind/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
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
	Store      store.Store
	Bucket     store.ZoneStore
	Forwarders []net.UDPAddr
	log		   log.Logger
	_started	bool
}



// Listen starts a DNS server on port 53
func (s *dnsService) Listen(ipAddress string, port int) error {
	var err error
	ipTokens:=strings.Split(ipAddress, ".")
	if len(ipTokens) < 4 {
		return errs.New(fmt.Sprintf("DNSServer: Invalid ip address: %s", ipAddress))
	}
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
	s.Conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: ip,Port: port})
	if err != nil {
		s.log.Fatal(err)
	}
	if s.Conn == nil {
		return errs.New(fmt.Sprintf("Nil Connection from dialing %s:%v", ip.String(), port));
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
	if ! p.Message.Header.Response {
		s.log.Debugf("Ip: %v", p.Addr.IP.String())
		s.log.Debugf("Port: %v", p.Addr.Port)
		zone := p.Addr.Zone
		if zone == "" {
			zone="default"
		}
		s.log.Debugf("Zone: %v (req:%s)", zone, p.Addr.Zone)
		s.log.Debugf("Response: %v", p.Message.Header.Response)
		jsonText, _ := json.Marshal(p)
		s.log.Debugf("PACKET=%s", jsonText)
		s.log.Debugf("pKey: %s", utils.PToString(p))
		q1 := p.Message.Questions[0]
		s.log.Debugf("qKey: %s", utils.QToString(q1))
		questionType := q1.Type.String()[4:]
		s.log.Debugf("qType: %s", questionType)

		//Recover by zone, the by question
		//TODO Fix it ( insert first cache check ....)

		//pKey := utils.PToString(p)
		if store, zId, err := s.Bucket.GetByZone(zone); err == nil {
			s.log.Debugf("Zone: <%s> discovered with id: %s", zone, zId)
			if store != nil {
				var added bool = false
				for _, qX := range p.Message.Questions {
					qKeyX := utils.QToString(qX)
					if len(qKeyX) > 0 {
						if qKeyX[len(qKeyX)-1] == '.' {
							qKeyX = qKeyX[:len(qKeyX)-1]
						}
					}
					if recs, err := store.Get(qKeyX); err==nil && err.Error() == nil {
						for _, rec := range recs {
							if rec.Type == questionType {
								for _, addr := range rec.Address {
									added = true
									sendPacket(s.Conn, p.Message, addr)
									//Insert in the store by zone, then split by
									// 1. NAME and 2. Query cache storage of answer
								}
							}
						}
					}
				}
				if ! added {
					for _, fwdr := range store.GetForwarders() {
						sendPacket(s.Conn, p.Message, fwdr)
					}
				}
			}
		}

//		if addrs, ok := s.Bucket.Get(pKey); ok {
//			for _, addr := range addrs {
//				go sendPacket(s.Conn, p.Message, addr)
//			}
//			s.Bucket.Remove(pKey)
//			go s.SaveBulk(utils.QToString(p.Message.Questions[0]), p.Message.Answers)
//		}
		return
	}

	// was checked before entering this routine
	q := p.Message.Questions[0]


	// answer the question
	val, ok := s.Store.Get(utils.QToString(q))

	if ok {
		p.Message.Answers = append(p.Message.Answers, val...)
		go sendPacket(s.Conn, p.Message, p.Addr)
	} else {
		// forwarding
		for i := 0; i < len(s.Forwarders); i++ {
			//TODO: FIXIT
			//s.Bucket.Set(utils.PToString(p), p.Addr)
			//go sendPacket(s.Conn, p.Message, s.Forwarders[i])
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
		Store:       store.NewStore(logger, rwDirPath),
		Bucket:      store.NewZoneStore(),
		Forwarders: forwarders,
		log: logger,
	}
}

// Start conveniently init every parts of DNS service.
func Start(rwDirPath string, ip string, port int, logger log.Logger, forwarders []net.UDPAddr) model.DNSServer {
	s := New(rwDirPath, logger, forwarders)
	s.(*dnsService).Store.Load()
	go s.Listen(ip, port)
	return s
}

func (s *dnsService) Save(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	ok := s.Store.Set(key, resource, old)
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
		for _, v := range r.Resources {
			body := v.Body.GoString()
			i := strings.Index(body, "{")
			recs = append(recs, model.Get{
				Host: v.Header.Name.String(),
				TTL:  v.Header.TTL,
				Type: v.Header.Type.String()[4:],
				Data: body[i : len(body)-1], // get content within "{" and "}"
			})
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
