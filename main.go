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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index", struct {
		Token        string
		RefreshToken string
		Scope        string
	}{
		Token:        tokenInfo.AccessToken,
		RefreshToken: tokenInfo.RefreshToken,
		Scope:        "None",
	})
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String()
	getParam := createGetParameter(map[string]string{
		"response_type": "code",
		"client_id":     clientInfo.id,
		"redirect_uri":  clientInfo.redirectURL,
		"score":         "read",
		"state":         state,
	})
	endpoint := authSeverInfo.authorizationEndPoint + "?" + getParam
	fmt.Println("Request:", endpoint)
	http.Redirect(w, r, endpoint, http.StatusFound)
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
	req.SetBasicAuth(clientInfo.id, clientInfo.secret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error Request:", err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		json.Unmarshal(body, &tokenInfo)
		hasToken = true
	} else {
		fmt.Printf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func fetchResourceHandler(w http.ResponseWriter, r *http.Request) {
	if !hasToken {
		fmt.Println("error: Missing access token.")
		return
	}
	req, _ := http.NewRequest("POST", protectedResource.resourceEndPoint, nil)
	req.Header.Add("Authorization", "Bearer "+tokenInfo.AccessToken)
	client := &http.Client{}
	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		fmt.Println(body)
		return
	}
	// 保護対象リソースからデータを取得できない時は、
	// リフレッシュトークンを使ってトークンを再度取得する。
	postParam := url.Values{}
	postParam.Set("grant_type", "refresh_token")
	postParam.Add("redirect_uri", clientInfo.redirectURL)
	postParam.Add("refresh_token", tokenInfo.RefreshToken)

	req, _ = http.NewRequest("POST", authSeverInfo.tokenEndPoint, strings.NewReader(postParam.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientInfo.id, clientInfo.secret)

	client = &http.Client{}
	client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		json.Unmarshal(body, &tokenInfo)
		hasToken = true
	} else {
		fmt.Printf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
	return
}

func resourceHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: 保護対象リソースは、OAuthサーバにトークンの検証を行う。
	fmt.Println("Resource OK!")
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/client", clientHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/fetch_resource", fetchResourceHandler)
	http.HandleFunc("/resource", resourceHandler)
	log.Fatal(http.ListenAndServe(":9000", nil))
}
