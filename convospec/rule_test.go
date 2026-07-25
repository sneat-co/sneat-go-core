package convospec

import (
	"regexp"
	"strings"
	"testing"
)

var addEntryDef = ActionDef{
	ID: "trackers.add_entry", Extension: "trackus",
	Args: []ArgDef{
		{Name: "trackerID", Type: ArgTypeString, Required: true},
		{Name: "value", Type: ArgTypeFloat, Required: true},
	},
}

var addItemsDef = ActionDef{
	ID: "lists.add_items", Extension: "listus",
	Args: []ArgDef{
		{Name: "items", Type: ArgTypeStringSlice, Required: true},
		{Name: "quantity", Type: ArgTypeFloat},
		{Name: "unit", Type: ArgTypeString, Enum: []string{"L", "kg", "pcs"}},
		{Name: "listID", Type: ArgTypeString},
	},
}

// Both phrasings of a measurement resolve to the same tracker and value.
func TestRuleMatch_bothWordOrders(t *testing.T) {
	rules := []Rule{
		{
			Pattern:  regexp.MustCompile(`^push-ups (\d+)$`),
			ActionID: "trackers.add_entry",
			Args:     map[string]any{"trackerID": "_push_ups", "value": "$1"},
		},
		{
			Pattern:  regexp.MustCompile(`^(\d+) push-ups$`),
			ActionID: "trackers.add_entry",
			Args:     map[string]any{"trackerID": "_push_ups", "value": "$1"},
		},
	}
	for i, text := range []string{"push-ups 20", "20 push-ups"} {
		args, matched := rules[i].Match(NormalizeText(text), addEntryDef)
		if !matched {
			t.Fatalf("%q did not match", text)
		}
		if args["trackerID"] != "_push_ups" {
			t.Errorf("%q trackerID = %#v", text, args["trackerID"])
		}
		if args["value"] != float64(20) {
			t.Errorf("%q value = %#v, want float64 20", text, args["value"])
		}
	}
}

// A constant in Args is not a template and must pass through untouched.
func TestRuleMatch_constantsAndCaptures(t *testing.T) {
	rule := Rule{
		Pattern:  regexp.MustCompile(`^(\w+) (\d+) liters$`),
		ActionID: "lists.add_items",
		Args: map[string]any{
			"items":    "$1",
			"quantity": "$2",
			"unit":     "L",
			"listID":   "buy!groceries",
		},
	}
	args, matched := rule.Match(NormalizeText("milk 2 liters"), addItemsDef)
	if !matched {
		t.Fatal("expected a match")
	}
	items, ok := args["items"].([]string)
	if !ok || len(items) != 1 || items[0] != "milk" {
		t.Errorf("items = %#v, want []string{milk}", args["items"])
	}
	if args["quantity"] != float64(2) {
		t.Errorf("quantity = %#v, want float64 2", args["quantity"])
	}
	if args["unit"] != "L" || args["listID"] != "buy!groceries" {
		t.Errorf("constants lost: unit=%#v listID=%#v", args["unit"], args["listID"])
	}
}

// A capture that cannot be typed must make the rule not match, so a
// lower-priority rule or clarify can take over — never a malformed call.
func TestRuleMatch_untypeableCaptureDoesNotMatch(t *testing.T) {
	rule := Rule{
		Pattern:  regexp.MustCompile(`^push-ups (\w+)$`),
		ActionID: "trackers.add_entry",
		Args:     map[string]any{"trackerID": "_push_ups", "value": "$1"},
	}
	if args, matched := rule.Match(NormalizeText("push-ups many"), addEntryDef); matched {
		t.Errorf("expected no match for an unparseable value, got %#v", args)
	}
}

// A rule may not smuggle a value past the declared enum.
func TestRuleMatch_enumViolationDoesNotMatch(t *testing.T) {
	rule := Rule{
		Pattern:  regexp.MustCompile(`^(\w+) (\d+) (\w+)$`),
		ActionID: "lists.add_items",
		Args:     map[string]any{"items": "$1", "quantity": "$2", "unit": "$3"},
	}
	if _, matched := rule.Match(NormalizeText("milk 2 liters"), addItemsDef); matched {
		t.Error(`"liters" is outside the declared unit vocabulary and must not match`)
	}
	if _, matched := rule.Match(NormalizeText("flour 2 kg"), addItemsDef); !matched {
		t.Error(`"kg" is in the declared vocabulary and should match`)
	}
}

// A rule referencing an argument the action does not declare is a rule/spec
// disagreement, not a match.
func TestRuleMatch_undeclaredArgDoesNotMatch(t *testing.T) {
	rule := Rule{
		Pattern:  regexp.MustCompile(`^push-ups (\d+)$`),
		ActionID: "trackers.add_entry",
		Args:     map[string]any{"trackerID": "_push_ups", "value": "$1", "nonsense": "x"},
	}
	if _, matched := rule.Match(NormalizeText("push-ups 20"), addEntryDef); matched {
		t.Error("a rule setting an undeclared argument must not match")
	}
}

func TestSubstituteGroups_outOfRangeReference(t *testing.T) {
	rule := Rule{
		Pattern:  regexp.MustCompile(`^push-ups (\d+)$`),
		ActionID: "trackers.add_entry",
		Args:     map[string]any{"trackerID": "$7", "value": "$1"},
	}
	args, matched := rule.Match(NormalizeText("push-ups 20"), addEntryDef)
	if !matched {
		t.Fatal("expected a match: an out-of-range reference resolves to empty for a string arg")
	}
	if args["trackerID"] != "" {
		t.Errorf("trackerID = %#v, want empty", args["trackerID"])
	}
	// The empty value then fails the action's own required-arg validation.
	if _, err := addEntryDef.ValidateArgs(args); err == nil {
		t.Error("expected ValidateArgs to reject the empty required trackerID")
	}
}

// ValidateRules catches a rename applied to an action but not to its rule.
func TestValidateRules(t *testing.T) {
	catalog := Catalog{ID: "trackus", Actions: []ActionDef{addEntryDef}}

	if err := catalog.ValidateRules([]Rule{{
		Pattern: regexp.MustCompile(`^x$`), ActionID: "trackers.add_entry",
		Args: map[string]any{"trackerID": "_push_ups"},
	}}); err != nil {
		t.Fatalf("valid rule rejected: %v", err)
	}

	for name, rule := range map[string]Rule{
		"unknown action": {Pattern: regexp.MustCompile(`^x$`), ActionID: "trackers.renamed"},
		"no pattern":     {ActionID: "trackers.add_entry"},
		"no action":      {Pattern: regexp.MustCompile(`^x$`)},
		"undeclared arg": {Pattern: regexp.MustCompile(`^x$`), ActionID: "trackers.add_entry", Args: map[string]any{"nope": 1}},
	} {
		err := catalog.ValidateRules([]Rule{rule})
		if err == nil {
			t.Errorf("%s: expected an error", name)
			continue
		}
		if !strings.Contains(err.Error(), "trackus") {
			t.Errorf("%s: error should name the catalog, got %v", name, err)
		}
	}
}
