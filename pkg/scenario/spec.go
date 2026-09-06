package scenario

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed specs/de_v1_58_structure.json
var specFS embed.FS

type Spec struct {
	Sections []NamedSection
}

func (s *Spec) section(name string) (SectionSpec, bool) {
	for _, section := range s.Sections {
		if section.Name == name {
			return section.Spec, true
		}
	}
	return SectionSpec{}, false
}

type NamedSection struct {
	Name string
	Spec SectionSpec
}

type SectionSpec struct {
	Retrievers []RetrieverSpec
	Structs    map[string]SectionSpec
}

type RetrieverSpec struct {
	Name         string
	Type         string                  `json:"type"`
	Repeat       int                     `json:"repeat"`
	IsList       *bool                   `json:"is_list"`
	Default      json.RawMessage         `json:"default"`
	Dependencies map[string][]Dependency `json:"-"`
}

type Dependency struct {
	Action string
	Target []Target
	Eval   string
}

type Target struct {
	Section string
	Name    string
}

func LoadCurrentDESpec() (*Spec, error) {
	data, err := specFS.ReadFile("specs/de_v1_58_structure.json")
	if err != nil {
		return nil, err
	}
	return parseSpec(data)
}

func LoadDESpecForVersion(version string) (*Spec, error) {
	spec, err := LoadCurrentDESpec()
	if err != nil {
		return nil, err
	}
	switch version {
	case "1.55":
		patchDE155To157EffectSpec(spec, version)
		patchDE155Spec(spec)
	case "1.56", "1.57":
		patchDE155To157EffectSpec(spec, version)
	case "1.58":
		// Current embedded spec.
	default:
		return nil, fmt.Errorf("unsupported scenario version %q", version)
	}
	return spec, nil
}

func patchDE155To157EffectSpec(spec *Spec, version string) {
	remove := map[string]bool{
		"object_filter":          true,
		"use_tag_color_for_icon": true,
	}
	switch version {
	case "1.55":
		remove["issue_group_command"] = true
		remove["queue_action"] = true
		remove["mutual_diplomacy"] = true
		remove["building_list"] = true
		remove["wall_x1"] = true
		remove["wall_y1"] = true
		remove["wall_x2"] = true
		remove["wall_y2"] = true
	case "1.56":
		remove["mutual_diplomacy"] = true
		remove["building_list"] = true
		remove["wall_x1"] = true
		remove["wall_y1"] = true
		remove["wall_x2"] = true
		remove["wall_y2"] = true
	}
	for i := range spec.Sections {
		if spec.Sections[i].Name != "Triggers" {
			continue
		}
		triggerSpec := spec.Sections[i].Spec.Structs["TriggerStruct"]
		effectSpec := triggerSpec.Structs["EffectStruct"]
		filtered := effectSpec.Retrievers[:0]
		for _, retriever := range effectSpec.Retrievers {
			if remove[retriever.Name] {
				continue
			}
			if retriever.Name == "static_value_83" {
				switch version {
				case "1.55":
					retriever.Name = "static_value_80"
				case "1.56":
					retriever.Name = "static_value_81"
				case "1.57":
					retriever.Name = "static_value_81"
				}
			}
			filtered = append(filtered, retriever)
		}
		effectSpec.Retrievers = filtered
		triggerSpec.Structs["EffectStruct"] = effectSpec
		spec.Sections[i].Spec.Structs["TriggerStruct"] = triggerSpec
		return
	}
}

func patchDE155Spec(spec *Spec) {
	for i := range spec.Sections {
		if spec.Sections[i].Name != "DataHeader" {
			continue
		}
		playerData := spec.Sections[i].Spec.Structs["PlayerDataOneStruct"]
		playerData.Retrievers = []RetrieverSpec{
			{Name: "active", Type: "u32", Repeat: 1, Dependencies: map[string][]Dependency{}},
			{Name: "human", Type: "u32", Repeat: 1, Dependencies: map[string][]Dependency{}},
			{Name: "civilization", Type: "u32", Repeat: 1, Dependencies: map[string][]Dependency{}},
			{Name: "architecture_set", Type: "u32", Repeat: 1, Dependencies: map[string][]Dependency{}},
			{Name: "cty_mode", Type: "u32", Repeat: 1, Dependencies: map[string][]Dependency{}},
		}
		spec.Sections[i].Spec.Structs["PlayerDataOneStruct"] = playerData
		return
	}
}

func parseSpec(data []byte) (*Spec, error) {
	sections, err := decodeOrderedSections(data)
	if err != nil {
		return nil, err
	}
	return &Spec{Sections: sections}, nil
}

func decodeOrderedSections(data []byte) ([]NamedSection, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("spec root is not an object")
	}
	var sections []NamedSection
	for dec.More() {
		nameTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		name, ok := nameTok.(string)
		if !ok {
			return nil, fmt.Errorf("section name token is %T", nameTok)
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		section, err := decodeSection(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		sections = append(sections, NamedSection{Name: name, Spec: section})
	}
	return sections, nil
}

func decodeSection(data []byte) (SectionSpec, error) {
	var raw struct {
		Retrievers json.RawMessage            `json:"retrievers"`
		Structs    map[string]json.RawMessage `json:"structs"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return SectionSpec{}, err
	}
	retrievers, err := decodeRetrievers(raw.Retrievers)
	if err != nil {
		return SectionSpec{}, err
	}
	structs := map[string]SectionSpec{}
	for name, structRaw := range raw.Structs {
		structSpec, err := decodeSection(structRaw)
		if err != nil {
			return SectionSpec{}, fmt.Errorf("struct %s: %w", name, err)
		}
		structs[name] = structSpec
	}
	return SectionSpec{Retrievers: retrievers, Structs: structs}, nil
}

func decodeRetrievers(data []byte) ([]RetrieverSpec, error) {
	if len(data) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return nil, fmt.Errorf("retrievers is not an object")
	}
	var out []RetrieverSpec
	for dec.More() {
		nameTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		name, ok := nameTok.(string)
		if !ok {
			return nil, fmt.Errorf("retriever name token is %T", nameTok)
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		var spec RetrieverSpec
		if err := json.Unmarshal(raw, &spec); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		if spec.Repeat == 0 {
			spec.Repeat = 1
		}
		spec.Name = name
		spec.Dependencies = map[string][]Dependency{}
		var depRaw struct {
			Dependencies map[string]json.RawMessage `json:"dependencies"`
		}
		if err := json.Unmarshal(raw, &depRaw); err != nil {
			return nil, err
		}
		for event, eventRaw := range depRaw.Dependencies {
			deps, err := decodeDependencyList(eventRaw)
			if err != nil {
				return nil, fmt.Errorf("%s dependency %s: %w", name, event, err)
			}
			spec.Dependencies[event] = deps
		}
		out = append(out, spec)
	}
	return out, nil
}

func decodeDependencyList(data []byte) ([]Dependency, error) {
	var many []json.RawMessage
	if len(data) > 0 && data[0] == '[' {
		if err := json.Unmarshal(data, &many); err != nil {
			return nil, err
		}
	} else {
		many = []json.RawMessage{data}
	}
	deps := make([]Dependency, 0, len(many))
	for _, raw := range many {
		var tmp struct {
			Action string          `json:"action"`
			Target json.RawMessage `json:"target"`
			Eval   string          `json:"eval"`
		}
		if err := json.Unmarshal(raw, &tmp); err != nil {
			return nil, err
		}
		targets, err := decodeTargets(tmp.Target)
		if err != nil {
			return nil, err
		}
		deps = append(deps, Dependency{Action: tmp.Action, Target: targets, Eval: tmp.Eval})
	}
	return deps, nil
}

func decodeTargets(data []byte) ([]Target, error) {
	if len(data) == 0 || string(data) == "null" {
		return nil, nil
	}
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		target, err := parseTarget(one)
		if err != nil {
			return nil, err
		}
		return []Target{target}, nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return nil, err
	}
	targets := make([]Target, 0, len(many))
	for _, raw := range many {
		target, err := parseTarget(raw)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func parseTarget(raw string) (Target, error) {
	for i := 0; i < len(raw); i++ {
		if raw[i] == ':' {
			return Target{Section: raw[:i], Name: raw[i+1:]}, nil
		}
	}
	return Target{}, fmt.Errorf("bad dependency target %q", raw)
}
