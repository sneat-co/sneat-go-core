package convospec

import (
	"regexp"
	"strconv"
	"strings"
)

// Rule is one deterministic interpretation: a pattern over the user's message
// and the action call it produces. Rules let an extension ship a reproducible
// stand-in for a language model alongside its action declarations, so tests do
// not depend on a network call or on a model's mood.
//
// The TYPE lives here, in the dependency-free specification module, because a
// contract lib may depend only on convospec — a rule type owned by the mock
// engine would make an extension's rules unexpressible in the one place they
// belong. This package owns the data; the engine that matches and ranks rules
// (convollm/llmmock) lives with the runtime and consumes it.
type Rule struct {
	// Pattern matches against the normalized message (see NormalizeText).
	// Capture groups feed Args via the $n substitution described below.
	Pattern *regexp.Regexp

	// ActionID is the action this rule resolves to. It MUST be declared in the
	// same catalog; binding validation rejects a rule naming an unknown action.
	ActionID string

	// Args maps argument name to a value template. A string value may reference
	// a capture group as "$1", "$2", … and is substituted before typing. Any
	// other value is used as-is, which is how a rule supplies a constant (a
	// default list ID, a standard tracker ID, a canonical unit).
	//
	// Values are typed according to the action's ArgDef: a "$1" bound to an int
	// argument becomes an int, to a float argument a float64, and to a
	// []string argument a single-element slice. A capture that cannot be typed
	// makes the rule not match, rather than producing a malformed call.
	Args map[string]any

	// Priority orders rules when several match the same message; higher wins.
	// Equal priorities are resolved by declaration order. Use it when a
	// specific phrasing must beat a general one ("bought milk" as a set-done
	// rather than as an add).
	Priority int
}

// Match applies the rule to already-normalized text. It reports whether the
// rule matched and, if so, the resolved arguments typed against def.
//
// It returns matched=false — rather than an error — when a capture cannot be
// typed against the action's declaration. A rule that produces a malformed call
// is a rule that does not apply; failing soft here lets a lower-priority rule
// or the clarify path take over instead of failing the turn.
func (r Rule) Match(text string, def ActionDef) (args map[string]any, matched bool) {
	if r.Pattern == nil {
		return nil, false
	}
	groups := r.Pattern.FindStringSubmatch(text)
	if groups == nil {
		return nil, false
	}
	args = make(map[string]any, len(r.Args))
	for name, template := range r.Args {
		argDef, declared := def.Arg(name)
		if !declared {
			return nil, false // rule and declaration disagree
		}
		value, ok := resolveTemplate(template, groups, argDef)
		if !ok {
			return nil, false
		}
		args[name] = value
	}
	return args, true
}

// resolveTemplate substitutes capture groups into a template value and types the
// result against the argument declaration.
func resolveTemplate(template any, groups []string, def ArgDef) (any, bool) {
	text, isString := template.(string)
	if !isString {
		return template, true // constants pass through untouched
	}
	substituted := substituteGroups(text, groups)
	switch def.Type {
	case ArgTypeString:
		if len(def.Enum) > 0 {
			for _, allowed := range def.Enum {
				if substituted == allowed {
					return substituted, true
				}
			}
			return nil, false
		}
		return substituted, true
	case ArgTypeStringSlice:
		return []string{substituted}, true
	case ArgTypeInt:
		n, err := strconv.Atoi(strings.TrimSpace(substituted))
		if err != nil {
			return nil, false
		}
		return n, true
	case ArgTypeFloat:
		f, err := strconv.ParseFloat(strings.TrimSpace(strings.Replace(substituted, ",", ".", 1)), 64)
		if err != nil {
			return nil, false
		}
		return f, true
	case ArgTypeBool:
		b, err := strconv.ParseBool(strings.TrimSpace(substituted))
		if err != nil {
			return nil, false
		}
		return b, true
	default:
		return nil, false
	}
}

// substituteGroups replaces $1..$9 with the corresponding capture group.
// An out-of-range reference resolves to the empty string, which then fails
// typing for every non-string argument rather than silently passing zero.
func substituteGroups(template string, groups []string) string {
	if !strings.Contains(template, "$") {
		return template
	}
	var b strings.Builder
	for i := 0; i < len(template); i++ {
		if template[i] != '$' || i+1 >= len(template) {
			b.WriteByte(template[i])
			continue
		}
		digit := template[i+1]
		if digit < '1' || digit > '9' {
			b.WriteByte(template[i])
			continue
		}
		index := int(digit - '0')
		if index < len(groups) {
			b.WriteString(groups[index])
		}
		i++ // consume the digit
	}
	return b.String()
}

// ValidateRules reports an error for every rule in the catalog that names an
// action the catalog does not declare, or that declares no pattern. Catalog
// binding calls this so a rename applied to an action but not to its rule is
// caught at construction rather than showing up as a silently unmatched
// message.
func (c Catalog) ValidateRules(rules []Rule) error {
	for i, rule := range rules {
		if rule.Pattern == nil {
			return &RuleError{Catalog: c.ID, Index: i, Reason: "rule has no pattern"}
		}
		if rule.ActionID == "" {
			return &RuleError{Catalog: c.ID, Index: i, Reason: "rule has no action ID"}
		}
		def, ok := c.Action(rule.ActionID)
		if !ok {
			return &RuleError{Catalog: c.ID, Index: i, ActionID: rule.ActionID, Reason: "rule names an action the catalog does not declare"}
		}
		for name := range rule.Args {
			if _, declared := def.Arg(name); !declared {
				return &RuleError{Catalog: c.ID, Index: i, ActionID: rule.ActionID, Reason: "rule sets argument " + strconv.Quote(name) + " that the action does not declare"}
			}
		}
	}
	return nil
}

// RuleError describes a rule that disagrees with the catalog it belongs to.
type RuleError struct {
	Catalog  string
	Index    int
	ActionID string
	Reason   string
}

func (e *RuleError) Error() string {
	var b strings.Builder
	b.WriteString("catalog ")
	b.WriteString(e.Catalog)
	b.WriteString(": rule ")
	b.WriteString(strconv.Itoa(e.Index))
	if e.ActionID != "" {
		b.WriteString(" (")
		b.WriteString(e.ActionID)
		b.WriteString(")")
	}
	b.WriteString(": ")
	b.WriteString(e.Reason)
	return b.String()
}
