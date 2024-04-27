package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/aws/aws-sdk-go-v2/service/shield/types"
)

type ShieldManager interface {
	CreateOrUpdateProtection(ctx context.Context, name, arn string) (string, error)
	DeleteProtection(ctx context.Context, resourceArn string) error
}

type shieldManager struct {
	client *shield.Client
}

var _ ShieldManager = &shieldManager{}

func NewShieldManager() ShieldManager {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatal(err)
	}

	client := shield.NewFromConfig(cfg)

	return &shieldManager{
		client: client,
	}
}

func (m *shieldManager) CreateOrUpdateProtection(ctx context.Context, name, resourceArn string) (string, error) {

	// Check if the resource protection already exists
	existing, err := m.client.DescribeProtection(ctx, &shield.DescribeProtectionInput{
		ResourceArn: aws.String(resourceArn),
	})

	// Protection already exists, update it if needed
	if err == nil {
		// TODO: manage tags
		return *existing.Protection.ProtectionArn, nil
	}

	var notFoundErr *types.ResourceNotFoundException
	if errors.As(err, &notFoundErr) {
		// Protection doesn't exist, create it
		_, err := m.client.CreateProtection(ctx, &shield.CreateProtectionInput{
			Name:        aws.String(name),
			ResourceArn: aws.String(resourceArn),
		})
		if err != nil {
			return "", err
		}
	} else {
		// An error occurred while checking if the protection exists
		return "", err
	}

	// Get the protection again so we can get its ARN
	protection, err := m.client.DescribeProtection(ctx, &shield.DescribeProtectionInput{
		ResourceArn: aws.String(resourceArn),
	})
	if err != nil {
		return "", fmt.Errorf("failed to describe protection after creation: %w", err)
	}

	return *protection.Protection.ProtectionArn, nil
}

func (m *shieldManager) DeleteProtection(ctx context.Context, protectionArn string) error {
	// Parse the ARN to get the protection id
	parsed, err := arn.Parse(protectionArn)
	if err != nil {
		return err
	}

	resource := strings.Split(parsed.Resource, "/")[1]

	_, err = m.client.DeleteProtection(ctx, &shield.DeleteProtectionInput{
		ProtectionId: aws.String(resource),
	})
	if err != nil {
		return err
	}

	return nil
}
