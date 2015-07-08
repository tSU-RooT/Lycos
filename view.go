package lycos

type TopView struct {
	Villages  []Village
	Login     bool
	LoginURL  string
	LogoutURL string
	Reader    User
}
type PreWriteView struct {
	Text        string
	HiddenText  string
	VillageNo   int64
	Author      string
	CharacterID string
	Face        string
	IsPublic    bool
	IsPersonal  bool
	IsWhisper   bool
	IsGraveyard bool
}

type VillageView struct {
	Posts              []Post
	Village            Village
	People             []Person
	CharacterSet       []Chara
	Chapters           []Chapter
	Indexes            []Page
	Result             []ResultCol
	Reader             Person
	No                 int64
	Day                int
	Login              bool
	Enter              bool
	Recent             bool
	ShowStartButton    bool
	ShowAbility        bool
	ShowAbilitySelect  bool
	ShowResult         bool
	UserName           string
	LoginURL           string
	LogoutURL          string
	UserFace           string
	AbilityDescription string
	UpdatetimeNotice   string
	NpcName            string
}

func (v VillageView) Collecting() bool {
	return v.Village.Day == 0
}
func (v VillageView) Opening() bool {
	return v.Village.Day > 0
}
func (v VillageView) IsEpilogue() bool {
	return v.Village.Day <= -1
}
func (v VillageView) ShowVoteForm() bool {
	return v.Village.Day >= 2
}
func (v VillageView) VoteTargetLists() []Person {
	targets := make([]Person, 0, 5)
	for _, p := range v.People {
		if !p.Dead && p.UserID != v.Reader.UserID {
			targets = append(targets, p)
		}
	}
	return targets
}
func (v VillageView) AbilityTargetLists() []Person {
	targets := make([]Person, 0, 5)
	if !v.Enter || v.Reader.Job == 0 {
		return targets
	}
	if v.Reader.Job == Wolf {
		if v.Village.Day == 1 {
			npc := Person{Name: v.NpcName, CharacterID: "NPC"}
			targets = append(targets, npc)
			return targets
		}
		for _, p := range v.People {
			if !p.Dead && p.Job != Wolf && p.UserID != v.Reader.UserID {
				targets = append(targets, p)
			}
		}
	} else {
		for _, p := range v.People {
			if !p.Dead && p.UserID != v.Reader.UserID {
				targets = append(targets, p)
			}
		}
	}
	return targets
}
func (v VillageView) VoteTargetsName() string {
	for _, p := range v.People {
		if !p.Dead && p.CharacterID == v.Reader.VoteTarget {
			return "あなたは" + p.Name + "を選択しています。"
		}
	}
	return "あなたは誰も選択していません。"
}

func (v VillageView) AbilityTargetsName() string {
	for _, p := range v.People {
		if !p.Dead && p.CharacterID == v.Reader.AbilityTarget {
			return "あなたは" + p.Name + "を選択しています。"
		}
	}
	if v.Reader.AbilityTarget == "NPC" && v.Reader.Job == Wolf {
		return "あなたは" + v.NpcName + "を選択しています。"
	}
	return "あなたは誰も選択していません。"
}

func (v VillageView) JobImage() string {
	if v.Reader.Job == Seer {
		return "Seer.png"
	} else if v.Reader.Job == Wolf {
		return "WereWolf.png"
	} else if v.Reader.Job == Medium {
		return "Medium.png"
	}
	return "face/" + v.Reader.Face
}

type ResultCol struct {
	Name    string
	Handle  string
	Dead    bool
	Victory bool
	Job     Job
	WantJob string
}

type Chapter struct {
	Name    string
	Day     int
	Invalid bool
}

type Page struct {
	Number  int
	Invalid bool
}
