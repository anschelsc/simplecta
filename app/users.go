package app

import (
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/user"
	"appengine/delay"
)

type subscribedItem struct {
	PubDate time.Time
}

var empty = new(struct{})

func userKey(c appengine.Context) *datastore.Key {
	userRoot := datastore.NewKey(c, "userRoot", "userRoot", 0, nil)
	return datastore.NewKey(c, "user", user.Current(c).ID, 0, userRoot)
}

func subscribe(c appengine.Context, fk *datastore.Key, populate bool) error {
	uk := userKey(c)
	_, err := datastore.Put(c, datastore.NewKey(c, "subscription", uk.Encode(), 0, fk), empty)
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
