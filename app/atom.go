package app

import (
	"time"
)

type Atom struct {
	Title string `xml:"title"`
	Link string `xml:"link>href,attr"`
	Entries []Entry `xml:"entry" datastore:"-"`
	IsAtom bool `xml:"-"`
}

type Entry struct {
	Title string `xml:"title"`
	Link string `xml:"link>href,attr"`
	GUID string `xml:"id"`
	PubDate time.Time `xml:"-"`
	RawPD string `xml:"updated" datastore:"-"`
}
