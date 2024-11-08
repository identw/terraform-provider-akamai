provider "akamai" {
  edgerc = "../../test/edgerc"
}

resource "akamai_edgeworkers_activation" "test" {
  edgeworker_id = 1234
  network       = "STAGING"
  version       = "test"
  note          = "note for edgeworkers activation"
}