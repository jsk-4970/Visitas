package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// FirebaseClient wraps Firebase Auth client
type FirebaseClient struct {
	authClient *auth.Client
	app        *firebase.App
}

// NewFirebaseClient initializes a new Firebase Admin SDK client
func NewFirebaseClient(ctx context.Context, credentialsPath string) (*FirebaseClient, error) {
	if credentialsPath == "" {
		return nil, fmt.Errorf("firebase credentials path is required")
	}

	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase Auth client: %w", err)
	}

	return &FirebaseClient{
		authClient: authClient,
		app:        app,
	}, nil
}

// VerifyIDToken verifies the Firebase ID token and returns the token claims
func (fc *FirebaseClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := fc.authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify ID token: %w", err)
	}
	return token, nil
}

// GetUser retrieves user information by UID
func (fc *FirebaseClient) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	user, err := fc.authClient.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// CreateUser creates a new user with email and password
func (fc *FirebaseClient) CreateUser(ctx context.Context, email, password string) (*auth.UserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email(email).
		Password(password).
		EmailVerified(false)

	user, err := fc.authClient.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing user
func (fc *FirebaseClient) UpdateUser(ctx context.Context, uid string, params *auth.UserToUpdate) (*auth.UserRecord, error) {
	user, err := fc.authClient.UpdateUser(ctx, uid, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// DeleteUser deletes a user by UID
func (fc *FirebaseClient) DeleteUser(ctx context.Context, uid string) error {
	if err := fc.authClient.DeleteUser(ctx, uid); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// SetCustomUserClaims sets custom claims for a user (e.g., roles, permissions)
func (fc *FirebaseClient) SetCustomUserClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	if err := fc.authClient.SetCustomUserClaims(ctx, uid, claims); err != nil {
		return fmt.Errorf("failed to set custom user claims: %w", err)
	}
	return nil
}

// Close closes the Firebase client (currently a no-op, but included for future compatibility)
func (fc *FirebaseClient) Close() error {
	// Firebase Admin SDK doesn't require explicit closing
	return nil
}
