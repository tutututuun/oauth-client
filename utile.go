package main

import (
	"fmt"
	"html/template"
	"net/http"
	"reflect"
)

func createGetParameter(param map[string]string) string {

	var getParam string
	for k, v := range param {
		getParam += fmt.Sprintf("%s=%s", k, v)
	}
	return getParam
}

var templates = template.Must(template.ParseFiles("login.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, value interface{}) {
	fv := reflect.ValueOf(value)
	err := templates.ExecuteTemplate(w, tmpl+".html", fv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
