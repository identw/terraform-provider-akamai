provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

data "akamai_gtm_asmap" "my_gtm_asmap" {
  map_name = "map1"
}