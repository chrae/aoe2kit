package main

import (
	"fmt"
	"os"
	"strings"

	xsauthor "aoe2kit/pkg/xs"
)

func runXS(args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: kit xs <bridge|shims|datagen|inspect> ...")
		os.Exit(2)
	}
	switch args[0] {
	case "bridge":
		runXSBridge(args[1:])
	case "shims":
		runXSShims(args[1:])
	case "datagen":
		runXSDatagen(args[1:])
	case "inspect", "xsdat":
		runXSInspect(args[1:])
	default:
		fmt.Fprintln(os.Stderr, "usage: kit xs <bridge|shims|datagen|inspect> ...")
		os.Exit(2)
	}
}

func runXSBridge(args []string) {
	input := ""
	outDir := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--out-dir":
			i++
			if i >= len(args) {
				die("kit xs bridge", fmt.Errorf("--out-dir needs a directory"))
			}
			outDir = args[i]
		default:
			if input != "" {
				die("kit xs bridge", fmt.Errorf("unexpected argument %q", args[i]))
			}
			input = args[i]
		}
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "usage: kit xs bridge <variables.json> [--out-dir DIR]")
		os.Exit(2)
	}
	spec, err := xsauthor.LoadBridgeSpec(input)
	if err != nil {
		die("kit xs bridge", err)
	}
	report, err := xsauthor.GenerateBridge(spec)
	if err != nil {
		die("kit xs bridge", err)
	}
	if err := xsauthor.WriteBridgeOutputs(report, outDir); err != nil {
		die("kit xs bridge", err)
	}
	printJSON(report)
	if !report.OK {
		os.Exit(1)
	}
}

func runXSShims(args []string) {
	out := ""
	prefix := "XS Shim"
	enabled := false
	includeParams := false
	paths := []string{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--out":
			i++
			if i >= len(args) {
				die("kit xs shims", fmt.Errorf("--out needs a recipe path"))
			}
			out = args[i]
		case "--prefix":
			i++
			if i >= len(args) {
				die("kit xs shims", fmt.Errorf("--prefix needs text"))
			}
			prefix = args[i]
		case "--enabled":
			enabled = true
		case "--include-params":
			includeParams = true
		default:
			paths = append(paths, args[i])
		}
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "usage: kit xs shims <file-or-dir>... [--out recipe.json] [--prefix TEXT] [--enabled] [--include-params]")
		os.Exit(2)
	}
	report, err := xsauthor.GenerateShims(xsauthor.ShimOptions{
		Paths:             paths,
		Prefix:            prefix,
		Enabled:           enabled,
		IncludeParamFuncs: includeParams,
	})
	if err != nil {
		die("kit xs shims", err)
	}
	if err := xsauthor.WriteShimRecipe(report, out); err != nil {
		die("kit xs shims", err)
	}
	printJSON(report)
}

func runXSInspect(args []string) {
	input := ""
	typeSpec := ""
	text := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--types":
			i++
			if i >= len(args) {
				die("kit xs inspect", fmt.Errorf("--types needs a comma-separated type list"))
			}
			typeSpec = args[i]
		case "--text":
			text = true
		default:
			if input != "" {
				die("kit xs inspect", fmt.Errorf("unexpected argument %q", args[i]))
			}
			input = args[i]
		}
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "usage: kit xs inspect <file.xsdat> [--types string,int,...] [--text]")
		os.Exit(2)
	}
	types, err := xsauthor.ParseDataTypes(typeSpec)
	if err != nil {
		die("kit xs inspect", err)
	}
	report, err := xsauthor.InspectDataFile(input, xsauthor.DataInspectOptions{Types: types})
	if err != nil {
		die("kit xs inspect", err)
	}
	if text {
		printXSInspectText(report)
	} else {
		printJSON(report)
	}
	if !report.OK {
		os.Exit(1)
	}
}

func printXSInspectText(report xsauthor.DataInspectReport) {
	fmt.Printf("%s: %d bytes mode=%s ok=%t\n", report.Path, report.SizeBytes, report.Mode, report.OK)
	if len(report.HeuristicWarnings) > 0 {
		fmt.Println("warnings:")
		for _, warning := range report.HeuristicWarnings {
			fmt.Printf("  - %s\n", warning)
		}
	}
	for _, value := range report.Values {
		fmt.Printf("[%02d] @%04d %-12s size=%d", value.Index, value.Offset, value.Type, value.Size)
		switch value.Type {
		case "string":
			fmt.Printf(" %q", value.String)
		case "int", "int_or_float":
			if value.Int != nil {
				fmt.Printf(" int=%d", *value.Int)
			}
			if value.UInt != nil {
				fmt.Printf(" uint=%d", *value.UInt)
			}
			if value.FloatCandidate != nil && value.Type == "int_or_float" {
				fmt.Printf(" float_candidate=%g", *value.FloatCandidate)
			}
		case "uint":
			if value.UInt != nil {
				fmt.Printf(" uint=%d", *value.UInt)
			}
		case "float":
			if value.Float != nil {
				fmt.Printf(" float=%g", *value.Float)
			}
		case "vector":
			parts := make([]string, 0, len(value.Vector))
			for _, component := range value.Vector {
				parts = append(parts, fmt.Sprintf("%g", component))
			}
			fmt.Printf(" vector=(%s)", strings.Join(parts, ", "))
		case "raw":
			fmt.Printf(" raw_hex=%s", value.RawHex)
		}
		if value.Note != "" {
			fmt.Printf(" note=%q", value.Note)
		}
		fmt.Println()
	}
	if len(report.Errors) > 0 {
		fmt.Println("errors:")
		for _, err := range report.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}
	if report.RemainingBytes > 0 {
		fmt.Printf("remaining @%04d bytes=%d hex=%s\n", report.RemainingOffset, report.RemainingBytes, report.RemainingHex)
	}
}

func runXSDatagen(args []string) {
	input := ""
	out := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--out":
			i++
			if i >= len(args) {
				die("kit xs datagen", fmt.Errorf("--out needs a module path"))
			}
			out = args[i]
		default:
			if input != "" {
				die("kit xs datagen", fmt.Errorf("unexpected argument %q", args[i]))
			}
			input = args[i]
		}
	}
	if input == "" {
		fmt.Fprintln(os.Stderr, "usage: kit xs datagen <arrays.json> [--out module.xs]")
		os.Exit(2)
	}
	spec, err := xsauthor.LoadDatagenSpec(input)
	if err != nil {
		die("kit xs datagen", err)
	}
	report, err := xsauthor.GenerateDatagen(spec)
	if err != nil {
		die("kit xs datagen", err)
	}
	if err := xsauthor.WriteDatagenModule(report, out); err != nil {
		die("kit xs datagen", err)
	}
	printJSON(report)
}
