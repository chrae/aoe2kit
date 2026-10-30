package xs

import "testing"

func TestFunctionDeclarationsIgnoreCommentsAndLiterals(t *testing.T) {
	decls := FunctionDeclarations(`
// void fake() {}
/* int alsoFake(int p) {} */
string example = "void inString() {}";
void tick() { xsChatData("not a function"); }
int withArgument(int value = 0) { return(value); }
`)
	if len(decls) != 2 || decls[0].Name != "tick" || decls[0].Parameters != 0 || decls[1].Name != "withArgument" || decls[1].Parameters != 1 {
		t.Fatalf("declarations=%+v", decls)
	}
}
