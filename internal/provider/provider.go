// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	unikraftcloud "sdk.kraft.cloud"
	"sdk.kraft.cloud/client"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UnikraftCloudProvider{
			version: version,
		}
	}
}

// UnikraftCloudProvider defines the provider implementation.
type UnikraftCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Ensure UnikraftCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &UnikraftCloudProvider{}

// UnikraftCloudModel describes the provider data model.
type UnikraftCloudModel struct {
	Metro types.String `tfsdk:"metro"`
	Token types.String `tfsdk:"token"`
}

// Metadata implements provider.Provider.
func (p *UnikraftCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unikraft-cloud"
	resp.Version = p.version
}

// Schema implements provider.Provider.
func (p *UnikraftCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Manage unikernel instances on Unikraft Cloud.",

		Attributes: map[string]schema.Attribute{
			"metro": schema.StringAttribute{
				MarkdownDescription: "API metro",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "API token",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

// Configure implements provider.Provider.
func (p *UnikraftCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data UnikraftCloudModel

	// Retrieve provider data from configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If a configuration value was provided for any of the attributes, it must
	// be a known value (either literal, or already resolved by Terraform).

	if data.Metro.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("metro"),
			"Unknown Unikraft Cloud API Metro",
			"The provider cannot create the Unikraft Cloud API client as there is an unknown configuration value for the Unikraft Cloud API metro. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the UKC_METRO environment variable.",
		)
	}

	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Unikraft Cloud API Token",
			"The provider cannot create the Unikraft Cloud API client as there is an unknown configuration value for the Unikraft Cloud API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the UKC_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Consider values from environment variables, but override with explicit
	// configuration values when provided.

	metro := os.Getenv("UKC_METRO")
	if metro == "" {
		metro = client.DefaultMetro
	}
	if !data.Metro.IsNull() {
		metro = data.Metro.ValueString()
	}

	token := os.Getenv("UKC_TOKEN")
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	// If any of the expected configurations are still missing at this point,
	// fail the provider's configuration phase.

	if metro == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("metro"),
			"Missing Unikraft Cloud API Metro",
			"The provider cannot create the Unikraft Cloud API client as there is a missing or empty configuration value for the Unikraft Cloud API metro. "+
				"Set the metro value in the configuration or use the UKC_METRO environment variable.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Unikraft Cloud API Token",
			"The provider cannot create the Unikraft Cloud API client as there is a missing or empty configuration value for the Unikraft Cloud API token. "+
				"Set the token value in the configuration or use the UKC_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Client configuration for data sources and resources
	client := unikraftcloud.NewClient(
		unikraftcloud.WithDefaultMetro(metro),
		unikraftcloud.WithToken(token),
	)

	resp.DataSourceData = client.Instances()
	resp.ResourceData = client.Instances()
}

// Resources describes the provider data model.
func (p *UnikraftCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
	}
}

// DataSources describes the provider data model.
func (p *UnikraftCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewInstanceDataSource,
		NewInstancesDataSource,
	}
}
