package app

import (
	"encoding/xml"
	"fmt"
	"net/http"

	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"appengine/urlfetch"
)

var updateFeed = delay.Func("updateFeed", func (c appengine.Context, fk *datastore.Key) error {
	cl := urlfetch.Client(c)
	resp, err := cl.Get(fk.StringID())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	var rfeed RSS
	err = datastore.Get(c, fk, &rfeed)
	if err != nil {
		return err
	}
	if rfeed.IsAtom {
		var afeed Atom
		err = decoder.Decode(&afeed)
		if err != nil {
			return err
		}
		return afeed.update(c, fk)
	} else {
		err = decoder.Decode(&rfeed)
		if err != nil {
			return err
		}
		return rfeed.update(c, fk)
	}
	panic("unreachable")
})

func updater(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	q := datastore.NewQuery("feed").Ancestor(feedRoot).KeysOnly()
	iter := q.Run(c)
	for {
		fk, err := iter.Next(c)
		if err == datastore.Done {
			break
		}
		if err != nil {
			handleError(w, err)
			return
		}
		updateFeed.Call(c, fk)
	}
	fmt.Fprintln(w, "Dispatched.")
}
