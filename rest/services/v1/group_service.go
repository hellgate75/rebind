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
	"net/http"
	"strings"
)

// DnsGroupService is an implementation of RestService interface.
type DnsGroupService struct {
	Pipe  net.NetPipe
	Store registry.Store
	Log   log.Logger
}

type DnsGroupResponse struct {
	Group		data.Group
	Resources	[]store.DNSRecord
}

func getGroup(r *http.Request) string {
	arr := strings.Split(r.URL.Path, "/")
	return arr[len(arr)-1]

}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *DnsGroupService) Create(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	response := model.Response{
		Status:  http.StatusMethodNotAllowed,
		Message: "Not allowed on specific dns group",
		Data: DnsGroupResponse{
		},
	}
	err := utils.RestParseResponse(w, r, &response)
	if err != nil {
		s.Log.Errorf("Error encording response: %v", err)
	}
}

// Read is HTTP handler of GET model.Request.
// Use for reading existed records on DNS server.
func (s *DnsGroupService) Read(w http.ResponseWriter, r *http.Request) {
	groupName := getGroup(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := model.Response{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("Group %s doesn't exist", groupName),
			Data: DnsGroupResponse{
			},
		}
		s.Log.Errorf("Group %s doesn't exist, Error: %v", groupName, err)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding group query response, Error: %v", err)
		}
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := model.Response{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("Error: %v", err),
			Data: DnsGroupResponse{
			},
		}
		s.Log.Errorf("Error recovering group store, Error: %v", err)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding group store response, Error: %v", err)
		}
		return
	}
	var recs = make([]store.DNSRecord, 0)
	for _, key := range gsd.Keys() {
		lst, _ := gsd.Get(key)
		if len(lst) > 0 {
			recs = append(recs, lst...)
		}
	}
	w.WriteHeader(http.StatusOK)
	response := model.Response{
		Status:  http.StatusOK,
		Message: "OK",
		Data: DnsGroupResponse{
			Group: group,
			Resources: recs,
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
func (s *DnsGroupService) Update(w http.ResponseWriter, r *http.Request) {
	groupName := getGroup(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := model.Response{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("Group %s doesn't exist", groupName),
			Data: DnsGroupResponse{
			},
		}
		s.Log.Errorf("Group %s doesn't exist, Error: %v", groupName, err)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding group query response, Error: %v", err)
		}
		return
	}
	s.Log.Infof("Group: %v", group)
	http.Error(w, "", http.StatusNotFound)
}

// Delete is HTTP handler of DELETE model.Request.
// Use for removing records on DNS server.
func (s *DnsGroupService) Delete(w http.ResponseWriter, r *http.Request) {
	groupName := getGroup(r)
	if groupName == utils.DEFAULT_GROUP_NAME {
		w.WriteHeader(http.StatusUnauthorized)
		response := model.Response{
			Status:  http.StatusUnauthorized,
			Message: fmt.Sprintf("Group %s cannot be deleted", groupName),
			Data: DnsGroupResponse{
			},
		}
		s.Log.Errorf("Group %s cannot be deleted", groupName)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding delete group response, Error: %v", err)
		}
		return
	}

	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response := model.Response{
			Status:  http.StatusNotFound,
			Message: fmt.Sprintf("Group %s doesn't exist", groupName),
			Data: DnsGroupResponse{
			},
		}
		s.Log.Errorf("Group %s doesn't exist, Error: %v", groupName, err)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding delete group response, Error: %v", err)
		}
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response := model.Response{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("Error: %v", err),
			Data: DnsGroupResponse{
			},
		}
		s.Log.Errorf("Error recovering group store, Error: %v", err)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding group store response, Error: %v", err)
		}
		return
	}
	state := s.Store.GetGroupBucket().Delete(group.Name)
	if ! state {
		w.WriteHeader(http.StatusInternalServerError)
		response := model.Response{
			Status:  http.StatusInternalServerError,
			Message: fmt.Sprintf("Group %s couldn't be deleted", groupName),
			Data: DnsGroupResponse{
			},
		}
		s.Log.Errorf("Group %s couldn't be deleted", groupName)
		err := utils.RestParseResponse(w, r, &response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorf("Error encoding delete group response, Error: %v", err)
		}
		return
	}
	s.Log.Infof("Group: %v was deleted!!", group)
	var recs = make([]store.DNSRecord, 0)
	for _, key := range gsd.Keys() {
		lst, _ := gsd.Get(key)
		if len(lst) > 0 {
			recs = append(recs, lst...)
		}
	}
	w.WriteHeader(http.StatusOK)
	response := model.Response{
		Status:  http.StatusOK,
		Message: "DELETED",
		Data: DnsGroupResponse{
			Group: group,
			Resources: recs,
		},
	}
	err = utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding group delete response, Error: %v", err)
	}
}
