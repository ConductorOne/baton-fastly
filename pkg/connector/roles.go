package connector

import (
	"context"
	"fmt"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	ent "github.com/conductorone/baton-sdk/pkg/types/entitlement"
	grant "github.com/conductorone/baton-sdk/pkg/types/grant"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/fastly/go-fastly/v8/fastly"
)

const (
	superUserRole = "Superuser"
	userRole      = "User"
	billingRole   = "Billing"
	engineerRole  = "Engineer"
)

var (
	roles = []string{superUserRole, userRole, billingRole, engineerRole}

	rolesWithAccessToAllServices = []string{superUserRole, userRole, billingRole}
)

type roleBuilder struct {
	resourceType *v2.ResourceType
	client       *fastly.Client
	customerId   string
}

func newRoleBuilder(client *fastly.Client, customerId string) *roleBuilder {
	return &roleBuilder{
		resourceType: roleResourceType,
		client:       client,
		customerId:   customerId,
	}
}

func (r *roleBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return roleResourceType
}

func newRoleResource(ctx context.Context, role string) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"name": role,
	}

	roleTraitOptions := []rs.RoleTraitOption{
		rs.WithRoleProfile(profile),
	}

	resource, err := rs.NewRoleResource(role, roleResourceType, role, roleTraitOptions)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func (r *roleBuilder) List(ctx context.Context, _ *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource
	for _, role := range roles {
		resource, err := newRoleResource(ctx, role)
		if err != nil {
			return nil, "", nil, err
		}

		resources = append(resources, resource)
	}

	return resources, "", nil, nil
}

func (o *roleBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assigmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("Assigned to %s role", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s role %s", resource.DisplayName, assignedEntitlement)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, assignedEntitlement, assigmentOptions...))

	return rv, "", nil, nil
}

func (o *roleBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	users, err := o.client.ListCustomerUsers(&fastly.ListCustomerUsersInput{CustomerID: o.customerId})
	if err != nil {
		return nil, "", nil, wrapError(err, "error listing users")
	}

	var rv []*v2.Grant
	for _, user := range users {
		if !strings.EqualFold(resource.DisplayName, user.Role) {
			continue
		}

		userResource, err := newUserResource(ctx, user)
		if err != nil {
			return nil, "", nil, wrapError(err, "error creating user resource")
		}

		rv = append(rv, grant.NewGrant(resource, assignedEntitlement, userResource.Id))
	}

	return rv, "", nil, nil
}
