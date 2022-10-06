package http

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	_http "net/http"

	"github.com/silvernodes/silvernode-go/utils/errutil"
)

const (
	GET  string = "GET"
	POST string = "POST"

	contentTypeJson string = "application/json"
)

type Client struct {
	_client *_http.Client
}

func NewClient() *Client {
	c := new(Client)
	c._client = &_http.Client{}
	return c
}

func (c *Client) Get(url string) ([]byte, error) {
	resp, err := c._client.Get(url)
	if err != nil {
		return nil, errutil.Extend("发送请求数据失败", err)
	}
	return parseHttpResp(resp)
}

func (c *Client) PostJson(url string, obj interface{}) ([]byte, error) {
	postData, err := json.Marshal(obj)
	if err != nil {
		return nil, errutil.ExtendWithCode(-1, "请求数据Json序列化失败", err)
	}
	return c.Post(url, postData, contentTypeJson)
}

func (c *Client) Post(url string, data []byte, contentType string) ([]byte, error) {
	if contentType == "" {
		contentType = contentTypeJson
	}
	var reader io.Reader = nil
	if data != nil {
		reader = bytes.NewReader(data)
	}
	resp, err := c._client.Post(url, contentType, reader)
	if err != nil {
		return nil, errutil.ExtendWithCode(-1, "发送请求数据失败", err)
	}
	return parseHttpResp(resp)
}

func (c *Client) Request(url string, method string, data []byte, header map[string]string) ([]byte, error) {
	var reader io.Reader = nil
	if data != nil {
		reader = bytes.NewReader(data)
	}
	request, err := _http.NewRequest(method, url, reader)
	if err != nil {
		return nil, errutil.ExtendWithCode(-1, "生成http请求失败", err)
	}
	for k, v := range header {
		request.Header.Set(k, v)
	}
	resp, err := c._client.Do(request)
	if err != nil {
		return nil, errutil.ExtendWithCode(-1, "发送请求数据失败", err)
	}
	return parseHttpResp(resp)
}

func (c *Client) CloseIdle() {
	c._client.CloseIdleConnections()
}

func parseHttpResp(resp *_http.Response) ([]byte, error) {
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errutil.ExtendWithCode(-1, "解析应答数据失败", err)
	}
	if resp.StatusCode != 200 {
		return nil, errutil.NewWithCode(resp.StatusCode, resp.Status)
	}
	return respBytes, nil
}
