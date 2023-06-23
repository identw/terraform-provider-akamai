provider "akamai" {
  edgerc        = "../../common/testutils/edgerc"
  cache_enabled = false
}

resource "akamai_botman_custom_bot_category_sequence" "test" {
  config_id    = 43253
  category_ids = ["d79285df-e399-43e8-bb0f-c0d980a88e4f", "afa309b8-4fd5-430e-a061-1c61df1d2ac2", "cc9c3f89-e179-4892-89cf-d5e623ba9dc7"]
}