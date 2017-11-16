package handlers

import (
	"net/http"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"crypto/tls"
	"errors"
	"strings"

	"github.com/toolsparty/groxy/conf"
	"github.com/toolsparty/groxy/logger"
	"github.com/toolsparty/groxy/encrypt"
)
// http handler
type Server struct {
	Config *conf.Configuration
	Logger *logger.FileLog
	Encoder *encrypt.Crypt
}

func (h Server) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	h.Logger.Write("Handle client request:", request.Method)

	// handle only post-requests
	switch request.Method {
	case "POST":
		data, err := ioutil.ReadAll(request.Body)
		defer request.Body.Close()
		if err != nil {
			h.Logger.WriteError(err)
			goto def
		}

		cr, err := h.Decode(data)

		switch cr.Method {
		case "GET":
			err := h.HandleGet(response, *cr)
			if err != nil {
				h.Logger.WriteError(err)
				goto def
			}
		case "POST":
			err := h.HandlePost(response, *cr)
			if err != nil {
				h.Logger.WriteError(err)
				goto def
			}
		default:
			goto def
		}

		// if success
		return
	default:
		goto def
	}

	// default response
	def:
		var body []byte
		headers := make(map[string][]string)
		sr := &ServerResponse{Body: body, Headers: headers}
		jsonData, _ := json.Marshal(sr)
		respData, _ := h.Encoder.Encrypt(jsonData)
		response.WriteHeader(404)
		fmt.Fprint(response, string(respData))
		return
}

// handle get requests
func (h Server) HandleGet(response http.ResponseWriter, cr ClientRequest) error {
	h.Logger.Write("Get data:", cr.GetUrl(h.Config.Prefix))
	// create get request
	request, err := http.NewRequest("GET", cr.GetUrl(h.Config.Prefix), nil)
	if err != nil {
		return err
	}

	return h.SendRequest(response, request, cr)
}

// handle post requests
func (h Server) HandlePost(response http.ResponseWriter, cr ClientRequest) error {
	h.Logger.Write("Post data:", cr.GetUrl(h.Config.Prefix))
	body := string(cr.Body)
	request, err := http.NewRequest("POST", cr.GetUrl(h.Config.Prefix), strings.NewReader(body))
	if err != nil {
		return err
	}

	return h.SendRequest(response, request, cr)
}

func (h Server) SendRequest(response http.ResponseWriter, request *http.Request, cr ClientRequest) error {
	client, err := getHttpClient(cr)
	if err != nil {
		return err
	}

	addHeaders(request, cr)

	resp, err := client.Do(request)
	if err != nil {
		return err
	}

	respData, err := h.Encode(resp)
	if err != nil {
		return err
	}

	// output
	response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(response, string(respData))
	return nil
}

func (h Server) Encode(response *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}

	headers := make(map[string][]string)
	for i, header := range response.Header {
		headers[i] = append(headers[i], header[0])
	}

	sr := &ServerResponse{Body: body, Headers: headers, Status: response.StatusCode}
	jsonData, err := json.Marshal(sr)
	if err != nil {
		return nil, err
	}

	return h.Encoder.Encrypt(jsonData)
}

func (h Server) Decode(data []byte) (*ClientRequest, error) {
	cr := &ClientRequest{}
	decData, err := h.Encoder.Decrypt(data)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decData, &cr)
	if err != nil {
		return nil, err
	}

	return cr, nil
}

func NewServer(conf *conf.Configuration, logger *logger.FileLog) (*Server, error) {
	encoder, err := encrypt.NewCrypt(conf.Encryption.GetKey(), conf.Encryption.GetIv())
	if err != nil {
		return nil, err
	}

	return &Server{Config: conf, Encoder: encoder, Logger: logger}, nil
}

func getHttpClient(cr ClientRequest) (*http.Client, error) {
	var client *http.Client
	var err error = nil

	switch cr.Proto {
	case "http":
		client = &http.Client{}
	case "https":
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}
		client = &http.Client{Transport: transCfg}
	default:
		client = nil
		err = errors.New("unknown protocol")
	}

	return client, err
}

// adding headers to request
func addHeaders(request *http.Request, cr ClientRequest)  {
	for key, header := range cr.Headers {
		for _, value := range header {
			request.Header.Add(key, string(value))
		}
	}
}

type Headers map[string][]string

// server response structure
type ServerResponse struct {
	Body []byte `json:"body"`
	Headers Headers `json:"headers"`
	Status int `json:"status"`
}