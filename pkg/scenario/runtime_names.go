package scenario

import (
	"aoe2kit/pkg/gamestrings"
	"strconv"
)

// RuntimeNames provides optional names from the user's installed game.
// A nil table means callers should use Kit's compatibility fallback names.
type RuntimeNames struct {
	Strings gamestrings.Table
}

func (n *RuntimeNames) effect(id int) string {
	if n != nil {
		if value, ok := n.Strings.Lookup(11550 + id); ok {
			return value
		}
	}
	return EffectTypeName(id)
}

// EffectName resolves an effect type using runtime strings, then Kit's
// compatibility vocabulary.
func (n *RuntimeNames) EffectName(id int) string { return n.effect(id) }

func (n *RuntimeNames) condition(id int) string {
	if n != nil {
		if value, ok := n.Strings.Lookup(10500 + id); ok {
			return value
		}
	}
	return ConditionTypeName(id)
}

// ConditionName resolves a condition type using runtime strings.
func (n *RuntimeNames) ConditionName(id int) string { return n.condition(id) }

func (n *RuntimeNames) attribute(id int) string {
	if n != nil {
		if value, ok := n.Strings.Lookup(15000 + id); ok {
			return value
		}
	}
	return "attribute_" + itoa(id)
}

// AttributeName resolves a resource/attribute id using the game's table.
func (n *RuntimeNames) AttributeName(id int) string { return n.attribute(id) }

func (n *RuntimeNames) objectGroup(id int) string {
	if n != nil {
		if value, ok := n.Strings.Lookup(13300 + id); ok {
			return value
		}
	}
	return "class_" + itoa(id)
}

// ObjectGroupName resolves a unit-class id using the game's table.
func (n *RuntimeNames) ObjectGroupName(id int) string { return n.objectGroup(id) }

func itoa(v int) string {
	return strconv.Itoa(v)
}
