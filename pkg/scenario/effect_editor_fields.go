package scenario

import (
	"fmt"
	"strings"
)

// EffectEditorSchema describes the fields shown by the DE scenario editor
// when an effect is first created. Some controls reveal additional fields
// after Set Area, Set Objects, or Set Location is used.
type EffectEditorSchema struct {
	Type          int               `json:"type"`
	Name          string            `json:"name"`
	VisibleFields []string          `json:"visible_fields"`
	RevealedBy    map[string]string `json:"revealed_by,omitempty"`
	Offered       bool              `json:"offered"`
}

// EffectEditorSchemaForType returns the observed default-state editor schema
// for an effect type. The returned slice is independent of the package table.
func EffectEditorSchemaForType(effectType int) (EffectEditorSchema, bool) {
	row, ok := effectEditorFields[effectType]
	if !ok {
		return EffectEditorSchema{}, false
	}
	fields := append([]string(nil), row.fields...)
	revealedBy := make(map[string]string, len(row.revealedBy))
	for field, control := range row.revealedBy {
		revealedBy[field] = control
	}
	return EffectEditorSchema{Type: effectType, Name: row.name, VisibleFields: fields, RevealedBy: revealedBy, Offered: row.offered}, true
}

// EffectEditorSchemas returns all observed effect schemas in type order.
func EffectEditorSchemas() []EffectEditorSchema {
	rows := make([]EffectEditorSchema, 0, len(effectEditorFields))
	for effectType := 1; effectType <= 109; effectType++ {
		if row, ok := EffectEditorSchemaForType(effectType); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

type effectEditorFieldRow struct {
	name       string
	fields     []string
	revealedBy map[string]string
	offered    bool
}

var effectEditorRevealedBy = map[int]map[string]string{
	32: {
		"X1":                 "Set Area",
		"Y1":                 "Set Area",
		"X2":                 "Set Area",
		"Y2":                 "Set Area",
		"Max Units Affected": "Set Area",
	},
}

// This is an owned observation of the editor UI. It is deliberately kept as
// labels rather than a dependency on the private Windows scan or its crops.
const effectEditorFieldsTSV = `
1	Change Diplomacy	Source Player; Diplomacy State; Target Player; Mutual
2	Research Technology	Source Player; Technologies; Force; Item ID
3	Send Chat	Source Player; String Table; Sound (Event) Name; Message
4	Play Sound	Source Player; Global Sound; Sound (Event) Name; Set Location
5	Tribute	Source Player; Target Player; Quantity; Tribute List
6	Unlock Gate	Set Gate
7	Lock Gate	Set Gate
8	Activate Trigger	Trigger List
9	Deactivate Trigger	Trigger List
10	AI Script Goal	AI Trigger Number
11	Create Object	Disable Sound; Source Player; Object List Type; Facet; Object List; Item ID; Set Location
12	Task Object	Source Player; Issue command to formation; Queue Action; Set Objects; Object Group; Object List Type; Action Type; Set Area; Object List; Disable Garrison Unload Sound; Object Type; Set Location
13	Declare Victory	Source Player; Victory
14	Kill Object	Source Player; Set Objects; Object Group; Object List Type; Set Area; Object List; Object Type
15	Remove Object	Source Player; Set Objects; Object State; Object Group; Object List Type; Set Area; Object List; Object Type
16	Change View	Source Player; Quantity; Scroll; Set Location
17	Unload	Source Player; Set Objects; Object Group; Object List Type; Object List; Object Type; Set Location
18	Change Ownership	Source Player; Target Player; Set Objects; Object Group; Object List Type; Set Area; Object List; Object Type; Flash Objects
19	Patrol	Source Player; Set Objects; Object Group; Object List Type; Set Area; Object List; Object Type; Set Location
20	Display Instructions	Source Player; Play Sound; Tag Color For Icon; String Table; Sound (Event) Name; Object List Type; Panel Location; Message; Object Icon; Object List; Timer
21	Clear Instructions	Panel Location
22	Freeze Object	Source Player; Set Objects; Object Group; Object List Type; Set Area; Object List; Object Type
23	Use Advanced Buttons	
24	Damage Object	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Object List; Object Type
25	Place Foundation	Source Player; Object List Type; Object List; Set Location
26	Change Object Name	Source Player; Set Objects; String Table ID; Object List Type; Set Area; Object List; Name
27	Change Object HP	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Object List; Operation; Object Type
28	Change Object Attack	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Armor/Attack Type; Object List; Operation; Object Type
29	Stop Object	Source Player; Set Objects; Object Group; Object List Type; Set Area; Object List; Object Type
30	Attack Move	Source Player; Set Objects; Object Group; Object List Type; Set Area; Object List; Object Type; Set Location
31	Change Object Armor	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Armor/Attack Type; Object List; Operation; Object Type
32	Change Object Range	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Object List; Operation; Object Type
33	Change Object Speed	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Object List; Object Type
34	Heal Object	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Object List; Object Type
35	Teleport Object	Source Player; Set Objects; Object Group; Object List Type; Set Area; Object List; Object Type; Set Location
36	Change Object Stance	Source Player; Set Objects; Object Group; Object List Type; Set Area; Set Attack Stance; Object List; Object Type
37	Display Timer	String Table; Timer Unit; Message; Timer ID; Timer
38	Enable/Disable Object	Source Player; Object List Type; Enabled; Object List; Item ID
39	Enable/Disable Technology	Source Player; Technologies; Enabled; Item ID
40	Change Object Cost	Source Player; Resource 1; Object List Type; Resource 2; Resource 3; Resource 1 Quantity; Object List; Resource 2 Quantity; Resource 3 Quantity
41	Set Player Visibility	Source Player; Target Player; Visiblity State
42	Change Object Icon	Source Player; Set Objects; Object Group; Object List Type; Object Icon; Set Area; Object List; Object Type
43	Replace Object	Source Player; Target Player; Set Objects; Object Group; Object List Type; Set Area; Object List; Object Type; Facet
44	Change Object Description	Source Player; String Table; Object List Type; Message; Object List
45	Change Player Name	Source Player; String Table ID; Name
46	Change Train Location	Source Player; Object List Type; Object List; Button Location
47	Change Technology Location	Source Player; Technologies; Object List Type; Object List; Button Location
48	Change Civilization Name	Source Player; String Table ID; Name
49	Create Garrisoned Object	Disable Sound; Source Player; Set Objects; Object List Type; Set Area; Object List
50	Acknowledge AI Signal	AI Signal Value
51	Modify Attribute	Source Player; Object List Type; Object Attributes; Name; Quantity; Object List; Operation; Item ID
52	Modify Resource	Source Player; Quantity; Tribute List; Operation; Item ID
53	Modify Resource By Variable	Source Player; Quantity; Tribute List; Operation; Item ID
54	Set Building Gather Point	Source Player; Set Objects; Object List Type; Set Area; Object List; Set Location
55	Script Call	String Table; Message
56	Change Variable	Variable; Quantity; Operation; Name
57	Clear Timer	Timer ID
58	Change Object Player Color	Source Player; Set Objects; Color; Object List Type; Set Area; Object List
59	Change Object Civilization Name	Source Player; Set Objects; String Table ID; Set Area; Name
60	Change Object Player Name	Source Player; Set Objects; String Table ID; Object List Type; Set Area; Object List; Name
61	Disable Unit Targetable State	Source Player; Set Objects; Object List Type; Set Area; Object List
62	Enable Unit Targetable State	Source Player; Set Objects; Object List Type; Set Area; Object List
63	Change Technology Cost	Source Player; Resource 1; Technologies; Resource 2; Resource 3; Resource 1 Quantity; Resource 2 Quantity; Resource 3 Quantity
64	Change Technology Research Time	Source Player; Technologies; Quantity
65	Change Technology Name	Source Player; String Table; Technologies; Message
66	Change Technology Description	Source Player; String Table; Technologies; Message
67	Enable Technology Stacking	Source Player; Technologies; Quantity
68	Disable Technology Stacking	Source Player; Technologies
69	Acknowledge Multiplayer AI Signal	AI Signal Value
70	Disable Object Selection	Source Player; Set Objects; Object List Type; Set Area; Object List
71	Enable Object Selection	Source Player; Set Objects; Object List Type; Set Area; Object List
72	Change Color Mood	Color Mood; Quantity
73	Enable Object Deletion	Source Player; Set Objects; Object List Type; Set Area; Object List
74	Disable Object Deletion	Source Player; Set Objects; Object List Type; Set Area; Object List
75	Train Unit	Source Player; Set Objects; Object List Type; Quantity; Set Area; Object List; Set Location
76	Initiate Research	Source Player; Set Objects; Technologies
77	Create Object Attack	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Armor/Attack Type; Object List; Operation; Object Type
78	Create Object Armor	Source Player; Set Objects; Object Group; Object List Type; Quantity; Set Area; Armor/Attack Type; Object List; Operation; Object Type
79	Modify Attribute By Variable	Source Player; Object List Type; Object Attributes; Name; Quantity; Object List; Operation; Item ID
80	Set Object Cost	Source Player; Object List Type; Quantity; Tribute List; Object List; Item ID
81	Load Key Value	Variable; Name; Quantity
82	Store Key Value	Destination Variable; Name
83	Delete Key	Name
84	Change Technology Icon	Source Player; Technologies; Quantity; Object Icon
85	Change Technology Hotkey	Source Player; Technologies; Quantity
86	Modify Variable By Resource	Source Player; Variable; Tribute List; Operation; Item ID
87	Modify Variable By Attribute	Source Player; Object List Type; Object Attributes; Name; Variable; Object List; Operation; Item ID
88	Change Object Caption	Source Player; Set Objects; String Table ID; Object List Type; Set Area; Object List; Name
89	Change Player Color	Source Player; Color
90	Create Decision	String Table; String Table ID; String; Table ID; Decision ID; Message; Option 1; Option
91	Achievement Trigger	(not offered in the editor dropdown)
92	Perth Tutorial Trigger	(not offered in the editor dropdown)
93	Create WS Prompt	(not offered in the editor dropdown)
94	Remove WS Prompt	(not offered in the editor dropdown)
95	Create Campaign Goal	(not offered in the editor dropdown)
96	Set Campaign Goal State	(not offered in the editor dropdown)
97	Open Campaign Goal Dialog	(not offered in the editor dropdown)
98	Disable Unit Attackable State	Source Player; Set Objects; Object List Type; Set Area; Object List
99	Enable Unit Attackable State	Source Player; Set Objects; Object List Type; Set Area; Object List
100	Modify Variable By Variable	Destination Variable; Modifier Variable; Operation
101	Count Units Into Variable	Source Player; Object Group; Object List Type; Variable; Set Area; Object List; Object Type
102	Add Train Location	Source Player; Object List Type; Object List; Button Location; Train Time; Hotkey
103	Research Local Technology	Source Player; Set Objects; Local Technologies; Object List Type; Set Area; Object List
104	Modify Attribute For Class	Source Player; Object List Type; Object Attributes; Name; Quantity; Object List; Operation; Item ID
105	Modify Object Attribute	Source Player; Object Modified State; Name; Set Objects; Object List Type; Object Attributes; Quantity; Set Area; Object List; Operation; Item ID
106	Modify Object Attribute By Variable	Source Player; Name; Object Modified State; Set Objects; Object List Type; Object Attributes; Quantity; Set Area; Object List; Operation; Item ID
107	Change Object Visibility	Source Player; Target Player; Set Objects; Visiblity State; Set Area
108	Build Object	Source Player; Issue command to formation; Queue Action; Set Objects; Object Group; Object List Type; Building List; Set Area; Object List; Object Type; Set Location; Set Wall
109	Mirror Diplomacy	Source Player; Target Player; Enabled
`

var effectEditorFields = parseEffectEditorFields(effectEditorFieldsTSV)

func parseEffectEditorFields(table string) map[int]effectEditorFieldRow {
	rows := make(map[int]effectEditorFieldRow)
	for _, line := range strings.Split(strings.TrimSpace(table), "\n") {
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		var effectType int
		if _, err := fmt.Sscanf(parts[0], "%d", &effectType); err != nil {
			continue
		}
		fields := []string(nil)
		if parts[2] != "" && !strings.HasPrefix(parts[2], "(") {
			for _, field := range strings.Split(parts[2], "; ") {
				fields = append(fields, field)
			}
		}
		rows[effectType] = effectEditorFieldRow{name: parts[1], fields: fields, revealedBy: effectEditorRevealedBy[effectType], offered: !strings.HasPrefix(parts[2], "(")}
	}
	return rows
}
