package aws

import "context"

type DiscoveryClient interface {
	Discover(context.Context, *DiscoveryRequest) (*DiscoveryResponse, error)
}

type DiscoveryProvider interface {
	Discover(context.Context, *DiscoveryRequest) (*DiscoveryResponse, error)
}

type DiscoveryRequest struct {
	ResourceTypes []string
	Regions       []string
}

type DiscoveryResponse struct {
	Resources []DiscoveredResource
}

type DiscoveredResource struct {
	Type string
	Arn  string
	Name string
}
