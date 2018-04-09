package swagger

import (
	"fmt"

	"net/http"

	"github.com/favclip/ucon"
)

var _ ucon.HTTPErrorResponse = &securityError{}
var _ error = &securityError{}

type securityError struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (ve *securityError) StatusCode() int {
	return ve.Code
}

func (ve *securityError) ErrorMessage() interface{} {
	return ve
}

func (ve *securityError) Error() string {
	return fmt.Sprintf("status code %d: %v", ve.StatusCode(), ve.Message)
}

func newSecurityError(code int, message string) *securityError {
	return &securityError{
		Code:    code,
		Type:    "https://github.com/favclip/ucon#swagger-security",
		Message: message,
	}
}

var (
	// ErrSecurityDefinitionsIsRequired is returned when security definition is missing in object or path items.
	ErrSecurityDefinitionsIsRequired = newSecurityError(http.StatusInternalServerError, "swagger: SecurityDefinitions is required")
	// ErrSecuritySettingsAreWrong is returned when required scope is missing in object or path items.
	ErrSecuritySettingsAreWrong = newSecurityError(http.StatusInternalServerError, "swagger: security settings are wrong")
	// ErrNotImplemented is returned when specified type is not implemented.
	ErrNotImplemented = newSecurityError(http.StatusInternalServerError, "swagger: not implemented")
	// ErrAccessDenied is returned when access user doesn't have a access grant.
	ErrAccessDenied = newSecurityError(http.StatusUnauthorized, "swagger: access denied")
)

// CheckSecurityRequirements about request.
func CheckSecurityRequirements(obj *Object, getScopes func(b *ucon.Bubble) ([]string, error)) ucon.MiddlewareFunc {

	return func(b *ucon.Bubble) error {
		op, ok := b.RequestHandler.Value(swaggerOperationKey{}).(*Operation)
		if !ok {
			return b.Next()
		}

		var secReqs []SecurityRequirement
		if op.Security != nil {
			// If len(op.Security) == 0, It overwrite top-level definition.
			// check by != nil.
			secReqs = op.Security

		} else {
			secReqs = obj.Security
		}

		// check security. It is ok if any one of the security passes.
		passed := false
		for _, req := range secReqs {
		sec_type:
			for name, oauth2ReqScopes := range req {
				if obj.SecurityDefinitions == nil {
					return ErrSecurityDefinitionsIsRequired
				}

				scheme, ok := obj.SecurityDefinitions[name]
				if !ok {
					return ErrSecurityDefinitionsIsRequired
				}

				switch scheme.Type {
				case "oauth2":
					scopes, err := getScopes(b)
					if err != nil {
						return err
					}

					// all scopes are required.
				outer:
					for _, reqScope := range oauth2ReqScopes {
						for _, scope := range scopes {
							if scope == reqScope {
								continue outer
							}
						}

						continue sec_type
					}

					passed = true

				case "basic":
					fallthrough
				case "apiKey":
					fallthrough
				default:
					if len(oauth2ReqScopes) != 0 {
						return ErrSecuritySettingsAreWrong
					}

					return ErrNotImplemented
				}
			}
		}
		if passed {
			return b.Next()
		}

		return ErrAccessDenied
	}
}
