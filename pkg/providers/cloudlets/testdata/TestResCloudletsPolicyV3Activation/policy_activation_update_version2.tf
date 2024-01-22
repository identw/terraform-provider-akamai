provider "akamai" {
  edgerc = "../../test/edgerc"
}

resource "akamai_cloudlets_policy_activation" "test" {
  policy_id = 1234
  network   = "staging"
  version   = 2
}

output "status" {
  value = akamai_cloudlets_policy_activation.test.status
}