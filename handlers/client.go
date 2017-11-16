package handlers

import (
	"net/http"
	"encoding/json"
	"strings"
	"fmt"
	"io/ioutil"

	"github.com/toolsparty/groxy/conf"
	"github.com/toolsparty/groxy/logger"
	"github.com/toolsparty/groxy/encrypt"
)

// http(s) handler
type Client struct {
	Ssl bool
	Config *conf.Configuration
	Logger *logger.FileLog
	Encoder *encrypt.Crypt
}

func (h Client) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	h.Logger.Write("Handle request:", request.Method, request.RequestURI)

	url := strings.Replace(request.URL.String(), h.GetProto() + "://" + request.Host, "", -1)
	req := &ClientRequest{Method: request.Method, Host: request.Host, Path: url, Proto: h.GetProto()}

	switch request.Method {
	case "GET":
		// skip
		break
	case "POST":
		// add post data
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			h.Logger.WriteError(err)
			h.ErrorResponse(response)
			return
		}
		defer request.Body.Close()
		req.Body = body
		req.ContentType = request.Header.Get("Content-Type")
	default:
		h.ErrorResponse(response)
		return
	}

	// add request headers
	headers := make(map[string][]string)
	for i, header := range request.Header {
		headers[i] = append(headers[i], header[0])
	}

	req.Headers = headers

	data, err := h.Encode(req)
	if err != nil {
		h.Logger.WriteError(err)
		h.ErrorResponse(response)
		return
	}

	// send data to server and get response
	resp, err := http.Post(h.Config.Server.GetUrl("/load").String(), "text/plain", strings.NewReader(string(data)))
	if err != nil {
		h.Logger.WriteError(err)
		h.ErrorResponse(response)
		return
	}

	// read response
	respData, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		h.Logger.WriteError(err)
		h.ErrorResponse(response)
		return
	}

	sr, err := h.Decode(respData)
	if err != nil {
		h.Logger.WriteError(err)
		h.ErrorResponse(response)
		return
	}

	for key, header := range sr.Headers {
		for _, value := range header {
			response.Header().Add(key, string(value))
		}
	}

	// output response
	response.WriteHeader(sr.Status)
	fmt.Fprint(response, string(sr.Body))
}

// for all errors
func (h Client) ErrorResponse(response http.ResponseWriter) {
	response.WriteHeader(404)
	fmt.Fprint(response, "Not Found")
	return
}

func (h Client) GetProto() string {
	var proto string

	if h.Ssl {
		proto = "https"
	} else {
		proto = "http"
	}

	return proto
}

// encode request
func (h Client) Encode(req *ClientRequest) ([]byte, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	return h.Encoder.Encrypt(data)
}

// decode response
func (h Client) Decode(body []byte) (*ServerResponse, error) {
	sr := &ServerResponse{}
	decData, err := h.Encoder.Decrypt(body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(decData, &sr)
	if err != nil {
		return nil, err
	}

	return sr, nil
}

// create client
func NewClient(conf *conf.Configuration, logger *logger.FileLog, ssl bool) (*Client, error) {
	encoder, err := encrypt.NewCrypt(conf.Encryption.GetKey(), conf.Encryption.GetIv())
	if err != nil {
		return nil, err
	}

	c := &Client{Config: conf, Ssl: ssl, Encoder: encoder, Logger: logger}
	return c, nil
}

// request structure
type ClientRequest struct {
	Method string `json:"method"`
	Host string `json:"host"`
	Path string `json:"path"`
	Proto string `json:"scheme"`
	Body []byte `json:"body"`
	ContentType string `json:"content_type"`
	Headers Headers `json:"headers"`
}

func (cr ClientRequest) GetUrl(prefix string) string {
	return cr.Proto + "://" + prefix + cr.Host + cr.Path
}

func (cr ClientRequest) GetContentType() string {
	if cr.ContentType == "" {
		cr.ContentType = "text/plain"
	}

	return cr.ContentType
}