provider "akamai" {
  edgerc        = "../../common/testutils/edgerc"
  cache_enabled = false
}

data "akamai_botman_bot_category_exception" "test" {
  config_id          = 43253
  security_policy_id = "AAAA_81230"
}