package v1

import (
	"encoding/json"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"net/http"
)

// DnsGroupService is an implementation of RestService interface.
type DnsGroupService struct {
	Pipe  net.NetPipe
	Store registry.Store
	Log   log.Logger
}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *DnsGroupService) Create(w http.ResponseWriter, r *http.Request) {
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
func (s *DnsGroupService) Read(w http.ResponseWriter, r *http.Request) {
	//TODO: Consider multiple paths
	//json.NewEncoder(w).Encode(s.Dn.All())
}

// Update is HTTP handler of PUT model.Request.
// Use for updating existed records on DNS server.
func (s *DnsGroupService) Update(w http.ResponseWriter, r *http.Request) {
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
func (s *DnsGroupService) Delete(w http.ResponseWriter, r *http.Request) {
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
