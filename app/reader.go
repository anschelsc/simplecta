package app

import (
	"net/http"

	"appengine"
	"appengine/channel"
	"appengine/datastore"
)

type alert struct {
	Ind  string
	Read bool
}

func markAlert(c appengine.Context, raw_key string, read bool, client, ind string) error {
	err := mark(c, raw_key, read)
	if err != nil {
		return err
	}
	return channel.SendJSON(c, client, &alert{Ind: ind, Read: read})
}

func mark(c appengine.Context, raw_key string, read bool) error {
	sk, err := datastore.DecodeKey(raw_key)
	if err != nil {
		return err
	}
	if read {
		return datastore.Delete(c, sk)
	}
	ik, err := datastore.DecodeKey(sk.StringID())
	if err != nil {
		return err
	}
	var it Item
	err = datastore.Get(c, ik, &it)
	if err != nil {
		return err
	}
	_, err = datastore.Put(c, sk, &subscribedItem{it.PubDate})
	return err
}

func reader(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, r.FormValue("link"), http.StatusFound)
	wrMark(w, r, true)
}

func readMarker(w http.ResponseWriter, r *http.Request) {
	wrMark(w, r, true)
}

func unreadMarker(w http.ResponseWriter, r *http.Request) {
	wrMark(w, r, false)
}

func wrMark(w http.ResponseWriter, r *http.Request, read bool) {
	c := appengine.NewContext(r)
	key := r.FormValue("key")
	client := r.FormValue("client")
	index := r.FormValue("index")
	err := markAlert(c, key, read, client, index)
	if err != nil {
		handleError(w, err)
	}
}
