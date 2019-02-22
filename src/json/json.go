package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
)

func Parse(filepath string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	t := reflect.TypeOf(data)
	if t.String() != "map[string]interface {}" {
		return nil, fmt.Errorf("the format of file is not json")
	}

	return data.(map[string]interface{}), nil
}

func Generate(packageName string, structName string, m map[string]interface{}) (string, error) {

	var sb strings.Builder
	var e *error

	sb.WriteString("package ")
	sb.WriteString(packageName)
	sb.WriteString("\n\n")

	sb.WriteString("type ")
	sb.WriteString(structName)
	sb.WriteString(" ")
	sb.WriteString(parseKeyVal(&e, "", m))

	if e != nil {
		return "", *e
	}

	sb.WriteRune('\n')
	return sb.String(), nil
}

func typeOfArray(e **error, k string, v []interface{}) string {
	if *e != nil {
		return ""
	}

	var t string
	for i, _v := range v {
		if i == 0 {
			t = typeOfVal(e, k, _v)
		} else if t != typeOfVal(e, k, _v) {
			err := fmt.Errorf("not all elements of key '%s' have the same data type '%s'", k, t)
			*e = &err
			return ""
		}
	}

	return fmt.Sprintf("[]%s", t)
}

func typeOfMap(e **error, k string, m map[string]interface{}) string {
	if *e != nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("struct {\n")
	nameToKeyMap := make(map[string]string)
	names := make([]string, 0, len(m))
	for _k := range m {
		name := nameOfKey(e, _k)
		if _kExists, exists := nameToKeyMap[name]; exists {
			err := fmt.Errorf("duplicate name '%s' for key '%s' and '%s' of key '%s'", name, _kExists, _k, k)
			*e = &err
			return ""
		}

		nameToKeyMap[name] = _k
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		_k := nameToKeyMap[name]
		v := m[_k]
		sb.WriteString(parseKeyVal(e, _k, v))
	}
	sb.WriteString("}")

	return sb.String()
}

func parseKeyVal(e **error, k string, v interface{}) string {
	if *e != nil {
		return ""
	}

	t := typeOfVal(e, k, v)

	if k == "" {
		return t + "\n"
	}

	return fmt.Sprintf("%s %s `json:\"%s\"`\n", nameOfKey(e, k), t, k)
}

func typeOfVal(e **error, k string, v interface{}) string {
	if *e != nil {
		return ""
	}

	var t string
	if v == nil {
		t = "interface {}"
	} else {
		t = reflect.TypeOf(v).String()
	}

	switch t {
	case "bool", "string", "float64", "interface {}":
		return t
	case "[]interface {}":
		return typeOfArray(e, k, v.([]interface{}))
	case "map[string]interface {}":
		return typeOfMap(e, k, v.(map[string]interface{}))
	default:
		err := fmt.Errorf("key '%s' has unsupported data type '%s'", k, t)
		*e = &err
		return ""
	}
}

func nameOfKey(e **error, key string) string {
	if *e != nil {
		return ""
	}

	var sb strings.Builder
	if len(key) < 1 {
		err := fmt.Errorf("field name is empty")
		*e = &err
		return ""
	}

	firstChar := []rune(key)[0]
	if !((firstChar >= 'a' && firstChar <= 'z') || (firstChar >= 'A' && firstChar <= 'Z')) {
		err := fmt.Errorf("first character of key '%s' is not english alphabet", key)
		*e = &err
		return ""
	}

	parts := splitByNonAlphabetNonNumber(key)
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		partRunes := []rune(part)
		firstCharOfPart := partRunes[0]
		if firstCharOfPart >= 'a' && firstCharOfPart <= 'z' {
			sb.WriteRune(firstCharOfPart + 'A' - 'a')
			sb.WriteString(part[1:])
		} else {
			sb.WriteString(part)
		}
	}

	return sb.String()
}

func splitByNonAlphabetNonNumber(s string) []string {
	f := func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0') && (r <= '9'))
	}
	return strings.FieldsFunc(s, f)
}
