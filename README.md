# go-json-schema-generator
[![codecov](https://codecov.io/gh/SteveRuble/go-json-schema-generator/branch/master/graph/badge.svg)](https://codecov.io/gh/SteveRuble/go-json-schema-generator)


Generate JSON Schema out of Golang schema

### Usage

First install the package
```
go get -u github.com/naveego/go-json-schema
```

Then create your generator file (see [Example](https://github.com/SteveRuble/go-json-schema-generator/blob/master/example) folder)
```go
package main

import (
	"fmt"
	"github.com/naveego/go-json-schema"
)

type Domain struct {
	Data string `json:"data"`
}

func main(){
	fmt.Println(jsonschema.NewGenerator().WithRoot(&Domain{}).MustGenerate())
}
```

### Definitions

To include a `"definitions"` element in your schema, you need to pass the definitions to the generator:

```go
package main

import (
	"fmt"
	"github.com/naveego/go-json-schema"
)

type Domain struct {
	Child Child `json:"child"`
}

type Child struct {
	Data string `json:"data"`
}

func main(){
    js, err := jsonschema.NewGenerator().
        WithRoot(&Domain{}).
        WithDefinition("child", &Child{}).
        Generate()
    if err != nil {
    	panic(err)
    }
    fmt.Println(js.String())
}
```

### Supported tags

* `required:"true"` - field will be marked as required
* `title:"Title"` - title will be added
* `description:"description"` - description will be added
* `extensions:"{\"enumNames\": [\"A\",\"B\",\"C\"] }"` - The JSON value of the tag will be merged into the resulting schema.

> On an unexported field, these tags will be added to the schema for the struct itself
> rather than to the property representing the field.

##### On string fields:

* `minLength:"5"` - Set the minimum length of the value
* `maxLength:"5"` - Set the maximum length of the value
* `enum:"apple|banana|pear"` - Limit the available values to a defined set, separated by vertical bars
* `const:"I need to be there"` - Require the field to have a specific value.

##### On numeric types (strings and floats)

* `min:"-4.141592"` -  Set a minimum value
* `max:"123456789"` -  Set a maximum value
* `exclusiveMin:"0"` - Values must be strictly greater than this value
* `exclusiveMax:"11"` - Values must be strictly smaller than this value
* `const:"42"` - Property must have exactly this value.

### Expected behaviour

If struct field is pointer to the primitive type, then schema will allow this type and null.
E.g.:

```
type Domain struct {
	NullableData *string `json:"nullableData"`
}
```
Output

```
{
    "$schema": "http://json-schema.org/schema#",
    "type": "object",
    "properties": {
        "nullableData": {
            "anyOf": [
                {
                    "type": "string"
                },
                {
                    "type": "null"
                }
            ]
        }
    }
}

```
