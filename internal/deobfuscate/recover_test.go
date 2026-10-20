package deobfuscate

import (
	"math"
	"strings"
	"testing"

	"recaptchav3research/internal/jsast"
)

const testFactory = `function(start,count,key,cache){return function(i,out,ch){
if(cache)return cache;
for(out="",i=0;i<count;i++){
ch=("abcdefghijklmnopqrstuvwxyz0123456789".charCodeAt(start+i)^key);
out+=String.fromCharCode(ch);
}
return cache=out;
};}`

func recoverTest(t *testing.T, text string, options Options) (*jsast.Source, *Report) {
	t.Helper()
	source, err := jsast.Parse("fixture.js", []byte(text))
	if err != nil {
		t.Fatal(err)
	}
	report, err := Recover(source, options)
	if err != nil {
		t.Fatal(err)
	}
	return source, report
}

func TestStructuralRecoveryAcrossSyntheticNamesAndLayouts(t *testing.T) {
	for _, name := range []string{"table", "renamed_92"} {
		t.Run(name, func(t *testing.T) {
			text := `var ` + name + `=[function(){return 0;},` + testFactory + `];var decoded=` + name + `[1+0](1,3,0);var answer=decoded();`
			source, report := recoverTest(t, text, Options{AssumeStandardIntrinsics: true, AssumeInitializedDecoders: true})
			if report.Recovered != 1 || len(report.Edits) != 1 || *report.Strings[0].Value != "bcd" {
				t.Fatalf("%+v", report)
			}
			output, err := Rewrite(source, report)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output, `var answer=("bcd")`) {
				t.Fatal(output)
			}
			if report.Strings[0].Invariant.Kind != "truthy-memoized-string" {
				t.Fatal(report.Strings[0])
			}
		})
	}
}

func TestRecoveryAssumptionsAreExplicit(t *testing.T) {
	text := `var T=[` + testFactory + `];var decoded=T[0](1,3,0);decoded();`
	_, report := recoverTest(t, text, Options{})
	if report.Recovered != 0 || len(report.Edits) != 0 || report.Strings[0].Reason != "STANDARD_INTRINSICS_NOT_ASSUMED" {
		t.Fatal(report)
	}
	_, report = recoverTest(t, text, Options{AssumeStandardIntrinsics: true})
	if report.Recovered != 1 || len(report.Edits) != 0 {
		t.Fatal(report)
	}
}

func TestMutationAndEscapesPreventRecovery(t *testing.T) {
	for _, mutation := range []string{`T[0]=function(){};`, `consume(T);`, `T.push(1);`, `delete T[0];`, `T[0].x=1;`, `decoded=function(){};`} {
		t.Run(mutation, func(t *testing.T) {
			_, report := recoverTest(t, `var T=[`+testFactory+`];var decoded=T[0](1,3,0);`+mutation+`decoded();`, Options{AssumeStandardIntrinsics: true, AssumeInitializedDecoders: true})
			if report.Recovered != 0 || len(report.Edits) != 0 {
				t.Fatalf("mutation accepted: %+v", report)
			}
		})
	}
}

func TestShadowingForwardDeclarationsAndWithScope(t *testing.T) {
	text := `var T=[` + testFactory + `];var decoded=T[0](1,3,0);function other(){decoded();var decoded=function(){return "private";};}decoded();with({decoded:custom}){decoded();}`
	_, report := recoverTest(t, text, Options{AssumeStandardIntrinsics: true, AssumeInitializedDecoders: true})
	if report.Recovered != 1 || len(report.Edits) != 1 {
		t.Fatalf("%+v", report)
	}
	_, shadowed := recoverTest(t, `var String=custom;var T=[`+testFactory+`];var decoded=T[0](1,3,0);decoded();`, Options{AssumeStandardIntrinsics: true, AssumeInitializedDecoders: true})
	if shadowed.Recovered != 0 || len(shadowed.Edits) != 0 || shadowed.Strings[0].Reason != "UNRESOLVED_STATIC_IDENTIFIER" {
		t.Fatal(shadowed)
	}
}

func TestBudgetUnknownGrammarAndStatefulClosure(t *testing.T) {
	for _, tc := range []struct {
		name, body, want string
		fuel             int
	}{
		{"fuel", testFactory, "STATIC_BUDGET_EXCEEDED", 20},
		{"property-write", strings.ReplaceAll(testFactory, `return cache=out;`, `cache=out;cache+="x";return cache;`), "STATEFUL_DECODER_UNSUPPORTED", 200000},
		{"unknown-call", strings.ReplaceAll(testFactory, `String.fromCharCode(ch)`, `danger(ch)`), "UNRESOLVED_STATIC_IDENTIFIER", 200000},
		{"nonconstant", testFactory, "NON_CONSTANT_FACTORY_ARGUMENT", 200000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			arg := "3"
			if tc.name == "nonconstant" {
				arg = "runtimeCount"
			}
			_, report := recoverTest(t, `var T=[`+tc.body+`];var decoded=T[0](1,`+arg+`,0);decoded();`, Options{AssumeStandardIntrinsics: true, AssumeInitializedDecoders: true, Fuel: tc.fuel})
			if report.Recovered != 0 || report.Strings[0].Reason != tc.want {
				t.Fatalf("%+v", report.Strings)
			}
		})
	}
}

func TestJavaScriptCoercionAndUTF16(t *testing.T) {
	checks := []struct {
		expr string
		want any
	}{
		{`"5"+2`, stringValue("52")}, {`"5"-2`, float64(3)}, {`(-1)>>>0`, float64(4294967295)},
		{`2147483648|0`, float64(-2147483648)}, {`null==void 0`, true}, {`null===void 0`, false},
		{`"10"<"2"`, true}, {`(0/0)<=1`, false}, {`2**3`, float64(8)},
	}
	for _, tc := range checks {
		t.Run(tc.expr, func(t *testing.T) {
			s, err := jsast.Parse("expr.js", []byte("var value=("+tc.expr+");"))
			if err != nil {
				t.Fatal(err)
			}
			v := makeIndex(s)
			got, ok := constant(v.declarations[0].Initializer.Expr)
			if !ok || !strictEqual(got, tc.want) {
				t.Fatalf("got %v (%T)", got, got)
			}
		})
	}
	r := newReducer(true, 1000)
	if got := r.intrinsicCall(intrinsic{"codePointAt", stringValue("😀")}, []any{float64(0)}, nil); got != float64(0x1f600) {
		t.Fatal(got)
	}
	if got := r.intrinsicCall(intrinsic{"charCodeAt", stringValue("😀")}, []any{float64(1)}, nil); got != float64(0xde00) {
		t.Fatal(got)
	}
	if got := r.intrinsicCall(intrinsic{name: "fromCharCode"}, []any{float64(-1)}, nil).(jsString); len(got) != 1 || got[0] != 65535 {
		t.Fatal(got)
	}
	if numberText(math.Copysign(0, -1)) != "0" || numberText(1e-7) != "1e-7" || numberText(1e-6) != "0.000001" {
		t.Fatal("Number string coercion")
	}
}
