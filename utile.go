package main

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"
)

func createGetParameter(param map[string]string) string {

	var getParam []string
	for k, v := range param {
		getParam = append(getParam, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(getParam, "&")
}

var templates = template.Must(template.ParseFiles("index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, value interface{}) {
	fv := reflect.ValueOf(value)
	err := templates.ExecuteTemplate(w, tmpl+".html", fv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
