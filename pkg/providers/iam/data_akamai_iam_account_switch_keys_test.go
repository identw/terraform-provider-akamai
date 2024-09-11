package iam

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v8/pkg/iam"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestAccountSwitchKeys(t *testing.T) {
	expectListAccountSwitchKeys := func(client *iam.Mock, clientID string, filter string, timesToRun int) {
		accountSwitchKeysResponse := iam.ListAccountSwitchKeysResponse{
			{
				AccountName:      "Internet Company",
				AccountSwitchKey: "1-EFGH",
			},
			{
				AccountName:      "Internet Company",
				AccountSwitchKey: "1-ABCD:Z-XYZ",
			},
			{
				AccountName:      "Digital Company",
				AccountSwitchKey: "1-ABCD:Z-PQR",
			},
		}
		client.On("ListAccountSwitchKeys", mock.Anything, iam.ListAccountSwitchKeysRequest{
			ClientID: clientID,
			Search:   filter,
		}).Return(accountSwitchKeysResponse, nil).Times(timesToRun)
	}

	expectListAccountSwitchKeysWithError := func(client *iam.Mock, timesToRun int) {
		client.On("ListAccountSwitchKeys", mock.Anything, iam.ListAccountSwitchKeysRequest{
			ClientID: "",
			Search:   "",
		}).Return(nil, fmt.Errorf("list account switch keys failed")).Times(timesToRun)
	}

	tests := map[string]struct {
		givenTF string
		init    func(*iam.Mock)
		error   *regexp.Regexp
	}{
		"happy path": {
			givenTF: "default.tf",
			init: func(m *iam.Mock) {
				expectListAccountSwitchKeys(m, "", "", 3)
			},
		},
		"happy path with correct filter": {
			givenTF: "default_correct_filter.tf",
			init: func(m *iam.Mock) {
				expectListAccountSwitchKeys(m, "XYZ", "ABC", 3)
			},
		},
		"incorrect filter": {
			givenTF: "incorrect_filter.tf",
			error:   regexp.MustCompile("Attribute filter string length must be at least 3, got: 2"),
		},
		"error listing account switch keys": {
			givenTF: "default.tf",
			init: func(m *iam.Mock) {
				expectListAccountSwitchKeysWithError(m, 1)
			},
			error: regexp.MustCompile("list account switch keys failed"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := &iam.Mock{}
			if test.init != nil {
				test.init(client)
			}

			useClient(client, func() {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
					IsUnitTest:               true,
					Steps: []resource.TestStep{
						{
							Config:      testutils.LoadFixtureString(t, fmt.Sprintf("testdata/TestDataAccountSwitchKeys/%s", test.givenTF)),
							Check:       checkAccountSwitchKeysAttrs(),
							ExpectError: test.error,
						},
					},
				})
			})
			client.AssertExpectations(t)
		})
	}
}

func checkAccountSwitchKeysAttrs() resource.TestCheckFunc {
	var checkFuncs []resource.TestCheckFunc

	checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr("data.akamai_iam_account_switch_keys.test", "account_switch_keys.#", "3"))
	checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr("data.akamai_iam_account_switch_keys.test", "account_switch_keys.0.account_name", "Internet Company"))
	checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr("data.akamai_iam_account_switch_keys.test", "account_switch_keys.0.account_switch_key", "1-EFGH"))
	checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr("data.akamai_iam_account_switch_keys.test", "account_switch_keys.1.account_name", "Internet Company"))
	checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr("data.akamai_iam_account_switch_keys.test", "account_switch_keys.1.account_switch_key", "1-ABCD:Z-XYZ"))
	checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr("data.akamai_iam_account_switch_keys.test", "account_switch_keys.2.account_name", "Digital Company"))
	checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr("data.akamai_iam_account_switch_keys.test", "account_switch_keys.2.account_switch_key", "1-ABCD:Z-PQR"))

	return resource.ComposeAggregateTestCheckFunc(checkFuncs...)
}