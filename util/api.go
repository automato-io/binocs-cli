package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const apiURLBase = "https://api.binocs.sh"
const storageDir = ".binocs"
const jwtFile = "auth.json"

var binocsAPIAccessToken string

// AuthResponse comes from the API
type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

// AccessTokenStorage as in the file
type AccessTokenStorage struct {
	AccessToken string `json:"access_token"`
}

type ApiErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func handleErr(err error) {
	sentry.CaptureException(err)
	sentry.Flush(10 * time.Second)
	fmt.Println(err)
	os.Exit(1)
}

func handleWarn(msg string) {
	sentry.CaptureMessage(msg)
	fmt.Println(msg)
}

// BinocsAPI is a gateway to the binocs REST API
func BinocsAPI(path, method string, data []byte) ([]byte, error) {
	var err error
	url, err := url.Parse(apiURLBase + path)
	if err != nil {
		return []byte{}, err
	}
	respBody, respStatusCode, err := makeBinocsAPIRequest(url, method, data)
	if err != nil {
		return []byte{}, err
	}
	if respStatusCode == http.StatusNotFound {
		return []byte{}, fmt.Errorf("The requested resource does not exist")
	}
	if respStatusCode == http.StatusBadRequest {
		var apiErrorResponse ApiErrorResponse
		err = json.Unmarshal(respBody, &apiErrorResponse)
		if err != nil {
			handleErr(err)
		}
		return []byte{}, fmt.Errorf(apiErrorResponse.Status + `: ` + apiErrorResponse.Error)
	}
	if respStatusCode == http.StatusUnauthorized {
		clientKey := viper.Get("client_key")
		if clientKey == nil {
			handleErr(fmt.Errorf("Cannot read Client Key"))
		}
		_ = BinocsAPIGetAccessToken(clientKey.(string))
		respBody, respStatusCode, err = makeBinocsAPIRequest(url, method, data)
		if err != nil {
			return []byte{}, err
		}
		if respStatusCode == http.StatusUnauthorized {
			return []byte{}, fmt.Errorf("Please login to your account using `binocs login` command.")
		}
	}
	return respBody, nil
}

// BinocsAPIGetAccessToken attempts to get an access token via API and stores it
func BinocsAPIGetAccessToken(clientKey string) error {
	url, err := url.Parse(apiURLBase + "/authenticate")
	if err != nil {
		return err
	}
	postData := []byte("{\"client_key\": \"" + clientKey + "\"}")
	respBody, respStatusCode, err := makeBinocsAPIRequest(url, http.MethodPost, postData)
	if err != nil {
		return err
	}
	if respStatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid credentials")
	}
	var respJSON AuthResponse
	err = json.Unmarshal(respBody, &respJSON)
	if err != nil {
		return err
	}
	err = storeAccessToken(&respJSON)
	if err != nil {
		return err
	}
	return nil
}

// ResetAccessToken removes the auth.json file that holds access_token
func ResetAccessToken() error {
	home, err := homedir.Dir()
	if err != nil {
		handleErr(err)
	}

	if _, err = os.Stat(home + "/" + storageDir + "/" + jwtFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(home + "/" + storageDir + "/" + jwtFile)
}

func VerifyAuthenticated() {
	_, err := BinocsAPI("/authd", http.MethodGet, []byte{})
	if err != nil {
		handleErr(err)
	}
}

func makeBinocsAPIRequest(url *url.URL, method string, data []byte) ([]byte, int, error) {
	req, err := http.NewRequest(method, url.String(), bytes.NewReader(data))
	if err != nil {
		return []byte{}, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	binocsAPIAccessToken, err = loadAccessToken()
	if err != nil {
		return []byte{}, 0, err
	}
	if len(binocsAPIAccessToken) > 0 {
		req.Header.Set("Authorization", "bearer "+binocsAPIAccessToken)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		handleErr(fmt.Errorf("Cannot reach Binocs API: %v\n", err))
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, 0, err
	}
	return respBody, resp.StatusCode, nil
}

func loadAccessToken() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		handleErr(err)
	}
	data, err := os.ReadFile(home + "/" + storageDir + "/" + jwtFile)
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
		handleErr(err)
	}

	if _, err = os.Stat(home + "/" + storageDir + "/" + jwtFile); os.IsNotExist(err) {
		err = os.MkdirAll(home+"/"+storageDir, 0755)
		if err != nil {
			return err
		}
	}
	authContent := []byte("{\"access_token\": \"" + d.AccessToken + "\"}")
	return os.WriteFile(home+"/"+storageDir+"/"+jwtFile, authContent, 0600)
}
