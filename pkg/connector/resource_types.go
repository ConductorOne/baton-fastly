package connector

import (
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

var userResourceType = &v2.ResourceType{
	Id:          "user",
	DisplayName: "User",
	Description: "User of Fatly customer.",
	Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
	Annotations: getSkippEntitlementsAndGrantsAnnotations(),
}
