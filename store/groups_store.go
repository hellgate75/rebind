// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package store

import (
	"errors"
	"fmt"
	rErrrors "github.com/hellgate75/rebind/errors"
	"github.com/hellgate75/rebind/utils"
	"sync"
	"time"
)

var DEFAULT_GROUP_CACHE_TIME_TO_LIVE time.Duration = 5 * time.Minute

// Defines the behaviour od a multi-zone tenant storage
type GroupsStore interface {
	// Retrieves a _zone storage by zone key
	Get(key string) (GroupStore, rErrrors.Error)
	// Retrieves a group storage and zone id by zone name
	GetByGroup(zone string) (GroupStore, string, rErrrors.Error)
	// Retrieves a group storage list matching on a given domain
	GetFirstByDomain(domain string) (GroupStore, string, rErrrors.Error)
	// Replaces a _zone storage by key
	Set(key string, addr GroupStore) rErrrors.Error
	// Remove entire sone
	Remove(key string) rErrrors.Error
	// Get Zone Keys
	Keys() []string
	// Removes expired groups
	Trim() error
}

type GroupStoreMeta struct {
	Store   GroupStore
	Created time.Time
	TTL     time.Duration
}

func (gsm GroupStoreMeta) IsValid() bool {
	return time.Now().Sub(gsm.Created).Nanoseconds() < gsm.TTL.Nanoseconds()
}

type GroupsStoreData struct {
	sync.RWMutex
	store map[string]GroupStoreMeta
}

func (b *GroupsStoreData) Keys() []string {
	var out []string = make([]string, 0)
	for k, v := range b.store {
		if v.IsValid() {
			out = append(out, k)
		}
	}
	return out
}

func (b *GroupsStoreData) Get(key string) (GroupStore, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("Key "+key+" doesn't exist"), int64(20), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(21), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.RLock()
	val, ok := b.store[key]
	var gStore GroupStore = nil
	if ok && val.IsValid() {
		internalErr = nil
		gStore = val.Store
	}
	return gStore, internalErr
}

func (b *GroupsStoreData) GetByGroup(group string) (GroupStore, string, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("_group "+group+" doesn't exist"), int64(29), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(30), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.RLock()
	var key string
	var val GroupStore
	for k, store := range b.store {
		if store.IsValid() && store.Store.GetGroup() == group {
			key = k
			val = store.Store
			internalErr = nil
			break
		}
	}
	return val, key, internalErr
}

func (b *GroupsStoreData) GetFirstByDomain(domain string) (GroupStore, string, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("_domain "+domain+" doesn't exist"), int64(29), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(30), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.RLock()
	var key string
	var val GroupStore
	for k, store := range b.store {
		isDefault := utils.IsDefaultGroupDomain(domain)
		//for _, group := range i.Groups {
		//	if ( isDefault && group.Name == utils.DEFAULT_GROUP_NAME ) ||
		//		( !isDefault && utils.StringsListContainItem(domain, group.Domains, false) ) {
		if store.IsValid() &&
			((isDefault && store.Store.GetGroup() == utils.DEFAULT_GROUP_NAME) ||
				utils.StringsListContainItem(domain, store.Store.GetDomains(), true)) {
			key = k
			val = store.Store
			internalErr = nil
			return val, key, internalErr
		}
	}
	return nil, "", internalErr
}

func (b *GroupsStoreData) GetAllByDomain(domain string) ([]GroupStore, rErrrors.Error) {
	internalErr := rErrrors.New(errors.New("_domain "+domain+" doesn't exist"), int64(29), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(30), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	b.RLock()
	var val []GroupStore = make([]GroupStore, 0)
	for _, store := range b.store {
		if store.IsValid() && ((len(store.Store.GetDomains()) == 0 && domain == "") ||
			utils.StringsListContainItem(domain, store.Store.GetDomains(), true)) {
			val = append(val, store.Store)
			internalErr = nil
		}
	}
	return val, internalErr
}

func (b *GroupsStoreData) Set(key string, store GroupStore) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key "+key+" already in use"), int64(22), rErrrors.StoreCreateErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(23), rErrrors.StoreCreateErrorType)
		}
		b.Unlock()
	}()
	b.Lock()
	if oldStroreMeta, ok := b.store[key]; ok && oldStroreMeta.IsValid() {
		oldStroreMeta.Created = time.Now()
		oldStroreMeta.Store = store
		b.store[key] = oldStroreMeta
	} else {
		b.store[key] = GroupStoreMeta{
			Store:   store,
			Created: time.Now(),
			TTL:     DEFAULT_GROUP_CACHE_TIME_TO_LIVE,
		}

		internalErr = nil
	}
	return internalErr
}

func (b *GroupsStoreData) Remove(key string) rErrrors.Error {
	internalErr := rErrrors.New(errors.New("Key "+key+" doesn't exist"), int64(24), rErrrors.StoreProcessErrorType)
	defer func() {
		if r := recover(); r != nil {
			internalErr = rErrrors.New(errors.New(fmt.Sprintf("Runtime error: %s", r)), int64(25), rErrrors.StoreProcessErrorType)
		}
		b.Unlock()
	}()
	_, ok := b.store[key]
	if ok {
		internalErr = nil
		delete(b.store, key)
	}
	return internalErr
}

func (b *GroupsStoreData) Trim() error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	for key, value := range b.store {
		if !value.IsValid() {
			delete(b.store, key)
		}
	}
	return err
}

func NewGroupsStore() GroupsStore {
	return &GroupsStoreData{
		store: make(map[string]GroupStoreMeta),
	}
}
