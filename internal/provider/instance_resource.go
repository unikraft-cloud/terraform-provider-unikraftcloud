// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"math"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"sdk.kraft.cloud/instance"
)

func NewInstanceResource() resource.Resource {
	return &InstanceResource{}
}

// InstanceResource defines the resource implementation.
type InstanceResource struct {
	client *instance.InstanceClient
}

// Ensure InstanceResource satisfies various resource interfaces.
var (
	_ resource.Resource                = &InstanceResource{}
	_ resource.ResourceWithImportState = &InstanceResource{}
)

// InstanceResourceModel describes the resource data model.
type InstanceResourceModel struct {
	Image        types.String `tfsdk:"image"`
	Args         types.List   `tfsdk:"args"`
	MemoryMB     types.Int64  `tfsdk:"memory_mb"`
	Port         types.Int64  `tfsdk:"port"`
	InternalPort types.Int64  `tfsdk:"internal_port"`
	Handlers     types.Set    `tfsdk:"handlers"`
	Autostart    types.Bool   `tfsdk:"autostart"`

	UUID         types.String `tfsdk:"uuid"`
	DNS          types.String `tfsdk:"dns"`
	PrivateIP    types.String `tfsdk:"private_ip"`
	State        types.String `tfsdk:"state"`
	CreatedAt    types.String `tfsdk:"created_at"`
	Env          types.Map    `tfsdk:"env"`
	ServiceGroup types.String `tfsdk:"service_group"`
	BootTimeUS   types.Int64  `tfsdk:"boot_time_us"`
}

// Metadata implements resource.Resource.
func (r *InstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Schema implements resource.Resource.
func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Allows the creation of KraftCloud instances.",
		MarkdownDescription: "Allows the creation of KraftCloud " +
			"[instances](https://docs.kraft.cloud/002-rest-api-v1-instances.html).",

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
			"port": schema.Int64Attribute{
				Required: true,
				Validators: []validator.Int64{
					int64validator.Between(1, math.MaxUint16),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"internal_port": schema.Int64Attribute{
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
				Default: setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("tls"),
				})),
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplaceIfConfigured(),
					setplanmodifier.UseStateForUnknown(),
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
			"env": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
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

// Configure implements resource.Resource.
func (r *InstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*instance.InstanceClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *instance.InstanceClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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
	// Default: int64default.StaticInt64(128)
	if data.MemoryMB.IsUnknown() || data.MemoryMB.IsNull() {
		data.MemoryMB = types.Int64Value(128)
	}

	// TODO(antoineco): the SDK should be sending a null when this is unset,
	// but currently sends 0 instead, which is invalid.
	// Set a default client-side for now until this is addressed.
	if data.InternalPort.IsUnknown() || data.InternalPort.IsNull() {
		data.InternalPort = data.Port
	}

	in := instance.CreateInstancePayload{
		Image:        data.Image.ValueString(),
		Memory:       data.MemoryMB.ValueInt64(),
		Port:         data.Port.ValueInt64(),
		InternalPort: data.InternalPort.ValueInt64(),
		Autostart:    data.Autostart.ValueBool(),
	}

	argVals := make([]types.String, 0, len(data.Args.Elements()))
	resp.Diagnostics.Append(data.Args.ElementsAs(ctx, &argVals, false)...)
	for _, v := range argVals {
		in.Args = append(in.Args, v.ValueString())
	}
	handlVals := make([]types.String, 0, len(data.Handlers.Elements()))
	resp.Diagnostics.Append(data.Handlers.ElementsAs(ctx, &handlVals, false)...)
	for _, v := range handlVals {
		in.Handlers = append(in.Handlers, v.ValueString())
	}
	if resp.Diagnostics.HasError() {
		return
	}

	ins, err := r.client.CreateInstance(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to create instance, got error: %v", err),
		)
		return
	}

	data.UUID = types.StringValue(ins.UUID)
	data.DNS = types.StringValue(ins.DNS)
	data.PrivateIP = types.StringValue(ins.PrivateIP)

	// Not all attributes are returned by CreateInstance
	ins, err = r.client.InstanceStatus(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to get instance status, got error: %v", err),
		)
		return
	}

	var diags diag.Diagnostics

	// NOTE(antoineco): although the Image attribute may be transformed by
	// KraftCloud (e.g. replace the tag with a digest), we must not update the
	// value read from the schema, otherwise Terraform fails to apply with the
	// following error:
	//
	//   Error: Provider produced inconsistent result after apply
	//   When applying changes to kraftcloud_instance.xyz, provider produced an unexpected new value: .image:
	//     was cty.StringVal("myimage/latest"), but now cty.StringVal("myimage/18a381f0062...").
	//
	data.State = types.StringValue(ins.Status)
	data.CreatedAt = types.StringValue(ins.CreatedAt)
	data.MemoryMB = types.Int64Value(int64(ins.MemoryMB))
	data.ServiceGroup = types.StringValue(ins.ServiceGroup)
	data.BootTimeUS = types.Int64Value(ins.BootTimeUS)

	data.Args, diags = types.ListValueFrom(ctx, types.StringType, ins.Args)
	resp.Diagnostics.Append(diags...)

	data.Env, diags = types.MapValueFrom(ctx, types.StringType, ins.Env)
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

	ins, err := r.client.InstanceStatus(ctx, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to get instance status, got error: %v", err),
		)
		return
	}

	var diags diag.Diagnostics

	// NOTE(antoineco): although the Image attribute may be transformed by
	// KraftCloud (e.g. replace the tag with a digest), we must not update the
	// value read from the schema, otherwise Terraform fails to apply with the
	// following error:
	//
	//   Error: Provider produced inconsistent result after apply
	//   When applying changes to kraftcloud_instance.xyz, provider produced an unexpected new value: .image:
	//     was cty.StringVal("myimage/latest"), but now cty.StringVal("myimage/18a381f0062...").
	//
	// However, we must still ensure that the Image attribute is populated by
	// "terraform import".
	if data.Image.IsNull() {
		data.Image = types.StringValue(ins.Image)
	}
	data.DNS = types.StringValue(ins.DNS)
	data.PrivateIP = types.StringValue(ins.PrivateIP)
	data.State = types.StringValue(ins.Status)
	data.CreatedAt = types.StringValue(ins.CreatedAt)
	data.MemoryMB = types.Int64Value(int64(ins.MemoryMB))
	data.ServiceGroup = types.StringValue(ins.ServiceGroup)
	data.BootTimeUS = types.Int64Value(ins.BootTimeUS)

	data.Args, diags = types.ListValueFrom(ctx, types.StringType, ins.Args)
	resp.Diagnostics.Append(diags...)

	data.Env, diags = types.MapValueFrom(ctx, types.StringType, ins.Env)
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

	err := r.client.DeleteInstance(ctx, data.UUID.ValueString())
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
