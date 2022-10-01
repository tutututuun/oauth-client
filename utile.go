package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
)

func randomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func base64URLEncode(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func createGetParameter(param map[string]string) string {

	var getParam []string
	for k, v := range param {
		getParam = append(getParam, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(getParam, "&")
}

var templates = template.Must(template.ParseFiles("index.html", "sorry.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, value interface{}) {
	fv := reflect.ValueOf(value)
	err := templates.ExecuteTemplate(w, tmpl+".html", fv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}
func sorryPage(w http.ResponseWriter, msg string) {
	renderTemplate(w, "sorry", struct {
		Message string
	}{
		Message: msg,
	})
	return
}
