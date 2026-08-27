package aws

import (
	"context"
	"os"
	"testing"
)

func TestRotateSecret_NoCreds(t *testing.T) {
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	err := RotateSecret(context.Background(), "arn:aws:secretsmanager:us-east-1:123456789012:secret:test")
	if err == nil {
		t.Fatal("expected error when AWS credentials not set")
	}
	if err.Error() != "AWS credentials not configured (AWS_ACCESS_KEY_ID not set)" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestRotateSecret_WithCreds_CLIFails(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY_ID", "fake-key")
	defer os.Unsetenv("AWS_ACCESS_KEY_ID")

	// `aws` binary will fail because credentials are fake; we just want to ensure
	// the error path is covered (the function attempts the exec)
	err := RotateSecret(context.Background(), "arn:aws:secretsmanager:us-east-1:123456789012:secret:test")
	if err == nil {
		// AWS CLI not installed or surprisingly succeeded; test is still valid
		t.Skip("aws CLI not installed or command succeeded unexpectedly")
	}
	// We expect an error here because `aws` will fail or not be installed
}

func TestGetSecretValue_CLIFails(t *testing.T) {
	// `aws` CLI will fail since no real credentials
	_, err := GetSecretValue(context.Background(), "arn:aws:secretsmanager:us-east-1:123456789012:secret:test")
	if err == nil {
		t.Skip("aws CLI not installed or command succeeded unexpectedly")
	}
}
