package util

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiURLBase = "https://binocs.sh"

// BinocsAPI is a gateway to the binocs REST API
func BinocsAPI(path string, method string, data url.Values) (string, int16, error) {
	url, err := url.Parse(apiURLBase + path)
	if err != nil {
		return "", 0, err
	}
	req, err := createRequest(url, method)
	if err != nil {
		return "", 0, err
	}
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			DualStack: true,
		}).DialContext,
		MaxIdleConns:        100,
		IdleConnTimeout:     20 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 30,
	}

	var resp *http.Response
	if method == http.MethodPost && len(data) > 0 {
		resp, err = client.PostForm(url.String(), data)
	} else {
		resp, err = client.Do(req)
	}
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	return string(respBody), int16(resp.StatusCode), nil
}

func createRequest(url *url.URL, method string) (*http.Request, error) {
	body := strings.NewReader("")
	return http.NewRequest(method, url.String(), body)
}

// BinocsAPI2 is another gateway to the binocs REST API
func BinocsAPI2(path, method string, data []byte) ([]byte, error) {
	var err error
	url, err := url.Parse(apiURLBase + path)
	if err != nil {
		return []byte{}, err
	}
	req, err := http.NewRequest(method, url.String(), bytes.NewReader(data))
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return respBody, nil
}
