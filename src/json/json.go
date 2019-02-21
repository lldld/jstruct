package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
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
	_generateMap(&e, &sb, structName, m)

	if e != nil {
		return "", *e
	}

	sb.WriteRune('\n')
	return sb.String(), nil
}

func GenerateMap(e **error, sb *strings.Builder, k string, m map[string]interface{}) {
	if *e != nil {
		return
	}
	_generateMap(e, sb, k, m)
	sb.WriteString(fmt.Sprintf(" `json:\"%s\"`\n", k))
}

func _generateMap(e **error, sb *strings.Builder, k string, m map[string]interface{}) {
	if *e != nil {
		return
	}
	sb.WriteString(nameOfKey(e, k))
	sb.WriteString(" struct {\n")

	nameToKeyMap := make(map[string]string)
	names := make([]string, 0, len(m))
	for k := range m {
		name := nameOfKey(e, k)
		if existsKey, exists := nameToKeyMap[name]; exists {
			err := fmt.Errorf("duplicate field name '%s' for json key: '%s' and '%s'", name, existsKey, k)
			*e = &err
			return
		}

		nameToKeyMap[name] = k
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		k := nameToKeyMap[name]
		v := m[k]
		var t string
		if v == nil {
			t = "interface {}"
		} else {
			t = reflect.TypeOf(v).String()
			if t == "float64" {
				f := v.(float64)
				if math.Trunc(f) == f {
					t = "int64"
				}
			}
		}
		switch t {
		case "bool", "string", "int64", "float64", "interface {}":
			GenerateBase(e, sb, k, t)
		case "[]interface {}":
			GenerateArray(e, sb, k, v.([]interface{}))
		case "map[string]interface {}":
			GenerateMap(e, sb, k, v.(map[string]interface{}))
		default:
			err := fmt.Errorf("field '%s' has invalid data type: %s", k, t)
			*e = &err
			return
		}
	}
	sb.WriteString("}")
}

func GenerateBase(e **error, sb *strings.Builder, k string, baseType string) {
	if *e != nil {
		return
	}
	sb.WriteString(fmt.Sprintf("%s %s `json:\"%s\"`\n", nameOfKey(e, k), baseType, k))
}

func GenerateArray(e **error, sb *strings.Builder, k string, v []interface{}) {
	if *e != nil {
		return
	}
	// todo
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
