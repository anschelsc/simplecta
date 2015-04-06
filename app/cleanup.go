package app

import (
	"fmt"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
)

const sixMonths = 4380 * time.Hour

func cleanup(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	since := time.Now().Add(-sixMonths)
	iq := datastore.NewQuery("item").Filter("PubDate<", since).KeysOnly()
	sq := datastore.NewQuery("subscribedItem").Filter("PubDate<", since).KeysOnly()
	iks, err := iq.GetAll(c, nil)
	if err != nil {
		handleError(w, err)
		return
	}
	sks, err := sq.GetAll(c, nil)
	if err != nil {
		handleError(w, err)
		return
	}
	if err = datastore.DeleteMulti(c, iks); err != nil {
		handleError(w, err)
		return
	}
	if err = datastore.DeleteMulti(c, sks); err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintln(w, "Success(?)")
}
