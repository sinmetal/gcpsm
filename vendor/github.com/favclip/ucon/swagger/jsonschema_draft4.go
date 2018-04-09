package swagger

// Schema is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#schemaObject
type Schema struct {
	Ref                  string             `json:"$ref,omitempty"`
	Format               string             `json:"format,omitempty"`
	Title                string             `json:"title,omitempty"`
	Description          string             `json:"description,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	Maximum              *int               `json:"maximum,omitempty"`
	ExclusiveMaximum     *bool              `json:"exclusiveMaximum,omitempty"`
	Minimum              *int               `json:"minimum,omitempty"`
	ExclusiveMinimum     *bool              `json:"exclusiveMinimum,omitempty"`
	MaxLength            *int               `json:"maxLength,omitempty"`
	MinLength            *int               `json:"minLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	MaxItems             *int               `json:"maxItems,omitempty"`
	MinItems             *int               `json:"minItems,omitempty"`
	UniqueItems          *bool              `json:"uniqueItems,omitempty"`
	MaxProperties        *int               `json:"maxProperties,omitempty"`
	MinProperties        *int               `json:"minProperties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	Enum                 []interface{}      `json:"enum,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Items                *Schema            `json:"items,omitempty"`
	AllOf                []*Schema          `json:"allOf,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty"`
	AdditionalProperties map[string]*Schema `json:"additionalProperties,omitempty"`
	Discriminator        string             `json:"discriminator,omitempty"`
	ReadOnly             *bool              `json:"readOnly,omitempty"`
	// Xml XML
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
	Example      interface{}            `json:"example,omitempty"`
}

// ShallowCopy returns a clone of *Schema.
func (schema *Schema) ShallowCopy() *Schema {
	return &Schema{
		Ref:                  schema.Ref,
		Format:               schema.Format,
		Title:                schema.Title,
		Description:          schema.Description,
		Default:              schema.Default,
		Maximum:              schema.Maximum,
		ExclusiveMaximum:     schema.ExclusiveMaximum,
		Minimum:              schema.Minimum,
		ExclusiveMinimum:     schema.ExclusiveMinimum,
		MaxLength:            schema.MaxLength,
		MinLength:            schema.MinLength,
		Pattern:              schema.Pattern,
		MaxItems:             schema.MaxItems,
		MinItems:             schema.MinItems,
		UniqueItems:          schema.UniqueItems,
		MaxProperties:        schema.MaxProperties,
		MinProperties:        schema.MinProperties,
		Required:             schema.Required,
		Enum:                 schema.Enum,
		Type:                 schema.Type,
		Items:                schema.Items,
		AllOf:                schema.AllOf,
		Properties:           schema.Properties,
		AdditionalProperties: schema.AdditionalProperties,
		Discriminator:        schema.Discriminator,
		ReadOnly:             schema.ReadOnly,
		ExternalDocs:         schema.ExternalDocs,
		Example:              schema.Example,
	}
}
