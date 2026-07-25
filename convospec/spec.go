// Package convospec defines the Typed Action Specification — the contract
// between AI models and Sneat. Definitions are plain data (JSON-serializable)
// so the same specification drives runtime prompts, CLI inspection, tests,
// documentation and future MCP/ChatGPT/Claude tool definitions.
//
// This package lives in its own dependency-free module so that an extension's
// contract lib can declare its conversational capability without inheriting an
// application dependency tree. It therefore imports nothing outside the
// standard library and knows nothing about DALgo, facades or transports.
package convospec

import (
	"fmt"
	"strings"
)

// ArgType enumerates supported argument types.
type ArgType string

const (
	ArgTypeString      ArgType = "string"
	ArgTypeInt         ArgType = "int"
	ArgTypeFloat       ArgType = "float"
	ArgTypeBool        ArgType = "bool"
	ArgTypeStringSlice ArgType = "[]string"
)

// ArgDef describes one argument of an action.
type ArgDef struct {
	Name        string   `json:"name"`
	Type        ArgType  `json:"type"`
	Required    bool     `json:"required,omitempty"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"` // for ArgTypeString: allowed values
}

// Example maps a user utterance to the action call it should produce.
// Examples feed LLM prompts, documentation and mock/regression tests.
type Example struct {
	UserText string         `json:"userText"`
	Args     map[string]any `json:"args,omitempty"`
}

// ActionDef is the specification of one typed action.
type ActionDef struct {
	// ID is the action identifier, e.g. "lists.add_items".
	ID string `json:"id"`
	// Extension is the owning Sneat extension, e.g. "listus". It is a plain
	// string rather than coretypes.ExtID so this module stays dependency-free.
	Extension string `json:"extension"`
	Summary   string `json:"summary"`
	// Description is shown to the LLM; write it for the model.
	Description string   `json:"description,omitempty"`
	Args        []ArgDef `json:"args,omitempty"`
	// Confirm marks the action as requiring explicit user confirmation
	// before execution (destructive or hard-to-reverse operations).
	Confirm bool `json:"confirm,omitempty"`
	// Examples of user text mapping to this action.
	Examples []Example `json:"examples,omitempty"`
	// Result describes the expected result shape in one line.
	Result string `json:"result,omitempty"`
}

// Arg returns the argument definition with the given name, if present.
func (d ActionDef) Arg(name string) (ArgDef, bool) {
	for _, a := range d.Args {
		if a.Name == name {
			return a, true
		}
	}
	return ArgDef{}, false
}

// Catalog groups the action definitions contributed by one extension,
// together with interpretation hints for the LLM.
type Catalog struct {
	// ID is the catalog identifier; by convention the extension ID.
	ID    string `json:"id"`
	Title string `json:"title"`
	// PromptHints are scope-specific interpretation rules, e.g. for listus:
	// "A bare item name like 'milk' means lists.add_items to the groceries list."
	PromptHints []string `json:"promptHints,omitempty"`
	// Triggers is the lowercase word/phrase vocabulary this extension is
	// listening for. It is a ROUTING INPUT ONLY: the runtime may use it to
	// narrow the action set offered to the model, and it never affects how an
	// action is validated or executed. An empty Triggers set is legal and
	// simply means this catalog never narrows the action set.
	Triggers []string    `json:"triggers,omitempty"`
	Actions  []ActionDef `json:"actions"`
}

// Action returns the action definition with the given ID, if present.
func (c Catalog) Action(id string) (ActionDef, bool) {
	for _, a := range c.Actions {
		if a.ID == id {
			return a, true
		}
	}
	return ActionDef{}, false
}

// MatchesTriggers reports whether any of the catalog's declared triggers occurs
// in text as a WHOLE WORD (or whole run of words, for multi-word triggers).
// text is expected to be already normalized; NormalizeText produces that form.
//
// Word-boundary rather than substring matching is deliberate. A routing
// prefilter that narrows on a single matching catalog can only ever misroute
// through a false positive, and substring matching manufactures them: a trigger
// "km" would match "kmart", sending a shopping-list message to a tracker. The
// cost is that morphological variants must be declared explicitly ("buy" does
// not match "buying"), which is the right trade — an explicitly listed trigger
// is reviewable, an accidental substring hit is not.
//
// A catalog with no triggers never matches, so it can only be reached through
// the unnarrowed action set.
func (c Catalog) MatchesTriggers(text string) bool {
	if text == "" {
		return false
	}
	padded := " " + text + " "
	for _, trigger := range c.Triggers {
		trigger = strings.TrimSpace(strings.ToLower(trigger))
		if trigger == "" {
			continue
		}
		if strings.Contains(padded, " "+trigger+" ") {
			return true
		}
	}
	return false
}

// NormalizeText lowercases text and collapses the punctuation that separates
// words in chat messages into single spaces, so a declared trigger like
// "push-ups" matches "Push-ups: 20" and "20 PUSH-UPS!". Hyphens are preserved
// because triggers themselves are hyphenated.
//
// A '.' or ',' BETWEEN DIGITS is preserved, because it is a decimal separator
// rather than punctuation: stripping it silently turned "80.5" into "80 5" and
// "2.5 km" into "2 5 km", which then failed to parse as a measurement.
func NormalizeText(text string) string {
	lowered := []rune(strings.ToLower(text))
	var b strings.Builder
	b.Grow(len(lowered))
	for i, r := range lowered {
		switch r {
		case '.', ',':
			if isDigit(prevRune(lowered, i)) && isDigit(nextRune(lowered, i)) {
				b.WriteRune(r) // decimal separator
			} else {
				b.WriteByte(' ')
			}
		case ':', ';', '!', '?', '(', ')', '[', ']', '"', '\'', '\n', '\t':
			b.WriteByte(' ')
		default:
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func prevRune(runes []rune, i int) rune {
	if i == 0 {
		return 0
	}
	return runes[i-1]
}

func nextRune(runes []rune, i int) rune {
	if i+1 >= len(runes) {
		return 0
	}
	return runes[i+1]
}

func isDigit(r rune) bool { return r >= '0' && r <= '9' }

// ValidateArgs checks the given arguments against the definition and returns
// a normalized copy: JSON numbers are converted for int args, []any of
// strings for []string args. Unknown argument names are rejected so that
// model drift is caught early instead of being silently ignored.
func (d ActionDef) ValidateArgs(args map[string]any) (map[string]any, error) {
	normalized := make(map[string]any, len(args))
	defs := make(map[string]ArgDef, len(d.Args))
	for _, a := range d.Args {
		defs[a.Name] = a
	}
	for name, value := range args {
		def, ok := defs[name]
		if !ok {
			return nil, fmt.Errorf("%w: action %s does not define argument %q", ErrInvalidArgs, d.ID, name)
		}
		v, err := normalizeArg(def, value)
		if err != nil {
			return nil, fmt.Errorf("%w: action %s argument %q: %v", ErrInvalidArgs, d.ID, name, err)
		}
		normalized[name] = v
	}
	for _, a := range d.Args {
		if a.Required {
			if v, ok := normalized[a.Name]; !ok || isEmptyValue(v) {
				return nil, fmt.Errorf("%w: action %s is missing required argument %q", ErrInvalidArgs, d.ID, a.Name)
			}
		}
	}
	return normalized, nil
}

// ErrInvalidArgs indicates an action call that does not match its definition.
var ErrInvalidArgs = fmt.Errorf("invalid action arguments")

func normalizeArg(def ArgDef, value any) (any, error) {
	switch def.Type {
	case ArgTypeString:
		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", value)
		}
		if len(def.Enum) > 0 {
			for _, e := range def.Enum {
				if s == e {
					return s, nil
				}
			}
			return nil, fmt.Errorf("value %q is not one of %v", s, def.Enum)
		}
		return s, nil
	case ArgTypeInt:
		switch n := value.(type) {
		case int:
			return n, nil
		case int64:
			return int(n), nil
		case float64: // JSON number
			if n != float64(int(n)) {
				return nil, fmt.Errorf("expected integer, got %v", n)
			}
			return int(n), nil
		default:
			return nil, fmt.Errorf("expected integer, got %T", value)
		}
	case ArgTypeFloat:
		switch n := value.(type) {
		case float64:
			return n, nil
		case float32:
			return float64(n), nil
		case int:
			return float64(n), nil
		case int64:
			return float64(n), nil
		default:
			return nil, fmt.Errorf("expected number, got %T", value)
		}
	case ArgTypeBool:
		b, ok := value.(bool)
		if !ok {
			return nil, fmt.Errorf("expected bool, got %T", value)
		}
		return b, nil
	case ArgTypeStringSlice:
		switch s := value.(type) {
		case []string:
			return s, nil
		case []any:
			result := make([]string, 0, len(s))
			for _, item := range s {
				str, ok := item.(string)
				if !ok {
					return nil, fmt.Errorf("expected string element, got %T", item)
				}
				result = append(result, str)
			}
			return result, nil
		default:
			return nil, fmt.Errorf("expected []string, got %T", value)
		}
	default:
		return nil, fmt.Errorf("unknown argument type %q", def.Type)
	}
}

func isEmptyValue(v any) bool {
	switch t := v.(type) {
	case string:
		return t == ""
	case []string:
		return len(t) == 0
	default:
		return v == nil
	}
}
