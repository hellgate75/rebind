package data

import (
	"github.com/hellgate75/rebind/log"
	"net"
	"sync"
)

type Group struct {
	Name       string        `yaml:"name" json:"name" xml:"name"`
	File       string        `yaml:"file" json:"file" xml:"file"`
	NumRecs    int64         `yaml:"numberOfRecords" json:"numberOfRecords" xml:"number-of-records"`
	Domains    []string      `yaml:"domains,omitempty" json:"domains,omitempty" xml:"domains,omitempty"`
	Forwarders []net.UDPAddr `yaml:"forwarders,omitempty" json:"forwarders,omitempty" xml:"forwarders,omitempty"`
}

type GroupsBucket struct {
	sync.Mutex
	storeMutex sync.Mutex
	log        log.Logger
	Folder     string           `yaml:"dataFolder" json:"dataFolder" xml:"data-folder"`
	Groups     map[string]Group `yaml:"groups" json:"groups" xml:"groups"`
}

type groupBucketPersitence struct {
	Groups map[string]Group `yaml:"groups" json:"groups" xml:"groups"`
}

func NewGroupsBucket(folder string, log log.Logger) GroupsBucket {
	return GroupsBucket{
		Folder: folder,
		log:    log,
		Groups: make(map[string]Group),
	}
}
