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
		$(".ajax_link").click(function() {
			var button = $(this);
			var url;
			var mark = button.data("mark")
			if (mark === "read") {
				url = "/markRead/";
			} else {
				url = "/markUnread/";
			}
			$.get(url, button.data("key"), function() {
				if (mark === "read") {
					mark = "unread";
				} else {
					mark = "read";
				}
				button.text("mark " + mark);
				button.data("mark", mark)
			});
		});
		$(".read_link").bind("mouseup", function() {
			var button = $(this).siblings("button");
			if (button.data("mark") === "read") {
				button.text("mark unread");
				button.data("mark", "unread");
			}
		});
	});
</script>
<head>
<title>Complexcta!</title>
<link rel="stylesheet" href="/static/main.css">
</head>
<body>
{{.Me}}
  <a class="admin" href="{{.Logout}}">log out</a> |
  <a class="admin" href="/feeds/">manage subscriptions</a>
<br>

<p>

{{range .Infos}}
<div class="item"><a class="feedlink" href="/feed/?{{.FeedID}}">{{.FeedTitle}}</a>
<div class="item_links"><a class="read_link" href="/read/?key={{.Key}}&link={{.ItemLink}}">{{.ItemTitle}}</a> <a class="peek" href="{{.ItemLink}}">(peek)</a> <button class="ajax_link" data-mark="read" data-key="{{.Key}}">mark read</button></div></div>
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
	count := 0			//Counter to prevent iter from requesting too many datasource request.
	for {				
		count++
		if count >=101 { //I've limited the count to the first 100 post that it can find.
			break		 //Will keep looking into alternatives for handling massive amounts of rss feeds.
		}
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
