package app

import (
	"html/template"
	"net/http"
	"sort"

	"appengine"
	"appengine/datastore"
)

const listerRaw = `
<html>
<head>
  <link rel="stylesheet" href="/static/main.css">
</head>
<body>
<a class="admin" href="/all/">home</a>
<p>

<form class="newfeed" action="/addAtom/" method="get">
  <fieldset>
    <legend>New Atom feed</legend>
    <input type="text" name="url"> <input type="submit" value="Subscribe">
  </fieldset>
</form>

<form class="newfeed" action="/addRSS/" method="get">
  <fieldset>
    <legend>New RSS feed</legend>
    <input type="text" name="url"> <input type="submit" value="Subscribe">
  </fieldset>
</form>
<p>

{{range .}}



<a class="largefeedlink" href="/feed/?{{.ID }}">{{.Title}}</a> (<a class="peek" href="/unsubscribe/?{{.SubID}}">unsubscribe</a>)<br>
{{end}}
</body>
</html>
`

type feedInfo struct {
	ID, Title, SubID string
}

type feedInfos []*feedInfo

func (f feedInfos) Len() int           { return len(f) }
func (f feedInfos) Less(i, j int) bool { return f[i].Title < f[j].Title }
func (f feedInfos) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func lister(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	feedRoot := datastore.NewKey(c, "feedRoot", "feedRoot", 0, nil)
	uk := userKey(c)
	q := datastore.NewQuery("subscription").Ancestor(feedRoot).Filter("User =", uk).KeysOnly()
	fc, err := q.Count(c)
	if err != nil {
		handleError(w, err)
		return
	}
	data := make(feedInfos, 0, fc)
	iter := q.Run(c)
	for {
		var f RSS
		sk, err := iter.Next(nil)
		if err == datastore.Done {
			break
		}
		if err != nil {
			handleError(w, err)
			return
		}
		k := sk.Parent()
		err = datastore.Get(c, k, &f)
		if err != nil {
			handleError(w, err)
			return
		}
		data = append(data, &feedInfo{
			ID:    k.StringID(),
			Title: f.Title,
			SubID: sk.Encode(),
		})
	}
	sort.Sort(data)
	templ, err := template.New("lister").Parse(listerRaw)
	if err != nil {
		handleError(w, err)
		return
	}
	err = templ.Execute(w, data)
	if err != nil {
		handleError(w, err)
		return
	}
}
