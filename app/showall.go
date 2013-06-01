package app

import (
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
)

const showRaw = `
<html>
<body>
<h1>All Items</h1>
{{range .}}
<p><a href="/feed/?{{.FeedID}}">{{.FeedTitle}}</a> <a href="{{.ItemLink}}">{{.ItemTitle}}</a></p>
{{end}}
</body>
</html>
`

type itemInfo struct {
	FeedID, FeedTitle   string
	ItemLink, ItemTitle string
}

func showAll(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("item").Order("-PubDate")
	ic, err := q.Count(c)
	if err != nil {
		handleError(w, err)
		return
	}
	infos := make([]*itemInfo, 0, ic)
	iter := q.Run(c)
	for {
		var it Item
		k, err := iter.Next(&it)
		if err == datastore.Done {
			break
		}
		if err != nil {
			handleError(w, err)
			return
		}
		toPut := &itemInfo{
			FeedID:    k.Parent().StringID(),
			ItemLink:  it.Link,
			ItemTitle: it.Title,
		}
		var f RSS
		err = datastore.Get(c, k.Parent(), &f)
		if err != nil {
			handleError(w, err)
			return
		}
		toPut.FeedTitle = f.Title
		infos = append(infos, toPut)
	}
	templ, err := template.New("all").Parse(showRaw)
	if err != nil {
		handleError(w, err)
		return
	}
	err = templ.Execute(w, infos)
	if err != nil {
		handleError(w, err)
		return
	}
}
