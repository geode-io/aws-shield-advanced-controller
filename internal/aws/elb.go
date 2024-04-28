package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
)

type ELBClient interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancing.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error)
}

type elbDiscoveryProvider struct {
	client ELBClient
	cache  AWSCache
}

var _ DiscoveryProvider = &elbDiscoveryProvider{}

func NewELBDiscoveryProvider(cfg aws.Config, cache AWSCache) DiscoveryProvider {
	return &elbDiscoveryProvider{
		client: elasticloadbalancing.NewFromConfig(cfg),
		cache:  cache,
	}
}

func (p *elbDiscoveryProvider) Discover(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResponse, error) {
	resources := []DiscoveredResource{}

	for _, region := range request.Regions {
		paginator := elasticloadbalancing.NewDescribeLoadBalancersPaginator(p.client, &elasticloadbalancing.DescribeLoadBalancersInput{})
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx, func(o *elasticloadbalancing.Options) {
				o.Region = region
			})
			if err != nil {
				return nil, fmt.Errorf("error describing ELB Classic Load Balancers in region %s: %v", region, err)
			}

			for _, lb := range output.LoadBalancerDescriptions {
				resources = append(resources, DiscoveredResource{
					Type: "elasticloadbalancing/loadbalancer/classic",
					Arn:  fmt.Sprintf("arn:aws:elasticloadbalancing:%s:%s:loadbalancer/%s", region, p.cache.GetAccountId(), *lb.LoadBalancerName),
					Name: *lb.LoadBalancerName,
				})
			}
		}
	}

	return &DiscoveryResponse{
		Resources: resources,
	}, nil
}
