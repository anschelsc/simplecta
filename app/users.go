package app

import (
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"appengine/user"
)

type subscribedItem struct {
	PubDate time.Time
}

type subscription struct {
	User *datastore.Key
}

var empty = new(struct{})

func userKey(c appengine.Context) *datastore.Key {
	userRoot := datastore.NewKey(c, "userRoot", "userRoot", 0, nil)
	return datastore.NewKey(c, "user", user.Current(c).ID, 0, userRoot)
}

func subscribe(c appengine.Context, fk *datastore.Key, populate bool) error {
	uk := userKey(c)
	_, err := datastore.Put(c, datastore.NewKey(c, "subscription", uk.Encode(), 0, fk), &subscription{uk})
	if err != nil {
		return err
	}
	if !populate {
		return nil
	}
	iter := datastore.NewQuery("item").Ancestor(fk).Order("-PubDate").Limit(10).Run(c)
	var k *datastore.Key
	var it Item
	for k, err = iter.Next(&it); err == nil; k, err = iter.Next(&it) {
		si := subscribedItem{it.PubDate}
		_, err = datastore.Put(c, datastore.NewKey(c, "subscribedItem", k.Encode(), 0, uk), &si)
		if err != nil {
			return err
		}
	}
	if err != datastore.Done {
		return err
	}
	return nil
}

var propagate = delay.Func("propagate", func(c appengine.Context, ik *datastore.Key) error {
	var it Item
	err := datastore.Get(c, ik, &it)
	if err != nil {
		return err
	}
	si := subscribedItem{it.PubDate}
	iter := datastore.NewQuery("subscription").Ancestor(ik.Parent()).KeysOnly().Run(c)
	var sk *datastore.Key
	for sk, err = iter.Next(nil); err == nil; sk, err = iter.Next(nil) {
		uk, err := datastore.DecodeKey(sk.StringID())
		if err != nil {
			return err
		}
		_, err = datastore.Put(c, datastore.NewKey(c, "subscribedItem", ik.Encode(), 0, uk), &si)
		if err != nil {
			return err
		}
	}
	if err != datastore.Done {
		return err
	}
	return nil
})

func unsubscriber(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	k, err := datastore.DecodeKey(r.URL.RawQuery)
	if err != nil {
		handleError(w, err)
		return
	}
	err = datastore.Delete(c, k)
	if err != nil {
		handleError(w, err)
		return
	}
	fk := k.Parent()
	uk, err := datastore.DecodeKey(k.StringID())
	if err != nil {
		handleError(w, err)
		return
	}
	iter := datastore.NewQuery("subscribedItem").Ancestor(uk).KeysOnly().Run(c)
	var sik *datastore.Key
	for sik, err = iter.Next(nil); err == nil; sik, err = iter.Next(nil) {
		ik, err := datastore.DecodeKey(sik.StringID())
		if err != nil {
			handleError(w, err)
			return
		}
		if ik.Parent().Equal(fk) {
			err = datastore.Delete(c, sik)
			if err != nil {
				handleError(w, err)
				return
			}
		}
	}
	if err != datastore.Done {
		handleError(w, err)
		return
	}
	http.Redirect(w, r, "/feeds/", http.StatusFound)
}
