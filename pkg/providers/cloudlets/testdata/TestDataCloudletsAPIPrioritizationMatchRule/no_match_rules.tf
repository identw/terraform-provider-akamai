provider "akamai" {
  edgerc = "../../test/edgerc"
}

data "akamai_cloudlets_api_prioritization_match_rule" "test" {}