package v1

import (
	"fmt"
	"github.com/hellgate75/rebind/data"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/utils"
	net2 "net"
	"net/http"
)

// DnsGroupsService is an implementation of RestService interface.
type DnsGroupsService struct {
	Pipe  net.NetPipe
	Store registry.Store
	Log   log.Logger
}

type DnsGroupsResponse struct {
	Groups []data.Group `yaml:"groups,omitempty" json:"groups,omitempty" xml:"groups,omitempty"`
}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *DnsGroupsService) Create(w http.ResponseWriter, r *http.Request) {
	var req model.GroupRequest
	err := utils.RestParseRequest(w, r, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := model.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Error: %v", err),
			Data:    DnsGroupsResponse{Groups: []data.Group{}},
		}
		s.Log.Errorf("Error decoding groups request, Error: %v", err)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding groups response, Error: %v", err)
		}
		return
	}
	s.Store.Load()
	if req.Name == "" || req.Domains == nil {
		w.WriteHeader(http.StatusBadRequest)
		response := model.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Group: %s has invalid or missing name / domains, please use proper api", req.Name),
			Data:    DnsGroupsResponse{Groups: []data.Group{}},
		}
		s.Log.Errorf("Error: Group: %s has invalid or missing name / domains, please use proper api", req.Name)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding groups response, Error: %v", err)
		}
		return
	}
	if s.Store.GetGroupBucket().Contains(req.Name) {
		w.WriteHeader(http.StatusConflict)
		response := model.Response{
			Status:  http.StatusConflict,
			Message: fmt.Sprintf("Group: %s already exists, please modify with proper api", req.Name),
			Data:    DnsGroupsResponse{Groups: []data.Group{}},
		}
		s.Log.Errorf("Error: Group: %s already exists, please modify with proper api", req.Name)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding groups response, Error: %v", err)
		}
		return
	}
	if req.Domains == nil {
		req.Domains = []string{}
	}
	if req.Forwarders == nil {
		req.Forwarders = []net2.UDPAddr{}
	}
	group, _, err := s.Store.GetGroupBucket().CreateAndPersistGroupAndStore(req.Name, req.Domains, req.Forwarders)
	if err == nil {
		s.Store.Save()
	}
	if err != nil {
		w.WriteHeader(http.StatusLocked)
		response := model.Response{
			Status:  http.StatusLocked,
			Message: fmt.Sprintf("Error: %v", err),
			Data:    DnsGroupsResponse{Groups: []data.Group{}},
		}
		s.Log.Errorf("Error creating new group, Error: %v", err)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding groups response, Error: %v", err)
		}
		return
	}

	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data:    DnsGroupsResponse{Groups: []data.Group{*group}},
	}
	w.WriteHeader(http.StatusCreated)
	err = utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding groups response, Error: %v", err)
	}
}

// Read is HTTP handler of GET model.Request.
// Use for reading existed records on DNS server.
func (s *DnsGroupsService) Read(w http.ResponseWriter, r *http.Request) {
	s.Store.Load()
	groups := s.Store.GetGroupBucket().ListGroups()
	var list = make([]data.Group, 0)
	for _, g := range groups {
		list = append(list, *g)
	}
	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data:    DnsGroupsResponse{Groups: list},
	}
	w.WriteHeader(http.StatusOK)
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding group response, Error: %v", err)
	}
}

// Update is HTTP handler of PUT model.Request.
// Use for updating existed records on DNS server.
func (s *DnsGroupsService) Update(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: "Not allowed on dns groups",
		Data: DnsGroupsResponse{
			Groups: []data.Group{},
		},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encording response: %v", err)
	}
}

// Delete is HTTP handler of DELETE model.Request.
// Use for removing records on DNS server.
func (s *DnsGroupsService) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: "Not allowed on dns groups",
		Data: DnsGroupsResponse{
			Groups: []data.Group{},
		},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encording response: %v", err)
	}
}
