---
layout: "akamai"
page_title: "Akamai: HostnameCoverage"
subcategory: "Application Security"
description: |-
 HostnameCoverage
---






# akamai_appsec_hostname_coverage

**Scopes**: Individual account

Returns information about the hostnames associated with your account; the returned data includes the hostname's protections, activation status, and other summary information. This information is described in the [HostnameCoverage members](https://developer.akamai.com/api/cloud_security/application_security/v1.html#getfailoverhostnames) section of the Application Security API.

**Related API Endpoint**: [/appsec/v1/hostname-coverage](https://developer.akamai.com/api/cloud_security/application_security/v1.html#gethostnamecoverage)

## Example Usage

Basic usage:

```
terraform {
  required_providers {
    akamai = {
      source = "akamai/akamai"
    }
  }
}

provider "akamai" {
  edgerc = "~/.edgerc"
}

// USE CASE: User wants to view hostname coverage data.

data "akamai_appsec_hostname_coverage" "hostname_coverage" {
}

output "hostname_coverage_list_json" {
  value = data.akamai_appsec_hostname_coverage.hostname_coverage.json
}

// USE CASE: User wants to display the returned data in a table.

output "hostname_coverage_list_output" {
  value = data.akamai_appsec_hostname_coverage.hostname_coverage.output_text
}
```

## Argument Reference

This data source does not support any arguments.

## Output Options

The following options can be used to determine the information returned, and how that returned information is formatted:

- `json`. JSON-formatted list of the hostname coverage information.
- `output_text`. Tabular report of the hostname coverage information.
