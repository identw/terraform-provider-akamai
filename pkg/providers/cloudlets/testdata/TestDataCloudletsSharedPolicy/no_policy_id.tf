provider "akamai" {
  edgerc = "../../test/edgerc"
}

data "akamai_cloudlets_shared_policy" "test" {}