package v1

import (
	"fmt"
	"github.com/hellgate75/rebind/data"
	"github.com/hellgate75/rebind/log"
	"github.com/hellgate75/rebind/model"
	"github.com/hellgate75/rebind/model/rest"
	"github.com/hellgate75/rebind/net"
	"github.com/hellgate75/rebind/registry"
	"github.com/hellgate75/rebind/store"
	"github.com/hellgate75/rebind/utils"
	net2 "net"
	"net/http"
	"reflect"
	"strings"
)

// DnsGroupService is an implementation of RestService interface.
type DnsGroupService struct {
	Pipe    net.NetPipe
	Store   registry.Store
	Log     log.Logger
	BaseUrl string
}

func toField(value string) rest.Field {
	return rest.Field(strings.ToLower(value))
}

func getGroup(r *http.Request) string {
	arr := strings.Split(r.URL.Path, "/")
	return arr[len(arr)-1]

}

// Create is HTTP handler of POST model.Request.
// Use for adding new record to DNS server.
func (s *DnsGroupService) Create(w http.ResponseWriter, r *http.Request) {
	s.Store.Load()
	groupName := getGroup(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err == nil {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "create-group", "group already exists", http.StatusConflict)
		return
	}
	var req rest.GroupCreationRequest
	err = utils.RestParseRequest(w, r, &req)
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "create-group", fmt.Sprintf("decoding group creation request, Error: %v", err), http.StatusBadRequest)
		return
	}
	if req.Domains == nil {
		req.Domains = []string{}
	}
	if req.Forwarders == nil {
		req.Forwarders = []net2.UDPAddr{}
	}
	group, _, err = s.Store.GetGroupBucket().CreateAndPersistGroupAndStore(groupName, req.Domains, req.Forwarders)
	if err == nil {
		s.Store.Save()
	}
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "create-group", fmt.Sprintf("creating new group, Error: %v", err), http.StatusLocked)
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
		s.Log.Errorf("Error encoding group creation response, Error: %v", err)
	}
}

// Read is HTTP handler of GET model.Request.
// Use for reading existed records on DNS server.
func (s *DnsGroupService) Read(w http.ResponseWriter, r *http.Request) {
	var action = r.URL.Query().Get("action")
	if strings.ToLower(action) == "template" {
		var templates = make([]rest.DnsTemplateDataType, 0)
		templates = append(templates, rest.DnsTemplateDataType{
			Method:  "POST",
			Header:  []string{},
			Query:   []string{},
			Request: rest.GroupCreationRequest{},
		})
		templates = append(templates, rest.DnsTemplateDataType{
			Method:  "PUT",
			Header:  []string{},
			Query:   []string{},
			Request: rest.DnsUpdateRequest{},
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
	groupName := getGroup(r)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "get-group", "group doesn't exists", http.StatusNotFound)
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "get-group", fmt.Sprintf("recovering group store, Error:", err), http.StatusInternalServerError)
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
		Data: rest.DnsGroupResponse{
			Group:     group,
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
	s.Log.Infof("Update Request for Group: %s", groupName)
	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "group doesn't exists", http.StatusNotFound)
		return
	}
	var req rest.DnsUpdateRequest
	err = utils.RestParseRequest(w, r, &req)
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("cannot parse request, Error: %v", err), http.StatusBadRequest)
		return
	}
	if req.Field == "" || !isCorrectAction(req.Action) {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "must have at least a field and and action", http.StatusBadRequest)
		return
	}
	if strings.ToLower(req.Field) == "name" {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "cannot update/delete group name", http.StatusUnauthorized)
		return
	}
	var field = toField(req.Field)
	if rest.DeleteResoource.Equals(req.Action) {
		//Request delete of an element
		if field.Equals(rest.Field("domain")) {
			value := req.Data.ListData.Value
			index := req.Data.ListData.Index
			if value == "" {
				if index < 0 || index >= len(group.Domains) {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("unset value and index out of bounds (%v) ->  0 <= n < %v", index, len(group.Domains)), http.StatusBadRequest)
					return
				}
			} else {
				index = -1
				for idx, domain := range group.Domains {
					if strings.ToLower(domain) == strings.ToLower(value) {
						index = idx
						break
					}
				}
				if index < 0 {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("given value %s is not present in the list", value), http.StatusBadRequest)
					return
				}
			}
			if index > 0 && index < len(group.Domains) {
				var out = make([]string, 0)
				out = append(out, group.Domains[0:index]...)
				out = append(out, group.Domains[index+1:]...)
				group.Domains = out
			} else if index == 0 {
				if len(group.Domains) > 1 {
					group.Domains = group.Domains[1:]
				} else {
					group.Domains = []string{}
				}
			} else {
				if len(group.Domains) > 1 {
					group.Domains = group.Domains[:len(group.Domains)-1]
				} else {
					group.Domains = []string{}
				}
			}
			s.Store.GetGroupBucket().UpdateExistingGroup(group)
			err = s.Store.GetGroupBucket().SaveMeta()
		} else if field.Equals(rest.Field("domains")) {
			group.Domains = []string{}
			s.Store.GetGroupBucket().UpdateExistingGroup(group)
			err = s.Store.GetGroupBucket().SaveMeta()
		} else if field.Equals(rest.Field("forwarder")) {
			value := req.Data.ListData.Value
			index := req.Data.ListData.Index
			if value == "" {
				if index < 0 && index >= len(group.Forwarders) {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("unset value and index out of bounds (%v) ->  0 <= n < %v", index, len(group.Forwarders)), http.StatusBadRequest)
					return
				}
			} else {
				index = -1
				for idx, forwarder := range group.Forwarders {
					if strings.Contains(strings.ToLower(forwarder.String()), value) {
						index = idx
						break
					}
				}
				if index < 0 {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("given value %s is not present in the list", value), http.StatusBadRequest)
					return
				}
			}
			if index > 0 && index < len(group.Forwarders) {
				var out = make([]net2.UDPAddr, 0)
				out = append(out, group.Forwarders[0:index]...)
				out = append(out, group.Forwarders[index+1:]...)
				group.Forwarders = out
			} else if index == 0 {
				if len(group.Forwarders) > 1 {
					group.Forwarders = group.Forwarders[1:]
				} else {
					group.Forwarders = []net2.UDPAddr{}
				}
			} else {
				if len(group.Forwarders) > 1 {
					group.Forwarders = group.Forwarders[:len(group.Domains)-1]
				} else {
					group.Forwarders = []net2.UDPAddr{}
				}
			}
			s.Store.GetGroupBucket().UpdateExistingGroup(group)
			err = s.Store.GetGroupBucket().SaveMeta()
		} else if field.Equals(rest.Field("forwarders")) {
			group.Forwarders = []net2.UDPAddr{}
			s.Store.GetGroupBucket().UpdateExistingGroup(group)
			err = s.Store.GetGroupBucket().SaveMeta()
		} else if field.Equals(rest.Field("data")) ||
			field.Equals(rest.Field("resources")) {
			gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
			if err == nil {
				gsd.ClearData()
				_, err = s.Store.GetGroupBucket().SaveGroup(gsd, group)
			}
		} else if field.Equals(rest.Field("resource")) {
			gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
			if err == nil {
				if req.Data.ListData.Value == "" {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "Request.Data.ListData.Value cannot be empty to delete a record", http.StatusBadRequest)
					return
				}
				rErr := gsd.Remove(req.Data.ListData.Value)
				if rErr != nil {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("unable to delete record names %s, Error: %v", req.Data.ListData.Value, rErr.Error()), http.StatusInternalServerError)
					return
				}
			}
		} else {
			writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("Cannot delete field type: %v", field), http.StatusNotImplemented)
			return
		}
	} else if rest.UpdateResoource.Equals(req.Action) {
		//Request Update of an element
		if field.Equals(rest.Field("domain")) ||
			field.Equals(rest.Field("domains")) {
			if req.Data.NewValue == nil ||
				req.Data.NewValue == "" {
				writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "Request.Data.NewValue connot be nil or empty, as update value", http.StatusBadRequest)
				return
			}
			value := req.Data.ListData.Value
			index := req.Data.ListData.Index
			if value == "" {
				if index < 0 || index >= len(group.Domains) {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("unset value and index out of bounds (%v) ->  0 <= n < %v", index, len(group.Domains)), http.StatusBadRequest)
					return
				}
			} else {
				index = -1
				for idx, domain := range group.Domains {
					if strings.ToLower(domain) == strings.ToLower(value) {
						index = idx
						break
					}
				}
				if index < 0 {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("given value %s is not present in the list", value), http.StatusBadRequest)
					return
				}
			}
			group.Domains[index] = fmt.Sprintf("%v", req.Data.NewValue)
			s.Store.GetGroupBucket().UpdateExistingGroup(group)
			err = s.Store.GetGroupBucket().SaveMeta()
		} else if field.Equals(rest.Field("forwarder")) ||
			field.Equals(rest.Field("forwarders")) {
			if req.Data.NewValue == nil ||
				req.Data.NewValue == "" {
				writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "Request.Data.NewValue connot be nil or empty, as update value", http.StatusBadRequest)
				return
			}
			value := req.Data.ListData.Value
			index := req.Data.ListData.Index
			if value == "" {
				if index < 0 && index >= len(group.Forwarders) {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("unset value and index out of bounds (%v) ->  0 <= n < %v", index, len(group.Forwarders)), http.StatusBadRequest)
					return
				}
			} else {
				index = -1
				for idx, forwarder := range group.Forwarders {
					if strings.Contains(strings.ToLower(forwarder.String()), value) {
						index = idx
						break
					}
				}
				if index < 0 {
					writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("given value %s is not present in the list", value), http.StatusBadRequest)
					return
				}
			}
			if reflect.TypeOf(req.Data.NewValue).AssignableTo(reflect.TypeOf(net2.UDPAddr{})) {
				group.Forwarders[index] = req.Data.NewValue.(net2.UDPAddr)
			} else {
				writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "Request.Data.NewValue is not type of net.UDPAddr, as update value", http.StatusBadRequest)
				return
			}
			s.Store.GetGroupBucket().UpdateExistingGroup(group)
			err = s.Store.GetGroupBucket().SaveMeta()
		} else if field.Equals(rest.Field("data")) ||
			field.Equals(rest.Field("resources")) {
			writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("Cannot update field type: %v", field), http.StatusNotImplemented)
			return
		} else if field.Equals(rest.Field("resource")) {
			writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("Cannot update field type: %v", field), http.StatusNotImplemented)
			return
		} else {
			writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("Cannot update field type: %v", field), http.StatusNotImplemented)
			return
		}
	} else if rest.AddResoource.Equals(req.Action) {
		//Request a new resource in an element
		if field.Equals(rest.Field("domain")) ||
			field.Equals(rest.Field("domains")) {
			if req.Data.NewValue == nil ||
				req.Data.NewValue == "" {
				writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "Request.Data.NewValue connot be nil or empty, as update value", http.StatusBadRequest)
				return
			}
			group.Domains = append(group.Domains, fmt.Sprintf("%v", req.Data.NewValue))
			s.Store.GetGroupBucket().UpdateExistingGroup(group)
			err = s.Store.GetGroupBucket().SaveMeta()
		} else if field.Equals(rest.Field("forwarder")) ||
			field.Equals(rest.Field("forwarders")) {
			if req.Data.NewValue == nil ||
				req.Data.NewValue == "" {
				writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "Request.Data.NewValue connot be nil or empty, as update value", http.StatusBadRequest)
				return
			}
			if reflect.TypeOf(req.Data.NewValue).AssignableTo(reflect.TypeOf(net2.UDPAddr{})) {
				group.Forwarders = append(group.Forwarders, req.Data.NewValue.(net2.UDPAddr))
			} else {
				writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", "Request.Data.NewValue is not type of net.UDPAddr, as update value", http.StatusBadRequest)
				return
			}
			s.Store.GetGroupBucket().UpdateExistingGroup(group)
			err = s.Store.GetGroupBucket().SaveMeta()
		} else if field.Equals(rest.Field("data")) ||
			field.Equals(rest.Field("resources")) {
			writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("Cannot update field type: %v", field), http.StatusNotImplemented)
			return
		} else if field.Equals(rest.Field("resource")) {
			writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("Cannot update field type: %v", field), http.StatusNotImplemented)
			return
		} else {
			writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("Cannot update field type: %v", field), http.StatusNotImplemented)
			return
		}
	} else {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("Unknown action %v on field type: %v", req.Action, field), http.StatusNotImplemented)
		return
	}
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, group.Name, "update-group", fmt.Sprintf("unable to save group data, Error:", err), http.StatusInternalServerError)
		return
	}
	http.Error(w, "", http.StatusNotFound)
}

func writeUpdateErrorResponse(w http.ResponseWriter, r *http.Request, logger log.Logger, groupName string, requestType string, messageSuffix string, httpStatus int) {
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

func isCorrectAction(act rest.Action) bool {
	return rest.AddResoource.Equals(act) ||
		rest.DeleteResoource.Equals(act) ||
		rest.UpdateResoource.Equals(act)
}

// Delete is HTTP handler of DELETE model.Request.
// Use for removing records on DNS server.
func (s *DnsGroupService) Delete(w http.ResponseWriter, r *http.Request) {
	groupName := getGroup(r)
	if groupName == utils.DEFAULT_GROUP_NAME {
		writeUpdateErrorResponse(w, r, s.Log, groupName, "delete-group", "default group cannot be delete", http.StatusUnauthorized)
		return
	}

	group, err := s.Store.GetGroupBucket().GetGroupById(groupName)
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, groupName, "delete-group", "requested group doesn't exist", http.StatusNotFound)
		return
	}
	gsd, err := s.Store.GetGroupBucket().GetGroupStore(group)
	if err != nil {
		writeUpdateErrorResponse(w, r, s.Log, groupName, "delete-group", fmt.Sprintf("recovering group store, Error: %v", err), http.StatusInternalServerError)
		return
	}
	state := s.Store.GetGroupBucket().Delete(group.Name)
	if !state {
		writeUpdateErrorResponse(w, r, s.Log, groupName, "delete-group", "requested group couldn't be deleted", http.StatusInternalServerError)
		return
	}
	s.Log.Infof("Group: %s has been deleted!!", group.Name)
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
		Data: rest.DnsGroupResponse{
			Group:     group,
			Resources: recs,
		},
	}
	err = utils.RestParseResponse(w, r, &response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorf("Error encoding group delete response, Error: %v", err)
	}
}
