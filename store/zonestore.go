// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package store

import (
	"fmt"
	rErrrors "github.com/hellgate75/rebind/errors"
	"net"
	"sync"
	"errors"
)

// Defines the behaviour od a multi-zone tenant storage
type ZoneStore interface {
	// Retrieves a Zone storage by zone key
	Get(key string) (DNSRecordStore, rErrrors.Error)
	// Retrieves a Zone storage and zone id by zone name
	GetByZone(zone string) (DNSRecordStore, string, rErrrors.Error)
	// Creates a New Zone storage for a key
	Create(key string, bindName string, subNames []string, forwarders []net.UDPAddr) rErrrors.Error
	// Replaces a Zone storage by key
	Set(key string, addr DNSRecordStore) rErrrors.Error
	// Remove entire sone
	Remove(key string) rErrrors.Error
}

type _zoneStore struct {
	sync.RWMutex
	store map[string]DNSRecordStore
}

func (b *_zoneStore) Get(key string) (DNSRecordStore, rErrrors.Error) {
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


func (b *_zoneStore) GetByZone(zone string) (DNSRecordStore, string, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("Zone " + zone + " doesn't exist"), int64(29), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(30), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.RLock()
	var key string
	var val DNSRecordStore
	for k, store := range b.store {
		if store.GetZone() == zone {
			key = k
			val = store
			internalErr = nil
			break
		}
	}
	return val, key, internalErr
}

func (b *_zoneStore) Create(key string, bindName string, subNames []string, forwarders []net.UDPAddr) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key " + key + " already in use"), int64(22), rErrrors.StoreCreateErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(23), rErrrors.StoreCreateErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	if _, ok := b.store[key]; !ok {
		b.store[key] = NewDNSRecordStore(bindName, subNames, forwarders)
		internalErr=nil
	}
	return internalErr
}

func (b *_zoneStore) Set(key string, store DNSRecordStore) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key " + key + " already in use"), int64(22), rErrrors.StoreCreateErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(23), rErrrors.StoreCreateErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	if _, ok := b.store[key]; !ok {
		b.store[key] = store
		internalErr=nil
	}
	return internalErr
}

func (b *_zoneStore) Remove(key string) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key " + key + " doesn't exist"), int64(24), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(25), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	_, ok := b.store[key]
	if ok {
		internalErr=nil
		delete(b.store, key)
	}
	return internalErr
}

func NewZoneStore() ZoneStore {
	return &_zoneStore{
		store: make(map[string]DNSRecordStore),
	}
}
