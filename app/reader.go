package app

import (
	"net/http"
	"fmt"

	"appengine"
	"appengine/datastore"
)

func mark(c appengine.Context, raw_key string, read bool) (string, error) {
	k, err := datastore.DecodeKey(raw_key)
	if err != nil {
		return "", err
	}
	var it Item
	err = datastore.Get(c, k, &it)
	if err != nil {
		return "", err
	}
	it.Read = read
	_, err = datastore.Put(c, k, &it)
	if err != nil {
		return "", err
	}
	return it.Link, nil
}

func reader(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	link, err := mark(c, r.URL.RawQuery, true)
	if err != nil {
		handleError(w, err)
		return
	}
	http.Redirect(w, r, link, http.StatusFound)
}

func readMarker(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	_, err := mark(c, r.URL.RawQuery, true)
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "OK")
}

func unreadMarker(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	_, err := mark(c, r.URL.RawQuery, false)
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "OK")
}
