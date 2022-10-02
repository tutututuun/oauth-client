package main

const (
	NUM_CODE_VERITIER     = 80
	CODE_CHALLENGE_METHOD = "S256"
	RESPONSE_TYPE         = "code"
	SCOPE                 = "read"
)

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

var CODE_VERITIER = ""
var STATE = ""
var HAS_TOKEN = false
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
	//テストなので、クライアントと同じポートで上げた。
	resourceEndPoint: "http://localhost:9000/resource",
}
