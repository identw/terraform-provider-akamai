provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

data "akamai_gtm_cidrmap" "gtm_cidrmap" {
  map_name = "mapTest"
}