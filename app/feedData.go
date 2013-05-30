package app

import (
	"encoding/xml"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
)

// Valid only if k has no descendents of its own kind.
func exists(c appengine.Context, k *datastore.Key) (bool, error) {
	count, err := datastore.NewQuery(k.Kind()).KeysOnly().Ancestor(k).Count(c)
	if err != nil {
		return false, err
	}
	return (count != 0), err
}

func fetchRSS(c appengine.Context, url string) (*RSS, error) {
	ret := new(RSS)
	cl := urlfetch.Client(c)
	resp, err := cl.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(ret)
	return ret, err
}

func (f *RSS) update(c appengine.Context, fk *datastore.Key) error {
	for _, it := range f.Items {
		if it.GUID == "" {
			it.GUID = it.Link
		}
		var err error
		it.PubDate, err = time.Parse(time.RFC1123Z, it.RawPD)
		if err != nil {
			it.PubDate, err = time.Parse(time.RFC822Z, it.RawPD)
			if err != nil {
				return err
			}
		}
		ik := datastore.NewKey(c, "item", it.GUID, 0, fk)
		done, err := exists(c, ik)
		if err != nil {
			return err
		}
		if !done {
			_, err := datastore.Put(c, ik, &it)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func addFeed(c appengine.Context, url string) error {
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	fk := datastore.NewKey(c, "feed", url, 0, feedRoot)
	return datastore.RunInTransaction(c, func(c appengine.Context) error {
		done, err := exists(c, fk)
		if err != nil {
			return err
		}
		if !done {
			f, err := fetchRSS(c, url)
			if err != nil {
				return err
			}
			_, err = datastore.Put(c, fk, f)
			if err != nil {
				return err
			}
			err = f.update(c, fk)
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	}, nil)
}
