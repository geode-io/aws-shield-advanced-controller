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

func TestCreateProtection(t *testing.T) {
	mockClient := new(mockShieldClient)
	manager := &shieldManager{client: mockClient}
	ctx := context.Background()

	resourceArn := "arn:aws:test:us-east-1:123456789012:test/test"
	protectionArn := "arn:aws:shield::123456789012:protection/abc123"

	mockClient.
		On("DescribeProtection", ctx, &shield.DescribeProtectionInput{ResourceArn: aws.String(resourceArn)}, mock.Anything).
		Return(&shield.DescribeProtectionOutput{}, &types.ResourceNotFoundException{}).
		Once()
	mockClient.
		On("CreateProtection", ctx, &shield.CreateProtectionInput{
			Name:        aws.String("my-protection"),
			ResourceArn: aws.String(resourceArn),
		}, mock.Anything).
		Return(&shield.CreateProtectionOutput{}, nil).
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

func TestUpdateProtection(t *testing.T) {
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

func TestDeleteProtection(t *testing.T) {
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
