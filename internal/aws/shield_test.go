package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/aws/aws-sdk-go-v2/service/shield/types"
)

type mockShieldClient struct {
	mock.Mock
}

func (m *mockShieldClient) ListProtections(ctx context.Context, input *shield.ListProtectionsInput, opts ...func(*shield.Options)) (*shield.ListProtectionsOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*shield.ListProtectionsOutput), args.Error(1)
}

func (m *mockShieldClient) DescribeProtection(ctx context.Context, input *shield.DescribeProtectionInput, opts ...func(*shield.Options)) (*shield.DescribeProtectionOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*shield.DescribeProtectionOutput), args.Error(1)
}

func (m *mockShieldClient) CreateProtection(ctx context.Context, input *shield.CreateProtectionInput, opts ...func(*shield.Options)) (*shield.CreateProtectionOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*shield.CreateProtectionOutput), args.Error(1)
}

func (m *mockShieldClient) DeleteProtection(ctx context.Context, input *shield.DeleteProtectionInput, opts ...func(*shield.Options)) (*shield.DeleteProtectionOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*shield.DeleteProtectionOutput), args.Error(1)
}

func (m *mockShieldClient) ListTagsForResource(ctx context.Context, input *shield.ListTagsForResourceInput, opts ...func(*shield.Options)) (*shield.ListTagsForResourceOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*shield.ListTagsForResourceOutput), args.Error(1)
}

func (m *mockShieldClient) TagResource(ctx context.Context, input *shield.TagResourceInput, opts ...func(*shield.Options)) (*shield.TagResourceOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*shield.TagResourceOutput), args.Error(1)
}

type mockAWSCache struct {
	mock.Mock
}

func (m *mockAWSCache) Init(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockAWSCache) GetAccountId() string {
	args := m.Called()
	return args.String(0)
}

func TestAWSShieldManager_CreateProtection(t *testing.T) {
	mockClient := new(mockShieldClient)
	mockCache := new(mockAWSCache)
	manager := &shieldManager{client: mockClient, cache: mockCache}
	ctx := context.Background()

	resourceArn := "arn:aws:test:us-east-1:123456789012:test/test"
	protectionArn := "arn:aws:shield::123456789012:protection/abc123"

	mockCache.
		On("GetAccountId").
		Return("123456789012").
		Once()

	mockClient.
		On("DescribeProtection", ctx, &shield.DescribeProtectionInput{ResourceArn: aws.String(resourceArn)}, mock.Anything).
		Return(&shield.DescribeProtectionOutput{}, &types.ResourceNotFoundException{}).
		Once()
	mockClient.
		On("CreateProtection", ctx, &shield.CreateProtectionInput{
			Name:        aws.String("my-protection"),
			ResourceArn: aws.String(resourceArn),
		}, mock.Anything).
		Return(&shield.CreateProtectionOutput{
			ProtectionId: aws.String("abc123"),
		}, nil).
		Once()
	mockClient.
		On("TagResource", ctx, &shield.TagResourceInput{
			ResourceARN: aws.String(protectionArn),
			Tags: []types.Tag{
				{Key: aws.String(OwnerTagKey), Value: aws.String(OwnerTagValue)},
			},
		}, mock.Anything).
		Return(&shield.TagResourceOutput{}, nil).
		Once()
	mockClient.
		On("DescribeProtection", ctx, &shield.DescribeProtectionInput{ResourceArn: aws.String(resourceArn)}, mock.Anything).
		Return(&shield.DescribeProtectionOutput{Protection: &types.Protection{ProtectionArn: aws.String(protectionArn)}}, nil).
		Once()

	arn, err := manager.CreateOrUpdateProtection(ctx, "my-protection", resourceArn)
	assert.NoError(t, err)
	assert.Equal(t, protectionArn, arn)

	mockClient.AssertExpectations(t)
}

func TestAWSShieldManager_UpdateProtection(t *testing.T) {
	mockClient := new(mockShieldClient)
	manager := &shieldManager{client: mockClient}
	ctx := context.Background()

	resourceArn := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-load-balancer/50dc6c495c0c9188"
	protectionArn := "arn:aws:shield::123456789012:protection/a1b2c3d4-5678-90ab-cdef-EXAMPLE11111"

	mockClient.
		On("DescribeProtection", ctx, &shield.DescribeProtectionInput{ResourceArn: aws.String(resourceArn)}, mock.Anything).
		Return(&shield.DescribeProtectionOutput{Protection: &types.Protection{ProtectionArn: aws.String(protectionArn)}}, nil).
		Once()

	arn, err := manager.CreateOrUpdateProtection(ctx, "my-protection", resourceArn)

	assert.NoError(t, err)
	assert.Equal(t, protectionArn, arn)

	mockClient.AssertExpectations(t)
}

func TestAWSShieldManager_DeleteProtection(t *testing.T) {
	mockClient := new(mockShieldClient)
	manager := &shieldManager{client: mockClient}
	ctx := context.Background()

	protectionArn := "arn:aws:shield::123456789012:protection/abc123"
	protectionId := "abc123"

	mockClient.
		On("DeleteProtection", ctx, &shield.DeleteProtectionInput{ProtectionId: aws.String(protectionId)}, mock.Anything).
		Return(&shield.DeleteProtectionOutput{}, nil).
		Once()

	err := manager.DeleteProtection(ctx, protectionArn)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
}
func TestAWSShieldManager_ProtectionArnToId(t *testing.T) {
	manager := &shieldManager{}

	tests := []struct {
		name          string
		protectionArn string
		expectedId    string
		expectError   bool
	}{
		{
			name:          "Valid ARN",
			protectionArn: "arn:aws:shield::123456789012:protection/a1b2c3d4-5678-90ab-cdef-EXAMPLE11111",
			expectedId:    "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111",
			expectError:   false,
		},
		{
			name:          "Invalid ARN format",
			protectionArn: "invalid-arn",
			expectedId:    "",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := manager.protectionArnToId(tt.protectionArn)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedId, id)
			}
		})
	}
}

func TestAWSShieldManager_ProtectionIdToArn(t *testing.T) {
	accountID := "123456789012"
	manager := &shieldManager{
		cache: &cache{AccountId: accountID},
	}

	tests := []struct {
		name         string
		protectionId string
		expectedArn  string
	}{
		{
			name:         "Valid protection ID",
			protectionId: "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111",
			expectedArn:  "arn:aws:shield::123456789012:protection/a1b2c3d4-5678-90ab-cdef-EXAMPLE11111",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arn := manager.protectionIdToArn(tt.protectionId)
			assert.Equal(t, tt.expectedArn, arn)
		})
	}
}
