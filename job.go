package lycos

/*
  Job:
    Villager - 村人
    Seer - 占い師
    Medium - 霊能者
    Guard - 守護者, 狩人
    Freemason - フリーメーソン, 共有者
    Possess - 狂人
    Wolf - 狼
    Fox - 妖狐
*/
type Job int

const (
	Villager Job = iota + 1
	Seer
	Medium
	Guard
	Freemason
	Possess
	Wolf
	Fox
)

func (j Job) String() string {
	switch j {
	case Villager:
		return "村人"
	case Seer:
		return "占い師"
	case Medium:
		return "霊能者"
	case Guard:
		return "守護者"
	case Freemason:
		return "共有者"
	case Possess:
		return "狂人"
	case Wolf:
		return "人狼"
	case Fox:
		return "妖狐"
	default:
		return "None"
	}
}

func (j Job) Description() string {
	switch j {
	case Villager:
		return "あなたは村人です。"
	case Seer:
		return "あなたは占い師です。"
	case Medium:
		return "あなたは霊能者です。"
	case Guard:
		return "あなたは狩人です。"
	case Freemason:
		return "あなたは共有者です。"
	case Possess:
		return "あなたは狂人です。"
	case Wolf:
		return "あなたは人狼です。"
	case Fox:
		return "あなたは妖狐です。"
	default:
		return "None"
	}
}

func (j Job) HasAbility() bool {
	switch j {
	case Seer:
		return true
	case Guard:
		return true
	case Wolf:
		return true
	default:
		return false
	}
}

func (j Job) IsEvil() bool {
	switch j {
	case Possess:
		return true
	case Wolf:
		return true
	case Fox:
		return true
	default:
		return false
	}
}

func (j Job) IsHuman() bool {
	switch j {
	case Wolf:
		return false
	case Fox:
		return false
	default:
		return true
	}
}

func (j Job) IsWolf() bool {
	return j == Wolf
}

func (j Job) IsBlack() bool {
	return j == Wolf
}

func (j Job) CanSpeakWhisper() bool {
	return j == Wolf
}

func (j Job) CanUseAbility(day int) bool {
	switch j {
	case Wolf:
		return day >= 2
	case Seer:
		return day >= 1
	case Guard:
		return day >= 2
	default:
		return false
	}
}

func (j Job) GotVictory(wonCamp int) bool {
	if wonCamp == 2 {
		return j == Wolf || j == Possess
	} else if wonCamp == 3 {
		return j == Fox
	} else if wonCamp == 1 {
		return j != Fox && j != Possess && j != Wolf
	}
	return false
}
