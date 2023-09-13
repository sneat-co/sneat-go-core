package flexidoc

type DocumentDefinitionBase struct {
	Fields []*Field `json:"fields" firestore:"fields"`
}

type Field struct {
	ID       string            `json:"id" firestore:"id"`
	Type     string            `json:"type" firestore:"id"` //e.g. "string", "int", "lookup"'
	Required bool              `json:"required" firestore:"required"`
	Titles   map[string]string `json:"titles" firestore:"titles"`
}
