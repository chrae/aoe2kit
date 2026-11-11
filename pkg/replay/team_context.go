package replay

// Lobby team choices are independent of slot order, player numbers, colors,
// trigger definitions and actual runtime alliances.
type TeamContext struct {
	Source                  string   `json:"source"`
	InGameAlliancesVerified bool     `json:"in_game_alliances_verified"`
	DiplomacyEffects        int      `json:"change_diplomacy_effects"`
	GraphTier               string   `json:"trigger_graph_decode_tier"`
	Warnings                []string `json:"warnings,omitempty"`
}

func replayTeamContext(rec *File) TeamContext {
	out := TeamContext{Source: "lobby_picks", GraphTier: rec.TriggerGraphTier}
	out.Warnings = append(out.Warnings, "team, selected_team_id and resolved_team_id are lobby picks; they do not establish in-game alliances or identify players by slot or color")
	if rec.TriggerGraph != nil {
		for _, trigger := range rec.TriggerGraph.Triggers {
			count := func(effect map[string]any) {
				if effect["type_name"] == "change_diplomacy" {
					out.DiplomacyEffects++
				}
			}
			switch effects := trigger["effects"].(type) {
			case []map[string]any:
				for _, effect := range effects {
					count(effect)
				}
			case []any:
				for _, value := range effects {
					if effect, ok := value.(map[string]any); ok {
						count(effect)
					}
				}
			}
		}
	}
	if out.DiplomacyEffects > 0 {
		out.Warnings = append(out.Warnings, "scenario trigger definitions contain change-diplomacy effects; lobby teams may be overridden, and trigger execution/current alliances are unverified")
	}
	if out.GraphTier != "triggergraph_count_and_order_validated" {
		out.Warnings = append(out.Warnings, "trigger graph coverage is unavailable or heuristic; missing diplomacy effects are not evidence that lobby teams persisted in game")
	}
	return out
}
