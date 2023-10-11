// Copyright (c) Unikraft GmbH
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// HTTPURL returns a validator which ensures that any configured attribute
// value is a HTTP(S) URL.
//
// Null (unconfigured) and unknown (known after apply) values are skipped.
func HTTPURL() validator.String {
	return httpURLValidator{}
}

var _ validator.String = httpURLValidator{}

// httpURLValidator validates that a string Attribute's value is a valid URL.
type httpURLValidator struct{}

// Description implements validator.Describer.
func (httpURLValidator) Description(context.Context) string {
	return "value must be a HTTP(S) URL"
}

// MarkdownDescription implements validator.Describer.
func (v httpURLValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString implements validator.String.
func (v httpURLValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		return
	}

	val := request.ConfigValue.ValueString()

	u, err := url.Parse(val)
	if err != nil || u.Scheme != "http" && u.Scheme != "https" {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueMatchDiagnostic(
			request.Path,
			v.Description(ctx),
			val,
		))
	}
}
