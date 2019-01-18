package jsonschema

import (
	"fmt"
	"math"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type propertySuite struct{}

var _ = Suite(&propertySuite{})

type ExampleJSONBasic struct {
	Omitted    string  `json:"-,omitempty"`
	Bool       bool    `json:",omitempty" default:"true"`
	Integer    int     `json:",omitempty" default:"42"`
	Integer8   int8    `json:",omitempty"`
	Integer16  int16   `json:",omitempty"`
	Integer32  int32   `json:",omitempty"`
	Integer64  int64   `json:",omitempty"`
	UInteger   uint    `json:",omitempty"`
	UInteger8  uint8   `json:",omitempty"`
	UInteger16 uint16  `json:",omitempty"`
	UInteger32 uint32  `json:",omitempty"`
	UInteger64 uint64  `json:",omitempty"`
	String     string  `json:",omitempty" default:"ok"`
	Bytes      []byte  `json:",omitempty"`
	Float32    float32 `json:",omitempty"`
	Float64    float64
	Interface  interface{} `required:"true"`
	Timestamp  time.Time   `json:",omitempty"`
}

func (self *propertySuite) TestLoad(c *C) {
	j, err := NewGenerator().WithRoot(&ExampleJSONBasic{}).Generate()
	c.Assert(err, IsNil)

	c.Assert(j, DeepEquals, &JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Property: Property{
			Type:     "object",
			Required: []string{"Interface"},
			Properties: map[string]*Property{
				"Bool":       &Property{Type: "boolean", Default:true},
				"Integer":    &Property{Type: "integer", Default: float64(42)},
				"Integer8":   &Property{Type: "integer"},
				"Integer16":  &Property{Type: "integer"},
				"Integer32":  &Property{Type: "integer"},
				"Integer64":  &Property{Type: "integer"},
				"UInteger":   &Property{Type: "integer"},
				"UInteger8":  &Property{Type: "integer"},
				"UInteger16": &Property{Type: "integer"},
				"UInteger32": &Property{Type: "integer"},
				"UInteger64": &Property{Type: "integer"},
				"String":     &Property{Type: "string", Default: "ok"},
				"Bytes":      &Property{Type: "string"},
				"Float32":    &Property{Type: "number"},
				"Float64":    &Property{Type: "number"},
				"Interface":  &Property{},
				"Timestamp":  &Property{Type: "string", Format: "date-time"},
			},
		},
	})
}

type ExampleJSONBasicWithTag struct {
	meta         string  `json:"-" title:"Title" description:"Description text."`
	Bool         bool    `json:"test" title:"BoolField"`
	String       string  `json:"string" description:"blah" minLength:"3" maxLength:"10" pattern:"m{3,10}"`
	Const        string  `json:"const" const:"blah"`
	Float        float32 `json:"float" min:"1.5" max:"42"`
	Int          int64   `json:"int" exclusiveMin:"-10" exclusiveMax:"0"`
	AnswerToLife int     `json:"answer" const:"42"`
	Fruit        string  `json:"fruit" enum:"apple|banana|pear"`
}

func (self *propertySuite) TestLoadWithTag(c *C) {
	j, err := NewGenerator().WithRoot(&ExampleJSONBasicWithTag{}).Generate()
	c.Assert(err, IsNil)

	c.Assert(j, DeepEquals, &JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Property: Property{
			Type:        "object",
			Title:       "Title",
			Description: "Description text.",
			Properties: map[string]*Property{
				"test": &Property{
					Type:  "boolean",
					Title: "BoolField",
				},
				"string": &Property{
					Type:        "string",
					MinLength:   int64ptr(3),
					MaxLength:   int64ptr(10),
					Pattern:     "m{3,10}",
					Description: "blah",
				},
				"const": &Property{
					Type:  "string",
					Const: "blah",
				},
				"float": &Property{
					Type:    "number",
					Minimum: float64ptr(1.5),
					Maximum: float64ptr(42),
				},
				"int": &Property{
					Type:             "integer",
					ExclusiveMinimum: float64ptr(-10),
					ExclusiveMaximum: float64ptr(0),
				},
				"answer": &Property{
					Type:  "integer",
					Const: int64(42),
				},
				"fruit": &Property{
					Type: "string",
					Enum: []string{"apple", "banana", "pear"},
				},
			},
		},
	})
}

type ExampleJSONBasicSlices struct {
	Slice            []string      `json:",foo,omitempty"`
	SliceOfInterface []interface{} `json:",foo" required:"true"`
}

func (self *propertySuite) TestLoadSliceAndContains(c *C) {
	j, err := NewGenerator().WithRoot(&ExampleJSONBasicSlices{}).Generate()
	c.Assert(err, IsNil)

	c.Assert(j, DeepEquals, &JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Property: Property{
			Type: "object",
			Properties: map[string]*Property{
				"Slice": &Property{
					Type:  "array",
					Items: &Property{Type: "string"},
				},
				"SliceOfInterface": &Property{
					Type: "array",
				},
			},

			Required: []string{"SliceOfInterface"},
		},
	})
}

type ExampleJSONExtensions struct {
	Value string `json:"value" enum:"a|b|c" extensions:"{\"enumNames\": [\"A\",\"B\",\"C\"] }"`
}

func (self *propertySuite) TestExtension(c *C) {
	j, err := NewGenerator().WithRoot(&ExampleJSONExtensions{}).Generate()
	c.Assert(err, IsNil)

	c.Assert(j, DeepEquals, &JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Property: Property{
			Type: "object",
			Properties: map[string]*Property{
				"value": &Property{
					Type: "string",
					Enum: []string{"a", "b", "c"},
					Extensions: map[string]interface{}{
						"enumNames": []interface{}{"A", "B", "C"},
					},
				},
			},
		},
	})
}

type ExampleJSONNestedSlice struct {
	Slice []struct {
		Foo string `required:"true"`
	} `json:"slice"`
}

func (self *propertySuite) TestLoadNested(c *C) {
	j, err := NewGenerator().WithRoot(&ExampleJSONNestedSlice{}).Generate()
	c.Assert(err, IsNil)

	c.Assert(j, DeepEquals, &JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Property: Property{
			Type: "object",
			Properties: map[string]*Property{
				"slice": &Property{
					Type: "array",
					Items: &Property{
						Type: "object",
						Properties: map[string]*Property{
							"Foo": &Property{Type: "string"},
						},
						Required: []string{"Foo"},
					},
				},
			},
		},
	})
}

type ExampleJSONNestedStructReferenceGrandParent struct {
	Child ExampleJSONNestedStructReferenceParent
}
type ExampleJSONNestedStructReferenceParent struct {
	Name  string
	Child ExampleJSONNestedStructReferenceChild
}
type ExampleJSONNestedStructReferenceChild struct {
	Foo string `required:"true"`
}

func (self *propertySuite) TestLoadNestedWithDefinitions(c *C) {
	j := NewGenerator().WithRoot(&ExampleJSONNestedStructReferenceGrandParent{}).
		WithDefinitions(map[string]interface{}{
			"parent": ExampleJSONNestedStructReferenceParent{},
			"child":  ExampleJSONNestedStructReferenceChild{},
		}).MustGenerate()

	k := JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Definitions: map[string]Property{
			"parent": Property{
				Type: "object",
				Properties: map[string]*Property{
					"Name": &Property{
						Type: "string",
					},
					"Child": &Property{
						Ref: "#/definitions/child",
					},
				},
			},
			"child": Property{
				Type: "object",
				Properties: map[string]*Property{
					"Foo": &Property{Type: "string"},
				},
				Required: []string{"Foo"},
			},
		},
		Property: Property{
			Type: "object",
			Properties: map[string]*Property{
				"Child": &Property{
					Ref: "#/definitions/parent",
				},
			},
		},
	}

	c.Assert(findDiff(j.String(), k.String()), Equals, "")
}

type ExampleJSONBasicMaps struct {
	Maps           map[string]string `json:",omitempty"`
	MapOfInterface map[string]interface{}
}

func (self *propertySuite) TestLoadMap(c *C) {
	j := NewGenerator().WithRoot(&ExampleJSONBasicMaps{}).MustGenerate()

	c.Assert(j, DeepEquals, &JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Property: Property{
			Type: "object",
			Properties: map[string]*Property{
				"Maps": &Property{
					Type: "object",
					Properties: map[string]*Property{
						".*": &Property{Type: "string"},
					},
				},
				"MapOfInterface": &Property{
					Type:                 "object",
					AdditionalProperties: true,
				},
			},
		},
	})
}

func (self *propertySuite) TestLoadNonStruct(c *C) {
	j := NewGenerator().WithRoot([]string{}).MustGenerate()

	c.Assert(j, DeepEquals, &JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Property: Property{
			Type:  "array",
			Items: &Property{Type: "string"},
		},
	})
}

func (self *propertySuite) TestString(c *C) {
	j := NewGenerator().WithRoot(true).MustGenerate()

	expected := "{\n" +
		"  \"$schema\": \"http://json-schema.org/schema#\",\n" +
		"  \"type\": \"boolean\"\n" +
		"}"

	c.Assert(j.String(), Equals, expected)
}

func (self *propertySuite) TestMarshal(c *C) {
	j := NewGenerator().WithRoot(10).MustGenerate()

	expected := "{\n" +
		"  \"$schema\": \"http://json-schema.org/schema#\",\n" +
		"  \"type\": \"integer\"\n" +
		"}"

	json := j.String()
	c.Assert(string(json), Equals, expected)
}

type ExampleJSONNestedSliceStruct struct {
	Struct  []ItemStruct
	Struct2 []*ItemStruct
}
type ItemStruct struct {
	Foo string `required:"true"`
}

func (self *propertySuite) TestLoadNestedSliceWithDefinitions(c *C) {
	j := NewGenerator().
		WithDefinitions(map[string]interface{}{
			"parent": ExampleJSONNestedSliceStruct{},
			"item":   &ItemStruct{},
		}).MustGenerate()

	k := JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Definitions: map[string]Property{
			"parent": Property{
				Type: "object",
				Properties: map[string]*Property{
					"Struct": &Property{
						Type: "array",
						Items: &Property{
							Ref: "#/definitions/item",
						},
					},
					"Struct2": &Property{
						Type: "array",
						Items: &Property{
							Ref: "#/definitions/item",
						},
					},
				},
			},
			"item": Property{
				Type: "object",
				Properties: map[string]*Property{
					"Foo": &Property{Type: "string"},
				},
				Required: []string{"Foo"},
			},
		},
	}

	c.Assert(findDiff(j.String(), k.String()), Equals, "")
}

func (self *propertySuite) TestLoadNestedSlice(c *C) {
	j := NewGenerator().WithRoot(&ExampleJSONNestedSliceStruct{}).MustGenerate()

	c.Assert(j, DeepEquals, &JSONSchema{
		Schema: DEFAULT_SCHEMA,
		Property: Property{
			Type: "object",
			Properties: map[string]*Property{
				"Struct": &Property{
					Type: "array",
					Items: &Property{
						Type: "object",
						Properties: map[string]*Property{
							"Foo": &Property{Type: "string"},
						},
						Required: []string{"Foo"},
					},
				},
				"Struct2": &Property{
					Type: "array",
					Items: &Property{
						Type: "object",
						Properties: map[string]*Property{
							"Foo": &Property{Type: "string"},
						},
						Required: []string{"Foo"},
					},
				},
			},
		},
	})
}

func findDiff(a, b string) string {
	var index int
	var different bool
	for ; index < len(a) && index < len(b); index++ {
		if a[index] != b[index] {
			different = true
			break
		}
	}

	if different {
		return fmt.Sprintf("found difference at index %d:\nactual: %q\nexpected: %q",
			index,
			getNearby(a, index),
			getNearby(b, index))
	}

	return ""
}

func getNearby(a string, index int) string {
	min := math.Max(0, float64(index)-10.0)
	max := math.Min(float64(len(a)), float64(index)+10.0)

	return a[int(min):int(max)]
}
