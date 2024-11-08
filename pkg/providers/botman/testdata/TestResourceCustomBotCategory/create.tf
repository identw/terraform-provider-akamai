provider "akamai" {
  edgerc        = "../../test/edgerc"
  cache_enabled = false
}

resource "akamai_botman_custom_bot_category" "test" {
  config_id = 43253
  custom_bot_category = jsonencode(
    {
      "testKey" : "testValue3"
    }
  )
}