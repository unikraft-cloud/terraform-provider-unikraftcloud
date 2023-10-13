// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"sdk.kraft.cloud/instance"
)

func NewInstanceDataSource() datasource.DataSource {
	return &InstanceDataSource{}
}

// InstanceDataSource defines the data source implementation.
type InstanceDataSource struct {
	client *instance.InstanceClient
}

// Ensure InstanceDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &InstanceDataSource{}

// InstanceDataSourceModel describes the data source data model.
type InstanceDataSourceModel struct {
	UUID types.String `tfsdk:"uuid"`

	DNS               types.String     `tfsdk:"dns"`
	PrivateIP         types.String     `tfsdk:"private_ip"`
	State             types.String     `tfsdk:"state"`
	CreatedAt         types.String     `tfsdk:"created_at"`
	Image             types.String     `tfsdk:"image"`
	MemoryMB          types.Int64      `tfsdk:"memory_mb"`
	Args              types.List       `tfsdk:"args"`
	Env               types.Map        `tfsdk:"env"`
	ServiceGroup      types.String     `tfsdk:"service_group"`
	NetworkInterfaces []netwIfaceModel `tfsdk:"network_interfaces"`
	BootTimeUS        types.Int64      `tfsdk:"boot_time_us"`
}

// netwIfaceModel describes the data model for an instance's network interface.
type netwIfaceModel struct {
	UUID      types.String `tfsdk:"uuid"`
	PrivateIP types.String `tfsdk:"private_ip"`
	MAC       types.String `tfsdk:"mac"`
}

// Metadata implements datasource.DataSource.
func (d *InstanceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Schema implements datasource.DataSource.
func (d *InstanceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provides status information about a KraftCloud instance.",

		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Unique identifier of the " +
					"[instance](https://docs.kraft.cloud/002-rest-api-v1-instances.html)",
			},
			"dns": schema.StringAttribute{
				Computed: true,
			},
			"private_ip": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"image": schema.StringAttribute{
				Computed: true,
			},
			"memory_mb": schema.Int64Attribute{
				Computed: true,
			},
			"args": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"env": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"network_interfaces": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed: true,
						},
						"private_ip": schema.StringAttribute{
							Computed: true,
						},
						"mac": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
			"service_group": schema.StringAttribute{
				Computed: true,
			},
			"boot_time_us": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

// Configure implements datasource.DataSource.
func (d *InstanceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*instance.InstanceClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *instance.InstanceClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read implements datasource.DataSource.
func (d *InstanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InstanceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ins, err := d.client.InstanceStatus(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to get instance status, got error: %v", err),
		)
		return
	}

	var diags diag.Diagnostics

	data.DNS = types.StringValue(ins.DNS)
	data.PrivateIP = types.StringValue(ins.PrivateIP)
	data.State = types.StringValue(ins.Status)
	data.CreatedAt = types.StringValue(ins.CreatedAt)
	data.Image = types.StringValue(ins.Image)
	data.MemoryMB = types.Int64Value(int64(ins.MemoryMB))
	data.ServiceGroup = types.StringValue(ins.ServiceGroup)
	data.BootTimeUS = types.Int64Value(ins.BootTimeUS)

	data.Args, diags = types.ListValueFrom(ctx, types.StringType, ins.Args)
	resp.Diagnostics.Append(diags...)

	data.Env, diags = types.MapValueFrom(ctx, types.StringType, ins.Env)
	resp.Diagnostics.Append(diags...)

	for _, net := range ins.NetworkInterfaces {
		data.NetworkInterfaces = append(data.NetworkInterfaces, netwIfaceModel{
			UUID:      types.StringValue(net.UUID),
			PrivateIP: types.StringValue(net.PrivateIP),
			MAC:       types.StringValue(net.MAC),
		})
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
