package lycos

import (
	gae "appengine"
	"appengine/datastore"
	"appengine/memcache"
	"appengine/user"
	"github.com/mjibson/goon"
	"net/http"
	"strconv"
	"time"
)

func createPageHandler(w http.ResponseWriter, r *http.Request) {
	if err := createPageTmpl.ExecuteTemplate(w, "base", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func buildPageHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	u := user.Current(c)
	if u == nil {
		if err := createPageTmpl.ExecuteTemplate(w, "base", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	var village Village
	if n := r.FormValue("Name"); n != "" {
		village.Name = n
	} else {
		village.Name = "Untitled"
	}
	village.CreatedTime = time.Now()
	village.NumberOfPeople = 0
	village.Day = 0
	village.Builder = u.ID
	village.IncludeFreemason = (r.FormValue("freemason") == "true")
	village.IncludeFox = (r.FormValue("fox") == "true")
	village.Chip = (r.FormValue("chip") == "true")
	village.UpdatetimeHour, _ = strconv.Atoi(r.FormValue("hour"))
	village.UpdatetimeMinute, _ = strconv.Atoi(r.FormValue("minute"))
	// Village.UpdatetimeMinute is allowable by 0 or 30 value
	if village.UpdatetimeMinute%30 != 0 {
		village.UpdatetimeMinute = (village.UpdatetimeMinute / 30) * 30
	}
	management := Management{Key: "management"}
	if err := g.Get(&management); err != nil && err != datastore.ErrNoSuchEntity {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if err == datastore.ErrNoSuchEntity {
		management.VillageNo = 0
	}
	village.No = management.VillageNo
	village.PublicPostNo = 0
	management.VillageNo++
	if _, err := g.Put(&management); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err := g.RunInTransaction(func(g *goon.Goon) error {
		if _, err := g.Put(&village); err != nil {
			return err
		}
		vKey := g.Key(village)
		p1 := Post{
			Text:      setting.PrologueSystemPost,
			Author:    "System",
			AuthorID:  "0",
			Type:      SystemMessage,
			Time:      time.Now(),
			ParentKey: vKey,
		}
		if _, err := g.Put(&p1); err != nil {
			return err
		}
		p2 := Post{
			Text:      setting.NpcFirstPost,
			Author:    setting.NpcName,
			Face:      setting.NpcFace,
			AuthorID:  "NPC",
			Type:      Public,
			Time:      time.Now(),
			NumberTag: "",
			ParentKey: vKey,
		}
		if _, err := g.Put(&p2); err != nil {
			return err
		}
		memcache.Delete(c, "Villages")
		return nil
	}, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/village/?vno="+strconv.FormatInt(village.No, 10)+"&day=recent", http.StatusFound)
}
