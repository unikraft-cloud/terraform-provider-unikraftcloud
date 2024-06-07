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

	"sdk.kraft.cloud/instances"
)

func NewInstanceDataSource() datasource.DataSource {
	return &InstanceDataSource{}
}

// InstanceDataSource defines the data source implementation.
type InstanceDataSource struct {
	client instances.InstancesService
}

// Ensure InstanceDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &InstanceDataSource{}

// InstanceDataSourceModel describes the data source data model.
type InstanceDataSourceModel struct {
	UUID types.String `tfsdk:"uuid"`

	Name              types.String `tfsdk:"name"`
	FQDN              types.String `tfsdk:"fqdn"`
	PrivateIP         types.String `tfsdk:"private_ip"`
	PrivateFQDN       types.String `tfsdk:"private_fqdn"`
	State             types.String `tfsdk:"state"`
	CreatedAt         types.String `tfsdk:"created_at"`
	Image             types.String `tfsdk:"image"`
	MemoryMB          types.Int64  `tfsdk:"memory_mb"`
	Args              types.List   `tfsdk:"args"`
	Env               types.Map    `tfsdk:"env"`
	ServiceGroup      *svcGrpModel `tfsdk:"service_group"`
	NetworkInterfaces types.List   `tfsdk:"network_interfaces"`
	BootTimeUS        types.Int64  `tfsdk:"boot_time_us"`
}

// Metadata implements datasource.DataSource.
func (d *InstanceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Schema implements datasource.DataSource.
func (d *InstanceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Provides state information about a KraftCloud instance.",

		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Required: true,
				MarkdownDescription: "Unique identifier of the " +
					"[instance](https://docs.kraft.cloud/002-rest-api-v1-instances.html)",
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"fqdn": schema.StringAttribute{
				Computed: true,
			},
			"private_ip": schema.StringAttribute{
				Computed: true,
			},
			"private_fqdn": schema.StringAttribute{
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
			"service_group": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"uuid": schema.StringAttribute{
						Computed: true,
					},
					"name": schema.StringAttribute{
						Computed: true,
					},
					"services": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"port": schema.Int64Attribute{
									Computed: true,
								},
								"destination_port": schema.Int64Attribute{
									Computed: true,
								},
								"handlers": schema.SetAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
							},
						},
					},
				},
			},
			"network_interfaces": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
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

	client, ok := req.ProviderData.(instances.InstancesService)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected instances.InstancesServices, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	insRaw, err := d.client.Get(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to get instance state, got error: %v", err),
		)
		return
	}
	ins := insRaw.Data.Entries[0]

	var diags diag.Diagnostics

	data.Name = types.StringValue(ins.Name)
	if ins.ServiceGroup != nil && len(ins.ServiceGroup.Domains) > 0 {
		data.FQDN = types.StringValue(ins.ServiceGroup.Domains[0].FQDN)
	}
	data.PrivateIP = types.StringValue(ins.PrivateIP)
	data.PrivateFQDN = types.StringValue(ins.PrivateFQDN)
	data.State = types.StringValue(string(ins.State))
	data.CreatedAt = types.StringValue(ins.CreatedAt)
	data.Image = types.StringValue(ins.Image)
	data.MemoryMB = types.Int64Value(int64(ins.MemoryMB))
	data.BootTimeUS = types.Int64Value(int64(ins.BootTimeUs))

	data.Args, diags = types.ListValueFrom(ctx, types.StringType, ins.Args)
	resp.Diagnostics.Append(diags...)

	data.Env, diags = types.MapValueFrom(ctx, types.StringType, ins.Env)
	resp.Diagnostics.Append(diags...)

	if ins.ServiceGroup != nil {
		data.ServiceGroup = &svcGrpModel{
			UUID:     types.StringValue(ins.ServiceGroup.UUID),
			Name:     types.StringValue(ins.ServiceGroup.Name),
			Services: make([]svcModel, len(ins.ServiceGroup.Domains)),
		}
	} else {
		data.ServiceGroup = &svcGrpModel{}
	}

	// TODO(craciunoiuc): Find out how this should be accessed now
	// for i, svc := range ins.ServiceGroup.Domains {
	// 	data.ServiceGroup.Services[i] = svcModel{
	// 		Port:            types.Int64Value(int64(svc.Port)),
	// 		DestinationPort: types.Int64Value(int64(svc.DestinationPort)),
	// 	}
	// 	data.ServiceGroup.Services[i].Handlers, diags = types.SetValueFrom(ctx, types.StringType, svc.Handlers)
	// 	resp.Diagnostics.Append(diags...)
	// }

	netwIfaces := make([]netwIfaceModel, len(ins.NetworkInterfaces))
	for i, net := range ins.NetworkInterfaces {
		netwIfaces[i].UUID = types.StringValue(net.UUID)
		netwIfaces[i].Name = types.StringValue(net.UUID) // No name in the API response
		netwIfaces[i].PrivateIP = types.StringValue(net.PrivateIP)
		netwIfaces[i].MAC = types.StringValue(net.MAC)
	}
	data.NetworkInterfaces, diags = types.ListValueFrom(ctx, netwIfaceModelType, netwIfaces)
	resp.Diagnostics.Append(diags...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
