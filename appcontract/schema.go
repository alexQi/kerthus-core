package appcontract

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"unicode/utf8"
)

// This deliberately bounded JSON Schema subset rejects unsupported keywords, rather than silently ignoring them.
func ValidateSchema(s map[string]any, depth int) error {
	if depth > 8 {
		return fmt.Errorf("schema nesting exceeds 8")
	}
	for k := range s {
		switch k {
		case "type", "description", "properties", "required", "additionalProperties", "items", "enum", "minimum", "maximum", "maxLength", "minLength", "maxItems", "minItems":
		default:
			return fmt.Errorf("unsupported schema keyword %s", k)
		}
	}
	kind, _ := s["type"].(string)
	allowed := map[string]bool{"type": true, "description": true, "enum": true}
	switch kind {
	case "object":
		for _, k := range []string{"properties", "required", "additionalProperties"} {
			allowed[k] = true
		}
		if s["additionalProperties"] != false {
			return fmt.Errorf("object requires additionalProperties: false")
		}
		props, ok := s["properties"].(map[string]any)
		if !ok || len(props) > 100 {
			return fmt.Errorf("invalid object properties")
		}
		for name, v := range props {
			if len(name) > 100 {
				return fmt.Errorf("property name too long")
			}
			child, ok := v.(map[string]any)
			if !ok {
				return fmt.Errorf("invalid property schema")
			}
			if e := ValidateSchema(child, depth+1); e != nil {
				return e
			}
		}
		if req, ok := s["required"]; ok {
			list, ok := req.([]any)
			if !ok {
				return fmt.Errorf("required must be array")
			}
			seen := map[string]bool{}
			for _, v := range list {
				name, ok := v.(string)
				if !ok || props[name] == nil || seen[name] {
					return fmt.Errorf("invalid required property")
				}
				seen[name] = true
			}
		}
	case "array":
		for _, k := range []string{"items", "maxItems", "minItems"} {
			allowed[k] = true
		}
		child, ok := s["items"].(map[string]any)
		if !ok {
			return fmt.Errorf("array items missing")
		}
		if e := ValidateSchema(child, depth+1); e != nil {
			return e
		}
		if n, ok := number(s["maxItems"]); !ok || n < 0 || n > 1000 {
			return fmt.Errorf("array needs maxItems <=1000")
		}
	case "string":
		allowed["minLength"] = true
		allowed["maxLength"] = true
		if n, ok := number(s["maxLength"]); !ok || n < 1 || n > 32768 {
			return fmt.Errorf("string needs maxLength <=32768")
		}
	case "number", "integer":
		allowed["minimum"] = true
		allowed["maximum"] = true
	case "boolean", "null":
	default:
		return fmt.Errorf("unsupported schema type")
	}
	for k := range s {
		if !allowed[k] {
			return fmt.Errorf("schema keyword %s invalid for %s", k, kind)
		}
	}
	for _, pair := range [][2]string{{"minimum", "maximum"}, {"minItems", "maxItems"}, {"minLength", "maxLength"}} {
		var lo, hi float64
		var hasLo, hasHi bool
		for i, k := range pair {
			if v, ok := s[k]; ok {
				n, ok := number(v)
				if !ok || math.IsNaN(n) || math.IsInf(n, 0) || (k != "minimum" && k != "maximum" && (n < 0 || n != math.Trunc(n))) {
					return fmt.Errorf("invalid numeric schema bound")
				}
				if i == 0 {
					lo = n
					hasLo = true
				} else {
					hi = n
					hasHi = true
				}
			}
		}
		if hasLo && hasHi && lo > hi {
			return fmt.Errorf("inverted schema bounds")
		}
	}
	if v, ok := s["enum"]; ok {
		list, ok := v.([]any)
		if !ok || len(list) == 0 || len(list) > 100 {
			return fmt.Errorf("invalid enum")
		}
	}
	return nil
}
func number(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, e := n.Float64()
		return f, e == nil
	}
	return 0, false
}
func ValidateValue(s map[string]any, v any) error {
	kind, _ := s["type"].(string)
	bound := func(n float64, min, max string) error {
		if lo, ok := number(s[min]); ok && n < lo {
			return fmt.Errorf("value below %s", min)
		}
		if hi, ok := number(s[max]); ok && n > hi {
			return fmt.Errorf("value above %s", max)
		}
		return nil
	}
	switch kind {
	case "object":
		obj, ok := v.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object")
		}
		props, _ := s["properties"].(map[string]any)
		req, _ := s["required"].([]any)
		for _, name := range req {
			if _, ok := obj[name.(string)]; !ok {
				return fmt.Errorf("required field %s", name)
			}
		}
		for name, value := range obj {
			child, ok := props[name].(map[string]any)
			if !ok {
				if allowed, _ := s["additionalProperties"].(bool); allowed {
					continue
				}
				return fmt.Errorf("unknown field %s", name)
			}
			if e := ValidateValue(child, value); e != nil {
				return fmt.Errorf("%s: %w", name, e)
			}
		}
	case "array":
		list, ok := v.([]any)
		if !ok {
			return fmt.Errorf("expected array")
		}
		if e := bound(float64(len(list)), "minItems", "maxItems"); e != nil {
			return e
		}
		child, _ := s["items"].(map[string]any)
		for _, value := range list {
			if e := ValidateValue(child, value); e != nil {
				return e
			}
		}
	case "string":
		str, ok := v.(string)
		if !ok {
			return fmt.Errorf("expected string")
		}
		if e := bound(float64(utf8.RuneCountInString(str)), "minLength", "maxLength"); e != nil {
			return e
		}
	case "number", "integer":
		n, ok := number(v)
		if !ok || math.IsNaN(n) || math.IsInf(n, 0) || (kind == "integer" && n != math.Trunc(n)) {
			return fmt.Errorf("expected %s", kind)
		}
		if e := bound(n, "minimum", "maximum"); e != nil {
			return e
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			return fmt.Errorf("expected boolean")
		}
	case "null":
		if v != nil {
			return fmt.Errorf("expected null")
		}
	default:
		return fmt.Errorf("invalid schema")
	}
	if list, ok := s["enum"].([]any); ok {
		for _, item := range list {
			if reflect.DeepEqual(item, v) {
				return nil
			}
		}
		return fmt.Errorf("value not in enum")
	}
	return nil
}
