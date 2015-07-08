package lycos

import (
	gae "appengine"
	"appengine/datastore"
	"appengine/memcache"
	"github.com/mjibson/goon"
	"net/http"
	"time"
)

type UpdateSchedule struct {
	VillageNo  int64          `datastore:"VillageNo" goon:"id"`
	ParentKey  *datastore.Key `datastore:"-" goon:"parent"`
	UpdateTime time.Time      `datastore:"UpdateTime"`
	Hour       int            `datastore:"Hour,noindex"`
	Minute     int            `datastore:"Minute,noindex"`
}

func (us *UpdateSchedule) SetNextUpdateTime(basetime time.Time) {
	next := time.Date(basetime.Year(), basetime.Month(), basetime.Day(), us.Hour, us.Minute, 0, 0, jst)
	for {
		s := next.Sub(basetime)
		if s <= time.Hour*12 {
			next = next.Add(time.Hour * 24)
		} else {
			break
		}
	}
	us.UpdateTime = next
}

func (us UpdateSchedule) IsUpdateTime(now time.Time) bool {
	s := now.Sub(us.UpdateTime)
	// 10分差まで許容
	if s < time.Minute*10 && s >= 0 {
		return true
	}
	return false
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	if r.Header.Get("X-AppEngine-Cron") != "true" {
		http.NotFound(w, r)
		return
	}
	var ups []UpdateSchedule
	if cache, err := memcache.Get(c, "UpdateSchedule"); err == memcache.ErrCacheMiss {
		q := datastore.NewQuery("UpdateSchedule")
		if _, err := g.GetAll(q, &ups); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if val, err := serialize(&ups); err == nil {
			item := memcache.Item{Key: "UpdateSchedule", Value: val, Expiration: time.Hour * 12}
			memcache.Add(c, &item)
		}
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		deserialize(cache.Value, &ups)
	}
	now := time.Now()
	for _, us := range ups {
		if us.IsUpdateTime(now) {
			if err := updateVillage(us, g); err != nil {
				c.Debugf("%v", err)
			}
			memcache.Delete(c, "UpdateSchedule")
		}
	}
	return
}

func updateVillage(us UpdateSchedule, g *goon.Goon) error {
	village := Village{No: us.VillageNo}
	if err := g.Get(&village); err != nil || village.Day <= 0 {
		return nil
	}
	village.Day++
	err := g.RunInTransaction(func(g *goon.Goon) error {
		vKey := g.Key(village)
		q3 := datastore.NewQuery("Person").Ancestor(vKey).Order("CreatedTime")
		people := make([]Person, 0, 10)
		_, err := g.GetAll(q3, &people)
		if err != nil {
			return err
		}
		// 3日目以降のみ投票処理
		var posts []Post
		if village.Day >= 3 {
			posts = Execute(people)
		} else if village.Day == 2 {
			t := setting.NpcName + "は無残な死体で見つかった。"
			p := Post{Author: "System", AuthorID: "0", Text: t, Time: time.Now(), Type: SystemMessage}
			posts = append(posts, p)
			p2 := Post{Author: "System", AuthorID: "0", Time: time.Now(), Type: SystemMessage}
			p2.Text = setting.SecondDaySystemPost
			posts = append(posts, p2)
		}
		before := make([]Person, len(people))
		copy(before, people)
		posts = append(posts, Fortune(people)...)
		Raid(people)
		// Check Dead Person
		peaceful := true
		for i := range people {
			if !before[i].Dead && people[i].Dead {
				peaceful = false
				t := people[i].Name + "は無残な死体で見つかった。"
				p := Post{Author: "System", AuthorID: "0", Text: t, Time: time.Now(), Type: SystemMessage}
				posts = append(posts, p)
			}
			people[i].VoteTarget = ""
			people[i].AbilityTarget = ""
		}
		if peaceful && village.Day >= 3 {
			t := "今日は誰も犠牲者がいないようだ。"
			p := Post{Author: "System", AuthorID: "0", Text: t, Time: time.Now(), Type: SystemMessage}
			posts = append(posts, p)
		}
		j := Judge(people)
		if j > 0 {
			village.Day *= -1
			var t string
			switch j {
			case 1:
				t = setting.VillagerWin
			case 2:
				t = setting.WerewolfWin
			case 3:
				t = setting.FoxWin
			}
			p := Post{Author: "System", AuthorID: "0", Text: t, Time: time.Now(), Type: SystemMessage}
			posts = append(posts, p)
		}
		_, err = g.Put(&village)
		if err != nil {
			return err
		}
		for i := range posts {
			posts[i].Day = village.Day
			if posts[i].Day <= -1 {
				posts[i].Day = -1
			}
			posts[i].ParentKey = vKey
			_, err = g.Put(&posts[i])
			if err != nil {
				return err
			}
		}
		for i := 0; i < len(people); i++ {
			people[i].ParentKey = vKey
			_, err = g.Put(&people[i])
			if err != nil {
				return err
			}
		}

		us.SetNextUpdateTime(time.Now())
		us.ParentKey = vKey
		_, err = g.Put(&us)
		if err != nil {
			return err
		}
		return nil
	}, nil)
	if err != nil {
		return err
	}
	return nil
}
