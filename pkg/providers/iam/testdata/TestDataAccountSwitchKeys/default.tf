provider "akamai" {
  edgerc = "../../common/testutils/edgerc"
}

data "akamai_iam_account_switch_keys" "test" {}