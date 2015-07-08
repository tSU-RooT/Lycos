package lycos

import (
	"appengine/datastore"
	"errors"
	"math/rand"
	"strings"
	"time"
)

type Person struct {
	UserID        string         `datastore:"UserID"`
	ParentKey     *datastore.Key `datastore:"-" goon:"parent"`
	CharacterID   string         `datastore:"CharacterID" goon:"id"`
	Face          string         `datastore:"Face,noindex"`
	Name          string         `datastore:"Name,noindex"`
	Job           Job            `datastore:"Job,noindex"`
	WantJob       string         `datastore:"WantJob,noindex"`
	AbilityTarget string         `datastore:"AbilityTarget,noindex"`
	VoteTarget    string         `datastore:"VoteTarget,noindex"`
	Dead          bool           `datastore:"Dead,noindex"`
	CreatedTime   time.Time      `datastore:"CreatedTime"`
}

func (p Person) HasWriteAuth(pt PostType) bool {
	switch pt {
	case Personal:
		return true
	case Public:
		return !p.Dead
	case Whisper:
		return p.Job.CanSpeakWhisper()
	case Graveyard:
		return p.Dead
	default:
		return false
	}
}

func DecideJob(people []Person, village Village, random *rand.Rand) ([]Person, error) {
	// 1:Auto 2:Villager Camp 3:Villager Camp(Has Ability) 4:Evil
	result := make([]Person, 0, len(people))
	restJobs := []Job{Villager}
	abilityJobs := []Job{Seer, Medium, Guard}
	evilJobs := []Job{Wolf, Possess}
	if village.IncludeFreemason {
		abilityJobs = append(abilityJobs, Freemason, Freemason)
	}
	if village.IncludeFox {
		evilJobs = append(evilJobs, Fox)
	}
	if village.NumberOfPeople >= 8 {
		evilJobs = append(evilJobs, Wolf)
		if village.NumberOfPeople >= 16 {
			evilJobs = append(evilJobs, Wolf)
		}
	}
	job_s := len(restJobs) + len(abilityJobs) + len(evilJobs)
	for i := village.NumberOfPeople - job_s; i > 0; i-- {
		restJobs = append(restJobs, Villager)
	}
	// 役欠け処理
	if village.Chip {
		if n := random.Intn(len(abilityJobs) + len(restJobs)); n < len(abilityJobs) {
			abilityJobs[n] = Villager
		}
	}

	auto := make([]Person, 0, 5)
	humans := make([]Person, 0, 5)
	abilities := make([]Person, 0, 5)
	nonHumans := make([]Person, 0, 5)
	for _, p := range people {
		switch p.WantJob {
		case "1":
			auto = append(auto, p)
		case "2":
			humans = append(humans, p)
		case "3":
			abilities = append(abilities, p)
		case "4":
			nonHumans = append(nonHumans, p)
		default:
			return nil, errors.New("DecideJob: Illegal Value of Person.WantJob")
		}
	}
	// Decide Job
	for {
		if len(nonHumans) <= 0 || len(evilJobs) <= 0 {
			break
		}
		r1 := random.Intn(len(nonHumans))
		r2 := random.Intn(len(evilJobs))
		nonHumans[r1].Job = evilJobs[r2]
		result = append(result, nonHumans[r1])
		nonHumans = append(nonHumans[:r1], nonHumans[r1+1:]...)
		evilJobs = append(evilJobs[:r2], evilJobs[r2+1:]...)
	}
	if len(nonHumans) > 0 {
		auto = append(auto, nonHumans...)
	}
	for {
		if len(abilities) <= 0 || len(abilityJobs) <= 0 {
			break
		}
		r1 := random.Intn(len(abilities))
		r2 := random.Intn(len(abilityJobs))
		abilities[r1].Job = abilityJobs[r2]
		result = append(result, abilities[r1])
		abilities = append(abilities[:r1], abilities[r1+1:]...)
		abilityJobs = append(abilityJobs[:r2], abilityJobs[r2+1:]...)
	}
	if len(abilities) > 0 {
		humans = append(humans, abilities...)
	} else if len(abilityJobs) > 0 {
		restJobs = append(restJobs, abilityJobs...)
	}
	for {
		if len(humans) <= 0 || len(restJobs) <= 0 {
			break
		}
		r1 := random.Intn(len(humans))
		r2 := random.Intn(len(restJobs))
		humans[r1].Job = restJobs[r2]
		result = append(result, humans[r1])
		humans = append(humans[:r1], humans[r1+1:]...)
		restJobs = append(restJobs[:r2], restJobs[r2+1:]...)
	}
	if len(humans) > 0 {
		auto = append(auto, humans...)
	}
	if len(evilJobs) > 0 {
		restJobs = append(restJobs, evilJobs...)
	}
	for _, nj := range restJobs {
		if len(auto) <= 0 {
			break
		}
		r := random.Intn(len(auto))
		auto[r].Job = nj
		result = append(result, auto[r])
		auto = append(auto[:r], auto[r+1:]...)
	}
	return result, nil
}

// Execute Player
func Execute(people []Person) []Post {
	result := make([]Post, 0)
	voteMessageText := ""
	var voteMap map[string]int = make(map[string]int)
	for _, p := range people {
		if p.Dead {
			continue
		} else if p.VoteTarget != "" {
			voteMap[p.VoteTarget] += 1
			voteMessageText += p.Name + "は" + p.VoteTarget + "に投票した。\n"
		} else {
			voteMessageText += p.Name + "は投票していない。\n"
		}
	}
	for _, p := range people {
		voteMessageText = strings.Replace(voteMessageText, p.CharacterID, p.Name, -1)
	}
	po1 := Post{Author: "System", AuthorID: "0", Text: voteMessageText, Time: time.Now(), Type: SystemMessage}
	result = append(result, po1)
	var execList []string
	numOfMostVotes := 0
	for s, t := range voteMap {
		if t > numOfMostVotes {
			execList = make([]string, 1)
			execList[0] = s
			numOfMostVotes = t
		} else if t == numOfMostVotes {
			execList = append(execList, s)
		}
	}
	rand.Seed(time.Now().UnixNano())
	target := ""
	if len(execList) == 1 {
		target = execList[0]
	} else if len(execList) > 1 {
		target = execList[rand.Intn(len(execList))]
	} else {
		return result
	}

	for i := range people {
		if people[i].CharacterID == target && !people[i].Dead {
			people[i].Dead = true
			t := people[i].Name + "は民衆によって処刑された。"
			po2 := Post{Author: "System", AuthorID: "0", Text: t, Time: time.Now(), Type: SystemMessage}
			result = append(result, po2)
			if po3, b := inspire(people, people[i]); b {
				result = append(result, po3)
			}
			return result
		}
	}
	return result
}

func Fortune(people []Person) []Post {
	result := make([]Post, 0)
	target := ""
	var seer Person
	for i := range people {
		if people[i].Job == Seer && !people[i].Dead {
			target = people[i].AbilityTarget
			seer = people[i]
		}
	}
	if target == "" || seer.Name == "" {
		return result
	}
	for i := range people {
		if people[i].CharacterID == target {
			if people[i].Job.IsBlack() {
				t := "占いをした結果、" + people[i].Name + "は人狼のようだ。"
				p := Post{Author: "System", AuthorID: seer.UserID, Text: t, Time: time.Now(), Type: SystemSecret}
				result = append(result, p)
			} else {
				if people[i].Job == Fox {
					people[i].Dead = true
				}
				t := "占いをした結果、" + people[i].Name + "は人狼ではないようだ。"
				p := Post{Author: "System", AuthorID: seer.UserID, Text: t, Time: time.Now(), Type: SystemSecret}
				result = append(result, p)
			}
		}
	}
	return result
}

// Raid by Werewolf
func Raid(people []Person) {
	wolfList := make([]Person, 0)
	var guard Person
	for _, p := range people {
		if p.Job == Wolf && !p.Dead && p.AbilityTarget != "" {
			wolfList = append(wolfList, p)
		}
		if p.Job == Guard && !p.Dead && p.AbilityTarget != "" {
			guard = p
		}
	}
	if len(wolfList) > 0 {
		rand.Seed(time.Now().UnixNano())
		wolf := wolfList[rand.Intn(len(wolfList))]
		if wolf.AbilityTarget == guard.AbilityTarget {
			return // 護衛成功
		}
		for i := range people {
			// 狐は除外
			if people[i].CharacterID == wolf.AbilityTarget &&
				!people[i].Dead && people[i].Job != Fox {
				people[i].Dead = true // 噛み
			}
		}
	}
}

// inspire is Ability Process for Medium.
func inspire(people []Person, deadman Person) (Post, bool) {
	for _, p := range people {
		if p.Job == Medium && !p.Dead {
			t := ""
			if deadman.Job.IsBlack() {
				t = "処刑された" + deadman.Name + "は人狼のようだ。"
			} else {
				t = "処刑された" + deadman.Name + "は人狼ではないようだ。"
			}
			p := Post{Author: "System", AuthorID: p.UserID, Text: t, Time: time.Now(), Type: SystemSecret}
			return p, true
		}
	}
	return Post{}, false
}

// Judge Game
func Judge(people []Person) int {
	// 0:continue 1:Human won 2:Werewolf won 3:Fox won
	h := 0 // Number of Humans
	w := 0 // Number of Werewolfs
	f := 0 // Number of Foxes
	for _, p := range people {
		if p.Dead {
			continue
		}
		if p.Job.IsHuman() {
			h++
		} else if p.Job.IsWolf() {
			w++
		} else if p.Job == Fox {
			f++
		}
	}
	if w >= h {
		if f >= 1 {
			return 3
		}
		return 2
	}
	if w == 0 && h >= 1 {
		if f >= 1 {
			return 3
		}
		return 1
	}
	return 0
}
