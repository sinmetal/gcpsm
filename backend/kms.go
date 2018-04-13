package backend

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// KMSService is KMS Serviceを提供するstruct
type KMSService struct {
	S *cloudkms.Service
}

// NewKMSService is KMS Serviceを作成
func NewKMSService(ctx context.Context) (*KMSService, error) {
	client, err := google.DefaultClient(ctx, cloudkms.CloudPlatformScope)
	if err != nil {
		return nil, errors.Wrap(err, "failed create google.DefaultClient: ")
	}

	// Create the KMS client.
	kmsService, err := cloudkms.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "failed cloudkms.New: ")
	}

	return &KMSService{
		S: kmsService,
	}, nil
}

// CryptKey is Cloud KMSのCryptKey Resourceの情報を保持
type CryptKey struct {
	ProjectID  string
	LocationID string
	KeyRingID  string
	KeyName    string
}

// Name is API実行時のCryptKey Resource文字列を返す
func (cryptKey *CryptKey) Name() string {
	return fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", cryptKey.ProjectID, cryptKey.LocationID, cryptKey.KeyRingID, cryptKey.KeyName)
}

// Encrypt is Cloud KMSでEncryptを行う
func (service *KMSService) Encrypt(cryptKey CryptKey, plaintext string) (ciphertext string, cryptoKey string, err error) {
	response, err := service.S.Projects.Locations.KeyRings.CryptoKeys.Encrypt(cryptKey.Name(), &cloudkms.EncryptRequest{
		Plaintext: base64.StdEncoding.EncodeToString([]byte(plaintext)),
	}).Do()
	if err != nil {
		return "", "", errors.Wrapf(err, "encrypt: failed to encrypt. CryptoKey=%s", cryptKey.Name())
	}

	return response.Ciphertext, response.Name, nil
}

// Decrypt is Cloud KMSでEncryptされた文字列をDecryptする
func (service *KMSService) Decrypt(cryptKey CryptKey, ciphertext string) (plaintext string, err error) {
	response, err := service.S.Projects.Locations.KeyRings.CryptoKeys.Decrypt(cryptKey.Name(), &cloudkms.DecryptRequest{
		Ciphertext: ciphertext,
	}).Do()
	if err != nil {
		return "", errors.Wrapf(err, "decrypt: failed to decrypt. CryptoKey=%s", cryptKey.Name())
	}

	t, err := base64.StdEncoding.DecodeString(response.Plaintext)
	if err != nil {
		return "", errors.Wrap(err, "decrypt: failed base64 decode")
	}
	return string(t), nil
}
