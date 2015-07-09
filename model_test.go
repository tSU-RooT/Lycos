package lycos

import (
	"testing"
)

// dataset for testing
var (
	people []Person
)

func init() {
	player1 := Person{
		UserID:      "1234567890",
		CharacterID: "e807f1fcf82d132f9bb018ca6738a19f", // like MD5 string
		Face:        "default/01.png",
		Name:        "Player1",
		Job:         Villager,
		WantJob:     "1",
	}
	player2 := Person{
		UserID:      "1234567891",
		CharacterID: "0f7e44a922df352c05c5f73cb40ba115",
		Face:        "default/02.png",
		Name:        "Player2",
		Job:         Seer,
		WantJob:     "1",
	}
	player3 := Person{
		UserID:      "1234567892",
		CharacterID: "893377c9d852e09874125b10a0e4f66",
		Face:        "default/03.png",
		Name:        "Player3",
		Job:         Wolf,
		WantJob:     "1",
	}
	player4 := Person{
		UserID:      "1234567893",
		CharacterID: "43042f668f07adfd174cb1823d4795e1",
		Face:        "default/04.png",
		Name:        "Player4",
		Job:         Fox,
		WantJob:     "1",
	}
	people = []Person{player1, player2, player3, player4}
}

func TestFortune(t *testing.T) {
	p := make([]Person, len(people))
	copy(p, people)
	// p[1](Player2) is Seer
	// Check White Case
	p[1].AbilityTarget = p[0].CharacterID
	res := Fortune(p)
	if res.Text != "占いをした結果、"+p[0].Name+"は人狼ではないようだ。" {
		t.Fatalf("unexpected fortune result")
	}
	// Check Black Case
	p[1].AbilityTarget = p[2].CharacterID
	res = Fortune(p)
	if res.Text != "占いをした結果、"+p[2].Name+"は人狼のようだ。" {
		t.Fatalf("unexpected fortune result")
	}
	// Check Fox Case
	p[1].AbilityTarget = p[3].CharacterID
	res = Fortune(p)
	if res.Text != "占いをした結果、"+p[3].Name+"は人狼ではないようだ。" {
		t.Fatalf("unexpected fortune result")
	}
	if !p[3].Dead {
		t.Fatalf("Fox must die:%v", p[3])
	}
}
