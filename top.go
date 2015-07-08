package lycos

import (
	gae "appengine"
	"appengine/datastore"
	"appengine/memcache"
	gaeuser "appengine/user"
	"github.com/mjibson/goon"
	"net/http"
	"strings"
	"time"
)

func topHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	management := Management{Key: "management"}
	if err := g.Get(&management); err != nil && err != datastore.ErrNoSuchEntity {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if err == datastore.ErrNoSuchEntity {
		management.VillageNo = 1
		g.Put(&management)
	}

	var villages []Village
	memVillageKey := "Villages"
	if cache, err := memcache.Get(c, memVillageKey); err == memcache.ErrCacheMiss {
		q := datastore.NewQuery("Village").Order("CreatedTime")
		if _, err := g.GetAll(q, &villages); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if val, err := serialize(&villages); err == nil {
			item := memcache.Item{Key: memVillageKey, Value: val, Expiration: time.Hour * 24}
			memcache.Add(c, &item)
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		deserialize(cache.Value, &villages)
	}
	topView := TopView{Villages: villages}
	u := gaeuser.Current(c)
	if u != nil {
		topView.Login = true
		topView.LogoutURL, _ = gaeuser.LogoutURL(c, r.URL.String())
		user := User{ID: u.ID}
		if err := g.Get(&user); err != nil {
			s := strings.Split(u.Email, "@")
			handle := "Noname"
			if len(s) > 0 {
				handle = s[0]
			}
			reader := User{Handle: handle, Email: u.Email, ID: u.ID}
			_, err = g.Put(&reader)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			topView.Reader = reader
		} else {
			topView.Reader = user
		}
	} else {
		topView.Login = false
		topView.LoginURL, _ = gaeuser.LoginURL(c, r.URL.String())
	}
	if err := topTmpl.ExecuteTemplate(w, "base", topView); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
