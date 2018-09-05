//admin.go - defines the admin op API of radosgw service

package radosgw

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultAdminPrefix = "/admin"
)

var globalHttpClient = &http.Client{Timeout: time.Second * 300}

// Client stands for the client to administrate the radosgw service
type Client struct {
	accessKeyId     string
	secretAccessKey string
	endpoint        string
	prefix          string
}

func NewClient(endpoint, ak, sk string) (*Client, error) {
	if len(endpoint) == 0 || len(ak) == 0 || len(sk) == 0 {
		return nil, fmt.Errorf("endpoint, ak and sk should not be empty")
	}
	if strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint[:len(endpoint)-1]
	}
	return &Client{ak, sk, endpoint, defaultAdminPrefix}, nil
}

func (c *Client) SetPrefix(p string) { c.prefix = p }

func (c *Client) sendRequest(method, uri string, args url.Values, headers map[string]string,
	body io.ReadCloser) (respBody []byte, status int, err error) {
	// Create http request and set the input params
	req := &http.Request{
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Method:     method,
		Header:     make(http.Header),
	}
	var requestUrl string
	if len(c.prefix) != 0 {
		requestUrl = fmt.Sprintf("%s%s%s?%s", c.endpoint, c.prefix, uri, args.Encode())
	} else {
		requestUrl = fmt.Sprintf("%s%s?%s", c.endpoint, uri, args.Encode())
	}
	req.URL, err = url.Parse(requestUrl)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	req.Host = req.URL.Host
	if headers != nil {
		for k := range headers {
			req.Header.Add(k, headers[k])
		}
	}
	if body != nil {
		req.Body = body
	}

	// Calculate the authorization string for AWS4 request to s3 service
	req = Sign(req, c.accessKeyId, c.secretAccessKey)

	// Do send the http request and get the result
	resp, err := globalHttpClient.Do(req)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	status = resp.StatusCode
	return
}
