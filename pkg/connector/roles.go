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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

const (
	superUserRole = "Superuser"
	userRole      = "User"
	billingRole   = "Billing"
	engineerRole  = "Engineer"
)

var (
	roles                        = []string{superUserRole, userRole, billingRole, engineerRole}
	rolesWithAccessToAllServices = []string{superUserRole, userRole, billingRole}
	revokedRole                  = userRole
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

func (o *roleBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	if principal.Id.ResourceType != userResourceType.Id {
		err := fmt.Errorf("baton-fastly: only users can be granted to roles")

		l.Warn(
			err.Error(),
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)
	}

	role := strings.ToLower(entitlement.Resource.Id.Resource)

	_, err := o.client.UpdateUser(&fastly.UpdateUserInput{
		ID:   principal.Id.Resource,
		Role: &role,
	})
	if err != nil {
		err = wrapError(err, "failed to grant role to user")

		l.Error(
			err.Error(),
			zap.String("role_id", entitlement.Resource.Id.Resource),
			zap.String("user_id", principal.Id.Resource),
		)
	}

	return nil, nil
}

func (o *roleBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	principal := grant.Principal

	if principal.Id.ResourceType != userResourceType.Id {
		err := fmt.Errorf("baton-fastly: only users can be granted to roles")

		l.Warn(
			err.Error(),
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)
	}

	role := strings.ToLower(revokedRole)

	_, err := o.client.UpdateUser(&fastly.UpdateUserInput{
		ID:   principal.Id.Resource,
		Role: &role,
	})
	if err != nil {
		err = wrapError(err, "failed to grant role to user")

		l.Error(
			err.Error(),
			zap.String("role_id", revokedRole),
			zap.String("user_id", principal.Id.Resource),
		)
	}

	return nil, nil
}
