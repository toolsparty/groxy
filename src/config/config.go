package config

import (
	"encoding/json"
	"os"
	"net/url"
)

type Interface interface {
	Load() error
}

type Configuration struct {
	isLoaded bool
	HttpHost string `json:"http_host"`
	HttpPort string `json:"http_port"`
	HttpsHost string `json:"https_host"`
	HttpsPort string `json:"https_port"`
	Server Server `json:"server"`
	Prefix string `json:"prefix"`
	Logger bool `json:"logger"`
	Encryption Encryption `json:"encryption"`
}

func (m *Configuration) Load() error {
	err := LoadFile("./conf/config.json", m)
	if err != nil {
		return err
	}

	m.isLoaded = true
	return nil
}

func (m *Configuration) GetAddr() string {
	return m.HttpHost + ":" + m.HttpPort
}

func (m *Configuration) GetAddrs() string {
	return m.HttpsHost + ":" + m.HttpsPort
}

func (m *Configuration) GetURL() string {
	return "http://" + m.HttpHost + ":" + m.HttpPort
}

type Encryption struct {
	Key string `json:"key"`
	Iv string `json:"iv"`
}

func (e *Encryption) GetKey() []byte {
	return []byte(e.Key)
}

func (e *Encryption) GetIv() []byte {
	return []byte(e.Iv)
}

type Server struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

func (s Server) GetUrl(path string) *url.URL {
	return &url.URL{Scheme: "http", Host: s.Host + ":" + s.Port, Path: path}
}

func (s Server) GetAddr() string {
	return s.Host + ":" + s.Port
}

func LoadFile(fileName string, v Interface) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	return decoder.Decode(&v)
}

var client Configuration = Configuration{}

func NewConfiguration() *Configuration {
	if client.isLoaded == true {
		return &client
	}

	client.Load()
	return &client
}