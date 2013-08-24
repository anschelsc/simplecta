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

func watashi(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	ks, err := datastore.NewQuery("feed").KeysOnly().GetAll(c, nil)
	if err != nil {
		handleError(w, err)
		return
	}
	for _, k := range ks {
		err = subscribe(c, k, false)
		if err != nil {
			handleError(w, err)
			return
		}
	}
	ks, err = datastore.NewQuery("item").Filter("Read =", false).KeysOnly().GetAll(c, nil)
	if err != nil {
		handleError(w, err)
		return
	}
	for _, k := range ks {
		propagate.Call(c, k)
		fmt.Fprintln(w, k)
	}
	fmt.Fprintln(w, "OK!")
}

func convertSubs(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	ks, err := datastore.NewQuery("subscription").KeysOnly().GetAll(c, nil)
	subs := make([]subscription, len(ks))
	for i, k := range ks {
		subs[i].User, err = datastore.DecodeKey(k.StringID())
		if err != nil {
			handleError(w, err)
			return
		}
	}
	_, err = datastore.PutMulti(c, ks, subs)
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintln(w, "OK!")
}
