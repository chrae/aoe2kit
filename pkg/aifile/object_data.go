package aifile

const objectDataSource = "https://airef.github.io/parameters/parameters-details.html#ObjectData"

// Meanings and caveats summarize the pinned reference. This table supplies
// source evidence only, never evidence that a runtime read succeeds.
func objectDataConstants() map[string]ReferenceConstant {
	entry := func(value int64, meaning, units, context string) ReferenceConstant {
		return ReferenceConstant{Value: value, ReferenceFamily: "UP", Source: objectDataSource, Meaning: meaning, Units: units, Context: context, RuntimeAvailability: "unverified"}
	}
	entries := map[string]ReferenceConstant{
		"object-data-id":               entry(0, "Object map identifier", "map object id", "Not a DAT unit type id."),
		"object-data-type":             entry(1, "Object type selector", "unit type id", "Guide reports base-line behavior outside scenarios; upgraded identity is context-dependent."),
		"object-data-class":            entry(2, "Object class", "class id", "Not the target's class; use object-data-target for that."),
		"object-data-action":           entry(5, "Current object action", "ActionId", "Guide reports -1 while idle or marching in formation; availability is not guaranteed."),
		"object-data-order":            entry(6, "Current object order", "OrderId", "Guide reports -1 when no order is executing or while marching in formation."),
		"object-data-target":           entry(7, "Class of the object's target", "ClassId", "Not a target object id. Formation/absent-target values differ in the guide: UP -1 versus DE -2."),
		"object-data-hitpoints":        entry(10, "Current hit points", "hit points", "Guide includes damage already suffered; this is not the base DAT HP field."),
		"object-data-maxhp":            entry(11, "Maximum hit points", "hit points", "Guide includes researched techs and civ bonuses; static DAT alone does not establish this runtime value."),
		"object-data-range":            entry(12, "Current range", "reference-unspecified", "Guide includes tech and civ effects; numeric scale is not independently established here."),
		"object-data-status":           entry(19, "Current object status", "ObjectStatus", "Object status is separate from action, order and attack stance."),
		"object-data-player":           entry(20, "Object player number", "player number", "Guide range 0..8; this catalogue does not establish current ownership."),
		"object-data-action-time":      entry(22, "Most recent movement time", "milliseconds", "Guide reports an UP attacking/reset ambiguity; this is not a verified universal last-action timestamp."),
		"object-data-target-id":        entry(23, "Identifier of the object's target", "map object id", "Guide reports -1 for no target or formation marching; not a DAT unit type id."),
		"object-data-formation-id":     entry(24, "Current formation state identifier", "formation state id", "Guide says this does not match the Formation parameter enum and may vary with action; reported -1 outside formation."),
		"object-data-distance":         entry(44, "Distance to the target point", "map distance", "Guide uses a normal target point and ignores obstacles; a precise target point needs a different selector."),
		"object-data-precise-distance": entry(45, "Scaled distance to the target point", "distance multiplied by 100", "Guide still expects a normal target point and ignores obstacles; scaled output does not imply a precise input point."),
		"object-data-reload-time":      entry(54, "Configured interval between attacks", "milliseconds", "Guide includes tech and civ effects; not the remaining attack cooldown."),
		"object-data-next-attack":      entry(55, "Countdown until the next attack is available", "milliseconds", "Guide says countdown starts at animation start; this is not a confirmed damage or projectile-release event."),
		"object-data-min-range":        entry(62, "Current minimum range", "reference-unspecified", "Guide includes tech and civ effects; no additional numeric scale is inferred."),
		"object-data-target-time":      entry(63, "Time attacking the current target", "milliseconds", "Guide reports -2 outside attacks and possible negative villager/herding values; no runtime behavior is verified."),
		"object-data-frame-delay":      entry(78, "Frames before projectile launch", "animation frames", "Not milliseconds. Frame duration varies; this does not independently establish an attack delay."),
	}
	delay := entry(90, "Attack delay", "milliseconds", "DE extension. The checker requires an explicit numeric defconst because builtin spelling availability is unverified; no engine damage-timing claim.")
	delay.ReferenceFamily = "DE"
	delay.DEOnly = true
	delay.NeedsDefinition = true
	entries["object-data-attack-delay"] = delay
	return entries
}
