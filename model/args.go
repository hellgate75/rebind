package model

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type ArgumentsList []string

func (i *ArgumentsList) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *ArgumentsList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *ArgumentsList) Get(index int) string {
	if index > 0 && index < len(*i) {
		return (*i)[index]
	}
	return ""
}

type ReWebConfig struct {
	DataDirPath         string `yaml:"dataDir" json:"dataDir" xml:"data-dir"`
	ConfigDirPath       string `yaml:"configDir" json:"configDir" xml:"config-dir"`
	ListenIP            string `yaml:"listenIp" json:"listenIp" xml:"listen-ip"`
	ListenPort          int    `yaml:"listenPort" json:"listenPort" xml:"listen-port"`
	DnsPipeIP           string `yaml:"dnsPipeIp" json:"dnsPipeIp" xml:"dns-pipe-ip"`
	DnsPipePort         int    `yaml:"dnsPipePort" json:"dnsPipePort" xml:"dns-pipe-port"`
	DnsPipeResponsePort int    `yaml:"dnsPipeResponsePort" json:"dnsPipeResponsePort" xml:"dns-pipe-response-port"`
	TlsCert             string `yaml:"tlsCertFilePath" json:"tlsCertFilePath" xml:"tls-cert-file-path"`
	TlsKey              string `yaml:"tlsKeyFilePath" json:"tlsKeyFilePath" xml:"tls-key-file-path"`
	EnableFileLogging   bool   `yaml:"enableFileLogging" json:"enableFileLogging" xml:"enable-file-logging"`
	LogVerbosity        string `yaml:"logVerbosity" json:"logVerbosity" xml:"log-verbosity"`
	LogFilePath         string `yaml:"logFilePath" json:"logFilePath" xml:"log-file-path"`
	EnableLogRotate     bool   `yaml:"enableLogRotate" json:"enableLogRotate" xml:"enable-log-rotate"`
	LogMaxFileSize      int64  `yaml:"logMaxFileSize" json:"logMaxFileSize" xml:"log-max-file-size"`
	LogFileCount        int    `yaml:"logFileCount" json:"logFileCount" xml:"log-file-count"`
}

type ReBindConfig struct {
	DataDirPath         string `yaml:"dataDir" json:"dataDir" xml:"data-dir"`
	ConfigDirPath       string `yaml:"configDir" json:"configDir" xml:"config-dir"`
	ListenIP            string `yaml:"listenIp" json:"listenIp" xml:"listen-ip"`
	ListenPort          int    `yaml:"listenPort" json:"listenPort" xml:"listen-port"`
	DnsPipeIP           string `yaml:"dnsPipeIp" json:"dnsPipeIp" xml:"dns-pipe-ip"`
	DnsPipePort         int    `yaml:"dnsPipePort" json:"dnsPipePort" xml:"dns-pipe-port"`
	DnsPipeResponsePort int    `yaml:"dnsPipeResponsePort" json:"dnsPipeResponsePort" xml:"dns-pipe-response-port"`
	EnableFileLogging   bool   `yaml:"enableFileLogging" json:"enableFileLogging" xml:"enable-file-logging"`
	LogVerbosity        string `yaml:"logVerbosity" json:"logVerbosity" xml:"log-verbosity"`
	LogFilePath         string `yaml:"logFilePath" json:"logFilePath" xml:"log-file-path"`
	EnableLogRotate     bool   `yaml:"enableLogRotate" json:"enableLogRotate" xml:"enable-log-rotate"`
	LogMaxFileSize      int64  `yaml:"logMaxFileSize" json:"logMaxFileSize" xml:"log-max-file-size"`
	LogFileCount        int    `yaml:"logFileCount" json:"logFileCount" xml:"log-file-count"`
}

func SaveConfig(path string, name string, config interface{}) error {
	if _, err := os.Stat(path); err != nil {
		_ = os.MkdirAll(path, 0660)
	}
	fileFullPath := fmt.Sprintf("%s%c%s.yaml", path, os.PathSeparator, name)
	bArr, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(fileFullPath, bArr, 0666)
	return err
}

func LoadConfig(path string, name string, config interface{}) error {
	if _, err := os.Stat(path); err != nil {
		return err
	}
	fileFullPath := fmt.Sprintf("%s%c%s.yaml", path, os.PathSeparator, name)
	if _, err := os.Stat(fileFullPath); err != nil {
		return err
	}
	bArr, err := ioutil.ReadFile(fileFullPath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(bArr, config)
	return err
}
