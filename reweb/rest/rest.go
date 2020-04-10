// Copyright 2020 Re-Bind Author (Fabrizio Torelli). All rights reserved.
// Use of this source code is governed by a LGPL-style
// license that can be found in the LICENSE file.

package rest

import (
	"encoding/json"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/net"
	"net/http"
)

// RestServer will do CRUD on DNS records
type RestServer interface {
	Create() http.HandlerFunc
	Read() http.HandlerFunc
	Update() http.HandlerFunc
	Delete() http.HandlerFunc
	GetService() RestService
}

type restServer struct {
	restService RestService
}

func (r *restServer) Create() http.HandlerFunc {
	return r.restService.Create
}
func (r *restServer) Read() http.HandlerFunc {
	return r.restService.Read
}
func (r *restServer) Update() http.HandlerFunc {
	return r.restService.Update
}
func (r *restServer) Delete() http.HandlerFunc {
	return r.restService.Delete
}
func (r *restServer) GetService() RestService {
	return r.restService
}

func NewRestServer(rs RestService) RestServer {
	return &restServer{
		restService: rs,
	}
}

type RestService interface {
	Create(w http.ResponseWriter, r *http.Request)
	Read(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

func NewRestService(pipe net.NetPipe) RestService {
	return &restService{
		Pipe: pipe,
	}
}

// restService is an implementation of RestService interface.
type restService struct {
	Pipe net.NetPipe
}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *restService) Create(w http.ResponseWriter, r *http.Request) {
	//TODO: Consider multiple paths
	var req model.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//resource, err := utils.ToResource(req)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}

	//s.Dn.Save(utils.NtToString(resource.Header.Name, resource.Header.Type), resource, nil)
	w.WriteHeader(http.StatusCreated)
}

// Read is HTTP handler of GET model.Request.
// Use for reading existed records on DNS server.
func (s *restService) Read(w http.ResponseWriter, r *http.Request) {
	//TODO: Consider multiple paths
	//json.NewEncoder(w).Encode(s.Dn.All())
}

// Update is HTTP handler of PUT model.Request.
// Use for updating existed records on DNS server.
func (s *restService) Update(w http.ResponseWriter, r *http.Request) {
	//TODO: Consider multiple paths
	var req model.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//oldReq := model.Request{Host: req.Host, Type: req.Type, Data: req.OldData}
	//old, err := utils.ToResource(oldReq)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//
	//resource, err := utils.ToResource(req)
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}

	//ok := s.Dn.Save(utils.NtToString(resource.Header.Name, resource.Header.Type), resource, &old)
	//if ok {
	//	w.WriteHeader(http.StatusOK)
	//	return
	//}

	http.Error(w, "", http.StatusNotFound)
}

// Delete is HTTP handler of DELETE model.Request.
// Use for removing records on DNS server.
func (s *restService) Delete(w http.ResponseWriter, r *http.Request) {
	//TODO: Consider multiple paths
	var req model.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ok := false
	//h, err := utils.ToResourceHeader(req.Host, req.Type)
	//if err == nil {
	//	ok = s.Dn.Remove(utils.NtToString(h.Name, h.Type), nil)
	//}

	if ok {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "", http.StatusNotFound)
}
