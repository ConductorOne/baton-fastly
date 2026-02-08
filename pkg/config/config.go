package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	AccessToken = field.StringField(
		"access-token",
		field.WithDescription("Fastly API token"),
		field.WithRequired(true),
		field.WithIsSecret(true),
		field.WithDisplayName("Access Token"),
	)
	BaseURLField = field.StringField(
		"base-url",
		field.WithDescription("Override the Fastly API URL (for testing)"),
	)

	// FieldRelationships defines relationships between the fields listed in
	// Config that can be automatically validated.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run ./gen
var Config = field.NewConfiguration([]field.SchemaField{
	AccessToken,
	BaseURLField,
})

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(cfg *Fastly) error {
	return nil
}
