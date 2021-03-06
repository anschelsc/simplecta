package app

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/delay"
	"appengine/memcache"
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
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	recentKey := datastore.NewKey(c, "recent", fk.StringID(), 0, feedRoot)
	var re Recent
	err = datastore.Get(c, recentKey, &re)
	if err == datastore.ErrNoSuchEntity {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = datastore.Put(c, datastore.NewKey(c, "subscribedItem", re.Item.Encode(), 0, uk), &subscribedItem{re.PubDate})
	return err
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

type Token struct {
	token []byte
}

func setUserToken(c appengine.Context, token []byte) error {
	uk := userKey(c)
	uid := user.Current(c).ID
	key := datastore.NewKey(c, "userToken", "userToken", 0, uk)
	_, err := datastore.Put(c, key, &Token{token})
	if err != nil {
		return err
	}
	_ = memcache.Set(c, &memcache.Item{Key: uid + "_token", Value: token})
	// Don't check the error on memcache.Set; if it didn't work, oh well
	return nil
}

func checkUserToken(c appengine.Context, token []byte) error {
	uid := user.Current(c).ID
	it, err := memcache.Get(c, uid+"_token")
	if err == nil {
		if bytes.Equal(token, it.Value) {
			return nil
		}
		return errors.New("Bad token.")
	}
	// memcache failed, try datastore
	uk := userKey(c)
	key := datastore.NewKey(c, "userToken", "userToken", 0, uk)
	var t Token
	err = datastore.Get(c, key, &t)
	if err == datastore.ErrNoSuchEntity {
		return errors.New("No token for this user.")
	}
	if err != nil {
		return err
	}
	if bytes.Equal(token, t.token) {
		return nil
	}
	return errors.New("Bad token")
}
