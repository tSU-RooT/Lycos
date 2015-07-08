package lycos

import (
	gae "appengine"
	"appengine/datastore"
	"appengine/memcache"
	"appengine/user"
	"github.com/mjibson/goon"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func villageStartHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	g := goon.FromContext(c)
	u := user.Current(c)
	no, err := strconv.ParseInt(r.URL.Query().Get("vno"), 10, 64)
	village := Village{No: no}
	if err := g.Get(&village); err != nil {
		bad(w)
		return
	}
	if u.ID != village.Builder {
		http.NotFound(w, r)
		return
	}
	vKey := g.Key(village)
	village.Day = 1
	q1 := datastore.NewQuery("Person").Ancestor(vKey).Order("CreatedTime")
	people := make([]Person, 0, 10)
	_, err = g.GetAll(q1, &people)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	now := time.Now()
	random := rand.New(rand.NewSource(now.UnixNano()))
	random.Seed(now.UnixNano())
	people, err = DecideJob(people, village, random)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = g.RunInTransaction(func(g *goon.Goon) error {
		for i := range people {
			if people[i].Job == Wolf {
				people[i].AbilityTarget = "NPC"
			}
			_, err := g.Put(&people[i])
			people[i].ParentKey = vKey
			if err != nil {
				return err
			}
		}
		post1 := Post{
			Day:       1,
			Text:      setting.FirstDaySystemPost,
			Time:      time.Now(),
			Author:    "System",
			Type:      SystemMessage,
			ParentKey: vKey,
		}
		detailText := "どうやらこの中には \n"
		post2 := Post{Day: 1, Time: time.Now(), Author: "0", Type: SystemMessage, ParentKey: vKey}
		if village.Chip {
			// 欠けありの場合、役職内訳を隠蔽します
			human := 0
			nonHuman := 0
			for i := range people {
				if people[i].Job.IsHuman() {
					human++
				} else {
					nonHuman++
				}
			}
			detailText += "人間が" + strconv.Itoa(human) + "人 \n"
			detailText += "人ならざる者が" + strconv.Itoa(nonHuman) + "人 \n"
			detailText += "いるようだ。"
		} else {
			jobDetail := make([]int, 8, 8)
			for _, p := range people {
				jobDetail[p.Job]++
			}
			for i := range jobDetail {
				if jobDetail[i] > 0 {
					detailText += Job(i).String() + "が" + strconv.Itoa(jobDetail[i]) + "人 "
				}
				if i%3 == 0 && i > 0 {
					detailText += "\n"
				}
			}
			detailText += "いるようだ。"
		}
		post2.Text = detailText
		post3 := Post{
			Day:       1,
			Text:      setting.NpcSecondPost,
			Author:    setting.NpcName,
			Face:      setting.NpcFace,
			AuthorID:  "NPC",
			Type:      Public,
			Time:      time.Now(),
			NumberTag: "",
			ParentKey: vKey,
		}

		_, err := g.Put(&post1)
		if err != nil {
			return err
		}
		_, err = g.Put(&post2)
		if err != nil {
			return err
		}
		_, err = g.Put(&post3)
		if err != nil {
			return err
		}

		_, err = g.Put(&village)
		if err != nil {
			return err
		}

		us := UpdateSchedule{VillageNo: no, Hour: village.UpdatetimeHour, Minute: village.UpdatetimeMinute, ParentKey: vKey}
		us.SetNextUpdateTime(time.Now())
		_, err = g.Put(&us)
		if err != nil {
			return err
		}
		memcache.Delete(c, "UpdateSchedule")
		return nil
	}, nil)
	if err != nil {
		bad(w)
		return
	}
	http.Redirect(w, r, "/village/?vno="+strconv.FormatInt(no, 10)+"&day=1", http.StatusFound)
}
