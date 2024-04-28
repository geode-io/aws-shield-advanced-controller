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

var (
	OwnerTagKey   = "shield.aws.geode.io/owner"
	OwnerTagValue = "aws-shield-advanced-controller"
)

type ShieldClient interface {
	ListProtections(ctx context.Context, input *shield.ListProtectionsInput, opts ...func(*shield.Options)) (*shield.ListProtectionsOutput, error)
	DescribeProtection(ctx context.Context, input *shield.DescribeProtectionInput, opts ...func(*shield.Options)) (*shield.DescribeProtectionOutput, error)
	CreateProtection(ctx context.Context, input *shield.CreateProtectionInput, opts ...func(*shield.Options)) (*shield.CreateProtectionOutput, error)
	DeleteProtection(ctx context.Context, input *shield.DeleteProtectionInput, opts ...func(*shield.Options)) (*shield.DeleteProtectionOutput, error)
	ListTagsForResource(ctx context.Context, input *shield.ListTagsForResourceInput, opts ...func(*shield.Options)) (*shield.ListTagsForResourceOutput, error)
	TagResource(ctx context.Context, input *shield.TagResourceInput, opts ...func(*shield.Options)) (*shield.TagResourceOutput, error)
}

type ShieldManager interface {
	ListOwnedProtections(ctx context.Context) ([]types.Protection, error)
	CreateOrUpdateProtection(ctx context.Context, name, resourceArn string) (string, error)
	DeleteProtection(ctx context.Context, protectionArn string) error
}

type shieldManager struct {
	client ShieldClient
	cache  Cache
}

var _ ShieldManager = &shieldManager{}

func NewShieldManager(cfg aws.Config, cache Cache) ShieldManager {
	return &shieldManager{
		client: shield.NewFromConfig(cfg),
		cache:  cache,
	}
}

func (m *shieldManager) ListOwnedProtections(ctx context.Context) ([]types.Protection, error) {
	log := log.FromContext(ctx)

	var protections []types.Protection

	// List all existing protections
	paginator := shield.NewListProtectionsPaginator(m.client, &shield.ListProtectionsInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list protections: %w", err)
		}

		for _, protection := range output.Protections {
			// Check if the protection is managed by the controller
			tags, err := m.client.ListTagsForResource(ctx, &shield.ListTagsForResourceInput{
				ResourceARN: protection.ProtectionArn,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to list tags for protection: %w", err)
			}

			var isManaged bool
			for _, tag := range tags.Tags {
				if aws.ToString(tag.Key) == OwnerTagKey && aws.ToString(tag.Value) == OwnerTagValue {
					isManaged = true
					break
				}
			}

			if isManaged {
				protections = append(protections, protection)
			}
		}
	}

	log.V(1).Info("Found existing AWS Shield Advanced protections", "count", len(protections))

	return protections, nil
}

func (m *shieldManager) CreateOrUpdateProtection(ctx context.Context, name, resourceArn string) (string, error) {
	log := log.FromContext(ctx)

	// Check if the resource protection already exists
	existing, err := m.client.DescribeProtection(ctx, &shield.DescribeProtectionInput{
		ResourceArn: aws.String(resourceArn),
	})

	// Protection already exists, update it if needed
	if err == nil {
		log.V(1).Info("Syncing existing AWS Shield Advanced protection", "name", name, "resourceArn", resourceArn)

		// Nothing to do here for now

		return *existing.Protection.ProtectionArn, nil
	}

	var notFoundErr *types.ResourceNotFoundException
	if errors.As(err, &notFoundErr) {
		// Protection doesn't exist, create it
		log.Info("Creating new AWS Shield Advanced protection", "name", name, "resourceArn", resourceArn)
		protection, err := m.client.CreateProtection(ctx, &shield.CreateProtectionInput{
			Name:        aws.String(name),
			ResourceArn: aws.String(resourceArn),
		})
		if err != nil {
			return "", err
		}

		// Tag with owner info
		_, err = m.client.TagResource(ctx, &shield.TagResourceInput{
			ResourceARN: aws.String(m.protectionIdToArn(*protection.ProtectionId)),
			Tags: []types.Tag{
				{Key: aws.String(OwnerTagKey), Value: aws.String(OwnerTagValue)},
			},
		})
		if err != nil {
			return "", fmt.Errorf("failed to tag protection: %w", err)
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

	log.Info("Deleting AWS Shield Advanced protection", "protectionArn", protectionArn)

	id, err := m.protectionArnToId(protectionArn)
	if err != nil {
		return err
	}

	_, err = m.client.DeleteProtection(ctx, &shield.DeleteProtectionInput{
		ProtectionId: aws.String(id),
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *shieldManager) protectionArnToId(protectionArn string) (string, error) {
	parsed, err := arn.Parse(protectionArn)
	if err != nil {
		return "", err
	}

	parts := strings.Split(parsed.Resource, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid protection ARN: %s", protectionArn)
	}

	return parts[1], nil
}

func (m *shieldManager) protectionIdToArn(protectionId string) string {
	return fmt.Sprintf("arn:aws:shield::%s:protection/%s", m.cache.GetAccountId(), protectionId)
}
