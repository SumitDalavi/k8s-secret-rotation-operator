package aws

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// RotateSecret triggers AWS Secrets Manager to rotate a secret.
// Uses the AWS CLI (aws secretsmanager rotate-secret) as the simplest integration approach.
func RotateSecret(ctx context.Context, secretARN string) error {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" {
		return fmt.Errorf("AWS credentials not configured (AWS_ACCESS_KEY_ID not set)")
	}
	cmd := exec.CommandContext(ctx, "aws", "secretsmanager", "rotate-secret",
		"--secret-id", secretARN,
		"--rotate-immediately",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("aws rotate-secret failed: %w: %s", err, string(out))
	}
	return nil
}

// GetSecretValue retrieves the current value of an AWS Secrets Manager secret.
func GetSecretValue(ctx context.Context, secretARN string) (string, error) {
	cmd := exec.CommandContext(ctx, "aws", "secretsmanager", "get-secret-value",
		"--secret-id", secretARN,
		"--query", "SecretString",
		"--output", "text",
	)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("aws get-secret-value: %w", err)
	}
	return string(out), nil
}
