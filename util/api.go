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
	"github.com/spf13/viper"
)

const apiURLBase = "https://api.binocs.sh"
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

// BinocsAPI is another gateway to the binocs REST API
func BinocsAPI(path, method string, data []byte) ([]byte, error) {
	var err error
	url, err := url.Parse(apiURLBase + path)
	if err != nil {
		return []byte{}, err
	}
	respBody, respStatusCode, err := makeBinocsAPIRequest(url, method, data, false)
	if err != nil {
		return []byte{}, err
	}
	if respStatusCode == http.StatusNotFound {
		return []byte{}, fmt.Errorf("requested resource does not exist")
	}
	if respStatusCode == http.StatusBadRequest {
		return []byte{}, fmt.Errorf("bad request")
	}
	if respStatusCode == http.StatusUnauthorized {
		BinocsAPIGetAccessToken(viper.Get("access_key_id").(string), viper.Get("secret_access_key").(string))
		respBody, respStatusCode, err = makeBinocsAPIRequest(url, method, data, true)
		if err != nil {
			return []byte{}, err
		}
		if respStatusCode == http.StatusUnauthorized {
			return []byte{}, fmt.Errorf("please login to your binocs account using `binocs login`")
		}
	}
	return respBody, nil
}

// BinocsAPIGetAccessToken attempts to get an access token via API and stores it
func BinocsAPIGetAccessToken(accessKeyID, secretAccessKey string) {
	postData := []byte("{\"access_key_id\": \"" + accessKeyID + "\", \"secret_access_key\": \"" + secretAccessKey + "\"}")
	respData, err := BinocsAPI("/authenticate", http.MethodPost, postData)
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
	err = storeAccessToken(&respJSON)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ResetAccessToken removes the auth.json file that holds access_token
func ResetAccessToken() error {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, err = os.Stat(home + "/" + storageDir + "/" + jwtFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(home + "/" + storageDir + "/" + jwtFile)
}

var binocsAPIAccessToken string

func makeBinocsAPIRequest(url *url.URL, method string, data []byte, forceAccessTokenReload bool) ([]byte, int, error) {
	req, err := http.NewRequest(method, url.String(), bytes.NewReader(data))
	if err != nil {
		return []byte{}, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if forceAccessTokenReload && binocsAPIAccessToken == "" {
		binocsAPIAccessToken, err = loadAccessToken()
	}
	if err != nil {
		return []byte{}, 0, err
	} else if len(binocsAPIAccessToken) > 0 {
		req.Header.Set("Authorization", "bearer "+binocsAPIAccessToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("cannot reach Binocs API: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 0, err
	}
	return respBody, resp.StatusCode, nil
}

func loadAccessToken() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

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

func storeAccessToken(d *AuthResponse) error {
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if _, err = os.Stat(home + "/" + storageDir + "/" + jwtFile); os.IsNotExist(err) {
		err = os.MkdirAll(home+"/"+storageDir, 0755)
		if err != nil {
			return err
		}
	}
	authContent := []byte("{\"access_token\": \"" + d.AccessToken + "\"}")
	return ioutil.WriteFile(home+"/"+storageDir+"/"+jwtFile, authContent, 0600)
}
