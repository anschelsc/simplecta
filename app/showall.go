package app

import (
	"code.google.com/p/go-uuid/uuid"
	"html/template"
	"net/http"

	"appengine"
	"appengine/datastore"
	"appengine/user"
)

const (
	tFile = "templates/showall"
	tHead = "templates/head"
)

type itemInfo struct {
	FeedID, FeedTitle   string
	ItemLink, ItemTitle string
	Key                 string
}

type showAllData struct {
	Infos  []*itemInfo
	Me     string
	Logout string
	Client string
}

func showAll(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if err := logUser(c); err != nil {
		handleError(w, err)
		return
	}
	// The limit of 100 keeps the page load time down to something reasonable.
	// In the future there should be a "### items remaining _next_" link somewhere.
	q := datastore.NewQuery("subscribedItem").Ancestor(userKey(c)).KeysOnly().Order("PubDate").Limit(100)
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
	templ, err := template.ParseFiles(tFile, tHead)
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
	client := uuid.New()
	err = templ.Execute(w, &showAllData{Infos: infos, Me: me, Logout: logout, Client: client})
	if err != nil {
		handleError(w, err)
		return
	}
}
