package swagger

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/favclip/ucon"
)

// TagSwagger is a struct tag for setting attributes of swagger.
type TagSwagger reflect.StructTag

// NewTagSwagger casts the tag to TagSwagger.
func NewTagSwagger(tag reflect.StructTag) TagSwagger {
	return TagSwagger(tag)
}

// Private returns whether the field is hidden from swagger.
func (swaggerTag TagSwagger) Private() bool {
	text := reflect.StructTag(swaggerTag).Get("swagger")
	name := strings.Split(text, ",")[0]

	return name == "-"
}

// Name returns the name for swagger.
// If the name is not given by `swagger` tag, `json` tag will be used instead.
func (swaggerTag TagSwagger) Name() string {
	{
		text := reflect.StructTag(swaggerTag).Get("swagger")
		name := strings.Split(text, ",")[0]

		if name != "" {
			return name
		}
	}
	{
		text := reflect.StructTag(swaggerTag).Get("json")
		name := strings.Split(text, ",")[0]

		if name != "" {
			return name
		}
	}

	return ""
}

// Empty returns whether the text of the tag is empty.
func (swaggerTag TagSwagger) Empty() bool {
	text := reflect.StructTag(swaggerTag).Get("swagger")
	if len(text) == 0 {
		return true
	}
	return false
}

// Default returns the default value of the field.
func (swaggerTag TagSwagger) Default() string {
	text := reflect.StructTag(swaggerTag).Get("swagger")
	texts := strings.Split(text, ",")[1:]

	for _, text := range texts {
		if strings.HasPrefix(text, "d=") {
			return text[2:]
		}
	}

	return ""
}

// In returns the place where the field will be put as parameter.
func (swaggerTag TagSwagger) In() string {
	text := reflect.StructTag(swaggerTag).Get("swagger")
	texts := strings.Split(text, ",")[1:]

	for _, text := range texts {
		if strings.HasPrefix(text, "in=") {
			return text[3:]
		}
	}

	return ""
}

// Required returns whether the field is required.
func (swaggerTag TagSwagger) Required() bool {
	text := reflect.StructTag(swaggerTag).Get("swagger")
	texts := strings.Split(text, ",")[1:]

	for _, text := range texts {
		if text == "req" {
			return true
		}
	}

	return false
}

// Enum returns the values allowed in the field.
func (swaggerTag TagSwagger) Enum() []string {
	text := reflect.StructTag(swaggerTag).Get("swagger")
	texts := strings.Split(text, ",")[1:]

	for _, text := range texts {
		if strings.HasPrefix(text, "enum=") {
			return strings.Split(text[5:], "|")
		}
	}

	return nil
}

// Object is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#swagger-object
type Object struct {
	Swagger             string                 `json:"swagger" swagger:",req,d=2.0"`
	Info                *Info                  `json:"info" swagger:",req"`
	Host                string                 `json:"host,omitempty"`
	BasePath            string                 `json:"basePath,omitempty"`
	Schemes             []string               `json:"schemes,omitempty" swagger:",d=https"`
	Consumes            []string               `json:"consumes,omitempty" swagger:",d=application/json"`
	Produces            []string               `json:"produces,omitempty" swagger:",d=application/json"`
	Paths               Paths                  `json:"paths" swagger:",req"`
	Definitions         Definitions            `json:"definitions,omitempty"`
	Parameters          ParametersDefinitions  `json:"parameters,omitempty"`
	Responses           ResponsesDefinitions   `json:"responses,omitempty"`
	SecurityDefinitions SecurityDefinitions    `json:"securityDefinitions,omitempty"`
	Security            []SecurityRequirement  `json:"security,omitempty"`
	Tags                []*Tag                 `json:"tags,omitempty"`
	ExternalDocs        *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// Info is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#infoObject
type Info struct {
	Title          string   `json:"title" swagger:",req"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Version        string   `json:"version" swagger:",req"`
}

// Paths is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#pathsObject
type Paths map[string]*PathItem

// Contact is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#contact-object
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#license-object
type License struct {
	Name string `json:"name" swagger:",req"`
	URL  string `json:"url,omitempty"`
}

// PathItem is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#pathItemObject
type PathItem struct {
	Ref        string       `json:"$ref,omitempty"`
	Get        *Operation   `json:"get,omitempty"`
	Put        *Operation   `json:"put,omitempty"`
	Post       *Operation   `json:"post,omitempty"`
	Delete     *Operation   `json:"delete,omitempty"`
	Options    *Operation   `json:"options,omitempty"`
	Head       *Operation   `json:"head,omitempty"`
	Patch      *Operation   `json:"patch,omitempty"`
	Parameters []*Parameter `json:"parameters,omitempty"`
}

// Operation is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#operation-object
type Operation struct {
	Tags         []string               `json:"tags,omitempty"`
	Summary      string                 `json:"summary,omitempty"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
	OperationID  string                 `json:"operationId,omitempty"`
	Consumes     []string               `json:"consumes,omitempty"`
	Produces     []string               `json:"produces,omitempty"`
	Parameters   []*Parameter           `json:"parameters,omitempty"`
	Responses    Responses              `json:"responses" swagger:",req"`
	Schemes      []string               `json:"schemes,omitempty"`
	Deprecated   bool                   `json:"deprecated,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty"`
}

// Parameter is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#parameterObject
type Parameter struct {
	Name        string `json:"name" swagger:",req"`
	In          string `json:"in" swagger:",req,enum=query|header|path|formData|body"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`

	// If in is "body":
	Schema *Schema `json:"schema,omitempty"`

	// If in is any value other than "body":
	Type             string        `json:"type,omitempty"`
	Format           string        `json:"format,omitempty"`
	AllowEmptyValue  bool          `json:"allowEmptyValue,omitempty"`
	Items            *Items        `json:"items,omitempty"`
	CollectionFormat string        `json:"collectionFormat,omitempty"`
	Default          interface{}   `json:"default,omitempty"`
	Maximum          *int          `json:"maximum,omitempty"`
	ExclusiveMaximum *bool         `json:"exclusiveMaximum,omitempty"`
	Minimum          *int          `json:"minimum,omitempty"`
	ExclusiveMinimum *bool         `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int          `json:"maxLength,omitempty"`
	MinLength        *int          `json:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty"`
	MaxItems         *int          `json:"maxItems,omitempty"`
	MinItems         *int          `json:"minItems,omitempty"`
	UniqueItems      *bool         `json:"uniqueItems,omitempty"`
	Enum             []interface{} `json:"enum,omitempty"`
	MultipleOf       *int          `json:"multipleOf,omitempty"`
}

// Items is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#items-object
type Items struct {
	Type             string        `json:"type" swagger:",req"`
	Format           string        `json:"format,omitempty"`
	Items            *Items        `json:"items,omitempty"`
	CollectionFormat string        `json:"collectionFormat,omitempty"`
	Default          interface{}   `json:"default,omitempty"`
	Maximum          *int          `json:"maximum,omitempty"`
	ExclusiveMaximum *bool         `json:"exclusiveMaximum,omitempty"`
	Minimum          *int          `json:"minimum,omitempty"`
	ExclusiveMinimum *bool         `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int          `json:"maxLength,omitempty"`
	MinLength        *int          `json:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty"`
	MaxItems         *int          `json:"maxItems,omitempty"`
	MinItems         *int          `json:"minItems,omitempty"`
	UniqueItems      *bool         `json:"uniqueItems,omitempty"`
	Enum             []interface{} `json:"enum,omitempty"`
	MultipleOf       *int          `json:"multipleOf,omitempty"`
}

// Responses is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responsesObject
type Responses map[string]*Response

// Response is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responseObject
type Response struct {
	Description string  `json:"description" swagger:",req"`
	Schema      *Schema `json:"schema,omitempty"`
	Headers     Headers `json:"headers,omitempty"`
	Examples    Example `json:"examples,omitempty"`

	// https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#reference-object
	Ref string `json:"$ref,omitempty"`
}

// Headers is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#headers-object
type Headers map[string]*Header

// Header is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#header-object
type Header struct {
	Description      string        `json:"description,omitempty"`
	Type             string        `json:"type" swagger:",req"`
	Format           string        `json:"format,omitempty"`
	Items            *Items        `json:"items,omitempty"`
	CollectionFormat string        `json:"collectionFormat,omitempty"`
	Default          interface{}   `json:"default,omitempty"`
	Maximum          *int          `json:"maximum,omitempty"`
	ExclusiveMaximum *bool         `json:"exclusiveMaximum,omitempty"`
	Minimum          *int          `json:"minimum,omitempty"`
	ExclusiveMinimum *bool         `json:"exclusiveMinimum,omitempty"`
	MaxLength        *int          `json:"maxLength,omitempty"`
	MinLength        *int          `json:"minLength,omitempty"`
	Pattern          string        `json:"pattern,omitempty"`
	MaxItems         *int          `json:"maxItems,omitempty"`
	MinItems         *int          `json:"minItems,omitempty"`
	UniqueItems      *bool         `json:"uniqueItems,omitempty"`
	Enum             []interface{} `json:"enum,omitempty"`
	MultipleOf       *int          `json:"multipleOf,omitempty"`
}

// Example is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#example-object
type Example map[string]interface{}

// SecurityRequirement is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#security-requirement-object
type SecurityRequirement map[string][]string

// Definitions is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#definitions-object
type Definitions map[string]*Schema

// ParametersDefinitions is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#parameters-definitions-object
type ParametersDefinitions map[string]*Parameter

// ResponsesDefinitions is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responses-definitions-object
type ResponsesDefinitions map[string]*Response

// SecurityDefinitions is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#security-definitions-object
type SecurityDefinitions map[string]*SecurityScheme

// SecurityScheme is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#security-scheme-object
type SecurityScheme struct {
	Type             string `json:"type" swagger:",req,enum=basic|apiKey|oauth2"`
	Description      string `json:"description,omitempty"`
	Name             string `json:"name" swagger:",req"`
	In               string `json:"in,omitempty" swagger:",enum=query|header"`
	Flow             string `json:"flow" swagger:",req"`
	AuthorizationURL string `json:"authorizationUrl" swagger:",req"`
	TokenURL         string `json:"tokenUrl" swagger:",req"`
	Scopes           Scopes `json:"scopes" swagger:",req"`
}

// Scopes is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#scopesObject
type Scopes map[string]string

// Tag is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#tag-object
type Tag struct {
	Name         string                 `json:"name" swagger:",req"`
	Description  string                 `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty"`
}

// ExternalDocumentation is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#externalDocumentationObject
type ExternalDocumentation struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url" swagger:",req"`
}

// Reference is https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#reference-object
type Reference struct {
	Ref string `json:"$ref,omitempty"`
}

func (o *Object) finish() error {
	err := checkSecurityDefinitions(o)
	if err != nil {
		return err
	}
	return checkObject(reflect.ValueOf(o))
}

func checkSecurityDefinitions(o *Object) error {
	checkSecReqs := func(secReqs []SecurityRequirement) error {
		for _, req := range secReqs {
			for name, oauth2ReqScopes := range req {
				if o.SecurityDefinitions == nil {
					return ErrSecurityDefinitionsIsRequired
				}

				scheme, ok := o.SecurityDefinitions[name]
				if !ok {
					return ErrSecurityDefinitionsIsRequired
				}

				if scheme.Type == "oauth2" {
					for _, reqScope := range oauth2ReqScopes {
						_, ok := scheme.Scopes[reqScope]
						if !ok {
							return ErrSecuritySettingsAreWrong
						}
					}
				}
			}
		}

		return nil
	}

	err := checkSecReqs(o.Security)
	if err != nil {
		return err
	}

	for _, pathItem := range o.Paths {
		if pathItem.Get != nil {
			err = checkSecReqs(pathItem.Get.Security)
			if err != nil {
				return err
			}
		}
		if pathItem.Put != nil {
			err = checkSecReqs(pathItem.Put.Security)
			if err != nil {
				return err
			}
		}
		if pathItem.Post != nil {
			err = checkSecReqs(pathItem.Post.Security)
			if err != nil {
				return err
			}
		}
		if pathItem.Delete != nil {
			err = checkSecReqs(pathItem.Delete.Security)
			if err != nil {
				return err
			}
		}
		if pathItem.Options != nil {
			err = checkSecReqs(pathItem.Options.Security)
			if err != nil {
				return err
			}
		}
		if pathItem.Head != nil {
			err = checkSecReqs(pathItem.Head.Security)
			if err != nil {
				return err
			}
		}
		if pathItem.Patch != nil {
			err = checkSecReqs(pathItem.Patch.Security)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func checkObject(refV reflect.Value) error {
	if refV.Kind() == reflect.Ptr {
		refV = refV.Elem()
	}

	var checkNext func(fV reflect.Value) error
	checkNext = func(fV reflect.Value) error {
		if fV.Kind() == reflect.Ptr && fV.IsNil() {
			return nil
		} else if fV.Kind() == reflect.Ptr {
			fV = fV.Elem()
		}

		switch fV.Kind() {
		case reflect.Struct:
			return checkObject(fV)
		case reflect.Slice, reflect.Array:
			for i, vLen := 0, fV.Len(); i < vLen; i++ {
				err := checkNext(fV.Index(i))
				if err != nil {
					return err
				}
			}
		case reflect.Map:
			for _, key := range fV.MapKeys() {
				err := checkNext(fV.MapIndex(key))
				if err != nil {
					return err
				}
			}
		case reflect.Interface:
			// through
		case reflect.String, reflect.Bool:
			// through
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// through
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// through
		case reflect.Float32, reflect.Float64:
			// through
		default:
			fmt.Println(fV)
			return fmt.Errorf("unsupported kind: %s", fV.Kind().String())
		}

		return nil
	}

	for i, numField := 0, refV.NumField(); i < numField; i++ {
		sf := refV.Type().Field(i)
		fV := refV.Field(i)
		swaggerTag := NewTagSwagger(sf.Tag)

		if d := swaggerTag.Default(); d != "" && ucon.IsEmpty(fV) {
			err := ucon.SetValueFromString(fV, d)
			if err != nil {
				return err
			}
		}
		if swaggerTag.Required() && ucon.IsEmpty(fV) {
			return fmt.Errorf("%s is required", sf.Name)
		}
		if enum := swaggerTag.Enum(); len(enum) != 0 && !ucon.IsEmpty(fV) {
			if sf.Type.Kind() != reflect.String {
				return fmt.Errorf("unsupported kind: %s", sf.Type.Kind().String())
			}

			str := fV.String()
			found := false
			for _, c := range enum {
				if c == str {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("invalid value: %s in %s", str, sf.Name)
			}
		}

		err := checkNext(fV)
		if err != nil {
			return err
		}
	}

	return nil
}
