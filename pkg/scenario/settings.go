package scenario

type SettingsReport struct {
	Path         string                 `json:"path,omitempty"`
	Version      string                 `json:"version"`
	Verification string                 `json:"verification"`
	PlayerCount  int                    `json:"player_count"`
	Players      []PlayerSettings       `json:"players,omitempty"`
	Diplomacy    DiplomacySettings      `json:"diplomacy"`
	Victory      map[string]int         `json:"victory,omitempty"`
	Resources    []PlayerResourceValues `json:"resources,omitempty"`
}

type PlayerSettings struct {
	Player           int                  `json:"player"`
	Active           bool                 `json:"active"`
	Human            bool                 `json:"human"`
	TribeName        string               `json:"tribe_name,omitempty"`
	Civilization     string               `json:"civilization,omitempty"`
	LockCivilization bool                 `json:"lock_civilization"`
	LockPersonality  bool                 `json:"lock_personality"`
	AIName           string               `json:"ai_name,omitempty"`
	AIType           int                  `json:"ai_type,omitempty"`
	Resources        PlayerResourceValues `json:"resources"`
	Diplomacy        []int                `json:"diplomacy,omitempty"`
	AlliedVictory    bool                 `json:"allied_victory"`
}

type PlayerResourceValues struct {
	Player     int `json:"player"`
	Gold       int `json:"gold"`
	Wood       int `json:"wood"`
	Food       int `json:"food"`
	Stone      int `json:"stone"`
	TradeGoods int `json:"trade_goods"`
}

type DiplomacySettings struct {
	Matrix                  [][]int `json:"matrix,omitempty"`
	AlliedVictory           []bool  `json:"allied_victory,omitempty"`
	LockTeams               bool    `json:"lock_teams"`
	AllowPlayersChooseTeams bool    `json:"allow_players_choose_teams"`
	RandomStartPoints       bool    `json:"random_start_points"`
	MaxNumberOfTeams        int     `json:"max_number_of_teams"`
}

func SettingsFile(path string) (SettingsReport, error) {
	file, err := Open(path)
	if err != nil {
		return SettingsReport{}, err
	}
	report := file.Settings()
	report.Path = path
	return report, nil
}

func (f *File) Settings() SettingsReport {
	report := SettingsReport{
		Path:         f.Path,
		Version:      f.Version,
		Verification: "structure_verified_not_engine_verified",
		PlayerCount:  f.PlayerCount,
		Players:      append([]PlayerSettings(nil), f.playerSettings()...),
		Diplomacy:    f.diplomacySettings(),
		Victory:      f.victorySettings(),
		Resources:    f.resourceSettings(),
	}
	for i := range report.Players {
		if i < len(report.Resources) {
			report.Players[i].Resources = report.Resources[i]
		}
		if i < len(report.Diplomacy.Matrix) {
			report.Players[i].Diplomacy = report.Diplomacy.Matrix[i]
		}
		if i < len(report.Diplomacy.AlliedVictory) {
			report.Players[i].AlliedVictory = report.Diplomacy.AlliedVictory[i]
		}
	}
	return report
}

func (f *File) playerSettings() []PlayerSettings {
	out := make([]PlayerSettings, 0, len(f.Players))
	for _, player := range f.Players {
		out = append(out, PlayerSettings{
			Player:           player.Player,
			Active:           player.Active,
			Human:            player.Human,
			TribeName:        player.TribeName,
			Civilization:     player.Civilization,
			LockCivilization: player.LockCivilization,
			LockPersonality:  player.LockPersonality,
			AIName:           player.AIName,
			AIType:           player.AIType,
		})
	}
	return out
}

func (f *File) resourceSettings() []PlayerResourceValues {
	playerDataTwo := f.root.section("PlayerDataTwo")
	if playerDataTwo == nil {
		return nil
	}
	resources := playerDataTwo.list("resources")
	out := make([]PlayerResourceValues, 0, len(resources))
	for i, resource := range resources {
		gold, _ := resource.intValue("gold")
		wood, _ := resource.intValue("wood")
		food, _ := resource.intValue("food")
		stone, _ := resource.intValue("stone")
		tradeGoods, _ := resource.intValue("trade_goods")
		out = append(out, PlayerResourceValues{
			Player:     i,
			Gold:       gold,
			Wood:       wood,
			Food:       food,
			Stone:      stone,
			TradeGoods: tradeGoods,
		})
	}
	return out
}

func (f *File) diplomacySettings() DiplomacySettings {
	diplomacy := f.root.section("Diplomacy")
	if diplomacy == nil {
		return DiplomacySettings{}
	}
	rows := diplomacy.list("per_player_diplomacy")
	matrix := make([][]int, 0, len(rows))
	for _, row := range rows {
		matrix = append(matrix, row.intList("stance_with_each_player"))
	}
	alliedRaw := diplomacy.intList("per_player_allied_victory")
	allied := make([]bool, 0, len(alliedRaw))
	for _, value := range alliedRaw {
		allied = append(allied, value != 0)
	}
	lockTeams, _ := diplomacy.intValue("lock_teams")
	allowTeams, _ := diplomacy.intValue("allow_players_choose_teams")
	randomStarts, _ := diplomacy.intValue("random_start_points")
	maxTeams, _ := diplomacy.intValue("max_number_of_teams")
	return DiplomacySettings{
		Matrix:                  matrix,
		AlliedVictory:           allied,
		LockTeams:               lockTeams != 0,
		AllowPlayersChooseTeams: allowTeams != 0,
		RandomStartPoints:       randomStarts != 0,
		MaxNumberOfTeams:        maxTeams,
	}
}

func (f *File) victorySettings() map[string]int {
	globalVictory := f.root.section("GlobalVictory")
	if globalVictory == nil {
		return nil
	}
	fields := []string{
		"conquest_required",
		"ruins",
		"artifacts_required",
		"discovery",
		"explored_percent_of_map_required",
		"gold_required",
		"all_custom_conditions_required",
		"mode",
		"required_score_for_score_victory",
		"time_for_timed_game_in_10ths_of_a_year",
	}
	out := map[string]int{}
	for _, field := range fields {
		value, ok := globalVictory.intValue(field)
		if ok {
			out[field] = value
		}
	}
	return out
}
