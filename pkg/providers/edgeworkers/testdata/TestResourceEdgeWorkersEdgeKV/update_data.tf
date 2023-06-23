provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

resource "akamai_edgekv" "test" {
  namespace_name       = "DevExpTest"
  network              = "staging"
  group_id             = 1234
  retention_in_seconds = 86401
  initial_data {
    key   = "new"
    value = "value"
  }
}