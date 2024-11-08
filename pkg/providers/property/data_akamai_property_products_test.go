package property

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/papi"
)

func TestVerifyProductsDataSourceSchema(t *testing.T) {
	t.Run("akamai_property_products - test data source required contract", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV5ProviderFactories: testAccProviders,
			IsUnitTest:               true,
			Steps: []resource.TestStep{{
				Config:      testConfig(""),
				ExpectError: regexp.MustCompile("The argument \"contract_id\" is required, but no definition was found"),
			}},
		})
	})
}

func TestOutputProductsDataSource(t *testing.T) {

	t.Run("akamai_property_products - input OK - output OK", func(t *testing.T) {
		client := &papi.Mock{}
		client.On("GetProducts", AnyCTX, mock.Anything).Return(&papi.GetProductsResponse{
			AccountID:  "act_anyAccount",
			ContractID: "ctr_AnyContract",
			Products: papi.ProductsItems{
				Items: []papi.ProductItem{{ProductName: "anyProduct", ProductID: "prd_anyProduct"}},
			},
		}, nil)

		useClient(client, nil, func() {
			resource.UnitTest(t, resource.TestCase{
				ProtoV5ProviderFactories: testAccProviders,
				IsUnitTest:               true,
				Steps: []resource.TestStep{{
					Config: testConfig("contract_id = \"ctr_test\""),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckOutput("product_name0", "anyProduct"),
						resource.TestCheckOutput("product_id0", "prd_anyProduct"),
					),
				}},
			})
		})
	})
}

func testConfig(contractIDConfig string) string {
	return fmt.Sprintf(`
	provider "akamai" {
		edgerc = "../../test/edgerc"
	}

	data "akamai_property_products" "example" { %s }

    output "product_name0" {
		value = "${data.akamai_property_products.example.products[0].product_name}"
	}

    output "product_id0" {
		value = "${data.akamai_property_products.example.products[0].product_id}"
	}
`, contractIDConfig)
}
