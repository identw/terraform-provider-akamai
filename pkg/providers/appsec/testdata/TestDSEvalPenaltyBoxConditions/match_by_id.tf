provider "akamai" {
  edgerc        = "../../test/edgerc"
  cache_enabled = false
}

data "akamai_appsec_eval_penalty_box_conditions" "test" {
  config_id          = 43253
  security_policy_id = "AAAA_81230"
}


