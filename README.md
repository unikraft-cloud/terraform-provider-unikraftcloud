<a href="https://kraft.cloud">
    <img src="https://avatars3.githubusercontent.com/kraftcloud" alt="KraftCloud logo" title="KraftCloud" align="right" height="50" />
</a>

# Terraform Provider for KraftCloud

The KraftCloud provider allows Terraform to manage unikernel instances on KraftCloud.

## Usage

Please refer to the [`kraftcloud` provider documentation][tfreg-docs] in the Terraform Registry.

## Development

This provider is built on top of the [Terraform Plugin Framework][tffw-home].

If you are unfamiliar with the framework, or with extending Terraform with providers in general, we recommend looking
into the [Custom Framework Providers tutorial][tffw-tuto] on the HashiCorp Developer portal as a starting point.

### Requirements

- [Terraform][tf-dl] >= 1.4
- [Go][go-dl] >= 1.21

### Local Installation

Terraform usually installs providers by downloading them from a registry. However, it can be configured to use a [local
development build][tffw-local] of the provider, which allows skipping the version and checksum checks typically
performed against release builds.

To achieve this, add a `dev_overrides` entry to your `~/.terraformrc` configuration file:

```hcl
provider_installation {

  dev_overrides {
      # The value is the path of $GOBIN (or $GOPATH/bin)
      "kraft.cloud/dev/kraftcloud" = "/home/myuser/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

> [!NOTE]
> We deliberately use the provider address `kraft.cloud/dev/kraftcloud` instead of
> `registry.terraform.io/kraftcloud/kraftcloud` in a development context. This allows switching more conveniently
> between the local development build and the release build of this provider.

Terraform will now resolve the provider source `kraft.cloud/dev/kraftcloud` to the local development build, instead of
the remote provider from the Terraform Registry, such as in the example below:

```hcl
terraform {
  required_providers {
    kraftcloud = {
      source = "kraft.cloud/dev/kraftcloud"
    }
  }
}
```

### Testing

This provider includes [Acceptance Tests][tffw-acc] that perform lifecycle actions using real Terraform configurations,
against real KraftCloud resources.

**It is mandatory to have access to a KraftCloud account to execute these tests.**

Acceptance tests are executed using the `testacc` Make target:

```sh
make testacc
```

All acceptance tests are run by default. The `TESTARGS` variable can be used to pass arbitrary arguments to `go test`
(see [Testing flags][gotest-flags]). For example, use the `-run` flag to run specific tests:

```sh
make testacc TESTARGS='-run=TestAccInstanceResource'
```

### Documentation

The documentation of this provider, including its resources and data sources, is generated using the [`tfplugindocs` CLI
tool][tfplugindocs].

To run the documentation generator, execute the `go generate` command. The output is written to the `docs/` directory.

The generator uses the templates from the `templates/` directory and the following data to populate the provider's
documentation pages:

- Examples inside the `examples/` directory
- Schema information from the provider, resources, and data sources


[tfreg-docs]: https://registry.terraform.io/providers/kraftcloud/kraftcloud/latest/docs

[tffw-home]: https://developer.hashicorp.com/terraform/plugin/framework
[tffw-tuto]: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework
[tffw-local]: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install
[tffw-acc]: https://developer.hashicorp.com/terraform/plugin/framework/acctests

[tfplugindocs]: https://github.com/hashicorp/terraform-plugin-docs

[tf-dl]: https://developer.hashicorp.com/terraform/downloads
[go-dl]: https://go.dev/doc/install

[gotest-flags]: https://pkg.go.dev/cmd/go#hdr-Testing_flags
