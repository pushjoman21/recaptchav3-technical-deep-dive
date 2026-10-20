package jsast

import "testing"

func TestSourceCoordinatesAndBindingLabels(t *testing.T) {
	source, err := Parse("authored.js", []byte("const prefix='😀';\n(A.prototype.first=function(){},A).prototype.second=function(x){return x};"))
	if err != nil {
		t.Fatal(err)
	}
	functions, err := Functions(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(functions) != 2 {
		t.Fatalf("functions=%d", len(functions))
	}
	if functions[1].Name != "A[\"prototype\"][\"second\"]" {
		t.Fatal(functions[1].Name)
	}
	if functions[0].Evidence.Start.Line != 2 || functions[0].Evidence.UTF8[0]-functions[0].Evidence.UTF16[0] != 2 {
		t.Fatal(functions[0].Evidence)
	}
}

func TestRejectMalformedSource(t *testing.T) {
	for _, data := range [][]byte{[]byte("function {"), {0xff}} {
		if _, err := Parse("bad.js", data); err == nil {
			t.Fatalf("accepted malformed source %q", data)
		}
	}
}
