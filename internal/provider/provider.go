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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"sdk.kraft.cloud/client"
	"sdk.kraft.cloud/instance"

	"github.com/kraftcloud/terraform-provider-kraftcloud/internal/validators"
)

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KraftCloudProvider{
			version: version,
		}
	}
}

// KraftCloudProvider defines the provider implementation.
type KraftCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Ensure KraftCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &KraftCloudProvider{}

// KraftCloudModel describes the provider data model.
type KraftCloudModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
	User     types.String `tfsdk:"user"`
	Token    types.String `tfsdk:"token"`
}

// Metadata implements provider.Provider.
func (p *KraftCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kraftcloud"
	resp.Version = p.version
}

// Schema implements provider.Provider.
func (p *KraftCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The KraftCloud provider allows Terraform to manage unikernel instances on KraftCloud.",

		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "API endpoint",
				Optional:            true,
				Validators: []validator.String{
					validators.HTTPURL(),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "API user",
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
func (p *KraftCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data KraftCloudModel

	// Retrieve provider data from configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If a configuration value was provided for any of the attributes, it must
	// be a known value (either literal, or already resolved by Terraform).

	if data.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown KraftCloud API Endpoint",
			"The provider cannot create the KraftCloud API client as there is an unknown configuration value for the KraftCloud API endpoint. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the KRAFTCLOUD_ENDPOINT environment variable.",
		)
	}

	if data.User.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("user"),
			"Unknown KraftCloud API User",
			"The provider cannot create the KraftCloud API client as there is an unknown configuration value for the KraftCloud API user. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the KRAFTCLOUD_USER environment variable.",
		)
	}

	if data.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown KraftCloud API Token",
			"The provider cannot create the KraftCloud API client as there is an unknown configuration value for the KraftCloud API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the KRAFTCLOUD_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Consider values from environment variables, but override with explicit
	// configuration values when provided.

	endpoint := os.Getenv("KRAFTCLOUD_ENDPOINT")
	if endpoint == "" {
		endpoint = client.BaseURL
	}
	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	user := os.Getenv("KRAFTCLOUD_USER")
	if !data.User.IsNull() {
		user = data.User.ValueString()
	}

	token := os.Getenv("KRAFTCLOUD_TOKEN")
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	// If any of the expected configurations are still missing at this point,
	// fail the provider's configuration phase.

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing KraftCloud API Endpoint",
			"The provider cannot create the KraftCloud API client as there is a missing or empty configuration value for the KraftCloud API endpoint. "+
				"Set the endpoint value in the configuration or use the KRAFTCLOUD_ENDPOINT environment variable.",
		)
	}

	if user == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user"),
			"Missing KraftCloud API User",
			"The provider cannot create the KraftCloud API client as there is a missing or empty configuration value for the KraftCloud API user. "+
				"Set the user value in the configuration or use the KRAFTCLOUD_USER environment variable.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing KraftCloud API Token",
			"The provider cannot create the KraftCloud API client as there is a missing or empty configuration value for the KraftCloud API token. "+
				"Set the token value in the configuration or use the KRAFTCLOUD_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Client configuration for data sources and resources
	client := instance.NewInstanceClient(client.NewHTTPClient(), endpoint, user, token)

	resp.DataSourceData = client
	resp.ResourceData = client
}

// Resources describes the provider data model.
func (p *KraftCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
	}
}

// DataSources describes the provider data model.
func (p *KraftCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewInstanceDataSource,
		NewInstancesDataSource,
	}
}
