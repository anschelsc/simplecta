package app

import (
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
	"appengine/user"
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
Logged in as {{.Me}}. <a href="{{.Logout}}">(logout)</a>
<form action="/addAtom/" method="get">
Add Atom feed: <input type="text" name="url"> <input type="submit" value="Add">
</form>
<form action="/addRSS/" method="get">
Add RSS feed: <input type="text" name="url"> <input type="submit" value="Add">
</form>
<h1>All Items (<a href="/feeds/">view feeds</a>)</h1>
{{range .Infos}}
<p><a href="/feed/?{{.FeedID}}">{{.FeedTitle}}</a> <a href="/read/?key={{.Key}}&link={{.ItemLink}}">{{.ItemTitle}}</a> <a href="{{.ItemLink}}">(keep unread)</a> <button class="ajax_read_link" data-key="{{.Key}}">mark read</button><button class="ajax_unread_link" data-key="{{.Key}}">mark unread</button></p>
{{end}}
</body>
</html>
`

type itemInfo struct {
	FeedID, FeedTitle   string
	ItemLink, ItemTitle string
	Key                 string
}

type showAllData struct {
	Infos  []*itemInfo
	Me     string
	Logout string
}

func showAll(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery("subscribedItem").Ancestor(userKey(c)).KeysOnly().Order("PubDate")
	ic, err := q.Count(c)
	if err != nil {
		handleError(w, err)
		return
	}
	infos := make([]*itemInfo, 0, ic)
	iter := q.Run(c)
	for {
		sk, err := iter.Next(empty)
		if err == datastore.Done {
			break
		}
		if err != nil {
			handleError(w, err)
			return
		}
		k, err := datastore.DecodeKey(sk.StringID())
		if err != nil {
			handleError(w, err)
			return
		}
		var it Item
		err = datastore.Get(c, k, &it)
		if err != nil {
			handleError(w, err)
			return
		}
		toPut := &itemInfo{
			FeedID:    k.Parent().StringID(),
			ItemLink:  it.Link,
			ItemTitle: it.Title,
			Key:       sk.Encode(),
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
	me := user.Current(c).String()
	logout, err := user.LogoutURL(c, r.URL.String())
	if err != nil {
		handleError(w, err)
		return
	}
	err = templ.Execute(w, &showAllData{Infos: infos, Me: me, Logout: logout})
	if err != nil {
		handleError(w, err)
		return
	}
}
