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

var updateFeed = delay.Func("updateFeed", func(c appengine.Context, fk *datastore.Key) error {
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
		return afeed.update(c, fk, true)
	} else {
		err = decoder.Decode(&rfeed)
		if err != nil {
			return err
		}
		return rfeed.update(c, fk, true)
	}
	panic("unreachable")
})

func updater(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	q := datastore.NewQuery("feed").Ancestor(feedRoot).KeysOnly()
	cu, err := q.Run(c).Cursor()
	if err != nil {
		handleError(w, err)
		return
	}
	updateBatch.Call(c, cu.String())
	fmt.Fprintln(w, "Dispatched.")
}

var (
	updateBatch *delay.Function
)

func init() {
	updateBatch = delay.Func("updateBatch", ubFunc)
}

func ubFunc(c appengine.Context, cuS string) error {
	cu, err := datastore.DecodeCursor(cuS)
	if err != nil {
		return err
	}
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	q := datastore.NewQuery("feed").Ancestor(feedRoot).KeysOnly().Start(cu)
	iter := q.Run(c)
	done := false
	for i := 0; i < 100; i++ {
		fk, err := iter.Next(c)
		if err == datastore.Done {
			done = true
			break
		}
		if err != nil {
			return err
		}
		updateFeed.Call(c, fk)
	}
	if !done {
		cu, err = iter.Cursor()
		if err != nil {
			return err
		}
		updateBatch.Call(c, cu.String())
	}
	return nil
}
