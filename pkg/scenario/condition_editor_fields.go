package scenario

import (
	"fmt"
	"strings"
)

// ConditionEditorSchema describes the fields shown by the DE scenario editor
// when a condition is first created.
type ConditionEditorSchema struct {
	Type          int               `json:"type"`
	Name          string            `json:"name"`
	VisibleFields []string          `json:"visible_fields"`
	RevealedBy    map[string]string `json:"revealed_by,omitempty"`
}

// ConditionEditorSchemaForType returns the observed default-state editor
// schema for a condition type.
func ConditionEditorSchemaForType(conditionType int) (ConditionEditorSchema, bool) {
	row, ok := conditionEditorFields[conditionType]
	if !ok {
		return ConditionEditorSchema{}, false
	}
	return ConditionEditorSchema{
		Type:          conditionType,
		Name:          row.name,
		VisibleFields: append([]string(nil), row.fields...),
		RevealedBy:    copyStringMap(row.revealedBy),
	}, true
}

// ConditionEditorSchemas returns all observed condition schemas in type order.
func ConditionEditorSchemas() []ConditionEditorSchema {
	rows := make([]ConditionEditorSchema, 0, len(conditionEditorFields))
	for conditionType := range conditionEditorFields {
		row, _ := ConditionEditorSchemaForType(conditionType)
		rows = append(rows, row)
	}
	for i := 1; i < len(rows); i++ {
		for j := i; j > 0 && rows[j].Type < rows[j-1].Type; j-- {
			rows[j], rows[j-1] = rows[j-1], rows[j]
		}
	}
	return rows
}

type conditionEditorFieldRow struct {
	name       string
	fields     []string
	revealedBy map[string]string
}

// This is an owned observation of the editor UI. It is deliberately kept as
// labels rather than a dependency on the private Windows scan or its crops.
const conditionEditorFieldsTSV = `
1	Bring Object to Area	Set Object; Set Area; Inverse Condition
2	Bring Object to Object	Set Object; Next Object; Inverse Condition
3	Own Objects	Source Player; Object Group; Object List Type; Quantity; Object List; Include Changeable Weapon Objects; Object Type
4	Own Fewer Objects	Source Player; Object Group; Object List Type; Quantity; Set Area; Object List; Include Changeable Weapon Objects; Object Type
5	Objects in Area	Source Player; Object State; Object Group; Object List Type; Quantity; Set Area; Inverse Condition; Object List; Include Changeable Weapon Objects; Object Type
6	Destroy Object	Set Object; Inverse Condition
7	Capture Object	Source Player; Set Object; Inverse Condition
8	Accumulate Attribute	Source Player; Quantity; Inverse Condition; Tribute List
9	Research Technology	Source Player; Technologies; Inverse Condition
10	Timer	Inverse Condition; Timer
11	Object Selected	Set Object; Inverse Condition
12	AI Signal	AI Signal Value; Inverse Condition
13	Player Defeated	Source Player; Inverse Condition
14	Object Has Target	Set Object; Object Group; Object List Type; Next Object; Inverse Condition; Object List; Object Type
15	Object Visible	Set Object
16	Object Not Visible	Set Object
17	Researching Tech	Source Player; Technologies; Inverse Condition
18	Units Garrisoned	Set Object; Quantity; Inverse Condition
19	Difficulty Level	Difficulty Level; Inverse Condition
20	Chance	Quantity
21	Technology State	Source Player; Technologies; Inverse Condition; Technology State
22	Variable Value	Destination Variable; Quantity; Comparison; Inverse Condition
23	Object HP	Set Object; Quantity; Comparison; Inverse Condition
24	Diplomacy State	Source Player; Diplomacy State; Target Player; Inverse Condition
25	Script Call	Message
26	Object Selected (Multiplayer)	Source Player; Set Object; Inverse Condition
27	Object Visible (Multiplayer)	Source Player; Set Object; Inverse Condition; Allow In Fog
28	Object Has Action	Set Object; Object Group; Object List Type; Unit AI Action; Next Object; Inverse Condition; Object List; Object Type
29	Condition Separator	
30	Multiplayer AI Signal	AI Signal Value; Inverse Condition
54	Building Is Trading	Set Object; Inverse Condition
55	Display Timer Triggered	Inverse Condition; Timer ID
56	Victory Timer	Source Player; Victory Timer Type; Quantity; Comparison; Inverse Condition
57	Condition Joiner	
75	Decision Triggered	Decision ID; Inverse Condition; Decision option
76	Object Attacked	Source Player; Set Object; Object Group; Object List Type; Quantity; Inverse Condition; Object List; Object Type
77	Hero Power Cast	Source Player
78	Compare Variables	Destination Variable; Modifier Variable; Comparison; Inverse Condition
79	Trigger Active	Trigger List; Inverse Condition
80	Local Tech Researched	Source Player; Set Object; Local Technologies; Quantity; Set Area; Inverse Condition
`

var conditionEditorFields = parseConditionEditorFields(conditionEditorFieldsTSV)

func parseConditionEditorFields(table string) map[int]conditionEditorFieldRow {
	rows := make(map[int]conditionEditorFieldRow)
	for _, line := range strings.Split(strings.TrimSpace(table), "\n") {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		var conditionType int
		if _, err := fmt.Sscanf(parts[0], "%d", &conditionType); err != nil {
			continue
		}
		fields := []string(nil)
		if parts[2] != "" {
			for _, field := range strings.Split(parts[2], "; ") {
				fields = append(fields, field)
			}
		}
		rows[conditionType] = conditionEditorFieldRow{name: parts[1], fields: fields}
	}
	return rows
}
