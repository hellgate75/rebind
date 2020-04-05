// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package store

import (
	"encoding/gob"
	"fmt"
	"github.com/hellgate75/rebind/utils"
	"io"
	"github.com/hellgate75/rebind/log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

const (
	_dnsStoreName   string = "_dnsStore"
	_dnsStoreBkName string = "_dnsStore_bk"
)
//TODO: Consider removing gof for file-save mode
//TODO: Link output to ZoneStore instead of the dnsmessage data

func init() {
	gob.Register(&dnsmessage.AResource{})
	gob.Register(&dnsmessage.NSResource{})
	gob.Register(&dnsmessage.CNAMEResource{})
	gob.Register(&dnsmessage.SOAResource{})
	gob.Register(&dnsmessage.PTRResource{})
	gob.Register(&dnsmessage.MXResource{})
	gob.Register(&dnsmessage.AAAAResource{})
	gob.Register(&dnsmessage.SRVResource{})
	gob.Register(&dnsmessage.TXTResource{})
	gob.Register(&dnsmessage.PTRResource{})
}

type DNSStore interface {
	Get(key string) ([]dnsmessage.Resource, bool)
	Set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool
	Override(key string, resources []dnsmessage.Resource)
	Remove(key string, r *dnsmessage.Resource) bool
	Save()
	Load()
	Clone() map[string]DNSEntry
}

// Create New Store with a logger and the rw config directory path
func NewDNSStore(logger log.Logger, rwDirPath string) DNSStore {
	return &_dnsStore{
		log: logger,
		store: make(map[string]DNSEntry),
		rwDirPath: rwDirPath,
	}
}

type _dnsStore struct {
	sync.RWMutex
	store     map[string]DNSEntry
	rwDirPath string
	log       log.Logger
}

type DNSEntry struct {
	Resources []dnsmessage.Resource
	TTL       uint32
	Created   int64
}

func (s *_dnsStore) Get(key string) ([]dnsmessage.Resource, bool) {
	s.RLock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Get::Runtime error: %v", r))
			}
		}
		s.RUnlock()
	}()
	e, ok := s.store[key]
	now := time.Now().Unix()
	if e.TTL > 1 && (e.Created+int64(e.TTL) < now) {
		s.Remove(key, nil)
		return nil, false
	}
	return e.Resources, ok
}

func (s *_dnsStore) Set(key string, resource dnsmessage.Resource, old *dnsmessage.Resource) bool {
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
	if _, ok := s.store[key]; ok {
		if old != nil {
			for i, rec := range s.store[key].Resources {
				if utils.RToString(rec) == utils.RToString(*old) {
					s.store[key].Resources[i] = resource
					changed = true
					break
				}
			}
		} else {
			e := s.store[key]
			e.Resources = append(e.Resources, resource)
			s.store[key] = e
			changed = true
		}
	} else {
		e := DNSEntry{
			Resources: []dnsmessage.Resource{resource},
			TTL:       resource.Header.TTL,
			Created:   time.Now().Unix(),
		}
		s.store[key] = e
		changed = true
	}

	return changed
}

func (s *_dnsStore) Override(key string, resources []dnsmessage.Resource) {
	s.Lock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Override::Runtime error: %v", r))
			}
		}
		s.Unlock()
	}()
	e := DNSEntry{
		Resources: resources,
		Created:   time.Now().Unix(),
	}
	if len(resources) > 0 {
		e.TTL = resources[0].Header.TTL
	}
	s.store[key] = e
}

func (s *_dnsStore) Remove(key string, r *dnsmessage.Resource) bool {
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
	if r == nil {
		_, ok = s.store[key]
		delete(s.store, key)
	} else {
		if _, ok = s.store[key]; ok {
			for i, rec := range s.store[key].Resources {
				if utils.RToString(rec) == utils.RToString(*r) {
					e := s.store[key]
					copy(e.Resources[i:], e.Resources[i+1:])
					var blank dnsmessage.Resource
					e.Resources[len(e.Resources)-1] = blank
					e.Resources = e.Resources[:len(e.Resources)-1]
					s.store[key] = e
					ok = true
					break
				}
			}
		}
	}
	return ok
}

func (s *_dnsStore) Save() {
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Save::Runtime error: %v", r))
			}
		}
	}()
	bk, err := os.OpenFile(filepath.Join(s.rwDirPath, _dnsStoreBkName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		if s.log != nil {
			s.log.Errorf("err open _dnsStore bak file %v",err)
		}
		return
	}
	defer bk.Close()

	dst, err := os.OpenFile(filepath.Join(s.rwDirPath, _dnsStoreName), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		if s.log != nil {
			s.log.Errorf("err open _dnsStore file %v",err)
		}
		return
	}
	defer dst.Close()

	// backing up current _dnsStore
	_, err = io.Copy(bk, dst)
	if err != nil {
		if s.log != nil {
			s.log.Errorf("err copy _dnsStore file %v",err)
		}
		return
	}

	enc := gob.NewEncoder(dst)
	book := s.Clone()
	if err = enc.Encode(book); err != nil {
		// main _dnsStore file is corrupted
		if s.log != nil {
			s.log.Fatal(err)
		}
	}
}

func (s *_dnsStore) Load() {
	fReader, err := os.Open(filepath.Join(s.rwDirPath, _dnsStoreName))
	if err != nil {
		if s.log != nil {
			s.log.Errorf("err load _dnsStore file %v maybe first start,please ignore",err)
		}
		return
	}
	defer fReader.Close()

	dec := gob.NewDecoder(fReader)

	s.Lock()
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Load::Runtime error: %v", r))
			}
		}
		s.Unlock()
	}()

	if err = dec.Decode(&s.store); err != nil {
		if s.log != nil {
			s.log.Fatalf("err decode _dnsStore file %v",err)
		}
	}
}

func (s *_dnsStore) Clone() map[string]DNSEntry {
	defer func() {
		if r := recover(); r != nil {
			if s.log != nil {
				s.log.Errorf(fmt.Sprintf("Store.Clone::Runtime error: %v", r))
			}
		}
		s.RUnlock()
	}()
	cp := make(map[string]DNSEntry)
	s.RLock()
	for k, v := range s.store {
		cp[k] = v
	}
	return cp
}
