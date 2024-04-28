package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type ELBV2Client interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
}

type elbv2DiscoveryProvider struct {
	client ELBV2Client
}

var _ DiscoveryProvider = &elbv2DiscoveryProvider{}

func NewELBv2DiscoveryProvider(cfg aws.Config) DiscoveryProvider {
	return &elbv2DiscoveryProvider{
		client: elasticloadbalancingv2.NewFromConfig(cfg),
	}
}

func (p *elbv2DiscoveryProvider) Discover(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResponse, error) {
	resources := []DiscoveredResource{}

	for _, region := range request.Regions {
		paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(p.client, &elasticloadbalancingv2.DescribeLoadBalancersInput{})
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx, func(o *elasticloadbalancingv2.Options) {
				o.Region = region
			})
			if err != nil {
				return nil, fmt.Errorf("error describing ELBv2 Application Load Balancers in region %s: %v", region, err)
			}

			for _, lb := range output.LoadBalancers {
				if lb.Type == types.LoadBalancerTypeEnumApplication {
					resources = append(resources, DiscoveredResource{
						Type: "elasticloadbalancing/loadbalancer/app",
						Arn:  *lb.LoadBalancerArn,
						Name: *lb.LoadBalancerName,
					})
				}
			}
		}
	}

	return &DiscoveryResponse{
		Resources: resources,
	}, nil
}
