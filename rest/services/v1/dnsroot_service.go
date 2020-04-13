package v1

import (
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/utils"
	"net/http"
)

type DnsRootResponse struct {
	Groups []string `yaml:"groups,omitempty" json:"groups,omitempty" xml:"groups,omitempty"`
}

// DnsRootService is an implementation of RestService interface.
type DnsRootService struct {
	Pipe    net.NetPipe
	Store   registry.Store
	Log     log.Logger
	BaseUrl string
}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *DnsRootService) Create(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: "Not allowed on dns root",
		Data:    DnsRootResponse{Groups: []string{}},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encoding response: %v", err)
	}
}

// Read is HTTP handler of GET model.Request.
// Use for reading existed records on DNS server.
func (s *DnsRootService) Read(w http.ResponseWriter, r *http.Request) {
	groups := s.Store.GetGroupBucket().ListGroups()
	var list = make([]string, 0)
	for _, g := range groups {
		list = append(list, g.Name)
	}
	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data:    DnsRootResponse{Groups: list},
	}
	w.WriteHeader(http.StatusOK)
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding response: %v", err)
	}
}

// Update is HTTP handler of PUT model.Request.
// Use for updating existed records on DNS server.
func (s *DnsRootService) Update(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: "Not allowed on dns root",
		Data:    DnsRootResponse{Groups: []string{}},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encoding response: %v", err)
	}
}

// Delete is HTTP handler of DELETE model.Request.
// Use for removing records on DNS server.
func (s *DnsRootService) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: "Not allowed on dns root",
		Data:    DnsRootResponse{Groups: []string{}},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encoding response: %v", err)
	}
}
