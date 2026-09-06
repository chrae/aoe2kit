package scenario

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ScenarioScanOptions struct {
	MaxInflatedBytes int
}

type ScenarioScanMeta struct {
	Index            int    `json:"index"`
	Name             string `json:"name,omitempty"`
	Description      string `json:"description,omitempty"`
	ShortDescription string `json:"short_description,omitempty"`
	Enabled          uint32 `json:"enabled"`
	Looping          int8   `json:"looping"`
	EffectCount      int    `json:"effect_count"`
	ConditionCount   int    `json:"condition_count"`
}

type triggerScanCallbacks struct {
	onMeta     func(ScenarioScanMeta)
	onEffect   func(ScenarioScanMeta, int, *parsedNode) error
	onVariable func(id int, name string)
}

type scanHeader struct {
	version         string
	headerBytes     int
	compressedBytes int
	inflatedBytes   int
}

func scanScenarioFile(path string, opts ScenarioScanOptions, callbacks triggerScanCallbacks) (scanHeader, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return scanHeader{}, err
	}
	return scanScenarioData(data, opts, callbacks)
}

func scanScenarioData(data []byte, opts ScenarioScanOptions, callbacks triggerScanCallbacks) (scanHeader, error) {
	currentSpec, err := LoadCurrentDESpec()
	if err != nil {
		return scanHeader{}, err
	}
	if len(data) < 8 {
		return scanHeader{}, errors.New("scenario too short")
	}
	fileHeaderSpec, ok := currentSpec.section("FileHeader")
	if !ok {
		return scanHeader{}, errors.New("embedded scenario spec has no FileHeader")
	}
	headerParser := parser{
		data:     data,
		sections: map[string]*parsedSection{},
	}
	headerNode, err := headerParser.parseNode("FileHeader", fileHeaderSpec, nil)
	if err != nil {
		return scanHeader{}, fmt.Errorf("parse FileHeader: %w", err)
	}
	version, _ := headerNode.stringValue("version")
	if !SupportsReadVersion(version) {
		return scanHeader{}, fmt.Errorf("unsupported scenario version %q; AoE2Kit read support currently covers DE 1.55 through 1.58", version)
	}
	bodySpec, err := LoadDESpecForVersion(version)
	if err != nil {
		return scanHeader{}, err
	}
	headerLen := headerParser.off
	maxBytes := opts.MaxInflatedBytes
	if maxBytes == 0 {
		maxBytes = maxInflatedScenarioBytes()
	}
	body, err := inflateRawLimited(data[headerLen:], maxBytes)
	if err != nil {
		return scanHeader{}, fmt.Errorf("inflate scenario body: %w", err)
	}
	p := parser{
		data:     body,
		sections: map[string]*parsedSection{},
	}
	for _, sectionSpec := range bodySpec.Sections {
		if sectionSpec.Name == "FileHeader" {
			continue
		}
		if sectionSpec.Name == "Triggers" {
			if err := p.scanTriggersSection(sectionSpec.Spec, callbacks); err != nil {
				return scanHeader{}, fmt.Errorf("section Triggers: %w", err)
			}
			return scanHeader{
				version:         version,
				headerBytes:     headerLen,
				compressedBytes: len(data) - headerLen,
				inflatedBytes:   len(body),
			}, nil
		}
		section, err := p.parseNodeDiscard(sectionSpec.Name, sectionSpec.Spec)
		if err != nil {
			return scanHeader{}, fmt.Errorf("section %s: %w", sectionSpec.Name, err)
		}
		p.sections[sectionSpec.Name] = section
	}
	return scanHeader{}, errors.New("missing Triggers section")
}

func (p *parser) scanTriggersSection(sectionSpec SectionSpec, callbacks triggerScanCallbacks) error {
	triggerSpec, ok := sectionSpec.Structs["TriggerStruct"]
	if !ok {
		return errors.New("missing TriggerStruct")
	}
	node := &parsedNode{Name: "Triggers", Start: p.off}
	for _, retriever := range sectionSpec.Retrievers {
		repeat, err := p.constructRepeat(retriever, node)
		if err != nil {
			return fmt.Errorf("%s: %w", retriever.Name, err)
		}
		if repeat < 0 {
			repeat = 0
		}
		field := &parsedNode{Name: retriever.Name, Start: p.off}
		switch retriever.Name {
		case "trigger_data":
			for i := 0; i < repeat; i++ {
				if err := p.scanTrigger(triggerSpec, i, callbacks); err != nil {
					return fmt.Errorf("TriggerStruct[%d]: %w", i, err)
				}
			}
			field.Value = []any{}
			field.End = p.off
		case "variable_data":
			variableSpec, ok := sectionSpec.Structs["VariableStruct"]
			if !ok {
				return errors.New("missing VariableStruct")
			}
			for i := 0; i < repeat; i++ {
				variable, err := p.parseNodeDiscard("VariableStruct", variableSpec)
				if err != nil {
					return fmt.Errorf("VariableStruct[%d]: %w", i, err)
				}
				if callbacks.onVariable != nil {
					id, _ := variable.intValue("variable_id")
					name, _ := variable.stringValue("variable_name")
					callbacks.onVariable(id, name)
				}
			}
			field.Value = []any{}
			field.End = p.off
		default:
			if strings.HasPrefix(retriever.Type, "struct:") {
				structName := strings.TrimPrefix(retriever.Type, "struct:")
				structSpec, ok := sectionSpec.Structs[structName]
				if !ok {
					return fmt.Errorf("missing struct spec %s", structName)
				}
				for i := 0; i < repeat; i++ {
					if _, err := p.parseNodeDiscard(structName, structSpec); err != nil {
						return fmt.Errorf("%s[%d]: %w", structName, i, err)
					}
				}
				field.Value = []any{}
				field.End = p.off
			} else if err := p.parsePrimitiveFieldDiscard(field, retriever, repeat); err != nil {
				return err
			}
		}
		node.Fields = append(node.Fields, field)
	}
	return nil
}

func (p *parser) scanTrigger(triggerSpec SectionSpec, index int, callbacks triggerScanCallbacks) error {
	effectSpec, ok := triggerSpec.Structs["EffectStruct"]
	if !ok {
		return errors.New("missing EffectStruct")
	}
	conditionSpec, ok := triggerSpec.Structs["ConditionStruct"]
	if !ok {
		return errors.New("missing ConditionStruct")
	}
	node := &parsedNode{Name: "TriggerStruct", Start: p.off}
	meta := ScenarioScanMeta{Index: index}
	for _, retriever := range triggerSpec.Retrievers {
		repeat, err := p.constructRepeat(retriever, node)
		if err != nil {
			return fmt.Errorf("%s: %w", retriever.Name, err)
		}
		if repeat < 0 {
			repeat = 0
		}
		field := &parsedNode{Name: retriever.Name, Start: p.off}
		switch retriever.Name {
		case "effect_data":
			for i := 0; i < repeat; i++ {
				effect, err := p.parseNodeDiscard("EffectStruct", effectSpec)
				if err != nil {
					return fmt.Errorf("EffectStruct[%d]: %w", i, err)
				}
				if callbacks.onEffect != nil {
					if err := callbacks.onEffect(meta, i, effect); err != nil {
						return err
					}
				}
			}
			field.Value = []any{}
			field.End = p.off
		case "condition_data":
			for i := 0; i < repeat; i++ {
				if _, err := p.parseNodeDiscard("ConditionStruct", conditionSpec); err != nil {
					return fmt.Errorf("ConditionStruct[%d]: %w", i, err)
				}
			}
			field.Value = []any{}
			field.End = p.off
		default:
			if strings.HasPrefix(retriever.Type, "struct:") {
				structName := strings.TrimPrefix(retriever.Type, "struct:")
				structSpec, ok := triggerSpec.Structs[structName]
				if !ok {
					return fmt.Errorf("missing struct spec %s", structName)
				}
				for i := 0; i < repeat; i++ {
					if _, err := p.parseNodeDiscard(structName, structSpec); err != nil {
						return fmt.Errorf("%s[%d]: %w", structName, i, err)
					}
				}
				field.Value = []any{}
				field.End = p.off
			} else if err := p.parsePrimitiveFieldDiscard(field, retriever, repeat); err != nil {
				return err
			}
		}
		switch retriever.Name {
		case "enabled":
			meta.Enabled = uint32(asInt(field.Value))
		case "looping":
			meta.Looping = int8(asInt(field.Value))
		case "trigger_description":
			meta.Description, _ = field.Value.(string)
		case "trigger_name":
			meta.Name, _ = field.Value.(string)
		case "short_description":
			meta.ShortDescription, _ = field.Value.(string)
		case "number_of_effects":
			meta.EffectCount = asInt(field.Value)
		case "number_of_conditions":
			meta.ConditionCount = asInt(field.Value)
		}
		node.Fields = append(node.Fields, field)
	}
	if callbacks.onMeta != nil {
		callbacks.onMeta(meta)
	}
	return nil
}

func effectTypeMatchesQuery(effectType int, query string) bool {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return false
	}
	if n, err := strconv.Atoi(query); err == nil {
		return effectType == n
	}
	name := strings.ToLower(EffectTypeName(effectType))
	if name == query {
		return true
	}
	return strings.Contains(name, query)
}

func sortedIntCounts(counts map[int]int) []IntCount {
	out := make([]IntCount, 0, len(counts))
	for id, count := range counts {
		out = append(out, IntCount{ID: id, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].ID < out[j].ID
		}
		return out[i].Count > out[j].Count
	})
	return out
}

func sortedStringCounts(counts map[string]int) []StringCount {
	out := make([]StringCount, 0, len(counts))
	for value, count := range counts {
		out = append(out, StringCount{Value: value, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Value < out[j].Value
		}
		return out[i].Count > out[j].Count
	})
	return out
}

type IntCount struct {
	ID    int `json:"id"`
	Count int `json:"count"`
}

type StringCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}
