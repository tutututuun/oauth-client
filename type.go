package main

type Client struct {
	id          string
	name        string
	redirectURL string
	secret      string
}

type Auth struct {
	authorizationEndPoint string
	tokenEndPoint         string
}

type Resource struct {
	resourceEndPoint string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

var hasToken bool = false
var tokenInfo TokenResponse

var clientInfo = Client{
	id:          "1234",
	name:        "test",
	redirectURL: "http://localhost:9000/callback",
	secret:      "secret",
}

var authSeverInfo = Auth{
	authorizationEndPoint: "http://localhost:8080/auth",
	tokenEndPoint:         "http://localhost:8080/token",
}

var protectedResource = Resource{
	//テストなので、クライアントと同じポートで上げた
	resourceEndPoint: "http://localhost:9000/resource",
}
