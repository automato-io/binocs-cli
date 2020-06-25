package util

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const apiURLBase = "https://binocs.sh/"

// BinocsAPI is a gateway to the binocs REST API
func BinocsAPI(path string, method string) (string, int16, error) {
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
	resp, err := client.Do(req)
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
	request, err := http.NewRequest(method, url.String(), body)
	if err != nil {
		// log.Printf("unable to create %s request for %s", method, url.String())
		return nil, err
	}
	return request, nil
}
