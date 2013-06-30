package app

import (
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
)

const showRaw = `
<html>
<script type="text/javascript" src="/static/jquery-1.10.1.min.js"></script>
<script type="text/javascript">
	$(function() {
		$(".ajax_read_link").click(function() {
			$.get("/markRead/", $(this).data("key"));
		});
		$(".ajax_unread_link").click(function() {
			$.get("/markUnread/", $(this).data("key"));
		});
	});
</script>
<body>
<h1>All Items (<a href="/list/">by feed</a>)</h1>
{{range .}}
<p><a href="/feed/?{{.FeedID}}">{{.FeedTitle}}</a> <a href="/read/?{{.Key}}">{{.ItemTitle}}</a> <a href="{{.ItemLink}}">(keep unread)</a> <a class="ajax_read_link" data-key="{{.Key}}">(mark read)</a> <a class="ajax_unread_link" data-key="{{.Key}}">(mark unread) </a></p>
{{end}}
</body>
</html>
`

type itemInfo struct {
	FeedID, FeedTitle   string
	ItemLink, ItemTitle string
	Key                 string
}

func showAll(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("item").Filter("Read =", false).Order("PubDate")
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
			Key:       k.Encode(),
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
