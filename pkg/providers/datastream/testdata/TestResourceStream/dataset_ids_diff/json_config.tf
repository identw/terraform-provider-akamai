provider "akamai" {
  edgerc = "../../test/edgerc"
}

resource "akamai_datastream" "splunk_stream" {
  active = false
  config {
    format = "JSON"
    frequency {
      time_in_sec = 30
    }
  }

  contract_id        = "test_contract"
  dataset_fields_ids = [1001, 1002]

  group_id      = 1337
  property_ids  = [1]
  stream_name   = "test_stream"
  stream_type   = "RAW_LOGS"
  template_name = "EDGE_LOGS"

  splunk_connector {
    compress_logs         = false
    connector_name        = "splunk_test_connector_name"
    event_collector_token = "splunk_event_collector_token"
    url                   = "splunk_url"
  }
}