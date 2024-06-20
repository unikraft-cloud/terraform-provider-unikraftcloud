// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"sdk.kraft.cloud/instances"
	"sdk.kraft.cloud/services"
)

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

// InstanceResource defines the resource implementation.
type InstanceResource struct {
	client instances.InstancesService
}

// Ensure InstanceResource satisfies various resource interfaces.
var (
	_ resource.Resource                = &InstanceResource{}
	_ resource.ResourceWithImportState = &InstanceResource{}
)

// InstanceResourceModel describes the resource data model.
type InstanceResourceModel struct {
	Image     types.String `tfsdk:"image"`
	Args      types.List   `tfsdk:"args"`
	MemoryMB  types.Int64  `tfsdk:"memory_mb"`
	Autostart types.Bool   `tfsdk:"autostart"`

	UUID              types.String `tfsdk:"uuid"`
	Name              types.String `tfsdk:"name"`
	FQDN              types.String `tfsdk:"fqdn"`
	PrivateIP         types.String `tfsdk:"private_ip"`
	PrivateFQDN       types.String `tfsdk:"private_fqdn"`
	State             types.String `tfsdk:"state"`
	CreatedAt         types.String `tfsdk:"created_at"`
	Env               types.Map    `tfsdk:"env"`
	ServiceGroup      *svcGrpModel `tfsdk:"service_group"`
	NetworkInterfaces types.List   `tfsdk:"network_interfaces"`
	BootTimeUS        types.Int64  `tfsdk:"boot_time_us"`
}

// Metadata implements resource.Resource.
func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Schema implements resource.Resource.
func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Allows the creation of KraftCloud instances.",

		Attributes: map[string]schema.Attribute{
			"image": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"args": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplaceIfConfigured(),
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"memory_mb": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Validators: []validator.Int64{
					int64validator.Between(16, 256),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"autostart": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplaceIfConfigured(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier of the instance",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"env": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"service_group": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"uuid": schema.StringAttribute{
						Computed: true,
					},
					"name": schema.StringAttribute{
						Computed: true,
					},
					"services": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"port": schema.Int64Attribute{
									Required: true,
									Validators: []validator.Int64{
										int64validator.Between(1, math.MaxUint16),
									},
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.RequiresReplace(),
									},
								},
								"destination_port": schema.Int64Attribute{
									Optional: true,
									Computed: true,
									Validators: []validator.Int64{
										int64validator.Between(1, math.MaxUint16),
									},
									PlanModifiers: []planmodifier.Int64{
										int64planmodifier.RequiresReplaceIfConfigured(),
										int64planmodifier.UseStateForUnknown(),
									},
								},
								"handlers": schema.SetAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Computed:    true,
									PlanModifiers: []planmodifier.Set{
										setplanmodifier.RequiresReplaceIfConfigured(),
										setplanmodifier.UseStateForUnknown(),
									},
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

// Configure implements resource.Resource.
func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(instances.InstancesService)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected instances.InstancesServices, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create implements resource.Resource.
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InstanceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO(antoineco): the SDK should be sending a null when this is unset,
	// but currently sends 0 instead, which is invalid.
	// Set a default client-side for now until this is addressed.
	if data.MemoryMB.IsUnknown() || data.MemoryMB.IsNull() {
		data.MemoryMB = types.Int64Value(128)
	}

	in := instances.CreateRequest{
		Image:    data.Image.ValueString(),
		MemoryMB: ptr(int(data.MemoryMB.ValueInt64())),
		ServiceGroup: &instances.CreateRequestServiceGroup{
			Services: make([]services.CreateRequestService, len(data.ServiceGroup.Services)),
		},
		Autostart: ptr(data.Autostart.ValueBool()),
	}

	argVals := make([]types.String, 0, len(data.Args.Elements()))
	resp.Diagnostics.Append(data.Args.ElementsAs(ctx, &argVals, false)...)
	for _, v := range argVals {
		in.Args = append(in.Args, v.ValueString())
	}

	for i, svc := range data.ServiceGroup.Services {
		in.ServiceGroup.Services[i].Port = int(svc.Port.ValueInt64())

		in.ServiceGroup.Services[i].DestinationPort = ptr(int(svc.DestinationPort.ValueInt64()))
		// TODO(antoineco): the SDK should be sending a null when this is unset,
		// but currently sends 0 instead, which is invalid.
		// Set a default client-side for now until this is addressed.
		if svc.DestinationPort.IsUnknown() || svc.DestinationPort.IsNull() {
			data.ServiceGroup.Services[i].DestinationPort = svc.Port
			in.ServiceGroup.Services[i].DestinationPort = ptr(int(svc.Port.ValueInt64()))
		}

		if !svc.Handlers.IsUnknown() {
			handlVals := make([]types.String, 0, len(svc.Handlers.Elements()))
			resp.Diagnostics.Append(svc.Handlers.ElementsAs(ctx, &handlVals, false)...)
			for _, v := range handlVals {
				in.ServiceGroup.Services[i].Handlers = append(in.ServiceGroup.Services[i].Handlers, services.Handler(v.ValueString()))
			}
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	insRaw, err := r.client.Create(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to create instance, got error: %v", err),
		)
		return
	}
	ins := insRaw.Data.Entries[0]

	data.UUID = types.StringValue(ins.UUID)
	data.Name = types.StringValue(ins.Name)
	if ins.ServiceGroup != nil && len(ins.ServiceGroup.Domains) > 0 {
		data.FQDN = types.StringValue(ins.ServiceGroup.Domains[0].FQDN)
	}
	data.PrivateIP = types.StringValue(ins.PrivateIP)
	data.PrivateFQDN = types.StringValue(ins.PrivateFQDN)

	// Not all attributes are returned by CreateInstance
	insRawFull, err := r.client.Get(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to get instance state, got error: %v", err),
		)
		return
	}
	insFull := insRawFull.Data.Entries[0]

	var diags diag.Diagnostics

	// NOTE(antoineco): although the Image attribute may be transformed by
	// KraftCloud (e.g. replace the tag with a digest), we must not update the
	// value read from the schema, otherwise Terraform fails to apply with the
	// following error:
	//
	//   Error: Provider produced inconsistent result after apply
	//   When applying changes to kraftcloud_instance.xyz, provider produced an unexpected new value: .image:
	//     was cty.StringVal("myimage:latest"), but now cty.StringVal("myimage@sha256:18a381f0062...").
	//
	data.State = types.StringValue(string(insFull.State))
	data.CreatedAt = types.StringValue(insFull.CreatedAt)
	data.MemoryMB = types.Int64Value(int64(insFull.MemoryMB))
	data.BootTimeUS = types.Int64Value(int64(insFull.BootTimeUs))

	data.Args, diags = types.ListValueFrom(ctx, types.StringType, insFull.Args)
	resp.Diagnostics.Append(diags...)

	data.Env, diags = types.MapValueFrom(ctx, types.StringType, insFull.Env)
	resp.Diagnostics.Append(diags...)

	if data.ServiceGroup == nil {
		data.ServiceGroup = &svcGrpModel{}
	}

	if insFull.ServiceGroup != nil {
		data.ServiceGroup.UUID = types.StringValue(insFull.ServiceGroup.UUID)
		data.ServiceGroup.Name = types.StringValue(insFull.ServiceGroup.Name)
	}

	netwIfaces := make([]netwIfaceModel, len(insFull.NetworkInterfaces))
	for i, net := range insFull.NetworkInterfaces {
		netwIfaces[i].UUID = types.StringValue(net.UUID)
		netwIfaces[i].Name = types.StringValue(insFull.Name)
		netwIfaces[i].PrivateIP = types.StringValue(net.PrivateIP)
		netwIfaces[i].MAC = types.StringValue(net.MAC)
	}
	data.NetworkInterfaces, diags = types.ListValueFrom(ctx, netwIfaceModelType, netwIfaces)
	resp.Diagnostics.Append(diags...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read implements resource.Resource.
func (r *InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	insRaw, err := r.client.Get(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to get instance state, got error: %v", err),
		)
		return
	}
	ins := insRaw.Data.Entries[0]

	var diags diag.Diagnostics

	// NOTE(antoineco): although the Image attribute may be transformed by
	// KraftCloud (e.g. replace the tag with a digest), we must not update the
	// value read from the schema, otherwise Terraform fails to apply with the
	// following error:
	//
	//   Error: Provider produced inconsistent result after apply
	//   When applying changes to kraftcloud_instance.xyz, provider produced an unexpected new value: .image:
	//     was cty.StringVal("myimage:latest"), but now cty.StringVal("myimage@sha256:18a381f0062...").
	//
	// However, we must still ensure that the Image attribute is populated by
	// "terraform import".
	if data.Image.IsNull() {
		data.Image = types.StringValue(ins.Image)
	}
	data.Name = types.StringValue(ins.Name)
	if ins.ServiceGroup != nil && len(ins.ServiceGroup.Domains) > 0 {
		data.FQDN = types.StringValue(ins.ServiceGroup.Domains[0].FQDN)
	}
	data.PrivateIP = types.StringValue(ins.PrivateIP)
	data.PrivateFQDN = types.StringValue(ins.PrivateFQDN)
	data.State = types.StringValue(string(ins.State))
	data.CreatedAt = types.StringValue(ins.CreatedAt)
	data.MemoryMB = types.Int64Value(int64(ins.MemoryMB))
	data.BootTimeUS = types.Int64Value(int64(ins.BootTimeUs))

	data.Args, diags = types.ListValueFrom(ctx, types.StringType, ins.Args)
	resp.Diagnostics.Append(diags...)

	data.Env, diags = types.MapValueFrom(ctx, types.StringType, ins.Env)
	resp.Diagnostics.Append(diags...)

	if data.ServiceGroup == nil {
		data.ServiceGroup = &svcGrpModel{}
	}

	if ins.ServiceGroup != nil {
		data.ServiceGroup.UUID = types.StringValue(ins.ServiceGroup.UUID)
		data.ServiceGroup.Name = types.StringValue(ins.ServiceGroup.Name)
	}

	netwIfaces := make([]netwIfaceModel, len(ins.NetworkInterfaces))
	for i, net := range ins.NetworkInterfaces {
		netwIfaces[i].UUID = types.StringValue(net.UUID)
		netwIfaces[i].Name = types.StringValue(net.UUID) // No name in the response
		netwIfaces[i].PrivateIP = types.StringValue(net.PrivateIP)
		netwIfaces[i].MAC = types.StringValue(net.MAC)
	}
	data.NetworkInterfaces, diags = types.ListValueFrom(ctx, netwIfaceModelType, netwIfaces)
	resp.Diagnostics.Append(diags...)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update implements resource.Resource.
func (r *InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unsupported",
		"This resource does not support updates. Configuration changes were expected to have triggered a replacement "+
			"of the resource. Please report this issue to the provider developers.",
	)
}

// Delete implements resource.Resource.
func (r *InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InstanceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to get delete instance, got error: %v", err),
		)
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}

func ptr[T comparable](v T) *T { return &v }
