package dns

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/dns"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestDataSourceDNSRecordSet_basic(t *testing.T) {

	t.Run("basic", func(t *testing.T) {
		client := &dns.Mock{}

		dataSourceName := "data.akamai_dns_record_set.test"
		outputName := "test_addrs"

		rdata := []string{"10.1.0.1", "10.2.0.1"}

		client.On("GetRdata",
			mock.Anything, // ctx is irrelevant for this test
			"exampleterraform.io",
			"exampleterraform.io",
			"A",
		).Return(rdata, nil)

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config: testutils.LoadFixtureString(t, "testdata/TestDataDnsRecordSet/basic.tf"),
						Check: resource.ComposeTestCheckFunc(
							// check the values set in dataSourceDNSRecordSetRead
							// rdata is an array that becomes rdata.0 and rdata.1 in tf state
							resource.TestCheckResourceAttrSet(dataSourceName, "rdata.0"),
							resource.TestCheckResourceAttrSet(dataSourceName, "rdata.1"),
							resource.TestCheckResourceAttrSet(dataSourceName, "id"),
							resource.TestCheckOutput(outputName, strings.Join(rdata, ",")),
						),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		client := &dns.Mock{}

		client.On("GetRdata",
			mock.Anything, // ctx is irrelevant for this test
			"exampleterraform.io",
			"exampleterraform.io",
			"A",
		).Return(nil, errors.New("invalid zone"))

		useClient(client, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testutils.NewProtoV6ProviderFactory(NewSubprovider()),
				Steps: []resource.TestStep{
					{
						Config:      testutils.LoadFixtureString(t, "testdata/TestDataDnsRecordSet/basic.tf"),
						ExpectError: regexp.MustCompile(`invalid zone`),
					},
				},
			})
		})

		client.AssertExpectations(t)
	})
}
