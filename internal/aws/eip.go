package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

type EC2Client interface {
	DescribeAddresses(ctx context.Context, params *ec2.DescribeAddressesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error)
}

type ec2EIPDiscoveryProvider struct {
	client EC2Client
	cache  AWSCache
}

var _ DiscoveryProvider = &ec2EIPDiscoveryProvider{}

func NewEC2EIPDiscoveryProvider(cfg aws.Config, cache AWSCache) DiscoveryProvider {
	return &ec2EIPDiscoveryProvider{
		client: ec2.NewFromConfig(cfg),
		cache:  cache,
	}
}

func (p *ec2EIPDiscoveryProvider) Discover(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResponse, error) {
	resources := []DiscoveredResource{}

	for _, region := range request.Regions {
		output, err := p.client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{}, func(o *ec2.Options) {
			o.Region = region
		})
		if err != nil {
			return nil, fmt.Errorf("error describing EC2 Elastic IPs in region %s: %v", region, err)
		}

		for _, addr := range output.Addresses {
			resources = append(resources, DiscoveredResource{
				Type: "ec2/eip",
				Arn:  fmt.Sprintf("arn:aws:ec2:%s:%s:eip-allocation/%s", region, p.cache.GetAccountId(), *addr.AllocationId),
				Name: *addr.PublicIp,
			})
		}
	}

	return &DiscoveryResponse{
		Resources: resources,
	}, nil
}
