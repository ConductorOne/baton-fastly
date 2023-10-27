package connector

import (
	"context"
	"strings"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	rs "github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/fastly/go-fastly/v8/fastly"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *fastly.Client
	customerId   string
}

func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

func newUserResource(ctx context.Context, user *fastly.User) (*v2.Resource, error) {
	firstName, lastName := parseName(user.Name)
	profile := map[string]interface{}{
		"customer_id": user.CustomerID,
		"login":       user.Login,
		"first_name":  firstName,
	}

	if lastName != "" {
		profile["last_name"] = lastName
	}

	var userStatus v2.UserTrait_Status_Status
	if user.Locked {
		userStatus = v2.UserTrait_Status_STATUS_DISABLED
	} else {
		userStatus = v2.UserTrait_Status_STATUS_ENABLED
	}

	userTraits := []rs.UserTraitOption{
		rs.WithUserProfile(profile),
		rs.WithStatus(userStatus),
		rs.WithUserLogin(user.Login),
	}

	resource, err := rs.NewUserResource(user.Name, userResourceType, user.ID, userTraits)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

func parseName(name string) (string, string) {
	names := strings.Split(name, " ")

	if len(names) == 1 {
		return names[0], ""
	}

	return names[0], names[1]
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	users, err := o.client.ListCustomerUsers(&fastly.ListCustomerUsersInput{CustomerID: o.customerId})
	if err != nil {
		return nil, "", nil, wrapError(err, "error listing users")
	}

	var resources []*v2.Resource
	for _, user := range users {
		resource, err := newUserResource(ctx, user)
		if err != nil {
			return nil, "", nil, wrapError(err, "error creating user resource")
		}

		resources = append(resources, resource)
	}

	return resources, "", nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func newUserBuilder(client *fastly.Client, customerId string) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
		customerId:   customerId,
	}
}
