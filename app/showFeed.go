package app

import (
	"html/template"
	"net/http"
	"net/url"

	"appengine"
	"appengine/datastore"
)

const feedPageRaw = `
<html>
<body>
<h1><a href="{{.Link}}">{{.Title | html}}</a></h1>
{{range .Items}}
<p><a href="{{.Link}}">{{.Title | html}}</a></p>
{{end}}
</body>
</html>
`

func showFeed(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	url, err := url.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		handleError(w, err)
		return
	}
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	fk := datastore.NewKey(c, "feed", url, 0, feedRoot)
	f := new(RSS)
	err = datastore.Get(c, fk, f)
	if err != nil {
		handleError(w, err)
		return
	}
	_, err = datastore.NewQuery("item").Ancestor(fk).Order("-PubDate").GetAll(c, &f.Items)
	if err != nil {
		handleError(w, err)
		return
	}
	templ, err := template.New("showFeed").Parse(feedPageRaw)
	if err != nil {
		handleError(w, err)
		return
	}
	err = templ.Execute(w, f)
	if err != nil {
		handleError(w, err)
		return
	}
}
