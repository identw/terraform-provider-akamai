provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

data "akamai_cloudaccess_key_properties" "test" {}