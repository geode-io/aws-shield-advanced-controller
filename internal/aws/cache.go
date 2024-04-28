package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Cache interface {
	Init(ctx context.Context) error
	GetAccountId() string
}

type STSClient interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

type cache struct {
	sts STSClient

	AccountId string
}

func NewCache(cfg aws.Config) Cache {
	stsClient := sts.NewFromConfig(cfg)
	return &cache{
		sts: stsClient,
	}
}

func (c *cache) Init(ctx context.Context) error {
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

func (c *cache) GetAccountId() string {
	return c.AccountId
}
