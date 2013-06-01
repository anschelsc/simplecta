package app

import (
	"fmt"
	"net/http"

	"appengine"
	"appengine/datastore"
)

func rehasher(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var items []*Item
	ks, err := datastore.NewQuery("item").GetAll(c, &items)
	if err != nil {
		handleError(w, err)
		return
	}
	_, err = datastore.PutMulti(c, ks, items)
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintln(w, "OK!")
}
