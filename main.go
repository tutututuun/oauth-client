package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func clientHandler(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String()
	getParam := createGetParameter(map[string]string{
		"response_type": "code",
		"client_id":     clientInfo.id,
		"redirect_url":  clientInfo.redirectURL,
		"state":         state,
	})
	endpoint := authSeverInfo.authorizationEndPoint + "&" + getParam
	fmt.Println("Request:", endpoint)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Println("Error Request:", err)
		return
	}

	var client *http.Client = &http.Client{}
	client.Do(req)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	code := query.Get("code")
	postParam := url.Values{}
	postParam.Set("grant_type", "authorization_code")
	postParam.Add("code", code)
	postParam.Add("redirect_uri", clientInfo.redirectURL)

	req, _ := http.NewRequest("POST", authSeverInfo.tokenEndPoint, strings.NewReader(postParam.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// server側は(r *http.Request)の引数から
	// clientID, clientSecret, ok := r.BasicAuth()で検証できる。
	req.SetBasicAuth(clientInfo.id, clientInfo.secret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error Request:", err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	// クローズしないとTCPコネクションが開いた状態のままになる。
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		json.Unmarshal(body, &tokenInfo)
		hasToken = true
	} else {
		fmt.Printf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode)
		return
	}
}

func resourceHandler(w http.ResponseWriter, r *http.Request) {
	if !hasToken {
		fmt.Println("error: Missing access token.")
		return
	}
	req, _ := http.NewRequest("POST", protectedResource.resourceEndPoint, nil)
	req.Header.Add("Authorization", "Bearer "+tokenInfo.AccessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error Request:", err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	// クローズしないとTCPコネクションが開いた状態のままになる。
	defer resp.Body.Close()
	fmt.Println(body)
}

func main() {
	http.HandleFunc("/client", clientHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/fetch_resource", resourceHandler)
	log.Fatal(http.ListenAndServe(":10080", nil))
}
