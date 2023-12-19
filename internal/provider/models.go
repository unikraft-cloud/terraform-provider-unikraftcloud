// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// svcGrpModel describes the data model for an instance's service group.
type svcGrpModel struct {
	UUID     types.String `tfsdk:"uuid"`
	Name     types.String `tfsdk:"name"`
	Services []svcModel   `tfsdk:"services"`
}

// svcModel describes the data model for a service group's service.
type svcModel struct {
	Port            types.Int64 `tfsdk:"port"`
	DestinationPort types.Int64 `tfsdk:"destination_port"`
	Handlers        types.Set   `tfsdk:"handlers"`
}

// netwIfaceModel describes the data model for an instance's network interface.
type netwIfaceModel struct {
	UUID      types.String `tfsdk:"uuid"`
	Name      types.String `tfsdk:"name"`
	PrivateIP types.String `tfsdk:"private_ip"`
	MAC       types.String `tfsdk:"mac"`
}

var netwIfaceModelType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"uuid":       types.StringType,
		"name":       types.StringType,
		"private_ip": types.StringType,
		"mac":        types.StringType,
	},
}
