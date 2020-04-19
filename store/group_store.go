// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package store

import (
	"errors"
	"fmt"
	rErrrors "github.com/hellgate75/rebind/rerrors"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"sync"
	"time"
)

type DNSRecord struct {
	NodeName string
	Type     string
	Addr     net.IP
	Data     string
	Resource dnsmessage.Resource
	TTL      uint32
	Created  time.Time
}

// This represent a single zone key store
// One server key can match more records ...
type GroupStore interface {
	// Gets all keys
	Keys() []string
	// Gets Records for a key
	Get(key string) ([]DNSRecord, rErrrors.Error)
	// Sets New Record for a key
	Set(key string, record DNSRecord) rErrrors.Error
	// Remove Records for a key
	Remove(key string) rErrrors.Error
	// Retrieve group name
	GetGroup() string
	// Retrieve references domains
	GetDomains() []string
	// Retrieve store name forwarders
	GetForwarders() []net.UDPAddr
	//Remove all records in the store
	ClearData()
}

type GroupStorePersistent struct {
	Store      map[string][]DNSRecord
	GroupName  string
	Domains    []string
	Forwarders []net.UDPAddr
}

type GroupStoreData struct {
	sync.RWMutex
	store      map[string][]DNSRecord
	GroupName  string
	Domains    []string
	Forwarders []net.UDPAddr
}

func (b *GroupStoreData) ClearData() {
	b.store = make(map[string][]DNSRecord)
}

func (b *GroupStoreData) PersistentData() GroupStorePersistent {
	return GroupStorePersistent{
		Store:      b.store,
		GroupName:  b.GroupName,
		Domains:    b.Domains,
		Forwarders: b.Forwarders,
	}
}

func (b *GroupStoreData) FromPersistentData(persistent GroupStorePersistent) {
	b.store = persistent.Store
	b.GroupName = persistent.GroupName
	b.Domains = persistent.Domains
	b.Forwarders = persistent.Forwarders
}

func (b *GroupStoreData) Keys() []string {
	var keys []string = make([]string, 0)
	for k, _ := range b.store {
		keys = append(keys, k)
	}
	return keys
}

func (b *GroupStoreData) GetGroup() string {
	return b.GroupName
}

func (b *GroupStoreData) GetDomains() []string {
	return b.Domains
}

func (b *GroupStoreData) GetForwarders() []net.UDPAddr {
	return b.Forwarders
}

func (b *GroupStoreData) Get(key string) ([]DNSRecord, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("Key "+key+" doesn't exist"), int64(20), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(21), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.RLock()
	val, ok := b.store[key]
	if ok {
		internalErr = nil
	}
	return val, internalErr
}

func (b *GroupStoreData) Set(key string, record DNSRecord) rErrrors.Error {
	var internalErr rErrrors.Error
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(23), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	if _, ok := b.store[key]; ok {
		b.store[key] = append(b.store[key], record)
	} else {
		b.store[key] = []DNSRecord{record}
	}
	return internalErr
}
func (b *GroupStoreData) Replace(key string, records []DNSRecord) rErrrors.Error {
	var internalErr rErrrors.Error
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(23), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	if _, ok := b.store[key]; ok {
		b.store[key] = records
	} else {
		b.store[key] = records
	}
	return internalErr
}

func (b *GroupStoreData) Remove(key string) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key "+key+" doesn't exist"), int64(24), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(25), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	_, ok := b.store[key]
	if ok {
		delete(b.store, key)
		internalErr = nil
	}
	return internalErr
}

// Generate New _zone Records Store
func NewGroupStore(groupName string, domains []string, forwarders []net.UDPAddr) GroupStore {
	return &GroupStoreData{
		GroupName:  groupName,
		Forwarders: forwarders,
		Domains:    domains,
		store:      make(map[string][]DNSRecord),
	}
}
