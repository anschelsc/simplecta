package app

import (
	"fmt"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/delay"
)

const sixMonths = 4380 * time.Hour

func cleanInfo(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	since := time.Now().Add(-sixMonths)
	iq := datastore.NewQuery("item").Filter("PubDate<", since).KeysOnly()
	sq := datastore.NewQuery("subscribedItem").Filter("PubDate<", since).KeysOnly()
	ic, err := iq.Count(c)
	if err != nil {
		handleError(w, err)
		return
	}
	sc, err := sq.Count(c)
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "%d old items and %d old subscribedItems", ic, sc)
}

var cleanHundredFunc *delay.Function

func cleanHundred(c appengine.Context, kind string, since time.Time, rawCursor string) {
	cursor, err := datastore.DecodeCursor(rawCursor)
	if err != nil {
		panic(err)
	}
	q := datastore.NewQuery(kind).Filter("PubDate<", since).KeysOnly().Start(cursor)
	it := q.Run(c)
	for i := 0; i < 100; i++ {
		k, err := it.Next(nil)
		if err == datastore.Done {
			return
		}
		if err != nil {
			panic(err)
		}
		err = datastore.Delete(c, k)
		if err != nil {
			panic(err)
		}
	}
	cursor, err = it.Cursor()
	if err != nil {
		panic(err)
	}
	cleanHundredFunc.Call(c, kind, since, cursor.String())
}

func init() {
	cleanHundredFunc = delay.Func("clean", cleanHundred)
}

func cleanup(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	since := time.Now().Add(-sixMonths)
	iq := datastore.NewQuery("item").Filter("PubDate<", since).KeysOnly()
	sq := datastore.NewQuery("subscribedItem").Filter("PubDate<", since).KeysOnly()
	ic, err := iq.Run(c).Cursor()
	if err != nil {
		handleError(w, err)
		return
	}
	sc, err := sq.Run(c).Cursor()
	if err != nil {
		handleError(w, err)
		return
	}
	cleanHundredFunc.Call(c, "item", since, ic.String())
	cleanHundredFunc.Call(c, "subscribedItem", since, sc.String())
	fmt.Fprintln(w, "Started...")
}
