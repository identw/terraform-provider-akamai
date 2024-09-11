package iam

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v8/pkg/iam"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/mock"
)

func TestResourceIAMIPAllowlistResource(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		configPath string
		init       func(*testing.T, *iam.Mock)
		steps      []resource.TestStep
		error      *regexp.Regexp
	}{
		"create - enable": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create
				mockReadIPAllowlistStatus(m, false)
				mockEnableIPAllowlist(m)
				// step 2 read
				mockReadIPAllowlistStatus(m, true)
				// step 3 delete - remove resource form state(no mock)
			},
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/enable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "true")),
				},
			},
		},
		"create - disable": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create
				mockReadIPAllowlistStatus(m, true)
				mockDisableIPAllowlist(m)
				// step 2 read
				mockReadIPAllowlistStatus(m, false)
				// step 3 delete - remove resource form state(no mock)
			},
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/disable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "false")),
				},
			},
		},
		"create - already enabled on server": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create - ip allowlist already enabled on server
				mockReadIPAllowlistStatus(m, true)
				// step 2 read
				mockReadIPAllowlistStatus(m, true)
				// step 3 delete - remove resource form state(no mock)
			},
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/enable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "true")),
				},
			},
		},
		"create - already disabled on server": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create - ip allowlist already disabled on server
				mockReadIPAllowlistStatus(m, false)
				// step 2 read
				mockReadIPAllowlistStatus(m, false)
				// step 3 delete - remove resource form state(no mock)
			},
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/disable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "false")),
				},
			},
		},
		"update - enable": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create
				mockReadIPAllowlistStatus(m, true)
				mockDisableIPAllowlist(m)
				// step 2 refresh
				mockReadIPAllowlistStatus(m, false)
				mockReadIPAllowlistStatus(m, false)
				// step 3 update
				mockReadIPAllowlistStatus(m, false)
				mockEnableIPAllowlist(m)
				// step 4 refresh
				mockReadIPAllowlistStatus(m, true)
				// step 5 delete - remove resource form state(no mock)
			},
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/disable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "false")),
				},
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/enable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "true")),
				},
			},
		},
		"update - disable": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create
				mockReadIPAllowlistStatus(m, false)
				mockEnableIPAllowlist(m)
				// step 2 refresh
				mockReadIPAllowlistStatus(m, true)
				mockReadIPAllowlistStatus(m, true)
				// step 3 update
				mockReadIPAllowlistStatus(m, true)
				mockDisableIPAllowlist(m)
				// step 4 refresh
				mockReadIPAllowlistStatus(m, false)
				// step 5 delete - remove resource form state(no mock)
			},
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/enable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "true")),
				},
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/disable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "false")),
				},
			},
		},
		"error - enable error ip not on allowlist": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create - error
				mockReadIPAllowlistStatus(m, false)
				mockEnableIPAllowlistError(m)
			},
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/enable.tf"),
					ExpectError: regexp.MustCompile("enable ip allowlist fail"),
				},
			},
		},
		"error - disable error IP not on allowlist": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create - error
				mockReadIPAllowlistStatus(m, true)
				mockDisableIPAllowlistError(m)
			},
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/disable.tf"),
					ExpectError: regexp.MustCompile("disable IP allowlist fail"),
				},
			},
		},
		"import": {
			init: func(t *testing.T, m *iam.Mock) {
				// step 1 create
				mockReadIPAllowlistStatus(m, false)
				mockEnableIPAllowlist(m)
				// step 2 refresh
				mockReadIPAllowlistStatus(m, true)

				// step 3 import
				mockReadIPAllowlistStatus(m, true)
				// step 4 refresh
				mockReadIPAllowlistStatus(m, true)
			},
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "./testdata/TestResIPAllowlist/enable.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_iam_ip_allowlist.test", "enable", "true")),
				},
				{
					ImportState:       true,
					ImportStateVerify: false,
					ResourceName:      "akamai_iam_ip_allowlist.test",
					ImportStateCheck:  checkImportEnabledIPAllowlistForSpecificUser(),
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := &iam.Mock{}
			if test.init != nil {
				test.init(t, client)
			}
			useClient(client, func() {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
					IsUnitTest:               true,
					Steps:                    test.steps,
				})
			})
			client.AssertExpectations(t)
		})
	}
}

func checkImportEnabledIPAllowlistForSpecificUser() resource.ImportStateCheckFunc {
	return func(s []*terraform.InstanceState) error {
		if len(s) == 0 {
			return errors.New("No Instance found")
		}
		if len(s) != 1 {
			return fmt.Errorf("Expected one Instance: %d", len(s))
		}

		state := s[0].Attributes

		attributes := map[string]string{
			"enable": "true",
		}

		invalidValues := []string{}
		for field, expectedVal := range attributes {
			if state[field] != expectedVal {
				invalidValues = append(invalidValues, fmt.Sprintf("field: %s, got: %s, expected: %s ", field, state[field], expectedVal))
			}
		}
		return nil
	}
}

func mockEnableIPAllowlist(m *iam.Mock) *mock.Call {
	return m.On("EnableIPAllowlist", mock.Anything).Return(nil).Once()
}

func mockEnableIPAllowlistError(m *iam.Mock) *mock.Call {
	return m.On("EnableIPAllowlist", mock.Anything).Return(iam.ErrEnableIPAllowlist).Once()
}

func mockDisableIPAllowlist(m *iam.Mock) *mock.Call {
	return m.On("DisableIPAllowlist", mock.Anything).Return(nil).Once()
}
func mockDisableIPAllowlistError(m *iam.Mock) *mock.Call {
	return m.On("DisableIPAllowlist", mock.Anything).Return(iam.ErrDisableIPAllowlist).Once()
}

func mockReadIPAllowlistStatus(m *iam.Mock, enabled bool) *mock.Call {
	if enabled {
		return m.On("GetIPAllowlistStatus", mock.Anything).Return(&iam.GetIPAllowlistStatusResponse{Enabled: true}, nil).Once()
	}
	return m.On("GetIPAllowlistStatus", mock.Anything).Return(&iam.GetIPAllowlistStatusResponse{Enabled: false}, nil).Once()
}