// Copyright Kozyrev Yury
// MIT license.
package jsonschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const DEFAULT_SCHEMA = "http://json-schema.org/schema#"

var rTypeInt64, rTypeFloat64 = reflect.TypeOf(int64(0)), reflect.TypeOf(float64(0))

type JSONSchema struct {
	Schema      string              `json:"$schema,omitempty"`
	Definitions map[string]Property `json:"definitions,omitempty"`
	Property
}

type knownTypes map[reflect.Type]string

func (k knownTypes) getReference(t reflect.Type) (string, bool) {
	if k != nil {
		if name, ok := k[t]; ok {
			return fmt.Sprintf("#/definitions/%s", name), true
		}
	}
	return "", false
}

type Generator struct {
	root        interface{}
	definitions map[string]interface{}
	options     Options
}

type Options struct {
	Schema string
}

func Generate(root interface{}) string {
	js, _ := NewGenerator().WithRoot(root).Generate()
	return js.String()
}

func NewGenerator(options ...Options) *Generator {
	g := &Generator{}
	if len(options) > 0 {
		g.options = options[0]
	}
	if g.options.Schema == "" {
		g.options.Schema = DEFAULT_SCHEMA
	}
	return g
}

func (g *Generator) WithRoot(r interface{}) *Generator {
	g.root = r
	return g
}

func (g *Generator) WithDefinitions(d map[string]interface{}) *Generator {
	for k, v := range d {
		g = g.WithDefinition(k, v)
	}
	return g
}

func (g *Generator) WithDefinition(name string, d interface{}) *Generator {
	if g.definitions == nil {
		g.definitions = map[string]interface{}{}
	}
	g.definitions[name] = d
	return g
}

func (g *Generator) MustGenerate() *JSONSchema {
	js, err := g.Generate()
	if err != nil {
		panic(err)
	}
	return js
}

// Generate generates a schema for the provided interface.
func (g *Generator) Generate() (*JSONSchema, error) {
	var err error
	d := &JSONSchema{
		Schema: g.options.Schema,
	}

	if g.definitions != nil {
		d.knownTypes = make(map[reflect.Type]string)
		d.Definitions = make(map[string]Property)

		for name, instance := range g.definitions {
			value := reflect.ValueOf(instance)
			defType := value.Type()
			if defType.Kind() == reflect.Ptr {
				defType = defType.Elem()
			}
			d.knownTypes[defType] = name
		}
	}

	for defType, name := range d.knownTypes {
		p := d.child()
		p.isDefinition = true
		err = p.read(defType)
		if err != nil {
			return nil, fmt.Errorf("error on type %s (%s): %s", defType, name, err)
		}
		d.Definitions[name] = *p
	}

	if g.root != nil {
		value := reflect.ValueOf(g.root)
		err = d.read(value.Type())
		if err != nil {
			return nil, fmt.Errorf("error on root type %T: %s", g.root, err)
		}
	}

	return d, nil
}

// String return the JSON encoding of the JSONSchema as a string
func (d JSONSchema) String() string {
	json, _ := json.MarshalIndent(d, "", "  ")
	return string(json)
}

func (d *JSONSchema) setDefaultSchema() {
	if d.Schema == "" {
		d.Schema = DEFAULT_SCHEMA
	}
}

type Property struct {
	Type                 string               `json:"type,omitempty"`
	Format               string               `json:"format,omitempty"`
	Items                *Property            `json:"items,omitempty"`
	Properties           map[string]*Property `json:"properties,omitempty"`
	Required             []string             `json:"required,omitempty"`
	AdditionalProperties bool                 `json:"additionalProperties,omitempty"`
	Description          string               `json:"description,omitempty"`
	AnyOf                []*Property          `json:"anyOf,omitempty"`
	OneOf                []*Property          `json:"oneOf,omitempty"`
	Dependencies         map[string]*Property `json:"dependencies,omitempty"`
	Default interface{} `json:"default,omitempty"`
	Extensions map[string]interface{} `json:"-"`

	// validation keywords:
	// For any number-valued fields, we're making them pointers, because
	// we want empty values to be omitted, but for numbers, 0 is seen as empty.

	// numbers validators
	MultipleOf       *float64 `json:"multipleOf,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	Minimum          *float64 `json:"minimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	// string validators
	MaxLength *int64 `json:"maxLength,omitempty"`
	MinLength *int64 `json:"minLength,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	// Enum is defined for arbitrary types, but I'm currently just implementing it for strings.
	Enum  []string `json:"enum,omitempty"`
	Title string   `json:"title,omitempty"`
	// Implemented for strings and numbers
	Const        interface{} `json:"const,omitempty"`
	Ref          string      `json:"$ref,omitempty"`
	knownTypes   knownTypes
	isDefinition bool
}

type marshallingProperty Property

func (p *Property) MarshalJSON() ([]byte, error) {
	if p == nil {
		return nil, nil
	}
	mp := marshallingProperty(*p)
	b, err := json.Marshal(mp)
	if err != nil {
		return nil, err
	}

	if p.Extensions == nil {
		return b, nil
	}

	// add extensions at the top level of the output
	var raw map[string]interface{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		return nil, err
	}

	for k, v := range p.Extensions {
		raw[k] = v
	}

	b, err = json.Marshal(raw)
	return b, err
}

func (p *Property) child() *Property {
	return &Property{knownTypes: p.knownTypes}
}

func (p *Property) read(t reflect.Type) error {
	jsType, format, kind := getTypeFromMapping(t)
	if jsType != "" {
		p.Type = jsType
	}
	if format != "" {
		p.Format = format
	}

	var err error

	switch kind {
	case reflect.Slice:
		err = p.readFromSlice(t)
	case reflect.Map:
		err = p.readFromMap(t)
	case reflect.Struct:
		err = p.readFromStruct(t)
	case reflect.Ptr:
		err = p.read(t.Elem())
	}

	if err != nil {
		return err
	}

	// say we have *int
	if kind == reflect.Ptr && isPrimitive(t.Elem().Kind()) {
		p.AnyOf = []*Property{
			{Type: p.Type},
			{Type: "null"},
		}
		p.Type = ""
	}

	return nil
}

func (p *Property) readFromSlice(t reflect.Type) error {
	jsType, _, kind := getTypeFromMapping(t.Elem())
	if kind == reflect.Uint8 {
		p.Type = "string"
	} else if jsType != "" || kind == reflect.Ptr {
		p.Items = p.child()
		return p.Items.read(t.Elem())
	}
	return nil
}

func (p *Property) readFromMap(t reflect.Type) error {
	jsType, format, _ := getTypeFromMapping(t.Elem())

	if jsType != "" {
		p.Properties = make(map[string]*Property, 0)
		p.Properties[".*"] = &Property{Type: jsType, Format: format}
	} else {
		p.AdditionalProperties = true
	}
	return nil
}

func (p *Property) readFromStruct(t reflect.Type) error {
	var err error
	var ok bool
	if !p.isDefinition {
		if p.Ref, ok = p.knownTypes.getReference(t); ok {
			p.Type = ""
			return nil
		}
	}

	p.Type = "object"
	p.Properties = make(map[string]*Property, 0)
	p.AdditionalProperties = false

	count := t.NumField()
	for i := 0; i < count; i++ {
		field := t.Field(i)

		tag := field.Tag.Get("json")

		name, opts := parseTag(tag)

		var target *Property
		if field.PkgPath == "" {
			// this is an exported property
			target = p.child()

			err := target.read(field.Type)
			if err != nil {
				return fmt.Errorf("property:%s:%s", field.Name, err)
			}
			if name == "" {
				name = field.Name
			}
			if name == "-" {
				continue
			}
			p.Properties[name] = target
		} else {
			// not an exported field, tags apply to this property
			target = p
		}

		target.Description = field.Tag.Get("description")
		target.Title = field.Tag.Get("title")
		target.addValidatorsFromTags(&field.Tag)

		dflt, ok := field.Tag.Lookup("default")
		if ok {
			switch target.Type {
			case "string": target.Default = dflt
			case "number", "integer":
				target.Default, err = strconv.ParseFloat(dflt, 64)
				if err != nil {
					return fmt.Errorf("could not parse %q to float64 for property %s", dflt, name)
				}
			case "boolean":
				target.Default, err = strconv.ParseBool(dflt)
				if err != nil {
					return fmt.Errorf("could not parse %q to bool for property %s", dflt, name)
				}
			default:
				return fmt.Errorf("default not supported for type %q on property %s", target.Type, name)
			}
		}

		extensionsRaw, hasExtensions := field.Tag.Lookup("extensions")
		if hasExtensions {
			var extensionsMap map[string]interface{}
			err := json.Unmarshal([]byte(extensionsRaw), &extensionsMap)
			if err != nil {
				return fmt.Errorf(`invalid "extensions" tag value %q: %s`, extensionsRaw, err)
			}
			target.Extensions = extensionsMap
		}

		_, required := field.Tag.Lookup("required")
		if opts.Contains("omitempty") || !required {
			continue
		}
		p.Required = append(p.Required, name)
	}

	return nil
}

func (p *Property) addValidatorsFromTags(tag *reflect.StructTag) {
	switch p.Type {
	case "string":
		p.addStringValidators(tag)
	case "number", "integer":
		p.addNumberValidators(tag)
	}
}

// Some helper functions for not having to create temp variables all over the place
func int64ptr(i interface{}) *int64 {
	v := reflect.ValueOf(i)
	if !v.Type().ConvertibleTo(rTypeInt64) {
		return nil
	}
	j := v.Convert(rTypeInt64).Interface().(int64)
	return &j
}

func float64ptr(i interface{}) *float64 {
	v := reflect.ValueOf(i)
	if !v.Type().ConvertibleTo(rTypeFloat64) {
		return nil
	}
	j := v.Convert(rTypeFloat64).Interface().(float64)
	return &j
}

func (p *Property) addStringValidators(tag *reflect.StructTag) {
	// min length
	mls := tag.Get("minLength")
	ml, err := strconv.ParseInt(mls, 10, 64)
	if err == nil {
		p.MinLength = int64ptr(ml)
	}
	// max length
	mls = tag.Get("maxLength")
	ml, err = strconv.ParseInt(mls, 10, 64)
	if err == nil {
		p.MaxLength = int64ptr(ml)
	}
	// pattern
	pat := tag.Get("pattern")
	if pat != "" {
		p.Pattern = pat
	}
	// enum
	en := tag.Get("enum")
	if en != "" {
		p.Enum = strings.Split(en, "|")
	}
	// const
	c := tag.Get("const")
	if c != "" {
		p.Const = c
	}
}

func (p *Property) addNumberValidators(tag *reflect.StructTag) {
	m, err := strconv.ParseFloat(tag.Get("multipleOf"), 64)
	if err == nil {
		p.MultipleOf = float64ptr(m)
	}
	m, err = strconv.ParseFloat(tag.Get("min"), 64)
	if err == nil {
		p.Minimum = float64ptr(m)
	}
	m, err = strconv.ParseFloat(tag.Get("max"), 64)
	if err == nil {
		p.Maximum = float64ptr(m)
	}
	m, err = strconv.ParseFloat(tag.Get("exclusiveMin"), 64)
	if err == nil {
		p.ExclusiveMinimum = float64ptr(m)
	}
	m, err = strconv.ParseFloat(tag.Get("exclusiveMax"), 64)
	if err == nil {
		p.ExclusiveMaximum = float64ptr(m)
	}
	c, err := parseType(tag.Get("const"), p.Type)
	if err == nil {
		p.Const = c
	}
}

func parseType(str, ty string) (interface{}, error) {
	var v interface{}
	var err error
	if ty == "number" {
		v, err = strconv.ParseFloat(str, 64)
	} else {
		v, err = strconv.ParseInt(str, 10, 64)
	}
	return v, err
}

var formatMapping = map[string][]string{
	"time.Time": []string{"string", "date-time"},
}

var kindMapping = map[reflect.Kind]string{
	reflect.Bool:    "boolean",
	reflect.Int:     "integer",
	reflect.Int8:    "integer",
	reflect.Int16:   "integer",
	reflect.Int32:   "integer",
	reflect.Int64:   "integer",
	reflect.Uint:    "integer",
	reflect.Uint8:   "integer",
	reflect.Uint16:  "integer",
	reflect.Uint32:  "integer",
	reflect.Uint64:  "integer",
	reflect.Float32: "number",
	reflect.Float64: "number",
	reflect.String:  "string",
	reflect.Slice:   "array",
	reflect.Struct:  "object",
	reflect.Map:     "object",
}

func isPrimitive(k reflect.Kind) bool {
	if v, ok := kindMapping[k]; ok {
		switch v {
		case "boolean":
		case "integer":
		case "number":
		case "string":
			return true
		}
	}
	return false
}

func getTypeFromMapping(t reflect.Type) (string, string, reflect.Kind) {
	if v, ok := formatMapping[t.String()]; ok {
		return v[0], v[1], reflect.String
	}

	if v, ok := kindMapping[t.Kind()]; ok {
		return v, "", t.Kind()
	}

	return "", "", t.Kind()
}

type structTag string

func parseTag(tag string) (string, structTag) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], structTag(tag[idx+1:])
	}
	return tag, structTag("")
}

func (o structTag) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}

	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

var _ fmt.Stringer = (*JSONSchema)(nil)
