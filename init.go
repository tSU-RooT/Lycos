package lycos

import (
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	jst, _       = time.LoadLocation("Asia/Tokyo")
	characterSet = make([]Chara, 0, 80)
	setting      = TextSetting{}

	topTmpl         *template.Template
	createPageTmpl  *template.Template
	villagePageTmpl *template.Template
	prewriteTmpl    *template.Template
	badTmpl         *template.Template

	characterSettingFilePath = "characters.yaml"
	textSettingFilePath      = "setting.yaml"
)

type TextSetting struct {
	PrologueSystemPost  string `yaml:"prologue_system_post"`
	FirstDaySystemPost  string `yaml:"first_day_system_post"`
	SecondDaySystemPost string `yaml:"second_day_system_post"`
	NpcName             string `yaml:"npc_name"`
	NpcFace             string `yaml:"npc_face"`
	NpcFirstPost        string `yaml:"npc_first_post"`
	NpcSecondPost       string `yaml:"npc_second_post"`
	VillagerWin         string `yaml:"villager_win"`
	WerewolfWin         string `yaml:"werewolf_win"`
	FoxWin              string `yaml:"fox_win"`
}

func init() {
	if b, err := ioutil.ReadFile(characterSettingFilePath); err == nil {
		if err = yaml.Unmarshal(b, &characterSet); err != nil {
			panic(err)
		}
		// Check Mistakes
		if len(characterSet) == 0 {
			panic(characterSettingFilePath + " is empty!")
		}
		for i := range characterSet {
			if characterSet[i].File == "" || characterSet[i].Name == "" {
				panic(characterSettingFilePath + " has Empty Element")
			}
		}
	} else {
		panic("Please Setup your " + characterSettingFilePath)
	}

	if b, err := ioutil.ReadFile(textSettingFilePath); err == nil {
		if err = yaml.Unmarshal(b, &setting); err != nil {
			panic(err)
		}
	} else {
		panic("Please Setup your " + textSettingFilePath +
			" if you don't want change from default setting, please reset this file.")
	}

	tf := template.FuncMap{
		"rawhtml": func(text string) template.HTML { return template.HTML(text) },
		"up":      func(i int) int { return i + 1 },
		"showNumberTag": func(tag string) string {
			if tag == "" {
				return ""
			} else {
				return "(No." + tag + ")"
			}
		},
	}
	topTmpl = template.Must(template.New("").Funcs(tf).ParseFiles("templates/top.html"))
	createPageTmpl = template.Must(template.New("").Funcs(tf).ParseFiles("templates/create.html"))
	villagePageTmpl = template.Must(template.New("").Funcs(tf).ParseFiles("templates/village.html"))
	prewriteTmpl = template.Must(template.New("").Funcs(tf).ParseFiles("templates/prewrite.html"))
	badTmpl = template.Must(template.New("").Funcs(tf).ParseFiles("templates/badreq.html"))

	http.HandleFunc("/", topHandler)
	http.HandleFunc("/village/", villageHandler)
	http.HandleFunc("/prewrite", villagePreWriteHandler)
	http.HandleFunc("/write", villagePostWriteHandler)
	http.HandleFunc("/enter", enterToVillageHandler)
	http.HandleFunc("/create", createPageHandler)
	http.HandleFunc("/rename", renameHandler)
	http.HandleFunc("/build", buildPageHandler)
	http.HandleFunc("/vote", changeVoteTargetHandler)
	http.HandleFunc("/change", changeAbilityTargetHandler)
	http.HandleFunc("/start/", villageStartHandler)
	http.HandleFunc("/update", updateHandler)
}
