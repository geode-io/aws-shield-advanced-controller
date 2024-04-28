package aws

import (
	"context"
	"fmt"
	"slices"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/sourcegraph/conc/pool"
)

type discoveryClient struct {
	providers map[string]DiscoveryProvider
	cache     Cache
}

var _ DiscoveryClient = &discoveryClient{}

func NewDiscoveryClient(cfg aws.Config, cache Cache) DiscoveryClient {

	// Resource types must match the enum defined in the API
	providers := map[string]DiscoveryProvider{
		"cloudfront/distribution":                   NewCloudfrontDiscoveryProvider(cfg),
		"route53/hostedzone":                        NewRoute53DiscoveryProvider(cfg),
		"globalaccelerator/accelerator":             NewGlobalAcceleratorDiscoveryProvider(cfg),
		"ec2/eip":                                   NewEC2EIPDiscoveryProvider(cfg, cache),
		"elasticloadbalancing/loadbalancer/app":     NewELBv2DiscoveryProvider(cfg),
		"elasticloadbalancing/loadbalancer/classic": NewELBDiscoveryProvider(cfg, cache),
	}

	return &discoveryClient{
		providers: providers,
		cache:     cache,
	}
}

func (d *discoveryClient) Discover(ctx context.Context, request *DiscoveryRequest) (*DiscoveryResponse, error) {
	log := log.FromContext(ctx)

	p := pool.
		NewWithResults[[]DiscoveredResource]().
		WithErrors().
		WithMaxGoroutines(4)

	for _, typ := range request.ResourceTypes {

		p.Go(func() ([]DiscoveredResource, error) {
			provider, ok := d.providers[typ]
			if !ok {
				return nil, fmt.Errorf("no discovery provider found for resource type: %s", typ)
			}

			log.V(1).Info("Discovering resource type", "type", typ)
			resp, err := provider.Discover(ctx, request)
			if err != nil {
				return nil, fmt.Errorf("error discovering resource type: %s, error: %v", typ, err)
			}
			log.V(1).Info("Discovered resources", "type", typ, "count", len(resp.Resources), "resources", resp.Resources)

			return resp.Resources, nil
		})
	}

	groups, err := p.Wait()
	if err != nil {
		return nil, err
	}

	return &DiscoveryResponse{
		Resources: slices.Concat(groups...),
	}, nil
}
