package app

import (
	"crypto/rand"
	"encoding/base64"
	"html/template"
	"net/http"
	"sort"

	"appengine"
	"appengine/datastore"
)

const tLister = "templates/lister"

type feedInfo struct {
	Title, SubID, URL string
}

type feedInfos []*feedInfo

func (f feedInfos) Len() int           { return len(f) }
func (f feedInfos) Less(i, j int) bool { return f[i].Title < f[j].Title }
func (f feedInfos) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func randBytes(size int) ([]byte, error) {
	bs := make([]byte, size)
	_, err := rand.Read(bs)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

const tokenSize = 16

func lister(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	uk := userKey(c)
	q := datastore.NewQuery("subscription").Ancestor(feedRoot).Filter("User =", uk).KeysOnly()
	fc, err := q.Count(c)
	if err != nil {
		handleError(w, err)
		return
	}
	feeds := make(feedInfos, 0, fc)
	iter := q.Run(c)
	for {
		var f RSS
		sk, err := iter.Next(nil)
		if err == datastore.Done {
			break
		}
		if err != nil {
			handleError(w, err)
			return
		}
		k := sk.Parent()
		err = datastore.Get(c, k, &f)
		if err != nil {
			handleError(w, err)
			return
		}
		feeds = append(feeds, &feedInfo{
			Title: f.Title,
			SubID: sk.Encode(),
			URL: k.StringID(),
		})
	}
	sort.Sort(feeds)
	token, err := randBytes(tokenSize)
	if err != nil {
		handleError(w, err)
		return
	}
	err = setUserToken(c, token)
	if err != nil {
		handleError(w, err)
		return
	}
	templ, err := template.ParseFiles(tLister, tHead)
	if err != nil {
		handleError(w, err)
		return
	}
	err = templ.Execute(w, &struct {
		Token string
		Feeds feedInfos
	}{base64.URLEncoding.EncodeToString(token), feeds})
	if err != nil {
		handleError(w, err)
		return
	}
}
