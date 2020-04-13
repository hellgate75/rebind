package v1

import (
	"fmt"
	"github.com/hellgate75/rebind/data"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/model/rest"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/utils"
	net2 "net"
	"net/http"
	"strings"
)

// DnsGroupsService is an implementation of RestService interface.
type DnsGroupsService struct {
	Pipe    net.NetPipe
	Store   registry.Store
	Log     log.Logger
	BaseUrl string
}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *DnsGroupsService) Create(w http.ResponseWriter, r *http.Request) {
	var req rest.GroupRequest
	err := utils.RestParseRequest(w, r, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response := model.Response{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Error: %v", err),
			Data:    rest.DnsGroupsResponse{Groups: []data.Group{}},
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
			Data:    rest.DnsGroupsResponse{Groups: []data.Group{}},
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
			Data:    rest.DnsGroupsResponse{Groups: []data.Group{}},
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
			Data:    rest.DnsGroupsResponse{Groups: []data.Group{}},
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
		Data:    rest.DnsGroupsResponse{Groups: []data.Group{group}},
	}
	w.WriteHeader(http.StatusCreated)
	err = utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding groups response, Error: %v", err)
	}
}

func writeGroupsErrorResponse(w http.ResponseWriter, r *http.Request, logger log.Logger, groupName string, requestType string, messageSuffix string, httpStatus int) {
	w.WriteHeader(httpStatus)
	response := model.Response{
		Status:  httpStatus,
		Message: fmt.Sprintf("Group %s request %s : %s", groupName, requestType, messageSuffix),
		Data:    nil,
	}
	logger.Errorf("Group %s %s request : %s", groupName, requestType, messageSuffix)
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Errorf("Error encoding group %s update query response, Error: %v", groupName, err)
	}
}

// Read is HTTP handler of GET model.Request.
// Use for reading existed records on DNS server.
func (s *DnsGroupsService) Read(w http.ResponseWriter, r *http.Request) {
	var action = r.URL.Query().Get("action")
	if strings.ToLower(action) == "template" {
		var templates = make([]rest.DnsTemplateDataType, 0)
		templates = append(templates, rest.DnsTemplateDataType{
			Method:  "POST",
			Header:  []string{},
			Query:   []string{},
			Request: rest.GroupRequest{},
		})
		templates = append(templates, rest.DnsTemplateDataType{
			Method:  "PUT",
			Header:  []string{},
			Query:   []string{},
			Request: model.Request{},
		})
		templates = append(templates, rest.DnsTemplateDataType{
			Method:  "DELETE",
			Header:  []string{},
			Query:   []string{},
			Request: nil,
		})
		templates = append(templates, rest.DnsTemplateDataType{
			Method:  "GET",
			Header:  []string{"Name: default", "Domain: my-dcomain.com", "Forwarder: 8.8.8.8", "Forwarder: 53"},
			Query:   []string{"action=template", "name=default", "domain=my-dcomain.com", "forwarder=8.8.8.8", "forwarder=53"},
			Request: nil,
		})
		tErr := utils.RestParseResponse(w, r, &rest.DnsTemplateResponse{
			Templates: templates,
		})
		if tErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding template(s) summary response, Error: %v", tErr)
		}
		return
	}
	groups := s.Store.GetGroupBucket().ListGroups()
	name := r.URL.Query().Get("name")
	if name == "" {
		name = r.Header.Get("Name")
	}
	if name == "" {
		name = r.Header.Get("name")
	}
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		domain = r.Header.Get("Domain")
	}
	if domain == "" {
		domain = r.Header.Get("domain")
	}
	forwarder := r.URL.Query().Get("forwarder")
	if forwarder == "" {
		forwarder = r.Header.Get("Forwarder")
	}
	if forwarder == "" {
		forwarder = r.Header.Get("forwarder")
	}
	var list = make([]data.Group, 0)
	for _, g := range groups {
		if (name == "" || name == g.Name) &&
			(domain == "" || utils.StringsListContainItem(domain, g.Domains, true)) &&
			(forwarder == "" || utils.UDPAddrListContainsValue(g.Forwarders, forwarder)) {
			list = append(list, g)
		}
	}
	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data:    rest.DnsGroupsResponse{Groups: list},
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
		Data: rest.DnsGroupsResponse{
			Groups: []data.Group{},
		},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encoding response: %v", err)
	}
}

// Delete is HTTP handler of DELETE model.Request.
// Use for removing records on DNS server.
func (s *DnsGroupsService) Delete(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: "Not allowed on dns groups",
		Data: rest.DnsGroupsResponse{
			Groups: []data.Group{},
		},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encoding response: %v", err)
	}
}
