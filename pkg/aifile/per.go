package aifile

import (
	"fmt"
	"strconv"
	"strings"
)

type perNode struct {
	text         string
	quoted, list bool
	line, column int
	children     []*perNode
}

// parsePER reads forms without treating punctuation inside strings or comments
// as syntax. Preprocessor branches are retained as an explicit coverage gap.
func parsePER(text, file string, report *LintReport) ([]*perNode, bool) {
	var roots []*perNode
	var stack []*perNode
	line, column, conditional := 1, 1, false
	advance := func(ch byte) {
		if ch == '\n' {
			line++
			column = 1
		} else {
			column++
		}
	}
	add := func(n *perNode) {
		if len(stack) == 0 {
			roots = append(roots, n)
		} else {
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, n)
		}
	}
	for i := 0; i < len(text); {
		ch := text[i]
		if ch == ';' || ch == '#' {
			if ch == '#' {
				conditional = true
			}
			for i < len(text) && text[i] != '\n' {
				advance(text[i])
				i++
			}
			continue
		}
		if strings.ContainsRune(" \t\r\n", rune(ch)) {
			advance(ch)
			i++
			continue
		}
		n := &perNode{line: line, column: column}
		switch ch {
		case '(':
			n.list = true
			add(n)
			stack = append(stack, n)
			advance(ch)
			i++
			if len(stack) > 256 {
				report.perIssue("error", "nesting_limit", file, n, "PER nesting exceeds 256 levels", "", "static_analysis")
				return roots, conditional
			}
		case ')':
			if len(stack) == 0 {
				report.perIssue("error", "unbalanced_parentheses", file, n, "extra closing parenthesis", "", "static_analysis")
			} else {
				stack = stack[:len(stack)-1]
			}
			advance(ch)
			i++
		case '"':
			n.quoted = true
			advance(ch)
			i++
			var value strings.Builder
			closed := false
			for i < len(text) {
				ch = text[i]
				advance(ch)
				i++
				if ch == '"' {
					closed = true
					break
				}
				if ch == '\\' && i < len(text) && (text[i] == '"' || text[i] == '\\') {
					ch = text[i]
					advance(ch)
					i++
				}
				value.WriteByte(ch)
			}
			n.text = value.String()
			add(n)
			if !closed {
				report.perIssue("error", "unterminated_string", file, n, "unterminated quoted string", "", "static_analysis")
			}
		default:
			start := i
			for i < len(text) && !strings.ContainsRune("()\"; \t\r\n", rune(text[i])) {
				advance(text[i])
				i++
			}
			n.text = text[start:i]
			add(n)
		}
	}
	for _, n := range stack {
		report.perIssue("error", "unbalanced_parentheses", file, n, "unclosed form", "", "static_analysis")
	}
	return roots, conditional
}

func (r *LintReport) perIssue(severity, code, file string, n *perNode, message, source, confidence string) {
	r.Issues = append(r.Issues, LintIssue{Severity: severity, Code: code, File: file, Line: n.line, Column: n.column, Message: message, Source: source, Confidence: confidence})
}

type perChecker struct {
	report      *LintReport
	file        string
	constants   map[string]*perNode
	conditional bool
	roots       []*perNode
	command     string
}

func checkPER(file, text string, report *LintReport) *perChecker {
	roots, conditional := parsePER(text, file, report)
	c := perChecker{report: report, file: file, constants: map[string]*perNode{}, conditional: conditional}
	c.roots = roots
	if conditional {
		report.perIssue("warning", "preprocessor_not_evaluated", file, &perNode{line: 1, column: 1}, "preprocessor branches are not evaluated; symbol and command coverage is partial", "", "static_analysis")
	}
	for _, n := range roots {
		if head(n) == "defconst" && len(n.children) == 3 && !n.children[1].list && !n.children[1].quoted {
			key := strings.ToLower(n.children[1].text)
			if _, exists := c.constants[key]; exists {
				report.perIssue("warning", "duplicate_constant", file, n.children[1], "multiple definitions of "+n.children[1].text+"; value is not resolved", commandSource("defconst"), "static_analysis")
				c.constants[key] = nil
			} else {
				c.constants[key] = n.children[2]
			}
		}
	}
	for _, n := range roots {
		c.check(n, "top")
	}
	return &c
}

func head(n *perNode) string {
	if !n.list || len(n.children) == 0 || n.children[0].list || n.children[0].quoted {
		return ""
	}
	return strings.ToLower(n.children[0].text)
}

func (c *perChecker) resolve(n *perNode) (*perNode, bool) {
	seen := map[string]bool{}
	for n != nil && !n.list && !n.quoted {
		if _, err := strconv.ParseInt(n.text, 10, 32); err == nil {
			return n, true
		}
		key := strings.ToLower(n.text)
		if seen[key] {
			return n, false
		}
		seen[key] = true
		next, found := c.constants[key]
		if !found || next == nil || c.conditional {
			return n, false
		}
		n = next
	}
	return n, n != nil && n.quoted
}

func (c *perChecker) resolveTyped(n *perNode, kind string) (*perNode, bool) {
	value, ok := c.resolve(n)
	if ok || value == nil || value.list || value.quoted {
		return value, ok
	}
	if _, declared := c.constants[strings.ToLower(value.text)]; declared {
		return value, false
	}
	lookup := func(constants map[string]ReferenceConstant) (*perNode, bool) {
		entry, found := constants[strings.ToLower(value.text)]
		if !found || entry.NeedsDefinition {
			return value, false
		}
		// Reference enum spellings are not proof of engine identifier availability.
		// Core resource names are documented engine constants; other bundled enums
		// remain useful for value checking but require visible numeric declarations.
		if kind != "stockpile-resource" && !isCoreResource(value.text) {
			c.report.perIssue("warning", "constant_availability_unverified", c.file, n, "recognized reference constant "+value.text+" has no visible numeric defconst or verified builtin source; load/include resolution is partial", commandSource(c.command), "unverified")
		}
		return &perNode{text: strconv.FormatInt(entry.Value, 10), line: n.line, column: n.column}, true
	}
	if kind == "constant" {
		for _, constants := range parameterConstants {
			if v, found := lookup(constants); found {
				return v, true
			}
		}
		return value, false
	}
	return lookup(parameterConstants[kind])
}

func isCoreResource(name string) bool {
	_, ok := parameterConstants["stockpile-resource"][strings.ToLower(name)]
	return ok
}

func parseCompareOperator(n *perNode) (string, bool) {
	if n == nil || n.list || n.quoted {
		return "", false
	}
	op, prefix := n.text, "c:"
	for _, p := range []string{"c:", "g:", "s:"} {
		if strings.HasPrefix(op, p) {
			prefix, op = p, strings.TrimPrefix(op, p)
			break
		}
	}
	return prefix, op == ">" || op == ">=" || op == "<" || op == "<=" || op == "==" || op == "!="
}

func (c *perChecker) check(n *perNode, context string) {
	name := head(n)
	previous := c.command
	c.command = name
	defer func() { c.command = previous }()
	if name == "" {
		c.report.perIssue("error", "expected_form", c.file, n, "expected a command form with an unquoted name", "", "static_analysis")
		return
	}
	if name == "defrule" {
		c.report.Coverage.CheckedForms++
		c.record(n, name, "(defrule <facts> => <actions>)")
		if context != "top" {
			c.issue(n, "invalid_context", "defrule must be top-level")
		}
		split := -1
		for i, child := range n.children[1:] {
			if !child.list && !child.quoted && child.text == "=>" {
				if split >= 0 {
					c.issue(child, "invalid_rule", "defrule has multiple => separators")
				}
				split = i + 1
			}
		}
		if split <= 1 || split == len(n.children)-1 {
			c.issue(n, "invalid_rule", "defrule requires facts, one => separator, and actions")
			return
		}
		for _, child := range n.children[1:split] {
			c.check(child, "fact")
		}
		for _, child := range n.children[split+1:] {
			c.check(child, "action")
		}
		return
	}
	if name == "and" || name == "or" || name == "not" || name == "nand" || name == "nor" || name == "xor" || name == "xnor" {
		c.report.Coverage.CheckedForms++
		c.record(n, name, "("+name+" <facts>)")
		if context != "fact" {
			c.issue(n, "invalid_context", name+" belongs in the facts section")
		}
		want := 2
		if name == "not" {
			want = 1
		}
		if len(n.children)-1 != want {
			c.issue(n, "wrong_arity", fmt.Sprintf("expected (%s %s)", name, strings.TrimSpace(strings.Repeat("<fact> ", want))))
		}
		for _, child := range n.children[1:] {
			c.check(child, "fact")
		}
		return
	}
	spec, known := perCatalogue[name]
	if !known {
		c.report.Coverage.UncheckedForms++
		severity, code, message, confidence := "warning", "command_not_catalogued", "command is outside the bundled signature catalogue: "+name, "unverified"
		for _, candidate := range typoCandidates(name) {
			severity, code, message, confidence = "error", "unknown_command", "unknown command "+name+"; did you mean "+perCatalogue[candidate].signature(candidate)+"?", "inferred_typo"
			break
		}
		c.report.perIssue(severity, code, c.file, n.children[0], message, catalogueSource, confidence)
		return
	}
	c.report.Coverage.CheckedForms++
	if !strings.Contains(spec.context, context) {
		c.issue(n, "invalid_context", fmt.Sprintf("%s is not valid in %s context; expected %s", name, context, spec.context))
	}
	args := n.children[1:]
	if name == "up-target-point" && len(args) == 4 {
		action, known := c.resolveTyped(args[1], "duc-action")
		if !known {
			c.record(n, name, spec.signature(name), spec)
			c.report.perIssue("warning", "unresolved_constant", c.file, args[1], "target-point action is unresolved or requires an explicit defconst; overload argument kinds remain unchecked", commandSource(name), "unverified")
			return
		}
		if !action.quoted {
			code, _ := strconv.ParseInt(action.text, 10, 32)
			if code == 10 || code == 21 {
				kind := "unit-id"
				if code == 21 {
					kind = "tech-id"
				}
				spec = commandSpec{context: "action", params: []string{"escrow-goal-id", "duc-action", "type-op", kind}, bounds: map[int][2]int64{0: {0, 16000}}, notes: "Non-combat train/research overload shape only; DE escrow behavior is reference-reported as uncertain."}
				c.report.perIssue("warning", "catalogue_uncertainty", c.file, n, spec.notes, commandSource(name), "unverified")
			}
		}
	}
	c.record(n, name, spec.signature(name), spec)
	if len(args) != len(spec.params) {
		c.issue(n, "wrong_arity", "expected "+spec.signature(name))
		return
	}
	for i, kind := range spec.params {
		arg := args[i]
		if kind == "compare-op" {
			if _, ok := parseCompareOperator(arg); !ok {
				c.issue(arg, "argument_kind", "expected comparison operator with optional c:/g:/s: prefix; "+spec.signature(name))
			}
			continue
		}
		if kind == "unit-id" && i > 0 {
			c.resolveUnit(arg, args[i-1])
		}
		if kind == "symbol" {
			if arg.list || arg.quoted || arg.text == "" {
				c.issue(arg, "argument_kind", "expected constant name; "+spec.signature(name))
			}
			continue
		}
		if kind == "type-op" {
			if arg.list || arg.quoted || (arg.text != "c:" && arg.text != "g:" && arg.text != "s:") {
				c.issue(arg, "argument_kind", "expected c:, g:, or s:; "+spec.signature(name))
			}
			continue
		}
		value, resolved := c.resolveTyped(arg, kind)
		if !resolved {
			if arg.list {
				c.issue(arg, "argument_kind", "expected "+kind+"; "+spec.signature(name))
			} else {
				c.report.perIssue("warning", "unresolved_constant", c.file, arg, "cannot establish constant value/type for "+arg.text+"; it may need a defconst or be outside the available constant catalogue", commandSource(name), "unverified")
			}
			continue
		}
		if kind == "constant" {
			continue
		}
		if (kind == "text") != value.quoted {
			c.issue(arg, "argument_kind", "expected "+kind+"; "+spec.signature(name))
			continue
		}
		if !value.quoted {
			number, _ := strconv.ParseInt(value.text, 10, 32)
			if name == "up-compare-goal" && i == 2 {
				prefix, valid := parseCompareOperator(args[1])
				if valid && prefix != "c:" {
					lo, hi := int64(1), int64(16000)
					if prefix == "s:" {
						lo, hi = 0, 511
					}
					if number < lo || number > hi {
						c.report.perIssue("warning", "reference_range", c.file, arg, fmt.Sprintf("%s comparison address outside %d..%d", prefix, lo, hi), commandSource(name), "reference_backed")
					}
					continue
				}
			}
			// An operand following g:/s: is an address, not its runtime value.
			if i > 0 && spec.params[i-1] == "type-op" && args[i-1].text != "c:" {
				continue
			}
			if bounds, ok := spec.bounds[i]; ok && !hasValue(spec.exceptions[i], number) && (number < bounds[0] || number > bounds[1]) {
				c.report.perIssue("warning", "reference_range", c.file, arg, fmt.Sprintf("%s value %d is outside documented DE range %d..%d; %s", name, number, bounds[0], bounds[1], spec.signature(name)), commandSource(name), "reference_backed")
			}
			if allowed := spec.allowed[i]; len(allowed) > 0 && !hasValue(allowed, number) {
				c.report.perIssue("warning", "reference_enum", c.file, arg, fmt.Sprintf("%s value %d is outside documented choices %v; %s", name, number, allowed, spec.signature(name)), commandSource(name), "reference_backed")
			}
			if kind == "object-data" {
				known := number == -1 && name == "up-clean-search"
				for _, entry := range parameterConstants[kind] {
					if entry.Value == number {
						known = true
					}
				}
				if !known {
					c.report.perIssue("warning", "enum_value_not_catalogued", c.file, arg, "object-data selector is outside the bundled subset; its meaning/availability remains unverified", commandSource(name), "unverified")
				}
			}
		}
	}
}

func (c *perChecker) issue(n *perNode, code, message string) {
	c.report.perIssue("error", code, c.file, n, message, commandSource(c.command), "reference_backed")
}

func (c *perChecker) record(n *perNode, name, signature string, specs ...commandSpec) {
	spec := perCatalogue[name]
	if len(specs) > 0 {
		spec = specs[0]
	}
	check := CommandCheck{File: c.file, Line: n.line, Column: n.column, Command: name, Signature: signature, Source: commandSource(name), Confidence: "reference_backed", ReferenceFamily: referenceFamily(name), Notes: spec.notes}
	if spec.mechanism != "" {
		check.Mechanism = spec.mechanism
		check.RuntimeSemantics = "unverified"
	}
	for i, kind := range spec.params {
		p := ParameterCheck{Kind: kind, Allowed: append([]int64(nil), spec.allowed[i]...), Exceptions: append([]int64(nil), spec.exceptions[i]...)}
		if bounds, found := spec.bounds[i]; found {
			p.Range = &bounds
		}
		if constants := parameterConstants[kind]; len(constants) > 0 {
			p.Constants = map[string]ReferenceConstant{}
			for name, entry := range constants {
				p.Constants[name] = entry
			}
		}
		check.Parameters = append(check.Parameters, p)
	}
	c.report.Checks = append(c.report.Checks, check)
}
