// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package registry

import (
	"fmt"
	"github.com/hellgate75/rebind/data"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/store"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"sync"
)

//TODO: Consider removing gof for file-save mode
//TODO: Link output to GroupsStore instead of the dnsmessage _store

type Store interface {
	Get(key string) ([]dnsmessage.Resource, bool)
	Set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool
	Override(key string, resources []dnsmessage.Resource)
	Remove(key string, r *dnsmessage.Resource) bool
	Save()
	Load()
	Clone() map[string]store.GroupStore
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

func (s *_store) Get(key string) ([]dnsmessage.Resource, bool) {
	var out []dnsmessage.Resource = make([]dnsmessage.Resource, 0)
	var ok bool = false
	s.RLock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Get::Runtime error: %v", r))
			}
		}
		s.RUnlock()
	}()
	// TODO: Here code to get all matching resources
	return out, ok
}

func (s *_store) Set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
	changed := false
	s.Lock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Set::Runtime error: %v", r))
			}
		}
		s.Unlock()
	}()
	// TODO: Here code for set/update (accumulate records...)
	return changed
}

func (s *_store) Override(key string, resources []dnsmessage.Resource) {
	s.Lock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Override::Runtime error: %v", r))
			}
		}
		s.Unlock()
	}()
	//TODO: Here code for overriding values...
}

func (s *_store) Remove(key string, r *dnsmessage.Resource) bool {
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

func (s *_store) Clone() map[string]store.GroupStore {
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Clone::Runtime error: %v", r))
			}
		}
		s.RUnlock()
	}()
	cp := make(map[string]store.GroupStore)
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
