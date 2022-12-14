package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var tokenString string
	if len(tokenInfo.AccessToken) == 0 {
		tokenString = "Header: None \n Payload: None \n Signature: None \n"
	} else {
		tokenParts := strings.Split(tokenInfo.AccessToken, ".")
		tokenString = fmt.Sprintf("Header: %s.\n Payload: %s.\n Signature: %s\n", tokenParts[0], tokenParts[1], tokenParts[2])
	}
	renderTemplate(w, "index", struct {
		Token        string
		RefreshToken string
		Scope        string
	}{
		Token:        tokenString,
		RefreshToken: tokenInfo.RefreshToken,
		Scope:        "None",
	})
}

func clientHandler(w http.ResponseWriter, r *http.Request) {
	HAS_TOKEN = false
	STATE = uuid.New().String()

	CODE_VERITIER = randomString(NUM_CODE_VERITIER)
	code_challenge := base64URLEncode(CODE_VERITIER)

	getParam := createGetParameter(map[string]string{
		"response_type":         RESPONSE_TYPE,
		"client_id":             clientInfo.id,
		"redirect_uri":          clientInfo.redirectURL,
		"scope":                 SCOPE,
		"state":                 STATE,
		"code_challenge":        code_challenge,
		"code_challenge_method": CODE_CHALLENGE_METHOD,
	})
	endpoint := authSeverInfo.authorizationEndPoint + "?" + getParam
	fmt.Println("Request:", endpoint)
	http.Redirect(w, r, endpoint, http.StatusFound)
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	cookie, err := r.Cookie("session")
	if err != nil {
		fmt.Println("Error cookies: ", err.Error())
		sorryPage(w, err.Error())
		return
	}

	code := query.Get("code")
	state := query.Get("state")
	if state != STATE {
		fmt.Println("Error state: Invalid state")
		sorryPage(w, fmt.Sprintf("Invalid state: %s", state))
		return
	}
	postParam := url.Values{}
	postParam.Set("grant_type", "authorization_code")
	postParam.Add("code", code)
	postParam.Add("redirect_uri", clientInfo.redirectURL)
	postParam.Add("code_verifier", CODE_VERITIER)

	req, _ := http.NewRequest("POST", authSeverInfo.tokenEndPoint, strings.NewReader(postParam.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientInfo.id, clientInfo.secret)
	req.AddCookie(cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error Request:", err)
		sorryPage(w, fmt.Sprintf("Error Request: %s", err.Error()))
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		json.Unmarshal(body, &tokenInfo)
		HAS_TOKEN = true
	} else {
		fmt.Printf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode)
		sorryPage(w, fmt.Sprintf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode))
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func fetchResourceHandler(w http.ResponseWriter, r *http.Request) {
	if !HAS_TOKEN {
		fmt.Println("error: Missing access token.")
		sorryPage(w, "error: Missing access token.")
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
		w.Write(body)
		return
	}
	// ?????????????????????????????????????????????????????????????????????
	// ??????????????????????????????????????????????????????????????????????????????
	postParam := url.Values{}
	postParam.Set("grant_type", "refresh_token")
	postParam.Add("redirect_uri", clientInfo.redirectURL)
	postParam.Add("refresh_token", tokenInfo.RefreshToken)

	req, _ = http.NewRequest("POST", authSeverInfo.tokenEndPoint, strings.NewReader(postParam.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientInfo.id, clientInfo.secret)

	client = &http.Client{}
	ref_resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(ref_resp.Body)
	defer ref_resp.Body.Close()

	if ref_resp.StatusCode >= 200 && ref_resp.StatusCode < 300 {
		json.Unmarshal(body, &tokenInfo)
		HAS_TOKEN = true
	} else {
		fmt.Printf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode)
		sorryPage(w, fmt.Sprintf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode))
		tokenInfo = TokenResponse{}
		HAS_TOKEN = false
	}
	http.Redirect(w, r, "/login", http.StatusFound)
	return
}

func resourceHandler(w http.ResponseWriter, r *http.Request) {
	bearerToken := r.Header.Get("Authorization")
	match, _ := regexp.MatchString(`Bearer\s`, bearerToken)
	if !match {
		log.Println("Invalid token format.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token := strings.Replace(bearerToken, "Bearer ", "", 1)
	if err := RSAVerify(token); err != nil {
		log.Println("RSAVerify: Invalid token verify.", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	postParam := url.Values{}
	postParam.Set("token", token)

	req, _ := http.NewRequest("POST", authSeverInfo.introspectEndPoint, strings.NewReader(postParam.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("abcd", "terces")

	client := &http.Client{}
	resp, _ := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	type _introspection struct {
		Active bool `json:"active"`
	}
	var introspectionResponse _introspection
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		json.Unmarshal(body, &introspectionResponse)
	} else {
		fmt.Printf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode)
		sorryPage(w, fmt.Sprintf("Unable to fetch access token, serverrespomce: %d\n", resp.StatusCode))
		tokenInfo = TokenResponse{}
		HAS_TOKEN = false
		return
	}
	if !introspectionResponse.Active {
		log.Println("Token is not active.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenParts := strings.Split(token, ".")
	payload_b, _ := base64.URLEncoding.DecodeString(tokenParts[1])
	type _Payload struct {
		Iss string `json:"iss"`
		Sub string `json:"sub"`
		Aud string `json:"aud"`
		Iat int64  `json:"iat"`
		Exp int64  `json:"exp"`
		Jti string `json:"jti"`
	}
	var payload _Payload
	err = json.Unmarshal(payload_b, &payload)
	if err != nil {
		log.Println("Invalid payload format.")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if payload.Iss == "http://localhost:8080/" && payload.Aud == "http://localhost:9000" {
		if now := time.Now().Unix(); payload.Iat <= now && payload.Exp >= now {
			fmt.Println("Resource OK!")
			w.WriteHeader(http.StatusOK)
			renderTemplate(w, "resource", struct {
				Message string
			}{
				Message: "Resource OK! Fooooooooooooooo!",
			})
			return
		}
	}
	log.Println("Resource NG!")
	w.WriteHeader(http.StatusBadRequest)
	return
}

func revokeHandler(w http.ResponseWriter, r *http.Request) {
	postParam := url.Values{}
	postParam.Set("token", tokenInfo.AccessToken)

	req, _ := http.NewRequest("POST", authSeverInfo.revokeEndPoint, strings.NewReader(postParam.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientInfo.id, clientInfo.secret)

	client := &http.Client{}
	_, err := client.Do(req)
	if err != nil {
		fmt.Println("Error Request:", err)
		sorryPage(w, fmt.Sprintf("Error Request: %s", err.Error()))
		return
	}
	tokenInfo.AccessToken = ""
	HAS_TOKEN = false
	http.Redirect(w, r, "/login", http.StatusFound)
}

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/client", clientHandler)
	http.HandleFunc("/callback", callbackHandler)
	http.HandleFunc("/fetch_resource", fetchResourceHandler)
	http.HandleFunc("/resource", resourceHandler)
	http.HandleFunc("/revoke", revokeHandler)
	log.Fatal(http.ListenAndServe(":9000", nil))
}
