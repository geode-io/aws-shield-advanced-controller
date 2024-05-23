package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSTSClient struct {
	mock.Mock
}

func (m *mockSTSClient) GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sts.GetCallerIdentityOutput), args.Error(1)
}

func TestAWSCache_Init(t *testing.T) {
	mockSTS := new(mockSTSClient)
	cache := &cache{
		sts: mockSTS,
	}

	ctx := context.Background()
	accountID := "123456789012"

	mockSTS.
		On("GetCallerIdentity", ctx, &sts.GetCallerIdentityInput{}, mock.Anything).
		Return(&sts.GetCallerIdentityOutput{Account: &accountID}, nil)

	err := cache.Init(ctx)
	assert.NoError(t, err)
	assert.Equal(t, accountID, cache.AccountId)

	mockSTS.AssertExpectations(t)
}

func TestAWSCache_GetAccountId(t *testing.T) {
	accountID := "123456789012"
	cache := &cache{
		AccountId: accountID,
	}

	assert.Equal(t, accountID, cache.GetAccountId())
}
