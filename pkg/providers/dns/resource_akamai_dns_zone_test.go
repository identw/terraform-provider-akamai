package dns

import (
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v9/pkg/dns"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResDNSZone(t *testing.T) {
	zone := &dns.GetZoneResponse{
		ContractID:      "ctr1",
		Zone:            "primaryexampleterraform.io",
		Type:            "primary",
		Comment:         "This is a test primary zone",
		SignAndServe:    false,
		ActivationState: "PENDING",
	}
	recordSetsResp := &dns.GetRecordSetsResponse{
		RecordSets: make([]dns.RecordSet, 2, 2),
	}

	t.Run("when group is not provided and there is no group for the user ", func(t *testing.T) {
		client := &dns.Mock{}

		client.On("ListGroups",
			mock.Anything,
			mock.AnythingOfType("dns.ListGroupRequest"),
		).Return(&dns.ListGroupResponse{}, nil)

		// work around to skip Delete which fails intentionally
		err := os.Setenv("DNS_ZONE_SKIP_DELETE", "")
		require.NoError(t, err)
		defer func() {
			err = os.Unsetenv("DNS_ZONE_SKIP_DELETE")
			require.NoError(t, err)
		}()
		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestResDnsZone/create_without_group.tf"),
						ExpectError: regexp.MustCompile("no group found. Please provide the group."),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	// This test performs a full life-cycle (CRUD) test
	t.Run("lifecycle test when group is not found and no. of group is 1", func(t *testing.T) {
		client := &dns.Mock{}
		groupListResponse := &dns.ListGroupResponse{
			Groups: []dns.Group{
				{
					GroupID:   1,
					GroupName: "name",
					ContractIDs: []string{
						"1", "2",
					},
					Permissions: []string{
						"DELETE", "READ", "WRITE", "ADD",
					},
				},
			},
		}

		client.On("ListGroups",
			mock.Anything,
			mock.AnythingOfType("dns.ListGroupRequest"),
		).Return(groupListResponse, nil)

		getCall := client.On("GetZone",
			mock.Anything,
			mock.AnythingOfType("dns.GetZoneRequest"),
		).Return(nil, &dns.Error{
			StatusCode: http.StatusNotFound,
		})

		client.On("CreateZone",
			mock.Anything,
			mock.AnythingOfType("dns.CreateZoneRequest"),
		).Return(nil).Run(func(args mock.Arguments) {
			getCall.ReturnArguments = mock.Arguments{zone, nil}
		})

		client.On("SaveChangeList",
			mock.Anything,
			mock.AnythingOfType("dns.SaveChangeListRequest"),
		).Return(nil)

		client.On("SubmitChangeList",
			mock.Anything,
			mock.AnythingOfType("dns.SubmitChangeListRequest"),
		).Return(nil)

		client.On("GetRecordSets",
			mock.Anything,
			mock.AnythingOfType("dns.GetRecordSetsRequest"),
		).Return(recordSetsResp, nil)

		dataSourceName := "akamai_dns_zone.test_without_group"

		// work around to skip Delete which fails intentionally
		err := os.Setenv("DNS_ZONE_SKIP_DELETE", "")
		require.NoError(t, err)
		defer func() {
			err = os.Unsetenv("DNS_ZONE_SKIP_DELETE")
			require.NoError(t, err)
		}()
		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResDnsZone/create_without_group.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "zone", "primaryexampleterraform.io"),
							resource.TestCheckResourceAttr(dataSourceName, "contract", "ctr1"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("when group is not provided and no. of group is more than 1 for the user ", func(t *testing.T) {
		client := &dns.Mock{}
		groupListResponse := &dns.ListGroupResponse{
			Groups: []dns.Group{
				{
					GroupID:   1,
					GroupName: "name",
					ContractIDs: []string{
						"1", "2",
					},
					Permissions: []string{
						"DELETE", "READ", "WRITE", "ADD",
					},
				},
				{
					GroupID:   2,
					GroupName: "name",
					ContractIDs: []string{
						"1", "2",
					},
					Permissions: []string{
						"DELETE", "READ", "WRITE", "ADD",
					},
				},
			},
		}

		client.On("ListGroups",
			mock.Anything,
			mock.AnythingOfType("dns.ListGroupRequest"),
		).Return(groupListResponse, nil)

		// work around to skip Delete which fails intentionally
		err := os.Setenv("DNS_ZONE_SKIP_DELETE", "")
		require.NoError(t, err)
		defer func() {
			err = os.Unsetenv("DNS_ZONE_SKIP_DELETE")
			require.NoError(t, err)
		}()
		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestResDnsZone/create_without_group.tf"),
						ExpectError: regexp.MustCompile("group is a required field when there is more than one group present."),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	// This test performs a full life-cycle (CRUD) test
	t.Run("lifecycle test with group", func(t *testing.T) {
		client := &dns.Mock{}

		getCall := client.On("GetZone",
			mock.Anything,
			mock.AnythingOfType("dns.GetZoneRequest"),
		).Return(nil, &dns.Error{
			StatusCode: http.StatusNotFound,
		})

		client.On("CreateZone",
			mock.Anything,
			mock.AnythingOfType("dns.CreateZoneRequest"),
		).Return(nil).Run(func(args mock.Arguments) {
			getCall.ReturnArguments = mock.Arguments{zone, nil}
		})

		client.On("UpdateZone",
			mock.Anything,
			mock.AnythingOfType("dns.UpdateZoneRequest"),
		).Return(nil).Run(func(args mock.Arguments) {
			zone.Comment = "This is an updated test primary zone"
		})

		client.On("SaveChangeList",
			mock.Anything,
			mock.AnythingOfType("dns.SaveChangeListRequest"),
		).Return(nil)

		client.On("SubmitChangeList",
			mock.Anything,
			mock.AnythingOfType("dns.SubmitChangeListRequest"),
		).Return(nil)

		client.On("GetRecordSets",
			mock.Anything,
			mock.AnythingOfType("dns.GetRecordSetsRequest"),
		).Return(recordSetsResp, nil)

		dataSourceName := "akamai_dns_zone.primary_test_zone"

		// work around to skip Delete which fails intentionally
		err := os.Setenv("DNS_ZONE_SKIP_DELETE", "")
		require.NoError(t, err)
		defer func() {
			err = os.Unsetenv("DNS_ZONE_SKIP_DELETE")
			require.NoError(t, err)
		}()
		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResDnsZone/create_primary.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "zone", "primaryexampleterraform.io"),
							resource.TestCheckResourceAttr(dataSourceName, "contract", "ctr1"),
							resource.TestCheckResourceAttr(dataSourceName, "comment", "This is a test primary zone"),
							resource.TestCheckResourceAttr(dataSourceName, "group", "grp1"),
						),
					},
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestResDnsZone/update_primary.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(dataSourceName, "zone", "primaryexampleterraform.io"),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})
}
