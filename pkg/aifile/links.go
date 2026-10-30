package aifile

import (
	"os"
	"path/filepath"
	"strings"

	"aoe2kit/pkg/xs"
)

type LintLink struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Target     string `json:"target,omitempty"`
	Status     string `json:"status"`
	Confidence string `json:"confidence"`
}

func (c *perChecker) checkLinks(path string) {
	var includes []*perNode
	var calls []*perNode
	var walk func(*perNode)
	walk = func(n *perNode) {
		switch head(n) {
		case "include":
			includes = append(includes, n)
		case "xs-script-call":
			calls = append(calls, n)
		case "load":
			c.report.perIssue("warning", "load_not_expanded", c.file, n, "PER load roots and loaded constants are not resolved by this catalogue slice", commandSource("load"), "unverified")
		}
		for _, child := range n.children {
			if child.list {
				walk(child)
			}
		}
	}
	for _, root := range c.roots {
		walk(root)
	}
	functions := map[string][]xs.FunctionDecl{}
	targets := map[string]string{}
	for _, n := range includes {
		if len(n.children) != 2 {
			continue
		}
		value, ok := c.resolve(n.children[1])
		if !ok || !value.quoted {
			continue
		}
		link := LintLink{File: c.file, Line: n.line, Kind: "xs_include", Name: value.text, Status: "unresolved", Confidence: "unverified"}
		if !strings.EqualFold(filepath.Ext(value.text), ".xs") {
			c.report.perIssue("error", "xs_include_extension", c.file, n, "XS include names require the .xs extension", commandSource("include"), "reference_backed")
		}
		// The reference resolves XS includes from an XS root, not the PER
		// directory. A sibling packaged xs directory is a project candidate;
		// its existence never proves mod/engine runtime resolution.
		xsRoot := ""
		for dir := filepath.Dir(path); ; dir = filepath.Dir(dir) {
			if strings.EqualFold(filepath.Base(dir), "ai") {
				xsRoot = filepath.Join(filepath.Dir(dir), "xs")
				break
			}
			if parent := filepath.Dir(dir); parent == dir || dir == "." {
				break
			}
		}
		if xsRoot != "" {
			candidate := filepath.Join(xsRoot, filepath.FromSlash(strings.ReplaceAll(value.text, `\`, "/")))
			data, err := os.ReadFile(candidate)
			if err == nil {
				link.Target = candidate
				link.Status = "project_file_found_runtime_unverified"
				link.Confidence = "static_analysis"
				for _, fn := range xs.FunctionDeclarations(string(data)) {
					functions[fn.Name] = append(functions[fn.Name], fn)
					targets[fn.Name] = candidate
				}
			}
		}
		c.report.Links = append(c.report.Links, link)
		message := "XS include root/runtime resolution is unverified; no packaged sibling xs file found"
		if link.Target != "" {
			message = "packaged XS file found; engine/mod include-root resolution and XS compilation remain unverified"
		}
		c.report.perIssue("warning", "xs_runtime_resolution_unverified", c.file, n, message, commandSource("include"), "unverified")
	}
	for _, n := range calls {
		if len(n.children) != 2 {
			continue
		}
		value, ok := c.resolve(n.children[1])
		if !ok || !value.quoted {
			continue
		}
		link := LintLink{File: c.file, Line: n.line, Kind: "xs_call", Name: value.text, Status: "unresolved", Confidence: "unverified"}
		decls := functions[value.text]
		if len(decls) == 1 {
			link.Target = targets[value.text]
			link.Status = "declaration_found_runtime_unverified"
			link.Confidence = "static_analysis"
			if decls[0].Parameters != 0 {
				link.Status = "requires_arguments"
				c.report.perIssue("error", "xs_function_requires_arguments", c.file, n, "xs-script-call requires an XS function with no parameters", commandSource("xs-script-call"), "reference_backed")
			}
		} else {
			c.report.perIssue("warning", "xs_function_unresolved", c.file, n, "cannot establish a unique zero-argument XS function named "+value.text+"; transitive XS includes are not expanded", commandSource("xs-script-call"), "unverified")
		}
		c.report.Links = append(c.report.Links, link)
	}
}
