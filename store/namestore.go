// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package store

import (
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	rErrrors "github.com/hellgate75/rebind/errors"
	"sync"
	"errors"
)

type DNSRecord struct {
	Address 	[]net.UDPAddr
	NodeName	string
	Type		string
	Resources 	[]dnsmessage.Resource
	TTL       	uint32
	Created   	int64
}

// This represent a single zone key store
// One server key can match more records ...
type DNSRecordStore interface {
	// Gets Records for a key
	Get(key string) ([]DNSRecord, rErrrors.Error)
	// Sets New Record for a key
	Set(key string, record DNSRecord) rErrrors.Error
	// Remove Records for a key
	Remove(key string) rErrrors.Error
	// Retrieve store name
	GetZone() string
	// Retrieve store name forwarders
	GetForwarders() []net.UDPAddr
}

type dnsRecordStore struct {
	sync.RWMutex
	store 			map[string][]DNSRecord
	BindName 		string
	SubNames 		[]string
	Forwarders 		[]net.UDPAddr
}

func (b *dnsRecordStore) GetZone() string {
	return b.BindName
}

func (b *dnsRecordStore) GetForwarders() []net.UDPAddr {
	return b.Forwarders
}


func (b *dnsRecordStore) Get(key string) ([]DNSRecord, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("Key " + key + " doesn't exist"), int64(20), rErrrors.StoreProcessErrorType)
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

func (b *dnsRecordStore) Set(key string, record DNSRecord) rErrrors.Error {
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

func (b *dnsRecordStore) Remove(key string) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key " + key + " doesn't exist"), int64(24), rErrrors.StoreProcessErrorType)
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

// Generate New Zone Records Store
func NewDNSRecordStore(bindName string, subNames []string, forwarders []net.UDPAddr) DNSRecordStore {
	return &dnsRecordStore{
		BindName: bindName,
		Forwarders: forwarders,
		SubNames: subNames,
		store: make(map[string][]DNSRecord),
	}
}