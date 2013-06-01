package app

import (
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
)

const listerRaw = `
<html>
<body>
<h1><a href="/all/">All Feeds</a></h1>
{{range .}}
<p><a href="/feed/?{{.ID }}">{{.Title}}</a></p>
{{end}}
</body>
</html>
`

type feedInfo struct {
	ID, Title string
}

func lister(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	q := datastore.NewQuery("feed").Ancestor(feedRoot).Order("Title")
	fc, err := q.Count(c)
	if err != nil {
		handleError(w, err)
		return
	}
	data := make([]*feedInfo, 0, fc)
	iter := q.Run(c)
	for {
		var f RSS
		k, err := iter.Next(&f)
		if err == datastore.Done {
			break
		}
		if err != nil {
			handleError(w, err)
			return
		}
		data = append(data, &feedInfo{ID: k.StringID(), Title: f.Title})
	}
	templ, err := template.New("lister").Parse(listerRaw)
	if err != nil {
		handleError(w, err)
		return
	}
	err = templ.Execute(w, data)
	if err != nil {
		handleError(w, err)
		return
	}
}
