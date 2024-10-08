---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "unikraft-cloud_instance Resource - terraform-provider-unikraft-cloud"
subcategory: ""
description: |-
  Allows the creation of Unikraft Cloud instances.
---

# unikraft-cloud_instance (Resource)

Allows the creation of Unikraft Cloud [instances][kc-instances].



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `image` (String)
- `service_group` (Attributes) (see [below for nested schema](#nestedatt--service_group))

### Optional

- `args` (List of String)
- `autostart` (Boolean)
- `memory_mb` (Number)

### Read-Only

- `boot_time_us` (Number)
- `created_at` (String)
- `env` (Map of String)
- `name` (String)
- `network_interfaces` (Attributes List) (see [below for nested schema](#nestedatt--network_interfaces))
- `private_fqdn` (String)
- `private_ip` (String)
- `state` (String)
- `uuid` (String) Unique identifier of the instance

<a id="nestedatt--service_group"></a>
### Nested Schema for `service_group`

Required:

- `services` (Attributes List) (see [below for nested schema](#nestedatt--service_group--services))

Optional:

- `domains` (Attributes List) (see [below for nested schema](#nestedatt--service_group--domains))

Read-Only:

- `name` (String)
- `uuid` (String)

<a id="nestedatt--service_group--services"></a>
### Nested Schema for `service_group.services`

Required:

- `port` (Number)

Optional:

- `destination_port` (Number)
- `handlers` (Set of String)


<a id="nestedatt--service_group--domains"></a>
### Nested Schema for `service_group.domains`

Required:

- `name` (String)

Optional:

- `certificate` (Attributes Map) (see [below for nested schema](#nestedatt--service_group--domains--certificate))

Read-Only:

- `fqdn` (String)

<a id="nestedatt--service_group--domains--certificate"></a>
### Nested Schema for `service_group.domains.certificate`

Read-Only:

- `name` (String)
- `state` (String)
- `uuid` (String)




<a id="nestedatt--network_interfaces"></a>
### Nested Schema for `network_interfaces`

Read-Only:

- `mac` (String)
- `name` (String)
- `private_ip` (String)
- `uuid` (String)

[kc-instances]: https://docs.kraft.cloud/002-rest-api-v1-instances.html
