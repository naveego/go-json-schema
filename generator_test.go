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
	Bool       bool    `json:",omitempty"`
	Integer    int     `json:",omitempty"`
	Integer8   int8    `json:",omitempty"`
	Integer16  int16   `json:",omitempty"`
	Integer32  int32   `json:",omitempty"`
	Integer64  int64   `json:",omitempty"`
	UInteger   uint    `json:",omitempty"`
	UInteger8  uint8   `json:",omitempty"`
	UInteger16 uint16  `json:",omitempty"`
	UInteger32 uint32  `json:",omitempty"`
	UInteger64 uint64  `json:",omitempty"`
	String     string  `json:",omitempty"`
	Bytes      []byte  `json:",omitempty"`
	Float32    float32 `json:",omitempty"`
	Float64    float64
	Interface  interface{} `required:"true"`
	Timestamp  time.Time   `json:",omitempty"`
}

func (self *propertySuite) TestLoad(c *C) {
	j, err := NewGenerator().WithRoot(&ExampleJSONBasic{}).Generate()
	c.Assert(err, IsNil)

	c.Assert(j, DeepEquals, JSONSchema{
		Schema: DEFAULT_SCHEMA,
		property: property{
			Type:     "object",
			Required: []string{"Interface"},
			Properties: map[string]*property{
				"Bool":       &property{Type: "boolean"},
				"Integer":    &property{Type: "integer"},
				"Integer8":   &property{Type: "integer"},
				"Integer16":  &property{Type: "integer"},
				"Integer32":  &property{Type: "integer"},
				"Integer64":  &property{Type: "integer"},
				"UInteger":   &property{Type: "integer"},
				"UInteger8":  &property{Type: "integer"},
				"UInteger16": &property{Type: "integer"},
				"UInteger32": &property{Type: "integer"},
				"UInteger64": &property{Type: "integer"},
				"String":     &property{Type: "string"},
				"Bytes":      &property{Type: "string"},
				"Float32":    &property{Type: "number"},
				"Float64":    &property{Type: "number"},
				"Interface":  &property{},
				"Timestamp":  &property{Type: "string", Format: "date-time"},
			},
		},
	})
}

type ExampleJSONBasicWithTag struct {
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

	c.Assert(j, DeepEquals, JSONSchema{
		Schema: DEFAULT_SCHEMA,
		property: property{
			Type: "object",
			Properties: map[string]*property{
				"test": &property{
					Type:  "boolean",
					Title: "BoolField",
				},
				"string": &property{
					Type:        "string",
					MinLength:   int64ptr(3),
					MaxLength:   int64ptr(10),
					Pattern:     "m{3,10}",
					Description: "blah",
				},
				"const": &property{
					Type:  "string",
					Const: "blah",
				},
				"float": &property{
					Type:    "number",
					Minimum: float64ptr(1.5),
					Maximum: float64ptr(42),
				},
				"int": &property{
					Type:             "integer",
					ExclusiveMinimum: float64ptr(-10),
					ExclusiveMaximum: float64ptr(0),
				},
				"answer": &property{
					Type:  "integer",
					Const: int64(42),
				},
				"fruit": &property{
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

	c.Assert(j, DeepEquals, JSONSchema{
		Schema: DEFAULT_SCHEMA,
		property: property{
			Type: "object",
			Properties: map[string]*property{
				"Slice": &property{
					Type:  "array",
					Items: &property{Type: "string"},
				},
				"SliceOfInterface": &property{
					Type: "array",
				},
			},

			Required: []string{"SliceOfInterface"},
		},
	})
}

type ExampleJSONNestedStruct struct {
	Struct struct {
		Foo string `required:"true"`
	}
}

type ExampleJSONNestedSlice struct {
	Slice []struct {
		Foo string `required:"true"`
	} `json:"slice"`
}

func (self *propertySuite) TestLoadNested(c *C) {
	j, err := NewGenerator().WithRoot(&ExampleJSONNestedSlice{}).Generate()
	c.Assert(err, IsNil)

	c.Assert(j, DeepEquals, JSONSchema{
		Schema: DEFAULT_SCHEMA,
		property: property{
			Type: "object",
			Properties: map[string]*property{
				"slice": &property{
					Type: "array",
					Items: &property{
						Type: "object",
						Properties: map[string]*property{
							"Foo": &property{Type: "string"},
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
		Definitions: map[string]property{
			"parent": property{
				Type: "object",
				Properties: map[string]*property{
					"Name": &property{
						Type: "string",
					},
					"Child": &property{
						Ref: "#/definitions/child",
					},
				},
			},
			"child": property{
				Type: "object",
				Properties: map[string]*property{
					"Foo": &property{Type: "string"},
				},
				Required: []string{"Foo"},
			},
		},
		property: property{
			Type: "object",
			Properties: map[string]*property{
				"Child": &property{
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

	c.Assert(j, DeepEquals, JSONSchema{
		Schema: DEFAULT_SCHEMA,
		property: property{
			Type: "object",
			Properties: map[string]*property{
				"Maps": &property{
					Type: "object",
					Properties: map[string]*property{
						".*": &property{Type: "string"},
					},
					AdditionalProperties: false,
				},
				"MapOfInterface": &property{
					Type:                 "object",
					AdditionalProperties: true,
				},
			},
		},
	})
}

func (self *propertySuite) TestLoadNonStruct(c *C) {
	j := NewGenerator().WithRoot([]string{}).MustGenerate()

	c.Assert(j, DeepEquals, JSONSchema{
		Schema: DEFAULT_SCHEMA,
		property: property{
			Type:  "array",
			Items: &property{Type: "string"},
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
		Definitions: map[string]property{
			"parent": property{
				Type: "object",
				Properties: map[string]*property{
					"Struct": &property{
						Type: "array",
						Items: &property{
							Ref: "#/definitions/item",
						},
					},
					"Struct2": &property{
						Type: "array",
						Items: &property{
							Ref: "#/definitions/item",
						},
					},
				},
			},
			"item": property{
				Type: "object",
				Properties: map[string]*property{
					"Foo": &property{Type: "string"},
				},
				Required: []string{"Foo"},
			},
		},
	}

	c.Assert(findDiff(j.String(), k.String()), Equals, "")
}

func (self *propertySuite) TestLoadNestedSlice(c *C) {
	j := NewGenerator().WithRoot(&ExampleJSONNestedSliceStruct{}).MustGenerate()

	c.Assert(j, DeepEquals, JSONSchema{
		Schema: DEFAULT_SCHEMA,
		property: property{
			Type: "object",
			Properties: map[string]*property{
				"Struct": &property{
					Type: "array",
					Items: &property{
						Type: "object",
						Properties: map[string]*property{
							"Foo": &property{Type: "string"},
						},
						Required: []string{"Foo"},
					},
				},
				"Struct2": &property{
					Type: "array",
					Items: &property{
						Type: "object",
						Properties: map[string]*property{
							"Foo": &property{Type: "string"},
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
