package v1

import (
	"fmt"
	"github.com/hellgate75/rebind/data"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/store"
	"github.com/hellgate75/rebind/utils"
	net2 "net"
	"net/http"
	"strings"
	"time"
)

// DnsGroupResourcesService is an implementation of RestService interface.
type DnsGroupResourcesService struct {
	Pipe  net.NetPipe
	Store registry.Store
	Log   log.Logger
}

type DnsGroupResourceType struct {
	Name    string  `yaml:"hostname,omitempty" json:"hostname,omitempty" xml:"hostname,omitempty"`
	Type    string  `yaml:"type,omitempty" json:"type,omitempty" xml:"type,omitempty"`
	Addr    net2.IP `yaml:"address,omitempty" json:"address,omitempty" xml:"address,omitempty"`
	RecData string  `yaml:"recordData,omitempty" json:"recordData,omitempty" xml:"record-data,omitempty"`
	Record  string  `yaml:"record,omitempty" json:"record,omitempty" xml:"record,omitempty"`
}

type DnsGroupResourcesBucker struct {
	Resources []DnsGroupResourceType `yaml:"resources" json:"resources" xml:"resources"`
}

func getParentGroup(r *http.Request) string {
	arr := strings.Split(r.URL.Path, "/")
	return arr[len(arr)-2]

}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *DnsGroupResourcesService) Create(w http.ResponseWriter, r *http.Request) {
	s.Store.Load()
	groupName := getParentGroup(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		writeResourcesErrorResponse(w, r, s.Log, group.Name, "create-resource", "group doesn't exists", http.StatusConflict)
		return
	}
	var req model.Request
	err = utils.RestParseRequest(w, r, &req)
	if err != nil {
		writeResourcesErrorResponse(w, r, s.Log, group.Name, "create-resource", fmt.Sprintf("decoding group creation request, Error: %v", err), http.StatusBadRequest)
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		writeResourcesErrorResponse(w, r, s.Log, group.Name, "create-resource", fmt.Sprintf("loading store for group %s, Error: %v", groupName, err), http.StatusInternalServerError)
		return
	}
	//	return req.Host, req.Type, udpAddr, rData, dnsmessage.Resource{
	host, typeS, ipAddr, recData, resource, err := utils.ToRecordData(req)
	_, _, bl := s.Store.Get(host)
	if bl {
		writeResourcesErrorResponse(w, r, s.Log, group.Name, "create-resource", fmt.Sprintf("loading store for group %s, Host %s already present in the store", groupName, host), http.StatusInternalServerError)
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
		writeResourcesErrorResponse(w, r, s.Log, group.Name, "create-resource", fmt.Sprintf("creating new group, Error: %v", err), http.StatusLocked)
		return
	}

	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data:    DnsGroupsResponse{Groups: []data.Group{group}},
	}
	w.WriteHeader(http.StatusCreated)
	err = utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding group creation response, Error: %v", err)
	}
}

// Read is HTTP handler of GET model.Request.
// Use for reading existed records on DNS server.
func (s *DnsGroupResourcesService) Read(w http.ResponseWriter, r *http.Request) {
	groupName := getParentGroup(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		writeResourcesErrorResponse(w, r, s.Log, group.Name, "get-resources", "group doesn't exists", http.StatusNotFound)
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		writeResourcesErrorResponse(w, r, s.Log, group.Name, "get-resources", fmt.Sprintf("recovering group store, Error:", err), http.StatusInternalServerError)
		return
	}
	var recs = make([]store.DNSRecord, 0)
	for _, key := range gsd.Keys() {
		lst, _ := gsd.Get(key)
		if len(lst) > 0 {
			recs = append(recs, lst...)
		}
	}
	var resources = make([]DnsGroupResourceType, 0)
	for _, key := range gsd.Keys() {
		lst, _ := gsd.Get(key)
		if len(lst) > 0 {
			for _, rec := range lst {
				resources = append(resources, DnsGroupResourceType{
					Name:    rec.NodeName,
					Type:    rec.Resource.Header.Type.String(),
					Addr:    rec.Addr,
					RecData: rec.Data,
					Record:  rec.Resource.Body.GoString(),
				})
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data: DnsGroupResourcesBucker{
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
func (s *DnsGroupResourcesService) Update(w http.ResponseWriter, r *http.Request) {
	groupName := getParentGroup(r)
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: fmt.Sprintf("Not allowed on dns group %s resources", groupName),
		Data: DnsGroupsResponse{
			Groups: []data.Group{},
		},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encoding group %s resources response: %v", groupName, err)
	}
}

func writeResourcesErrorResponse(w http.ResponseWriter, r *http.Request, logger log.Logger, groupName string, requestType string, messageSuffix string, httpStatus int) {
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
func (s *DnsGroupResourcesService) Delete(w http.ResponseWriter, r *http.Request) {
	groupName := getParentGroup(r)
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: fmt.Sprintf("Not allowed on dns group %s resources", groupName),
		Data: DnsGroupsResponse{
			Groups: []data.Group{},
		},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encoding group %s resources response: %v", groupName, err)
	}
}
