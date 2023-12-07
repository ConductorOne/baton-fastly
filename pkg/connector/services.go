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

type serviceBuilder struct {
	resourceType *v2.ResourceType
	client       *fastly.Client
	customerId   string
}

const (
	ReadOnlyPermission    = "read_only"
	PurgeSelectPermission = "purge_select"
	PurgeAllPermission    = "purge_all"
	FullAccessPermission  = "full"
)

var (
	permissionEntitlementMap = map[string][]string{
		ReadOnlyPermission:    {readStatsAndConfigurationEntitlement},
		PurgeSelectPermission: {readStatsAndConfigurationEntitlement, purgeSelectedContentEntitlement},
		PurgeAllPermission:    {readStatsAndConfigurationEntitlement, purgeSelectedContentEntitlement, purgeAllEntitlement},
		FullAccessPermission:  {readStatsAndConfigurationEntitlement, purgeSelectedContentEntitlement, purgeAllEntitlement, fullAccessEntitlement},
	}
	entitlementPermissionMap = map[string]string{
		readStatsAndConfigurationEntitlement: ReadOnlyPermission,
		purgeSelectedContentEntitlement:      PurgeSelectPermission,
		purgeAllEntitlement:                  PurgeAllPermission,
		fullAccessEntitlement:                FullAccessPermission,
	}
	revokeEntitlementMap = map[string]string{
		purgeSelectedContentEntitlement: readStatsAndConfigurationEntitlement,
		purgeAllEntitlement:             purgeSelectedContentEntitlement,
		fullAccessEntitlement:           purgeAllEntitlement,
	}
)

func newServiceBuilder(client *fastly.Client, customerId string) *serviceBuilder {
	return &serviceBuilder{
		resourceType: serviceResourceType,
		client:       client,
		customerId:   customerId,
	}
}

func (o *serviceBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return serviceResourceType
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

func (o *serviceBuilder) List(ctx context.Context, _ *v2.ResourceId, pagination *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
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

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return resources, nextPage, nil, nil
}

func (o *serviceBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var rv []*v2.Entitlement

	assigmentOptions := []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("Can read stats and analytics of %s", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s of %s", readStatsAndAnalyticsEntitlement, resource.DisplayName)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, readStatsAndAnalyticsEntitlement, assigmentOptions...))

	assigmentOptions = []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("Access billing of %s", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s of %s", accessBillingEntitlement, resource.DisplayName)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, accessBillingEntitlement, assigmentOptions...))

	assigmentOptions = []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("manage users and accounts of %s", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s of %s", manageUsersAndAccountsEntitlement, resource.DisplayName)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, manageUsersAndAccountsEntitlement, assigmentOptions...))

	assigmentOptions = []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("Read configuration of %s", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s of %s", readStatsAndConfigurationEntitlement, resource.DisplayName)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, readStatsAndConfigurationEntitlement, assigmentOptions...))

	assigmentOptions = []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("Write configuration of %s", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s of %s", purgeSelectedContentEntitlement, resource.DisplayName)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, purgeSelectedContentEntitlement, assigmentOptions...))

	assigmentOptions = []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("Purge configuration of %s", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s of %s", purgeAllEntitlement, resource.DisplayName)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, purgeAllEntitlement, assigmentOptions...))

	assigmentOptions = []ent.EntitlementOption{
		ent.WithGrantableTo(userResourceType),
		ent.WithDescription(fmt.Sprintf("Activate configuration of %s", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s of %s", fullAccessEntitlement, resource.DisplayName)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, fullAccessEntitlement, assigmentOptions...))

	assigmentOptions = []ent.EntitlementOption{
		ent.WithGrantableTo(roleResourceType),
		ent.WithDescription(fmt.Sprintf("Access %s", resource.DisplayName)),
		ent.WithDisplayName(fmt.Sprintf("%s of %s", accessEntitlement, resource.DisplayName)),
	}
	rv = append(rv, ent.NewAssignmentEntitlement(resource, accessEntitlement, assigmentOptions...))

	return rv, "", nil, nil
}

func (o *serviceBuilder) Grants(ctx context.Context, resource *v2.Resource, pagination *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	bag, page, err := parsePageToken(pagination.Token, &v2.ResourceId{ResourceType: o.resourceType.Id})
	if err != nil {
		return nil, "", nil, err
	}

	var rv []*v2.Grant

	// Handle grants without pagination
	if page == 0 {
		grants, err := grantRoles(ctx, resource)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to grant roles")
		}
		rv = append(rv, grants...)

		grants, err = o.grantUsers(ctx, resource)
		if err != nil {
			return nil, "", nil, wrapError(err, "failed to grant users")
		}

		rv = append(rv, grants...)
	}

	authorizations, err := o.client.ListServiceAuthorizations(&fastly.ListServiceAuthorizationsInput{PageNumber: page, PageSize: resourcePageSize})
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to list service authorizations")
	}

	grants, err := o.grantEngineer(ctx, resource, authorizations.Items)
	if err != nil {
		return nil, "", nil, wrapError(err, "failed to process service authorizations")
	}
	rv = append(rv, grants...)

	if isLastPage(len(authorizations.Items), resourcePageSize) {
		return rv, "", nil, nil
	}

	nextPage, err := getPageTokenFromPage(bag, page+1)
	if err != nil {
		return nil, "", nil, err
	}

	return rv, nextPage, nil, nil
}

func grantRoles(ctx context.Context, resource *v2.Resource) ([]*v2.Grant, error) {
	var rv []*v2.Grant

	for _, role := range rolesWithAccessToAllServices {
		roleResource, err := newRoleResource(ctx, role)
		if err != nil {
			return nil, err
		}

		rv = append(rv, grant.NewGrant(resource, accessEntitlement, roleResource.Id))
	}

	return rv, nil
}

func (o *serviceBuilder) grantUsers(ctx context.Context, service *v2.Resource) ([]*v2.Grant, error) {
	var rv []*v2.Grant

	users, err := o.client.ListCustomerUsers(&fastly.ListCustomerUsersInput{CustomerID: o.customerId})
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		userResource, err := newUserResource(ctx, user)
		if err != nil {
			return nil, err
		}

		switch strings.ToLower(user.Role) {
		case strings.ToLower(superUserRole):
			rv = append(rv, grantSuperuser(service, userResource)...)
		case strings.ToLower(userRole):
			rv = append(rv, grantUser(service, userResource)...)
		case strings.ToLower(billingRole):
			rv = append(rv, grantBilling(service, userResource)...)
		case strings.ToLower(engineerRole):

		default:
			return nil, fmt.Errorf("unknown role %s", user.Role)
		}
	}

	return rv, nil
}

func grantSuperuser(service *v2.Resource, user *v2.Resource) []*v2.Grant {
	rv := []*v2.Grant{
		grant.NewGrant(service, readStatsAndAnalyticsEntitlement, user.Id),
		grant.NewGrant(service, accessBillingEntitlement, user.Id),
		grant.NewGrant(service, manageUsersAndAccountsEntitlement, user.Id),
	}

	return rv
}

func grantUser(service *v2.Resource, user *v2.Resource) []*v2.Grant {
	rv := []*v2.Grant{
		grant.NewGrant(service, readStatsAndAnalyticsEntitlement, user.Id),
	}

	return rv
}

func grantBilling(service *v2.Resource, user *v2.Resource) []*v2.Grant {
	rv := []*v2.Grant{
		grant.NewGrant(service, readStatsAndAnalyticsEntitlement, user.Id),
		grant.NewGrant(service, accessBillingEntitlement, user.Id),
	}

	return rv
}

func (o *serviceBuilder) grantEngineer(ctx context.Context, service *v2.Resource, authorizations []*fastly.ServiceAuthorization) ([]*v2.Grant, error) {
	var rv []*v2.Grant

	for _, authorization := range authorizations {
		if authorization.Service.ID == service.Id.Resource {
			user, err := o.client.GetUser(&fastly.GetUserInput{ID: authorization.User.ID})
			if err != nil {
				return nil, err
			}

			userResource, err := newUserResource(ctx, user)
			if err != nil {
				return nil, err
			}

			if entitlements, exists := permissionEntitlementMap[authorization.Permission]; exists {
				for _, entitlement := range entitlements {
					rv = append(rv, grant.NewGrant(service, entitlement, userResource.Id))
				}
			} else {
				return nil, fmt.Errorf("unknown permission %s", authorization.Permission)
			}
		}
	}

	return rv, nil
}

func (o *serviceBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlement *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	permission, exists := entitlementPermissionMap[entitlement.Slug]
	if !exists {
		err := fmt.Errorf("baton-fastly: unable to grant %s entitlement", entitlement.Slug)

		l.Warn(
			err.Error(),
			zap.String("entitlement_id", entitlement.Slug),
		)

		return nil, err
	}

	err := o.validateGrantOperation(principal, entitlement, l)
	if err != nil {
		return nil, err
	}

	_, err = o.upsertServiceAuthorizationForUser(entitlement.Resource.Id.Resource, principal.Id.Resource, permission, l)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (o *serviceBuilder) getServiceAuthorizationForUser(serviceId, userId string) (*fastly.ServiceAuthorization, error) {
	pageNumber := 1

	for {
		serviceAuthorizations, err := o.client.ListServiceAuthorizations(&fastly.ListServiceAuthorizationsInput{
			PageNumber: pageNumber,
			PageSize:   resourcePageSize,
		})
		if err != nil {
			return nil, err
		}

		for _, serviceAuthorization := range serviceAuthorizations.Items {
			if serviceAuthorization.Service.ID == serviceId && serviceAuthorization.User.ID == userId {
				return serviceAuthorization, nil
			}
		}

		if pageNumber >= serviceAuthorizations.Info.Meta.TotalPages {
			break
		}
	}

	return nil, nil
}

// Service authorization for user can already exist with different permission.
// In this case we need to update it.
func (o *serviceBuilder) upsertServiceAuthorizationForUser(serviceId, userId, permission string, l *zap.Logger) (*fastly.ServiceAuthorization, error) {
	serviceAuthorization, err := o.getServiceAuthorizationForUser(serviceId, userId)
	if err != nil {
		return nil, wrapError(err, "failed to get service authorization")
	}

	if serviceAuthorization != nil {
		if serviceAuthorization.Permission == permission {
			return serviceAuthorization, nil
		}

		serviceAuthorization, err := o.client.UpdateServiceAuthorization(&fastly.UpdateServiceAuthorizationInput{
			ID:         serviceAuthorization.ID,
			Permission: permission,
		})
		if err != nil {
			err = wrapError(err, "failed to update permission to user")

			l.Error(
				err.Error(),
				zap.String("permission", permission),
				zap.String("user_id", userId),
				zap.String("service_id", serviceId),
			)
		}

		return serviceAuthorization, nil
	} else {
		serviceAuthorization, err := o.client.CreateServiceAuthorization(&fastly.CreateServiceAuthorizationInput{
			Service: &fastly.SAService{
				ID: serviceId,
			},
			User: &fastly.SAUser{
				ID: userId,
			},
			Permission: permission,
		})
		if err != nil {
			err = wrapError(err, "failed to grant permission to user")

			l.Error(
				err.Error(),
				zap.String("permission", permission),
				zap.String("user_id", userId),
				zap.String("service_id", serviceId),
			)
		}

		return serviceAuthorization, nil
	}
}

func (o *serviceBuilder) validateGrantOperation(principal *v2.Resource, entitlement *v2.Entitlement, l *zap.Logger) error {
	if principal.Id.ResourceType != userResourceType.Id {
		err := fmt.Errorf("baton-fastly: only users can be granted to service")

		l.Warn(
			err.Error(),
			zap.String("principal_id", principal.Id.Resource),
			zap.String("principal_type", principal.Id.ResourceType),
		)

		return err
	}

	user, err := o.client.GetUser(&fastly.GetUserInput{ID: principal.Id.Resource})
	if err != nil {
		err := wrapError(err, "failed to get user")

		l.Error(
			err.Error(),
			zap.String("user_id", principal.Id.Resource),
		)

		return err
	}

	if user.Role != strings.ToLower(engineerRole) {
		err := fmt.Errorf("baton-fastly: only users with role %s can be granted to service", engineerRole)

		l.Warn(
			err.Error(),
			zap.String("user_id", principal.Id.Resource),
			zap.String("user_role", user.Role),
		)

		return err
	}

	return nil
}

func (o *serviceBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	principal := grant.Principal
	entitlement := grant.Entitlement

	revokedEntitlement, exists := revokeEntitlementMap[entitlement.Slug]
	if !exists {
		err := fmt.Errorf("baton-fastly: unable to revoke %s entitlement", entitlement.Slug)

		l.Warn(
			err.Error(),
			zap.String("entitlement_id", entitlement.Slug),
		)

		return nil, err
	}

	revokedPermission, exists := entitlementPermissionMap[revokedEntitlement]
	if !exists {
		err := fmt.Errorf("baton-fastly: unable to map %s entitlement to permission", revokedEntitlement)

		l.Warn(
			err.Error(),
			zap.String("entitlement_id", revokedEntitlement),
		)

		return nil, err
	}

	err := o.validateGrantOperation(principal, entitlement, l)
	if err != nil {
		return nil, err
	}

	_, err = o.upsertServiceAuthorizationForUser(entitlement.Resource.Id.Resource, principal.Id.Resource, revokedPermission, l)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
