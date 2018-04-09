package backend

import (
	"context"
	"net/http"

	"google.golang.org/appengine/user"
	"google.golang.org/appengine/log"
	"github.com/favclip/ucon/swagger"
	"github.com/favclip/ucon"
)

func setupSecretAPI(swPlugin *swagger.Plugin) {
	api := &SecretAPI{}
	tag := swPlugin.AddTag(&swagger.Tag{Name: "Secret", Description: "Secret API list"})
	var hInfo *swagger.HandlerInfo

	hInfo = swagger.NewHandlerInfo(api.Post)
	ucon.Handle(http.MethodPost, "/api/1/secret", hInfo)
	hInfo.Description, hInfo.Tags = "post to secret", []string{tag.Name}
}

// SecretAPI is API to register and acquire Secret
type SecretAPI struct{}

// Post is Secret registration handler
func (api *SecretAPI) Post(ctx context.Context) (error) {
	u := user.Current(ctx)
	if u == nil {
		return &HTTPError{Code: http.StatusForbidden, Message:"You do not have permission."}
	}
	log.Infof(ctx, "%+v", u)

	return nil
}