package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
)

type CloudFrontClient interface {
	ListDistributions(ctx context.Context, params *cloudfront.ListDistributionsInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error)
}

type cloudfrontDiscoveryProvider struct {
	client CloudFrontClient
}

var _ DiscoveryProvider = &cloudfrontDiscoveryProvider{}

func NewCloudfrontDiscoveryProvider(cfg aws.Config) DiscoveryProvider {
	return &cloudfrontDiscoveryProvider{
		client: cloudfront.NewFromConfig(cfg),
	}
}

func (p *cloudfrontDiscoveryProvider) Discover(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResponse, error) {
	resources := []DiscoveredResource{}

	paginator := cloudfront.NewListDistributionsPaginator(p.client, &cloudfront.ListDistributionsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing CloudFront distributions: %v", err)
		}

		for _, dist := range page.DistributionList.Items {
			resources = append(resources, DiscoveredResource{
				Type: "cloudfront/distribution",
				Arn:  *dist.ARN,
				Name: *dist.Id,
			})
		}
	}

	return &DiscoveryResponse{
		Resources: resources,
	}, nil
}
