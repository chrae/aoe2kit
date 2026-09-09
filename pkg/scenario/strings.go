package scenario

import (
	"regexp"
	"sort"
	"strings"
)

var scenarioMarkupTagRE = regexp.MustCompile(`<([^<>]+)>`)

type StringReport struct {
	Path          string             `json:"path,omitempty"`
	Version       string             `json:"version"`
	Verification  string             `json:"verification"`
	Messages      map[string]string  `json:"messages"`
	StringTable   []StringTableEntry `json:"string_table,omitempty"`
	Variables     []VariableEntry    `json:"variables,omitempty"`
	TriggerText   []TriggerTextEntry `json:"trigger_text,omitempty"`
	EffectText    []EffectTextEntry  `json:"effect_text,omitempty"`
	MarkupSummary MarkupSummary      `json:"markup_summary"`
	Warnings      []string           `json:"warnings,omitempty"`
}

type StringTableEntry struct {
	ID     int    `json:"id"`
	Source string `json:"source"`
	Text   string `json:"text"`
}

type VariableEntry struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TriggerTextEntry struct {
	TriggerIndex int    `json:"trigger_index"`
	TriggerName  string `json:"trigger_name,omitempty"`
	Field        string `json:"field"`
	Text         string `json:"text"`
}

type EffectTextEntry struct {
	TriggerIndex int    `json:"trigger_index"`
	TriggerName  string `json:"trigger_name,omitempty"`
	EffectIndex  int    `json:"effect_index"`
	Type         int    `json:"type"`
	StringID     int    `json:"string_id,omitempty"`
	Text         string `json:"text"`
}

type MarkupSummary struct {
	ColorTags         map[string]int `json:"color_tags"`
	VariableRefs      []string       `json:"variable_refs"`
	VariableRefCounts map[string]int `json:"variable_ref_counts"`
}

func (f *File) Strings() StringReport {
	report := StringReport{
		Path:         f.Path,
		Version:      f.Version,
		Verification: "structure_verified_not_engine_verified",
		Messages:     map[string]string{},
	}
	if f.headerRoot == nil || f.root == nil {
		report.Warnings = append(report.Warnings, "scenario parse tree unavailable")
		return report
	}
	if text, ok := f.headerRoot.stringValue("scenario_instructions"); ok && text != "" {
		report.Messages["scenario_instructions"] = text
	}
	report.addMessageSection(f.root.section("Messages"))
	report.addMessageSection(f.root.section("Cinematics"))
	report.StringTable = scenarioStringTable(f.root)
	report.Variables = scenarioVariables(f.root)
	report.TriggerText = scenarioTriggerText(f.root)
	report.EffectText = scenarioEffectText(f.root)
	report.MarkupSummary = summarizeScenarioMarkup(report.allText())
	return report
}

func (r *StringReport) addMessageSection(section *parsedSection) {
	if section == nil {
		return
	}
	for _, field := range section.Fields {
		text, ok := field.Value.(string)
		if !ok || text == "" {
			continue
		}
		r.Messages[field.Name] = text
	}
}

func scenarioStringTable(root *parsedRoot) []StringTableEntry {
	section := root.section("PlayerDataTwo")
	if section == nil {
		return nil
	}
	values := section.stringList("strings")
	out := make([]StringTableEntry, 0, len(values))
	for i, text := range values {
		if text == "" {
			continue
		}
		out = append(out, StringTableEntry{ID: i, Source: "PlayerDataTwo.strings", Text: text})
	}
	return out
}

func scenarioVariables(root *parsedRoot) []VariableEntry {
	section := root.section("Triggers")
	if section == nil {
		return nil
	}
	var out []VariableEntry
	for _, variable := range section.list("variable_data") {
		id, _ := variable.intValue("variable_id")
		name, _ := variable.stringValue("variable_name")
		if name == "" {
			continue
		}
		out = append(out, VariableEntry{ID: id, Name: name})
	}
	return out
}

func scenarioTriggerText(root *parsedRoot) []TriggerTextEntry {
	section := root.section("Triggers")
	if section == nil {
		return nil
	}
	var out []TriggerTextEntry
	for triggerIndex, trigger := range section.list("trigger_data") {
		triggerName, _ := trigger.stringValue("trigger_name")
		for _, field := range []string{"trigger_description", "short_description"} {
			text, _ := trigger.stringValue(field)
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			out = append(out, TriggerTextEntry{
				TriggerIndex: triggerIndex,
				TriggerName:  triggerName,
				Field:        field,
				Text:         text,
			})
		}
	}
	return out
}

func scenarioEffectText(root *parsedRoot) []EffectTextEntry {
	section := root.section("Triggers")
	if section == nil {
		return nil
	}
	var out []EffectTextEntry
	for triggerIndex, trigger := range section.list("trigger_data") {
		triggerName, _ := trigger.stringValue("trigger_name")
		for effectIndex, effect := range trigger.list("effect_data") {
			message, _ := effect.stringValue("message")
			message = strings.TrimSpace(message)
			if message == "" {
				continue
			}
			effectType, _ := effect.intValue("effect_type")
			stringID, _ := effect.intValue("string_id")
			entry := EffectTextEntry{
				TriggerIndex: triggerIndex,
				TriggerName:  triggerName,
				EffectIndex:  effectIndex,
				Type:         effectType,
				StringID:     stringID,
				Text:         message,
			}
			out = append(out, entry)
		}
	}
	return out
}

func (r StringReport) allText() []string {
	var out []string
	for _, text := range r.Messages {
		out = append(out, text)
	}
	for _, entry := range r.StringTable {
		out = append(out, entry.Text)
	}
	for _, entry := range r.TriggerText {
		out = append(out, entry.Text)
	}
	for _, entry := range r.EffectText {
		out = append(out, entry.Text)
	}
	return out
}

func summarizeScenarioMarkup(texts []string) MarkupSummary {
	colors := map[string]int{}
	variables := map[string]int{}
	for _, text := range texts {
		for _, match := range scenarioMarkupTagRE.FindAllStringSubmatch(text, -1) {
			if len(match) < 2 {
				continue
			}
			tag := strings.TrimSpace(match[1])
			if tag == "" {
				continue
			}
			upper := strings.ToUpper(tag)
			if scenarioColorTag(upper) {
				colors[upper]++
				continue
			}
			if strings.HasPrefix(upper, "VARIABLE ") || strings.HasPrefix(upper, "KILLCOUNT") {
				variables[tag]++
			}
		}
	}
	return MarkupSummary{ColorTags: colors, VariableRefs: sortedStringKeys(variables), VariableRefCounts: variables}
}

func scenarioColorTag(tag string) bool {
	switch tag {
	case "RED", "GREEN", "YELLOW", "BLUE", "AQUA", "PURPLE", "ORANGE", "GREY":
		return true
	default:
		return false
	}
}

func sortedStringKeys(values map[string]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
