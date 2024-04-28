package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
)

type GlobalAcceleratorClient interface {
	ListAccelerators(ctx context.Context, params *globalaccelerator.ListAcceleratorsInput, optFns ...func(*globalaccelerator.Options)) (*globalaccelerator.ListAcceleratorsOutput, error)
}

type globalAcceleratorDiscoveryProvider struct {
	client GlobalAcceleratorClient
}

var _ DiscoveryProvider = &globalAcceleratorDiscoveryProvider{}

func NewGlobalAcceleratorDiscoveryProvider(cfg aws.Config) DiscoveryProvider {
	client := globalaccelerator.NewFromConfig(cfg)
	return &globalAcceleratorDiscoveryProvider{
		client: client,
	}
}

func (p *globalAcceleratorDiscoveryProvider) Discover(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResponse, error) {
	resources := []DiscoveredResource{}

	paginator := globalaccelerator.NewListAcceleratorsPaginator(p.client, &globalaccelerator.ListAcceleratorsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing Global Accelerator accelerators: %v", err)
		}

		for _, accelerator := range page.Accelerators {
			resources = append(resources, DiscoveredResource{
				Type: "globalaccelerator/accelerator",
				Arn:  *accelerator.AcceleratorArn,
				Name: *accelerator.Name,
			})
		}
	}

	return &DiscoveryResponse{
		Resources: resources,
	}, nil
}
