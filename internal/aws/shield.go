package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/aws/aws-sdk-go-v2/service/shield/types"
)

type ShieldClient interface {
	DescribeProtection(ctx context.Context, input *shield.DescribeProtectionInput, opts ...func(*shield.Options)) (*shield.DescribeProtectionOutput, error)
	CreateProtection(ctx context.Context, input *shield.CreateProtectionInput, opts ...func(*shield.Options)) (*shield.CreateProtectionOutput, error)
	DeleteProtection(ctx context.Context, input *shield.DeleteProtectionInput, opts ...func(*shield.Options)) (*shield.DeleteProtectionOutput, error)
}

type ShieldManager interface {
	CreateOrUpdateProtection(ctx context.Context, name, resourceArn string) (string, error)
	DeleteProtection(ctx context.Context, protectionArn string) error
}

type shieldManager struct {
	client ShieldClient
}

var _ ShieldManager = &shieldManager{}

func NewShieldManager(cfg aws.Config) ShieldManager {
	client := shield.NewFromConfig(cfg)

	return &shieldManager{
		client: client,
	}
}

func (m *shieldManager) CreateOrUpdateProtection(ctx context.Context, name, resourceArn string) (string, error) {
	log := log.FromContext(ctx)

	// Check if the resource protection already exists
	existing, err := m.client.DescribeProtection(ctx, &shield.DescribeProtectionInput{
		ResourceArn: aws.String(resourceArn),
	})

	// Protection already exists, update it if needed
	if err == nil {
		log.Info("Updating existing AWS Shield Advanced protection", "name", name, "resourceArn", resourceArn)

		// TODO: manage tags
		return *existing.Protection.ProtectionArn, nil
	}

	var notFoundErr *types.ResourceNotFoundException
	if errors.As(err, &notFoundErr) {
		// Protection doesn't exist, create it
		log.Info("Creating new AWS Shield Advanced protection", "name", name, "resourceArn", resourceArn)
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
	log := log.FromContext(ctx)

	// Parse the ARN to get the protection id
	parsed, err := arn.Parse(protectionArn)
	if err != nil {
		return err
	}
	resource := strings.Split(parsed.Resource, "/")[1]

	log.Info("Deleting AWS Shield Advanced protection", "protectionArn", protectionArn)
	_, err = m.client.DeleteProtection(ctx, &shield.DeleteProtectionInput{
		ProtectionId: aws.String(resource),
	})
	if err != nil {
		return err
	}

	return nil
}
