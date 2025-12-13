package encryption

import (
	"context"
	"encoding/base64"
	"fmt"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
)

// KMSEncryptor handles encryption/decryption using Google Cloud KMS
type KMSEncryptor struct {
	client     *kms.KeyManagementClient
	keyName    string
	projectID  string
	locationID string
	keyRingID  string
	keyID      string
}

// NewKMSEncryptor creates a new KMS encryptor
func NewKMSEncryptor(ctx context.Context, projectID, locationID, keyRingID, keyID string) (*KMSEncryptor, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}

	keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		projectID, locationID, keyRingID, keyID)

	return &KMSEncryptor{
		client:     client,
		keyName:    keyName,
		projectID:  projectID,
		locationID: locationID,
		keyRingID:  keyRingID,
		keyID:      keyID,
	}, nil
}

// Close closes the KMS client
func (e *KMSEncryptor) Close() error {
	return e.client.Close()
}

// EncryptMyNumber encrypts a My Number using KMS AEAD
// Returns base64-encoded ciphertext
// Deprecated: Use EncryptMyNumberWithPatient for better security
func (e *KMSEncryptor) EncryptMyNumber(ctx context.Context, plaintext string) (string, error) {
	return e.EncryptMyNumberWithPatient(ctx, plaintext, "")
}

// EncryptMyNumberWithPatient encrypts a My Number using KMS AEAD with patient-specific AAD
// Returns base64-encoded ciphertext
func (e *KMSEncryptor) EncryptMyNumberWithPatient(ctx context.Context, plaintext, patientID string) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("plaintext cannot be empty")
	}

	// Additional Authenticated Data (AAD) for My Number
	// Include patient ID for stronger security binding
	aad := buildMyNumberAAD(patientID)

	req := &kmspb.EncryptRequest{
		Name:                        e.keyName,
		Plaintext:                   []byte(plaintext),
		AdditionalAuthenticatedData: aad,
	}

	result, err := e.client.Encrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	// Encode ciphertext to base64 for storage
	ciphertext := base64.StdEncoding.EncodeToString(result.Ciphertext)

	return ciphertext, nil
}

// DecryptMyNumber decrypts a My Number using KMS AEAD
// Expects base64-encoded ciphertext
// Deprecated: Use DecryptMyNumberWithPatient for better security
func (e *KMSEncryptor) DecryptMyNumber(ctx context.Context, ciphertextBase64 string) (string, error) {
	return e.DecryptMyNumberWithPatient(ctx, ciphertextBase64, "")
}

// DecryptMyNumberWithPatient decrypts a My Number using KMS AEAD with patient-specific AAD
// Expects base64-encoded ciphertext
func (e *KMSEncryptor) DecryptMyNumberWithPatient(ctx context.Context, ciphertextBase64, patientID string) (string, error) {
	if ciphertextBase64 == "" {
		return "", fmt.Errorf("ciphertext cannot be empty")
	}

	// Decode base64 ciphertext
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	// Additional Authenticated Data (AAD) must match encryption
	aad := buildMyNumberAAD(patientID)

	req := &kmspb.DecryptRequest{
		Name:                        e.keyName,
		Ciphertext:                  ciphertext,
		AdditionalAuthenticatedData: aad,
	}

	result, err := e.client.Decrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(result.Plaintext), nil
}

// buildMyNumberAAD builds the Additional Authenticated Data for My Number encryption
// This binds the ciphertext to the patient, preventing copy attacks
func buildMyNumberAAD(patientID string) []byte {
	if patientID == "" {
		// Fallback for backward compatibility
		return []byte("mynumber")
	}
	// Format: "mynumber:patient_id:{uuid}"
	return []byte("mynumber:patient_id:" + patientID)
}

// Encrypt encrypts data with optional AAD
// Returns base64-encoded ciphertext
func (e *KMSEncryptor) Encrypt(ctx context.Context, plaintext string, aad []byte) (string, error) {
	if plaintext == "" {
		return "", fmt.Errorf("plaintext cannot be empty")
	}

	req := &kmspb.EncryptRequest{
		Name:                        e.keyName,
		Plaintext:                   []byte(plaintext),
		AdditionalAuthenticatedData: aad,
	}

	result, err := e.client.Encrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt: %w", err)
	}

	ciphertext := base64.StdEncoding.EncodeToString(result.Ciphertext)
	return ciphertext, nil
}

// Decrypt decrypts data with optional AAD
// Expects base64-encoded ciphertext
func (e *KMSEncryptor) Decrypt(ctx context.Context, ciphertextBase64 string, aad []byte) (string, error) {
	if ciphertextBase64 == "" {
		return "", fmt.Errorf("ciphertext cannot be empty")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 ciphertext: %w", err)
	}

	req := &kmspb.DecryptRequest{
		Name:                        e.keyName,
		Ciphertext:                  ciphertext,
		AdditionalAuthenticatedData: aad,
	}

	result, err := e.client.Decrypt(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(result.Plaintext), nil
}
