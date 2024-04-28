package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type AWSCache interface {
	Init(ctx context.Context) error
	GetAccountId() string
}

type STSClient interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type awsCache struct {
	sts STSClient

	AccountId string
}

func NewAWSCache(cfg aws.Config) AWSCache {
	stsClient := sts.NewFromConfig(cfg)
	return &awsCache{
		sts: stsClient,
	}
}

func (c *awsCache) Init(ctx context.Context) error {
	// init account id
	if c.AccountId == "" {
		output, err := c.sts.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
		if err != nil {
			return fmt.Errorf("error caching account ID: %v", err)
		}
		c.AccountId = *output.Account
	}

	return nil
}

func (c *awsCache) GetAccountId() string {
	return c.AccountId
}
