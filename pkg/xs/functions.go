package xs

// FunctionDecl is source evidence only; it does not establish that the engine
// can compile or load the containing file.
type FunctionDecl struct {
	Name       string
	ReturnType string
	Parameters int
	Line       int
}

// FunctionDeclarations reuses the XS lint lexer so comments and string
// examples cannot masquerade as callable function definitions.
func FunctionDeclarations(source string) []FunctionDecl {
	ts := lexLintSource(source)
	var out []FunctionDecl
	depth := 0
	for i := 0; i < len(ts); i++ {
		if depth == 0 && i+2 < len(ts) && !ts[i].literal && isXSReturnType(ts[i].text) && ts[i+2].text == "(" {
			end := matchingParen(ts, i+2)
			if end >= 0 && end+1 < len(ts) && ts[end+1].text == "{" {
				out = append(out, FunctionDecl{Name: ts[i+1].text, ReturnType: ts[i].text, Parameters: len(splitCallArguments(ts[i+3 : end])), Line: ts[i].line})
			}
		}
		if !ts[i].literal {
			if ts[i].text == "{" {
				depth++
			}
			if ts[i].text == "}" {
				depth--
			}
		}
	}
	return out
}
