package app

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"

	"appengine"
	"appengine/urlfetch"
)

func handleError(w http.ResponseWriter, err error) {
	http.Error(w, fmt.Sprintf("internal server error: %s", err),
		http.StatusInternalServerError)
}

func init() {
	http.HandleFunc("/", sender)
	http.HandleFunc("/feed/", showFeed)
	http.HandleFunc("/list/", lister)
}

const pageRaw = `
<html>
<body>
<h1><a href="{{.Link}}">{{.Title | html}}</a></h1>
{{range .Items}}
<p><a href="{{.Link}}">{{.Title | html}}</a></p>
{{end}}
`

func sender(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("page").Parse(pageRaw)
	if err != nil {
		handleError(w, err)
		return
	}
	c := appengine.NewContext(r)
	cl := urlfetch.Client(c)
	resp, err := cl.Get("http://xkcd.com/rss.xml")
	if err != nil {
		handleError(w, err)
		return
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	var parsed RSS
	if err = decoder.Decode(&parsed); err != nil {
		handleError(w, err)
		return
	}
	tmpl.Execute(w, &parsed)
}
