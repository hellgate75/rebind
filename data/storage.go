package data

import (
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/hellgate75/rebind/store"
	"github.com/hellgate75/rebind/utils"
	"golang.org/x/net/dns/dnsmessage"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net"
	"os"
)

func init() {
	gob.Register(&dnsmessage.AResource{})
	gob.Register(&dnsmessage.NSResource{})
	gob.Register(&dnsmessage.CNAMEResource{})
	gob.Register(&dnsmessage.SOAResource{})
	gob.Register(&dnsmessage.PTRResource{})
	gob.Register(&dnsmessage.MXResource{})
	gob.Register(&dnsmessage.AAAAResource{})
	gob.Register(&dnsmessage.SRVResource{})
	gob.Register(&dnsmessage.TXTResource{})
	gob.Register(&dnsmessage.PTRResource{})

	gob.Register(&dnsmessage.Resource{})
	gob.Register(&dnsmessage.ResourceHeader{})
	gob.Register(&dnsmessage.Name{})
	gob.Register(&dnsmessage.Header{})
	gob.Register(&dnsmessage.Message{})
	gob.Register(&dnsmessage.Option{})
	gob.Register(&dnsmessage.OPTResource{})
	gob.Register(&dnsmessage.Question{})

	gob.Register(&store.GroupsStoreData{})
	gob.Register(&store.GroupStoreData{})
	gob.Register(&store.AnswersCacheStoreData{})
	gob.Register(&store.GroupStorePersistent{})

	gob.Register(&Group{})
	gob.Register(&GroupsBucket{})
	gob.Register(&GroupBlock{})
	gob.Register(&groupBucketPersitence{})
	gob.Register(&store.DNSRecord{})
	gob.Register(&net.UDPAddr{})
}

var (
	__sepPath string = fmt.Sprintf("%c", os.PathSeparator)
)

func (i *GroupsBucket) Load(defaultForwarders []net.UDPAddr) error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprintf("%v", r))
		}
	}()
	fileName := fmt.Sprintf("%s%sgroups.yaml", i.Folder, __sepPath)
	if _, err := os.Stat(fileName); err != nil {
		if i.log != nil {
			i.log.Infof("Missing default group storage: Creating file: %s", fileName)
		} else {
			fmt.Printf("[info ] Missing default group storage: Creating file: %s\n", fileName)
		}
		_, _, err = i.CreateAndPersistGroupAndStore("default", []string{}, defaultForwarders)
		if err == nil {
			err = i.SaveMeta()
		}
	} else {
		err = i.ReLoad()
	}
	return err
}

func (i *GroupsBucket) ReLoad() error {
	fileName := fmt.Sprintf("%s%sgroups.yaml", i.Folder, __sepPath)
	if _, err := os.Stat(fileName); err != nil {
		return errors.New(fmt.Sprintf("Unable to find file: %s", fileName))
	}
	arr, rErr := ioutil.ReadFile(fileName)
	if rErr != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error reading groups main index at: %s, Error: %v", fileName, rErr)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error reading groups main index at: %s, Error: %v", fileName, rErr)
		}
		return rErr
	}
	var bucket groupBucketPersitence
	mErr := yaml.Unmarshal(arr, &bucket)
	if mErr != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error converting groups main index, Error: %v", mErr)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error converting groups main index, Error: %v", mErr)
		}
		return mErr
	}
	i.Groups = bucket.Groups
	return nil
}

func (i *GroupsBucket) Contains(group string) bool {
	if _, ok := i.Groups[utils.ConvertKeyToId(group)]; ok {
		return true
	}
	return false
}

func (i *GroupsBucket) Keys() []string {
	var out []string = make([]string, 0)
	for k, _ := range i.Groups {
		out = append(out, k)
	}
	return out
}
func (i *GroupsBucket) CreateAndPersistGroupAndStore(key string, domains []string,
	forwarders []net.UDPAddr) (Group, store.GroupStoreData, error) {
	groupRef := i.CreateUnboundGroup(key, domains, forwarders)
	groupStore := store.NewGroupStore(groupRef.Name, domains, forwarders)
	var gsd = *(groupStore.(*store.GroupStoreData))
	groupRef, err := i.SaveGroup(gsd, groupRef)
	if err != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error creating and persisting store: %s, Error: %v", key, err)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error creating and persisting store: %s, Error: %v", key, err)
		}
		return Group{}, store.GroupStoreData{}, err
	}
	i.Groups[groupRef.Name] = groupRef
	err = i.SaveMeta()
	if err != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error creating and persisting group: %s, Error: %v", key, err)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error creating and persisting group: %s, Error: %v", key, err)
		}
		return Group{}, store.GroupStoreData{}, err
	}
	return groupRef,
		gsd,
		err
}

func (i *GroupsBucket) SaveMeta() error {
	var err error
	defer func() {
		if r := recover(); r != nil {
			if i.log != nil {
				i.log.Errorf("GroupsBucket:: [ERROR] Runtime Error saving groups main index, Error: %v", r)
			} else {
				fmt.Sprintf("GroupsBucket:: [ERROR] Runtime Error saving groups main index, Error: %v", r)
			}
			err = errors.New(fmt.Sprintf("Runtime Error saving groups main index, Error: %v", r))
		}
		i.Unlock()
	}()
	i.Lock()
	persistence := groupBucketPersitence{
		Groups: i.Groups,
	}
	arr, mErr := yaml.Marshal(&persistence)
	if mErr != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error converting groups main index, Error: %v", mErr)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error converting groups main index, Error: %v", mErr)
		}
		return mErr
	}
	fileName := fmt.Sprintf("%s%sgroups.yaml", i.Folder, __sepPath)
	sErr := ioutil.WriteFile(fileName, arr, 0666)
	if sErr != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error saving groups main index at: %s, Error: %v", fileName, sErr)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error saving groups main index at: %s, Error: %v", fileName, sErr)
		}
		return sErr
	}
	if i.log != nil {
		i.log.Infof("GroupsBucket:: [INFO] Successfully saved groups main index at: %s", fileName)
	} else {
		fmt.Sprintf("GroupsBucket:: [INFO] Successfully saved groups main index at: %s", fileName)
	}
	return err
}
func (i *GroupsBucket) ConvertToGroupLikeKey(key string) string {
	return utils.ConvertKeyToId(key)
}
func (i *GroupsBucket) CreateUnboundGroup(key string, domains []string,
	forwarders []net.UDPAddr) Group {
	def := utils.ConvertKeyToId(key)
	return Group{
		Name:       def,
		Domains:    domains,
		Forwarders: forwarders,
		NumRecs:    int64(0),
		File:       fmt.Sprintf("gob-%s.dat", def),
	}
}

func (i *GroupsBucket) Delete(groupName string) bool {
	if g, ok := i.Groups[groupName]; ok {
		fileName := fmt.Sprintf("%s%s%s", i.Folder, __sepPath, g.File)
		err := os.Remove(fileName)
		if err != nil {
			return false
		}
		delete(i.Groups, groupName)
		err = i.SaveMeta()
		if err != nil {
			return false
		}
	}
	return true
}

func (i *GroupsBucket) GetGroupById(id string) (Group, error) {
	if group, ok := i.Groups[id]; ok {
		return group, nil
	}
	return Group{}, errors.New(fmt.Sprintf("Unable to find group by id: %s", id))
}

func (i *GroupsBucket) GetGroupByName(name string) (Group, error) {
	for _, group := range i.Groups {
		if group.Name == name {
			return group, nil
		}
	}
	return Group{}, errors.New(fmt.Sprintf("Unable to find group by name: %s", name))
}

func (i *GroupsBucket) GetGroupsByDomain(domain string) ([]Group, error) {
	var out = make([]Group, 0)
	isDefault := utils.IsDefaultGroupDomain(domain)
	for _, group := range i.Groups {
		if (isDefault && group.Name == utils.DEFAULT_GROUP_NAME) ||
			utils.StringsListContainItem(domain, group.Domains, false) {
			out = append(out, group)
		}
	}
	if len(out) > 0 {
		return out, nil
	}
	return out, errors.New(fmt.Sprintf("Unable to find group by domain: %s", domain))
}

type GroupBlock struct {
	Group Group
	Data  store.GroupStoreData
}

var cache = make(map[string]GroupBlock)

func (i *GroupsBucket) GetGroupStore(group Group) (store.GroupStoreData, error) {
	if gs, ok := cache[group.Name]; ok {
		return gs.Data, nil
	}
	var err error
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Runtime Error loading groups store file, Error: %v", r)
			if i.log != nil {
				i.log.Errorf("GroupsBucket:: [ERROR]", message)
			} else {
				fmt.Println("GroupsBucket:: [ERROR]", message)
			}
			err = errors.New(message)
		}
		i.storeMutex.Unlock()
	}()
	i.storeMutex.Lock()
	fileName := fmt.Sprintf("%s%s%s", i.Folder, __sepPath, group.File)
	if _, err = os.Stat(fileName); err != nil {
		message := fmt.Sprintf("Error loading group groupStore file at: %s, File doesn't exist", fileName)
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR]", message)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] %s", message)
		}
		return store.GroupStoreData{}, errors.New(message)
	}
	f, cErr := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if cErr != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error reading group groupStore file at: %s, Error: %v", fileName, cErr)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error reading group groupStore file at: %s, Error: %v", fileName, cErr)
		}
		return store.GroupStoreData{}, cErr
	}
	defer f.Close()
	groupStore := store.GroupStoreData{}
	var load store.GroupStorePersistent
	err = gob.NewDecoder(f).Decode(&load)
	groupStore.FromPersistentData(load)
	if err != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error loading group groupStore file at: %s, Error: %v", fileName, err)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error loading group groupStore file at: %s, Error: %v", fileName, err)
		}
		return store.GroupStoreData{}, err
	}
	groupStore.Forwarders = group.Forwarders
	groupStore.Domains = group.Domains
	groupStore.GroupName = group.Name
	cache[group.Name] = GroupBlock{
		Group: group,
		Data:  groupStore,
	}
	return groupStore, err
}

func (i *GroupsBucket) SaveGroups(groupsStore store.GroupsStoreData) error {
	keys := groupsStore.Keys()
	for _, key := range keys {
		def := utils.ConvertKeyToId(key)
		zStore, err := groupsStore.Get(key)
		var forwards = make([]net.UDPAddr, 0)
		var length int64 = 0
		if err == nil {
			forwards = utils.RemoveDuplicatesInUpdAddrList(append(forwards, zStore.GetForwarders()...))
			for _, key := range zStore.Keys() {
				el, _ := zStore.Get(key)
				length += int64(len(el))
			}
		}
		if _, ok := i.Groups[def]; !ok {
			i.Groups[key] = Group{
				Name:       def,
				Domains:    []string{key},
				Forwarders: forwards,
				NumRecs:    length,
				File:       fmt.Sprintf("gob-%s.dat", def),
			}
		}
	}

	for _, key := range keys {
		def := utils.ConvertKeyToId(key)
		if groupCfg, ok := i.Groups[def]; ok {
			zStore, err := groupsStore.Get(key)
			if err != nil {
				if i.log != nil {
					i.log.Errorf("GroupsBucket:: [ERROR] Error retriving group groupsStore for key: %s, Error: %v", key, err.Error())
				} else {
					fmt.Sprintf("GroupsBucket:: [ERROR] Error retriving group groupsStore for key: %s, Error: %v", key, err.Error())
				}
				return err.Error()
			}
			group, sErr := i.SaveGroup(*(zStore.(*store.GroupStoreData)), groupCfg)
			if sErr != nil {
				if i.log != nil {
					i.log.Errorf("GroupsBucket:: [ERROR] Error saving group groupsStore for key: %s, Error: %v", key, err.Error())
				} else {
					fmt.Sprintf("GroupsBucket:: [ERROR] Error saving group groupsStore for key: %s, Error: %v", key, err.Error())
				}
				return err.Error()
			}
			i.Groups[def] = group
		} else {
			if i.log != nil {
				i.log.Errorf("GroupsBucket:: [WARN ] Unable to find config for group: %s", key)
			} else {
				fmt.Sprintf("GroupsBucket:: [WARN ] Unable to find config for group: %s", key)
			}
		}
	}
	return i.SaveMeta()
}

func (i *GroupsBucket) UpdateExistingGroup(group Group) bool {
	if _, ok := i.Groups[group.Name]; ok {
		i.Groups[group.Name] = group
		return true
	}
	return false
}

func (i *GroupsBucket) ListGroups() []Group {
	var out = make([]Group, 0)
	for _, g := range i.Groups {
		out = append(out, g)
	}
	return out
}

func (i *GroupsBucket) SaveGroup(groupStore store.GroupStoreData, group Group) (Group, error) {
	var err error
	//if group == nil {
	//	return nil, rerrors.New("Unable to save nil group ...")
	//}
	//if groupStore == nil {
	//	return nil, rerrors.New("Unable to save nil group store ...")
	//}
	defer func() {
		if r := recover(); r != nil {
			message := fmt.Sprintf("Runtime Error saving groups store file, Error: %v", r)
			if i.log != nil {
				i.log.Errorf("GroupsBucket:: [ERROR]", message)
			} else {
				fmt.Println("GroupsBucket:: [ERROR]", message)
			}
			err = errors.New(message)
		}
		i.Unlock()
		i.storeMutex.Unlock()
	}()
	i.Lock()
	i.storeMutex.Lock()
	fileName := fmt.Sprintf("%s%s%s", i.Folder, __sepPath, group.File)
	if _, err = os.Stat(fileName); err == nil {
		rErr := os.Remove(fileName)
		if rErr != nil {
			if i.log != nil {
				i.log.Errorf("GroupsBucket:: [ERROR] Error saving removing old group file at: %s, Error: %v", fileName, rErr)
			} else {
				fmt.Sprintf("GroupsBucket:: [ERROR] Error saving removing old group file at: %s, Error: %v", fileName, rErr)
			}
			return Group{}, rErr
		}
	}
	f, err := os.OpenFile(fileName, os.O_RDWR+os.O_CREATE, 0666)
	if err != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error creating new group file at: %s, Error: %v", fileName, err)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error creating new group file at: %s, Error: %v", fileName, err)
		}
		return Group{}, err
	}
	defer f.Close()
	var save = groupStore.PersistentData()
	err = gob.NewEncoder(f).Encode(&save)
	if err != nil {
		if i.log != nil {
			i.log.Errorf("GroupsBucket:: [ERROR] Error saving new group file at: %s, Error: %v", fileName, err)
		} else {
			fmt.Sprintf("GroupsBucket:: [ERROR] Error saving new group file at: %s, Error: %v", fileName, err)
		}
		return Group{}, err
	}
	cache[group.Name] = GroupBlock{
		Group: group,
		Data:  groupStore,
	}
	return group, err
}
