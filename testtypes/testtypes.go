package testtypes


type ExampleJSONBasicWithTag struct {
	meta string `json:"-" schema-title:"Title" schema-description:"Description text."`
	Bool         bool    `json:"test" title:"BoolField"`
	String       string  `json:"string" description:"blah" minLength:"3" maxLength:"10" pattern:"m{3,10}"`
	Const        string  `json:"const" const:"blah"`
	Float        float32 `json:"float" min:"1.5" max:"42"`
	Int          int64   `json:"int" exclusiveMin:"-10" exclusiveMax:"0"`
	AnswerToLife int     `json:"answer" const:"42"`
	Fruit        string  `json:"fruit" enum:"apple|banana|pear"`
}