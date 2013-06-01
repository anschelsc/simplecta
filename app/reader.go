package app

import (
	"net/http"

	"appengine"
	"appengine/datastore"
)

func reader(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	k, err := datastore.DecodeKey(r.URL.RawQuery)
	if err != nil {
		handleError(w, err)
		return
	}
	var it Item
	err = datastore.Get(c, k, &it)
	it.Read = true
	_, err = datastore.Put(c, k, &it)
	if err != nil {
		handleError(w, err)
		return
	}
	http.Redirect(w, r, it.Link, http.StatusFound)
}
