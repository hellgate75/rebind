// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package registry

import (
	"fmt"
	"github.com/hellgate75/rebind/data"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/store"
	"github.com/hellgate75/rebind/utils"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"strings"
	"sync"
	"time"
)

//TODO: Consider removing gof for file-save mode
//TODO: Link output to GroupsStore instead of the dnsmessage _store

type Store interface {
	Get(hostname string) ([]dnsmessage.Resource, []net.UDPAddr, bool)
	Set(hostname string, resource dnsmessage.Resource, addr net.IP, recordData string, old *dnsmessage.Resource) bool
	Override(hostname string, resources []dnsmessage.Resource)
	Remove(hostname string, r *dnsmessage.Resource) bool
	Save()
	Load()
	Clone() map[string]store.GroupStoreData
	GetGroupBucket() *data.GroupsBucket
}

// Create New Store with a logger and the rw config directory path
func NewStore(logger log.Logger, rwDirPath string,
	forwarders []net.UDPAddr) Store {
	return &_store{
		log:        logger,
		store:      data.NewGroupsBucket(rwDirPath, logger),
		cache:      store.NewGroupsStore(),
		forwarders: forwarders,
		rwDirPath:  rwDirPath,
	}
}

type _store struct {
	sync.RWMutex
	store      data.GroupsBucket
	cache      store.GroupsStore
	rwDirPath  string
	log        log.Logger
	forwarders []net.UDPAddr
}

func (s *_store) GetGroupBucket() *data.GroupsBucket {
	return &s.store
}

func (s *_store) Get(hostname string) ([]dnsmessage.Resource, []net.UDPAddr, bool) {
	var ok = true
	s.RLock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Get::Runtime error: %v", r))
				ok = false
			}
		}
		s.RUnlock()
	}()
	domains := utils.SplitDomainsFromHostname(hostname)
	var res = make([]dnsmessage.Resource, 0)
	var fwd = make([]net.UDPAddr, 0)

	var groups = make(map[string]data.Group, 0)
	for _, domain := range domains {
		gr, err := s.store.GetGroupsByDomain(domain)
		if err == nil || len(gr) == 0 {
			s.log.Error("Store.Get:: Unable to get Store from domain/sub-domain: %s, due to Error: %v", domain, err)
			ok = false
			continue
		}
		for _, g := range gr {
			groups[g.Name] = g
		}
	}
	for _, g := range groups {
		fwd = append(fwd, g.Forwarders...)
		sg, err := s.store.GetGroupStore(g)
		if err != nil {
			s.log.Error("Store.Get:: Unable to get Store from file, due to Error: %v", err)
			continue
		}
		recs, errR := sg.Get(hostname)
		if errR != nil {
			s.log.Error("Store.Get:: Unable to discover for host: %s, due to Error: %v", hostname, errR.Error())
			continue
		}
		for _, r := range recs {
			res = append(res, r.Resource)
		}
	}
	return res, fwd, ok
}

func (s *_store) GetGroupsFromHost(hostname string) ([]data.Group, error) {
	return nil, nil
	server := strings.Split(hostname, ".")[0]
	domains := utils.SplitDomainsFromHostname(hostname)
	if len(domains) == 1 && (domains[0] == "" || utils.IsDefaultGroupDomain(domains[0])) {
		hostname = server
	}
	var groups = make([]data.Group, 0)
	for _, domain := range domains {
		gr, err := s.store.GetGroupsByDomain(domain)
		if err == nil || len(gr) == 0 {
			s.log.Error("Store.GetGroupsFromHost:: Unable to get Store from domain/sub-domain: %s, due to Error: %v", domain, err)
			continue
		}
		groups = append(groups, gr...)
	}
	return groups, nil
}

func (s *_store) Set(hostname string, resource dnsmessage.Resource, addr net.IP, recordData string, old *dnsmessage.Resource) bool {
	ok := true
	changed := false
	s.Lock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Set::Runtime error: %v", r))
				ok = false
			}
		}
		s.Unlock()
	}()
	server := strings.Split(hostname, ".")[0]
	domains := utils.SplitDomainsFromHostname(hostname)
	if len(domains) == 1 && (domains[0] == "" || utils.IsDefaultGroupDomain(domains[0])) {
		hostname = server
	}

	var fwd = make([]net.UDPAddr, 0)
	var groups = make(map[string]data.Group, 0)
	for _, domain := range domains {
		gr, err := s.store.GetGroupsByDomain(domain)
		if err == nil || len(gr) == 0 {
			s.log.Error("Store.Set:: Unable to get Store from domain/sub-domain: %s, due to Error: %v", domain, err)
			ok = false
			continue
		}
		for _, g := range gr {
			groups[g.Name] = g
		}
	}
	var change = false
	for _, g := range groups {
		fwd = append(fwd, g.Forwarders...)
		sg, err := s.store.GetGroupStore(g)
		if err != nil {
			s.log.Error("Store.Set:: Unable to get Store from file, due to Error: %v", err)
			continue
		}
		sg.Forwarders = append(sg.Forwarders, fwd...)
		sg.Forwarders = utils.RemoveDuplicatesInUpdAddrList(sg.Forwarders)
		recs, errR := sg.Get(hostname)
		if errR != nil || len(recs) == 0 {
			sg.Set(hostname, store.DNSRecord{
				Resource: resource,
				TTL:      resource.Header.TTL,
				Created:  time.Now(),
				Data:     recordData,
				Addr:     addr,
				Type:     resource.Header.Type.String(),
				NodeName: hostname,
			})
		} else {
			sg.Set(hostname, store.DNSRecord{
				Resource: resource,
				TTL:      resource.Header.TTL,
				Created:  time.Now(),
				Data:     recordData,
				Addr:     addr,
				Type:     resource.Header.Type.String(),
				NodeName: hostname,
			})
		}
		g, err = s.store.SaveGroup(sg, g)
		if s.store.UpdateExistingGroup(g) {
			change = true
		}
	}
	if change {
		s.store.SaveMeta()
	}
	return ok && changed
}

func (s *_store) Override(hostname string, resources []dnsmessage.Resource) {
	s.Lock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Override::Runtime error: %v", r))
			}
		}
		s.Unlock()
	}()
	if len(resources) == 0 {
		s.log.Error("Store.Override:: error: Resource are empty, please remove instead of update empty")
		return
	}
	var dnsRecords = make([]store.DNSRecord, 0)
	for _, resource := range resources {
		var addr net.IP = nil
		rec := store.DNSRecord{
			NodeName: hostname,
			Type:     resource.Header.Type.String(),
			Resource: resource,
			TTL:      resource.Header.TTL,
			Created:  time.Now(),
			Data:     resource.Body.GoString(),
			Addr:     addr,
		}

		dnsRecords = append(dnsRecords, rec)
	}
	server := strings.Split(hostname, ".")[0]
	domains := utils.SplitDomainsFromHostname(hostname)
	if len(domains) == 1 && (domains[0] == "" || utils.IsDefaultGroupDomain(domains[0])) {
		hostname = server
	}

	var groups = make([]data.Group, 0)
	for _, domain := range domains {
		gr, err := s.store.GetGroupsByDomain(domain)
		if err == nil || len(gr) == 0 {
			s.log.Error("Store.Override:: Unable to get Store from domain/sub-domain: %s, due to Error: %v", domain, err)
			continue
		}
		groups = append(groups, gr...)
	}
	var change = false
	var fwd = make([]net.UDPAddr, 0)
	for _, g := range groups {
		fwd = append(fwd, g.Forwarders...)
		sg, err := s.store.GetGroupStore(g)
		if err != nil {
			s.log.Error("Store.Override:: Unable to get Store from file, due to Error: %v", err)
			continue
		}
		sg.Forwarders = append(sg.Forwarders, fwd...)
		sg.Forwarders = utils.RemoveDuplicatesInUpdAddrList(sg.Forwarders)
		errR := sg.Replace(hostname, dnsRecords)
		if errR == nil {
			g, err = s.store.SaveGroup(sg, g)
			if s.store.UpdateExistingGroup(g) {
				change = true
			}
		}
	}
	if change {
		s.store.SaveMeta()
	}
}

func (s *_store) Remove(hostname string, r *dnsmessage.Resource) bool {
	ok := false
	s.Lock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Remove::Runtime error: %v", r))
			}
		}
		s.Unlock()
	}()
	//TODO: Here code for removing element in all marching groups
	return ok
}

func (s *_store) Save() {
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Save::Runtime error: %v", r))
			}
		}
	}()
	err := s.store.SaveMeta()
	if err != nil {
		if s.log != nil {
			s.log.Errorf("Store.Save:: saving meta -> Error: %v", err)
		}
	}
}

func (s *_store) Load() {
	err := s.store.Load(s.forwarders)
	if err != nil {
		if s.log != nil {
			s.log.Errorf("Store.Load:: loading meta -> err Error: %v maybe first start,please ignore", err)
		}
		return
	}
	if s.log != nil {
		s.log.Info("Store.Load:: loading meta complete!!")
	}
}

func (s *_store) Clone() map[string]store.GroupStoreData {
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Clone::Runtime error: %v", r))
			}
		}
		s.RUnlock()
	}()
	cp := make(map[string]store.GroupStoreData)
	s.RLock()
	for _, key := range s.store.Keys() {
		group, err := s.store.GetGroupByName(s.store.ConvertToGroupLikeKey(key))
		if err == nil {
			groupStore, err := s.store.GetGroupStore(group)
			if err == nil {
				cp[key] = groupStore
			}
		}
	}
	return cp
}
