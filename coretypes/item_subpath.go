// Copyright 2026 Sneat Co.
package coretypes

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	ItemSubPathMaxBytes    = 2048
	ItemSubPathMaxSegments = 16
)

// ItemSubPathSegment is either an object Field or an array selection by a
// stable string Key and Value. Array order is never part of an item's identity.
type ItemSubPathSegment struct {
	Field string
	Key   string
	Value string
}

// FormatItemSubPath creates the bounded Sneat Linkage subpath syntax. It is
// inspired by JSON Pointer, with an explicit @key=value array selector:
// /items/@id=abc and /latestReceipt/obligations/@lineID=xyz. The selector value
// uses ~0 for tilde and ~1 for slash. This is not arbitrary JSONPath: numeric
// indices, wildcards, recursive descent and executable expressions are absent.
func FormatItemSubPath(segments ...ItemSubPathSegment) (string, error) {
	if len(segments) > ItemSubPathMaxSegments {
		return "", fmt.Errorf("subpath has too many segments")
	}
	parts := make([]string, len(segments))
	for i, segment := range segments {
		if segment.Field != "" {
			if segment.Key != "" || segment.Value != "" || !subPathField(segment.Field) {
				return "", fmt.Errorf("invalid subpath field at segment %d", i)
			}
			parts[i] = segment.Field
		} else {
			if i == 0 || !subPathField(segment.Key) || !subPathValue(segment.Value) {
				return "", fmt.Errorf("invalid stable-key selector at segment %d", i)
			}
			parts[i] = "@" + segment.Key + "=" + strings.NewReplacer("~", "~0", "/", "~1").Replace(segment.Value)
		}
	}
	if len(parts) == 0 {
		return "", nil
	}
	path := "/" + strings.Join(parts, "/")
	if len(path) > ItemSubPathMaxBytes {
		return "", fmt.Errorf("subpath exceeds %d bytes", ItemSubPathMaxBytes)
	}
	return path, nil
}

// ParseItemSubPath validates canonical syntax only. Resolving a selector must
// find exactly one matching string key; zero or multiple matches are errors.
// The owner must additionally allow the particular path and authorize access
// to the containing document and embedded item. Never fall back to the parent
// when a subpath cannot be resolved.
func ParseItemSubPath(path string) ([]ItemSubPathSegment, error) {
	if path == "" {
		return nil, nil
	}
	if len(path) > ItemSubPathMaxBytes || !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("invalid subpath size or prefix")
	}
	parts := strings.Split(path[1:], "/")
	if len(parts) > ItemSubPathMaxSegments {
		return nil, fmt.Errorf("subpath has too many segments")
	}
	segments := make([]ItemSubPathSegment, len(parts))
	for i, part := range parts {
		if strings.HasPrefix(part, "@") {
			key, encoded, ok := strings.Cut(part[1:], "=")
			if !ok {
				return nil, fmt.Errorf("invalid selector at segment %d", i)
			}
			var value strings.Builder
			for j := 0; j < len(encoded); j++ {
				if encoded[j] != '~' {
					value.WriteByte(encoded[j])
					continue
				}
				j++
				if j >= len(encoded) {
					return nil, fmt.Errorf("incomplete subpath escape")
				}
				switch encoded[j] {
				case '0':
					value.WriteByte('~')
				case '1':
					value.WriteByte('/')
				default:
					return nil, fmt.Errorf("unsupported subpath escape")
				}
			}
			segments[i] = ItemSubPathSegment{Key: key, Value: value.String()}
		} else {
			segments[i] = ItemSubPathSegment{Field: part}
		}
	}
	canonical, err := FormatItemSubPath(segments...)
	if err != nil {
		return nil, err
	}
	if canonical != path {
		return nil, fmt.Errorf("subpath must use canonical encoding")
	}
	return segments, nil
}

func subPathField(value string) bool {
	if len(value) == 0 || len(value) > 100 {
		return false
	}
	for i, c := range value {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' || i > 0 && c >= '0' && c <= '9' {
			continue
		}
		return false
	}
	return true
}

func subPathValue(value string) bool {
	if value == "" || len(value) > 512 || !utf8.ValidString(value) || value != strings.TrimSpace(value) {
		return false
	}
	for _, c := range value {
		if unicode.IsControl(c) {
			return false
		}
	}
	return true
}
