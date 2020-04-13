package v1

import (
	"fmt"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/model/rest"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/store"
	"github.com/hellgate75/rebind/utils"
	"net/http"
	"strings"
	"time"
)

// DnsGroupResourceDetailsService is an implementation of RestService interface.
type DnsGroupResourceDetailsService struct {
	Pipe    net.NetPipe
	Store   registry.Store
	Log     log.Logger
	BaseUrl string
}

func getResourceGroup(r *http.Request) string {
	arr := strings.Split(r.URL.Path, "/")
	return arr[len(arr)-3]

}

func getResourceHost(r *http.Request) string {
	arr := strings.Split(r.URL.Path, "/")
	return arr[len(arr)-1]

}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *DnsGroupResourceDetailsService) Create(w http.ResponseWriter, r *http.Request) {
	s.Store.Load()
	groupName := getResourceGroup(r)
	hostname := getResourceHost(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "create-resource-data", "group doesn't exists", http.StatusConflict)
		return
	}
	var req model.Request
	err = utils.RestParseRequest(w, r, &req)
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "create-resource-data", fmt.Sprintf("decoding group resource creation request, Error: %v", err), http.StatusBadRequest)
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "create-resource-data", fmt.Sprintf("loading store for group %s resource host %s, Error: %v", groupName, hostname, err), http.StatusInternalServerError)
		return
	}
	//	return req.Host, req.Type, udpAddr, rData, dnsmessage.Resource{
	host, typeS, ipAddr, recData, resource, err := utils.ToRecordData(req)
	if host != hostname {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "create-resource-data", fmt.Sprintf("loading store for group %s resource, Host %s doesn't match with the path %s", groupName, host, hostname), http.StatusBadRequest)
		return
	}
	_, _, bl := s.Store.Get(host)
	if bl {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "create-resource-data", fmt.Sprintf("loading store for group %s resource, Host %s already present in the store", groupName, host), http.StatusInternalServerError)
		return
	}
	rec := store.DNSRecord{
		Data:     recData,
		Addr:     ipAddr,
		Resource: resource,
		Type:     typeS,
		NodeName: host,
		TTL:      resource.Header.TTL,
		Created:  time.Now(),
	}
	sErr := gsd.Set(host, rec)
	if sErr != nil {
		err = sErr.Error()
	}
	if err == nil {
		s.Store.Save()
	}
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "create-resource-data", fmt.Sprintf("creating new group resource, Error: %v", err), http.StatusLocked)
		return
	}

	var resourceAnswer = DnsGroupResourceType{
		Type:    typeS,
		Addr:    ipAddr,
		Name:    host,
		RecData: recData,
		Record:  resource.Body.GoString(),
	}
	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data:    DnsGroupResourcesBucket{Resources: []DnsGroupResourceType{resourceAnswer}},
	}
	w.WriteHeader(http.StatusCreated)
	err = utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding group resource %s creation response, Error: %v", hostname, err)
	}
}

// Read is HTTP handler of GET model.Request.
// Use for reading existed records on DNS server.
func (s *DnsGroupResourceDetailsService) Read(w http.ResponseWriter, r *http.Request) {
	var action = r.URL.Query().Get("action")
	if strings.ToLower(action) == "template" {
		var templates = make([]rest.DnsTemplateDataType, 0)
		templates = append(templates, rest.DnsTemplateDataType{
			Method:  "POST",
			Header:  []string{},
			Query:   []string{},
			Request: model.Request{},
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
			Header:  []string{},
			Query:   []string{"action=template"},
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
	groupName := getResourceGroup(r)
	hostname := getResourceHost(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "get-resource-datas", "group doesn't exists", http.StatusNotFound)
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "get-resource-datas", fmt.Sprintf("recovering group store, Error:", err), http.StatusInternalServerError)
		return
	}
	dnsRecs, sErr := gsd.Get(hostname)
	if sErr != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "get-resource-datas", fmt.Sprintf("recovering group resources, Error:", sErr.Error()), http.StatusInternalServerError)
		return
	}
	var resources = make([]DnsGroupResourceType, 0)
	for _, rec := range dnsRecs {
		resources = append(resources, DnsGroupResourceType{
			Type:    rec.Type,
			Addr:    rec.Addr,
			Name:    rec.NodeName,
			RecData: rec.Data,
			Record:  rec.Resource.Body.GoString(),
		})
	}
	w.WriteHeader(http.StatusOK)
	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data: DnsGroupResourcesBucket{
			Resources: resources,
		},
	}
	err = utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding group store response, Error: %v", err)
	}
}

// Update is HTTP handler of PUT model.Request.
// Use for updating existed records on DNS server.
func (s *DnsGroupResourceDetailsService) Update(w http.ResponseWriter, r *http.Request) {
	s.Create(w, r)
}

func writeResourceDetailsErrorResponse(w http.ResponseWriter, r *http.Request, logger log.Logger, groupName string, requestType string, messageSuffix string, httpStatus int) {
	w.WriteHeader(httpStatus)
	response := model.Response{
		Status:  httpStatus,
		Message: fmt.Sprintf("Group %s Resources request %s : %s", groupName, requestType, messageSuffix),
		Data:    nil,
	}
	logger.Errorf("Group %s Resources %s request : %s", groupName, requestType, messageSuffix)
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		logger.Errorf("Error encoding group %s resource query response, Error: %v", groupName, err)
	}
}

// Delete is HTTP handler of DELETE model.Request.
// Use for removing records on DNS server.
func (s *DnsGroupResourceDetailsService) Delete(w http.ResponseWriter, r *http.Request) {
	groupName := getResourceGroup(r)
	hostname := getResourceHost(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "get-resource-datas", "group doesn't exists", http.StatusNotFound)
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "get-resource-datas", fmt.Sprintf("recovering group store, Error:", err), http.StatusInternalServerError)
		return
	}
	dErr := gsd.Remove(hostname)
	if dErr != nil {
		err = dErr.Error()
	}
	if err != nil {
		writeResourceDetailsErrorResponse(w, r, s.Log, group.Name, "get-resource-datas", fmt.Sprintf("deleting dns record for host: %s, Error:", hostname, err), http.StatusInternalServerError)
		return
	}
	s.Log.Infof("Group: %v -> Resource: %s has been deleted!!", group.Name, hostname)
	w.WriteHeader(http.StatusOK)
	response := model.Response{
		Status:  http.StatusOK,
		Message: fmt.Sprintf("Removed on dns group %s the resource: %s", groupName, hostname),
		Data:    nil,
	}
	err = utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encoding group %s resources %s deletion response: %v", groupName, hostname, err)
	}
}
