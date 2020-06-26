package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/mitchellh/go-homedir"
)

const apiURLBase = "https://binocs.sh"
const storageDir = ".binocs"
const jwtFile = "auth.json"

// AuthResponse comes from the API
type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

// AccessTokenStorage as in the file
type AccessTokenStorage struct {
	AccessToken string `json:"access_token"`
}

// BinocsAPI2 is another gateway to the binocs REST API
func BinocsAPI2(path, method string, data []byte) ([]byte, error) {
	var err error
	url, err := url.Parse(apiURLBase + path)
	if err != nil {
		return []byte{}, err
	}
	respBody, respStatusCode, err := makeBinocsAPIRequest(url, method, data)
	if err != nil {
		return []byte{}, err
	}
	if respStatusCode == http.StatusUnauthorized {
		// authRespBody, authRespStatusCode, err := BinocsAPIGetAccessToken()
		// 	// login with keys and save jwt
		// 	// if ok then proceed to # 1
		// 	// if nok then display error and suggest authorizing using /login (fixing incorrect or missing credentials)
	}
	return respBody, nil
}

// BinocsAPIGetAccessToken attempts to get an access token via API and stores it
func BinocsAPIGetAccessToken(accessKeyID, secretAccessKey string) {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	postData := []byte("{\"access_key_id\": \"" + accessKeyID + "\", \"secret_access_key\": \"" + secretAccessKey + "\"}")
	respData, err := BinocsAPI2("/authenticate", http.MethodPost, postData)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var respJSON AuthResponse
	err = json.Unmarshal(respData, &respJSON)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = storeAccessToken(home, &respJSON)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func makeBinocsAPIRequest(url *url.URL, method string, data []byte) ([]byte, int, error) {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req, err := http.NewRequest(method, url.String(), bytes.NewReader(data))
	if err != nil {
		return []byte{}, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	accessToken, err := loadAccessToken(home)
	if err != nil {
		return []byte{}, 0, err
	} else if len(accessToken) > 0 {
		req.Header.Set("Authorization", "bearer "+accessToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, 0, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 0, err
	}
	return respBody, resp.StatusCode, nil
}

func loadAccessToken(home string) (string, error) {
	data, err := ioutil.ReadFile(home + "/" + storageDir + "/" + jwtFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var accessTokenData AccessTokenStorage
	err = json.Unmarshal(data, &accessTokenData)
	if err != nil {
		return "", err
	}
	return accessTokenData.AccessToken, nil
}

func storeAccessToken(home string, d *AuthResponse) error {
	var err error
	if _, err = os.Stat(home + "/" + storageDir + "/" + jwtFile); os.IsNotExist(err) {
		err = os.MkdirAll(home+"/"+storageDir, 0755)
		if err != nil {
			return err
		}
	}
	authContent := []byte("{\"access_token\": \"" + d.AccessToken + "\"}")
	return ioutil.WriteFile(home+"/"+storageDir+"/"+jwtFile, authContent, 0600)
}
