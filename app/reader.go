package app

import (
	"fmt"
	"net/http"

	"appengine"
	"appengine/datastore"
)

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
	c := appengine.NewContext(r)
	err := mark(c, r.URL.Query()["key"][0], true)
	if err != nil {
		handleError(w, err)
		return
	}
	link := r.URL.Query()["link"][0]
	http.Redirect(w, r, link, http.StatusFound)
}

func readMarker(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := mark(c, r.URL.RawQuery, true)
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "OK")
}

func unreadMarker(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	err := mark(c, r.URL.RawQuery, false)
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "OK")
}
