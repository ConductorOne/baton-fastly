package connector

import (
	"context"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/fastly/go-fastly/v8/fastly"
)

type serviceBuilder struct {
	resourceType *v2.ResourceType
	client       *fastly.Client
}

func (o *serviceBuilder) ResourceType() *v2.ResourceType {
	return o.resourceType
}

func newServiceResource(service *fastly.Service) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"name":    service.Name,
		"comment": service.Comment,
		"type":    service.Type,
	}

	serviceTraits := []rs.AppTraitOption{
		rs.WithAppProfile(profile),
	}

	resource, err := rs.NewAppResource(service.Name, serviceResourceType, service.ID, serviceTraits)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (o *serviceBuilder) List(ctx context.Context, _ *v2.RequestId, pagination *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pagination.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	services, err := o.client.ListServices(&fastly.ListServicesInput{Page: page, PerPage: resourcePageSize})
	if err != nil {
		return nil, "", nil, err
	}

	var resources []*v2.Resource
	for _, service := range services {
		resource, err := newServiceResource(service)
		if err != nil {
			return nil, "", nil, err
		}

		resources = append(resources, resource)
	}

	if isLastPage(len(services), resourcePageSize) {
		return resources, "", nil, nil
	}

	nextPage, err := handleNextPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

func (o *serviceBuilder) Entitlements(ctx context.Context, resource *v2.RequestId, _ *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (o *serviceBuilder) Grants(ctx context.Context, resource *v2.RequestId, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
