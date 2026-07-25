package convospec

import (
	"errors"
	"testing"
)

var testDef = ActionDef{
	ID: "test.action",
	Args: []ArgDef{
		{Name: "title", Type: ArgTypeString, Required: true},
		{Name: "count", Type: ArgTypeInt},
		{Name: "amount", Type: ArgTypeFloat},
		{Name: "done", Type: ArgTypeBool},
		{Name: "items", Type: ArgTypeStringSlice},
		{Name: "kind", Type: ArgTypeString, Enum: []string{"a", "b"}},
	},
}

func TestValidateArgs_normalizesTypes(t *testing.T) {
	args, err := testDef.ValidateArgs(map[string]any{
		"title":  "hello",
		"count":  float64(3), // JSON number
		"amount": 2,          // int widened to float
		"done":   true,
		"items":  []any{"x", "y"},
		"kind":   "a",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if args["count"] != 3 {
		t.Errorf("count: want int 3, got %#v", args["count"])
	}
	if args["amount"] != float64(2) {
		t.Errorf("amount: want float64 2, got %#v", args["amount"])
	}
	items, ok := args["items"].([]string)
	if !ok || len(items) != 2 || items[0] != "x" {
		t.Errorf("items: want []string{x,y}, got %#v", args["items"])
	}
}

func TestValidateArgs_missingRequired(t *testing.T) {
	if _, err := testDef.ValidateArgs(map[string]any{}); !errors.Is(err, ErrInvalidArgs) {
		t.Fatalf("want ErrInvalidArgs, got %v", err)
	}
}

func TestValidateArgs_unknownArg(t *testing.T) {
	if _, err := testDef.ValidateArgs(map[string]any{"title": "x", "nope": 1}); !errors.Is(err, ErrInvalidArgs) {
		t.Fatalf("want ErrInvalidArgs, got %v", err)
	}
}

func TestValidateArgs_enumViolation(t *testing.T) {
	if _, err := testDef.ValidateArgs(map[string]any{"title": "x", "kind": "z"}); !errors.Is(err, ErrInvalidArgs) {
		t.Fatalf("want ErrInvalidArgs, got %v", err)
	}
}

func TestValidateArgs_wrongTypes(t *testing.T) {
	for name, args := range map[string]map[string]any{
		"int as string":     {"title": "x", "count": "3"},
		"fractional number": {"title": "x", "count": 3.5},
		"float as string":   {"title": "x", "amount": "2"},
		"bool as string":    {"title": "x", "done": "yes"},
		"non-string slice":  {"title": "x", "items": []any{1}},
		"required empty":    {"title": ""},
	} {
		if _, err := testDef.ValidateArgs(args); !errors.Is(err, ErrInvalidArgs) {
			t.Errorf("%s: want ErrInvalidArgs, got %v", name, err)
		}
	}
}

func TestCatalogAction(t *testing.T) {
	c := Catalog{ID: "test", Actions: []ActionDef{testDef}}
	if _, ok := c.Action("test.action"); !ok {
		t.Error("expected to find test.action")
	}
	if _, ok := c.Action("missing"); ok {
		t.Error("did not expect to find missing action")
	}
}

func TestActionDefArg(t *testing.T) {
	if def, ok := testDef.Arg("count"); !ok || def.Type != ArgTypeInt {
		t.Errorf("Arg(count) = %+v, %v", def, ok)
	}
	if _, ok := testDef.Arg("missing"); ok {
		t.Error("did not expect to find missing arg")
	}
}

func TestNormalizeText(t *testing.T) {
	for input, want := range map[string]string{
		"Push-ups: 20":     "push-ups 20",
		"20 PUSH-UPS!":     "20 push-ups",
		"milk, 2 liters":   "milk 2 liters",
		"  bought   milk ": "bought milk",
		"What's on it?":    "whats on it",
		"who's going?":     "whos going",
		"'quoted' words":   "quoted words",
		// A decimal separator is not punctuation: stripping it turned "80.5"
		// into "80 5" and the measurement stopped parsing.
		"my weight is 80.5": "my weight is 80.5",
		"ran 2.5 km":        "ran 2.5 km",
		"80,5 kg":           "80,5 kg",
		// A period that is NOT between digits is still punctuation.
		"bought milk. bought bread": "bought milk bought bread",
		"2. milk":                   "2 milk",
	} {
		if got := NormalizeText(input); got != want {
			t.Errorf("NormalizeText(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCatalogMatchesTriggers(t *testing.T) {
	trackus := Catalog{ID: "trackus", Triggers: []string{"push-ups", "pull-ups", "km"}}
	for _, text := range []string{"push-ups 20", "20 push-ups", "10 pull-ups"} {
		if !trackus.MatchesTriggers(NormalizeText(text)) {
			t.Errorf("expected %q to match trackus triggers", text)
		}
	}
	if trackus.MatchesTriggers(NormalizeText("milk 2 liters")) {
		t.Error("did not expect milk to match trackus triggers")
	}
}

// A catalog that declares no triggers must never match, so it is only ever
// reachable through the unnarrowed action set.
func TestCatalogWithoutTriggersNeverMatches(t *testing.T) {
	quiet := Catalog{ID: "quiet"}
	for _, text := range []string{"anything at all", ""} {
		if quiet.MatchesTriggers(NormalizeText(text)) {
			t.Errorf("catalog with no triggers matched %q", text)
		}
	}
}

// An empty-string trigger must not turn into a match-everything rule, which is
// what strings.Contains would do if it were not skipped.
func TestEmptyTriggerIsIgnored(t *testing.T) {
	sloppy := Catalog{ID: "sloppy", Triggers: []string{"", "   "}}
	if sloppy.MatchesTriggers(NormalizeText("milk")) {
		t.Error("empty trigger must not match")
	}
}

// Triggers match whole words only. Substring matching would let a short trigger
// like "km" claim "kmart" and misroute the message, and a single false-positive
// catalog is the only way a fail-open prefilter can route wrongly.
func TestTriggersMatchWholeWordsOnly(t *testing.T) {
	trackus := Catalog{ID: "trackus", Triggers: []string{"km", "kg", "ran"}}
	for _, text := range []string{"ran 10 km", "weighed 80 kg", "10 KM"} {
		if !trackus.MatchesTriggers(NormalizeText(text)) {
			t.Errorf("expected %q to match", text)
		}
	}
	for _, text := range []string{
		"buy socks at kmart",  // km inside kmart
		"branding guidelines", // ran inside branding
		"kgb documentary",     // kg inside kgb
	} {
		if trackus.MatchesTriggers(NormalizeText(text)) {
			t.Errorf("substring match leaked: %q must not match", text)
		}
	}
}

// A multi-word trigger must match as a contiguous run of words.
func TestMultiWordTrigger(t *testing.T) {
	listus := Catalog{ID: "listus", Triggers: []string{"shopping list"}}
	if !listus.MatchesTriggers(NormalizeText("add milk to my shopping list, please")) {
		t.Error("expected multi-word trigger to match")
	}
	if listus.MatchesTriggers(NormalizeText("shopping for a list of books")) {
		t.Error("multi-word trigger must match contiguous words only")
	}
}
