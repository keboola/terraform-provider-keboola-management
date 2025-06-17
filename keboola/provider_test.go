package keboola

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

// TestProviderSchema verifies the provider schema is correct.
func TestProviderSchema(t *testing.T) {
	p := Provider()
	assert.NotNil(t, p)
	assert.Contains(t, p.Schema, "url")
	assert.Contains(t, p.Schema, "token")
}

// TestProviderConfigure verifies the provider can be configured with minimal settings.
func TestProviderConfigure(t *testing.T) {
	p := Provider()
	// Use real Manage API token from environment if available
	token := os.Getenv("KBC_MANAGE_API_TOKEN")
	if token == "" {
		token = "test-token" // fallback for local/dev testing
	}
	d := schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{
		"url":   "connection.keboola.com",
		"token": token,
	})
	client, diags := p.ConfigureContextFunc(context.Background(), d)
	assert.Nil(t, diags)
	assert.NotNil(t, client)
}
