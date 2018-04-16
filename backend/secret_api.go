package backend

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/favclip/ucon"
	"github.com/favclip/ucon/swagger"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/user"
)

func setupSecretAPI(swPlugin *swagger.Plugin) {
	api := &SecretAPI{}
	tag := swPlugin.AddTag(&swagger.Tag{Name: "Secret", Description: "Secret API list"})
	var hInfo *swagger.HandlerInfo

	hInfo = swagger.NewHandlerInfo(api.Post)
	ucon.Handle(http.MethodPost, "/api/1/secret", hInfo)
	hInfo.Description, hInfo.Tags = "post to secret", []string{tag.Name}

	hInfo = swagger.NewHandlerInfo(api.Get)
	ucon.Handle(http.MethodGet, "/api/1/secret/{key}", hInfo)
	hInfo.Description, hInfo.Tags = "get from secret", []string{tag.Name}
}

// LogEntry is Output Request Log
type LogEntry struct {
	User string `json:"user"`
}

// Secret is Datastore Entity
type Secret struct {
	Value string `datastore:",noindex"`
}

// SecretAPI is API to register and acquire Secret
type SecretAPI struct{}

// SecretAPIPostRequest is SecretAPI Post Request
type SecretAPIPostRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Post is Secret registration handler
func (api *SecretAPI) Post(ctx context.Context, form *SecretAPIPostRequest, r *http.Request) error {
	le := &LogEntry{}
	defer outputRequestLog(ctx, le)

	u := user.Current(ctx)
	if u == nil {
		return &HTTPError{Code: http.StatusForbidden, Message: "You do not have permission."}
	}
	le.User = u.Email

	ds, err := FromContext(ctx)
	if err != nil {
		return err
	}

	kms, err := NewKMSService(ctx)
	if err != nil {
		return err
	}
	appID := appengine.AppID(ctx)
	ev, _, err := kms.Encrypt(CryptKey{
		ProjectID:  appID,
		LocationID: "global",
		KeyRingID:  "testkey",
		KeyName:    "testCryptKey",
	}, form.Value)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		return err
	}

	k := ds.NameKey("Secret", form.Key, nil)
	s := &Secret{
		Value: ev,
	}
	_, err = ds.Put(ctx, k, s)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		return err
	}

	return nil
}

// SecretAPIGetRequest is SecretAPI Get Request
type SecretAPIGetRequest struct {
	Key string `json:"key" swagger:",in=query"`
}

// SecretAPIGetResponse is SecretAPI Get Response
type SecretAPIGetResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Get is Secret registration handler
func (api *SecretAPI) Get(ctx context.Context, form *SecretAPIGetRequest, r *http.Request) (*SecretAPIGetResponse, error) {
	le := &LogEntry{}
	defer outputRequestLog(ctx, le)

	u := user.Current(ctx)
	if u == nil {
		return nil, &HTTPError{Code: http.StatusForbidden, Message: "You do not have permission."}
	}
	le.User = u.Email

	ds, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}

	k := ds.NameKey("Secret", form.Key, nil)
	s := &Secret{}
	if err := ds.Get(ctx, k, s); err != nil {
		log.Errorf(ctx, "%+v", err)
		return nil, err
	}

	kms, err := NewKMSService(ctx)
	if err != nil {
		return nil, err
	}
	appID := appengine.AppID(ctx)
	pt, err := kms.Decrypt(CryptKey{
		ProjectID:  appID,
		LocationID: "global",
		KeyRingID:  "testkey",
		KeyName:    "testCryptKey",
	}, s.Value)
	if err != nil {
		log.Errorf(ctx, "%+v", err)
		return nil, err
	}

	return &SecretAPIGetResponse{
		Key:   form.Key,
		Value: pt,
	}, nil
}

func outputRequestLog(ctx context.Context, e *LogEntry) {
	j, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	log.Infof(ctx, "LogEntry=%s", j)
}
