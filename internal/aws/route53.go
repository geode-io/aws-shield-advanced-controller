package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

type Route53Client interface {
	ListHostedZones(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error)
}

type route53DiscoveryProvider struct {
	client Route53Client
}

var _ DiscoveryProvider = &route53DiscoveryProvider{}

func NewRoute53DiscoveryProvider(cfg aws.Config) DiscoveryProvider {
	return &route53DiscoveryProvider{
		client: route53.NewFromConfig(cfg),
	}
}

func (p *route53DiscoveryProvider) Discover(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResponse, error) {
	resources := []DiscoveredResource{}

	paginator := route53.NewListHostedZonesPaginator(p.client, &route53.ListHostedZonesInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("error listing Route53 hosted zones: %v", err)
		}

		for _, zone := range page.HostedZones {
			id := strings.TrimPrefix(*zone.Id, "/")
			resources = append(resources, DiscoveredResource{
				Type: "route53/hostedzone",
				Arn:  fmt.Sprintf("arn:aws:route53:::%s", id),
				Name: *zone.Name,
			})
		}
	}

	return &DiscoveryResponse{
		Resources: resources,
	}, nil
}
