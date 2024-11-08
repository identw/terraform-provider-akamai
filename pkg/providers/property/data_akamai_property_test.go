package property

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/papi"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestDataProperty(t *testing.T) {
	tests := map[string]struct {
		givenTF            string
		init               func(*papi.Mock)
		expectedAttributes map[string]string
		withError          *regexp.Regexp
	}{
		"valid rules, no version provided": {
			givenTF: "no_version.tf",
			init: func(m *papi.Mock) {
				m.On("SearchProperties", mock.Anything, papi.SearchRequest{
					Key:   papi.SearchKeyPropertyName,
					Value: "property_name",
				}).Return(&papi.SearchResponse{
					Versions: papi.SearchItems{
						Items: []papi.SearchItem{
							{
								ContractID: "ctr_1",
								GroupID:    "grp_1",
								PropertyID: "prp_123",
							},
						},
					},
				}, nil)
				m.On("GetProperty", mock.Anything, papi.GetPropertyRequest{
					ContractID: "ctr_1",
					GroupID:    "grp_1",
					PropertyID: "prp_123",
				}).Return(&papi.GetPropertyResponse{
					Properties: papi.PropertiesItems{Items: []*papi.Property{
						{
							PropertyID:    "prp_123",
							LatestVersion: 1,
							ContractID:    "ctr_1",
							GroupID:       "grp_1",
						},
					}},
				}, nil)
				m.On("GetRuleTree", mock.Anything, papi.GetRuleTreeRequest{
					PropertyID:      "prp_123",
					PropertyVersion: 1,
					ContractID:      "ctr_1",
					GroupID:         "grp_1",
				}).Return(&papi.GetRuleTreeResponse{
					Response: papi.Response{
						ContractID: "ctr_1",
						GroupID:    "grp_1",
					},
					PropertyID:      "prp_123",
					PropertyVersion: 1,
					Rules: papi.Rules{
						Behaviors: []papi.RuleBehavior{
							{
								Name: "beh 1",
							},
						},
						Name:                "rule 1",
						CriteriaMustSatisfy: "all",
					},
				}, nil)
			},
			expectedAttributes: map[string]string{
				"name":  "property_name",
				"rules": compactJSON(testutils.LoadFixtureBytes(t, "testdata/TestDataProperty/no_version_rules.json")),
			},
		},
		"valid rules, with version provided": {
			givenTF: "with_version.tf",
			init: func(m *papi.Mock) {
				m.On("SearchProperties", mock.Anything, papi.SearchRequest{
					Key:   papi.SearchKeyPropertyName,
					Value: "property_name",
				}).Return(&papi.SearchResponse{
					Versions: papi.SearchItems{
						Items: []papi.SearchItem{
							{
								ContractID: "ctr_1",
								GroupID:    "grp_1",
								PropertyID: "prp_123",
							},
						},
					},
				}, nil)
				m.On("GetProperty", mock.Anything, papi.GetPropertyRequest{
					ContractID: "ctr_1",
					GroupID:    "grp_1",
					PropertyID: "prp_123",
				}).Return(&papi.GetPropertyResponse{
					Properties: papi.PropertiesItems{Items: []*papi.Property{
						{
							PropertyID:    "prp_123",
							LatestVersion: 1,
							ContractID:    "ctr_1",
							GroupID:       "grp_1",
						},
					}},
				}, nil)
				m.On("GetRuleTree", mock.Anything, papi.GetRuleTreeRequest{
					PropertyID:      "prp_123",
					PropertyVersion: 2,
					ContractID:      "ctr_1",
					GroupID:         "grp_1",
				}).Return(&papi.GetRuleTreeResponse{
					Response: papi.Response{
						ContractID: "ctr_1",
						GroupID:    "grp_1",
					},
					PropertyID:      "prp_123",
					PropertyVersion: 2,
					Rules: papi.Rules{
						Behaviors: []papi.RuleBehavior{
							{
								Name: "beh 1",
							},
						},
						Name:                "rule 1",
						CriteriaMustSatisfy: "all",
					},
				}, nil)
			},
			expectedAttributes: map[string]string{
				"name":  "property_name",
				"rules": compactJSON(testutils.LoadFixtureBytes(t, "testdata/TestDataProperty/with_version_rules.json")),
			},
		},
		"error searching for property": {
			givenTF: "with_version.tf",
			init: func(m *papi.Mock) {
				m.On("SearchProperties", mock.Anything, papi.SearchRequest{
					Key:   papi.SearchKeyPropertyName,
					Value: "property_name",
				}).Return(nil, fmt.Errorf("oops"))
			},
			withError: regexp.MustCompile("oops"),
		},
		"error fetching property": {
			givenTF: "with_version.tf",
			init: func(m *papi.Mock) {
				m.On("SearchProperties", mock.Anything, papi.SearchRequest{
					Key:   papi.SearchKeyPropertyName,
					Value: "property_name",
				}).Return(&papi.SearchResponse{
					Versions: papi.SearchItems{
						Items: []papi.SearchItem{
							{
								ContractID: "ctr_1",
								GroupID:    "grp_1",
								PropertyID: "prp_123",
							},
						},
					},
				}, nil)
				m.On("GetProperty", mock.Anything, papi.GetPropertyRequest{
					ContractID: "ctr_1",
					GroupID:    "grp_1",
					PropertyID: "prp_123",
				}).Return(nil, fmt.Errorf("oops"))
			},
			withError: regexp.MustCompile("oops"),
		},
		"property not found": {
			givenTF: "with_version.tf",
			init: func(m *papi.Mock) {
				m.On("SearchProperties", mock.Anything, papi.SearchRequest{
					Key:   papi.SearchKeyPropertyName,
					Value: "property_name",
				}).Return(&papi.SearchResponse{
					Versions: papi.SearchItems{
						Items: []papi.SearchItem{},
					},
				}, nil)
			},
			withError: regexp.MustCompile("property not found"),
		},
		"error fetching rules": {
			givenTF: "with_version.tf",
			init: func(m *papi.Mock) {
				m.On("SearchProperties", mock.Anything, papi.SearchRequest{
					Key:   papi.SearchKeyPropertyName,
					Value: "property_name",
				}).Return(&papi.SearchResponse{
					Versions: papi.SearchItems{
						Items: []papi.SearchItem{
							{
								ContractID: "ctr_1",
								GroupID:    "grp_1",
								PropertyID: "prp_123",
							},
						},
					},
				}, nil)
				m.On("GetProperty", mock.Anything, papi.GetPropertyRequest{
					ContractID: "ctr_1",
					GroupID:    "grp_1",
					PropertyID: "prp_123",
				}).Return(&papi.GetPropertyResponse{
					Properties: papi.PropertiesItems{Items: []*papi.Property{
						{
							PropertyID:    "prp_123",
							LatestVersion: 1,
							ContractID:    "ctr_1",
							GroupID:       "grp_1",
						},
					}},
				}, nil)
				m.On("GetRuleTree", mock.Anything, papi.GetRuleTreeRequest{
					PropertyID:      "prp_123",
					PropertyVersion: 2,
					ContractID:      "ctr_1",
					GroupID:         "grp_1",
				}).Return(nil, fmt.Errorf("oops"))
			},
			withError: regexp.MustCompile("property rules not found"),
		},
		"error name not provided": {
			givenTF:   "no_name.tf",
			init:      func(m *papi.Mock) {},
			withError: regexp.MustCompile("Missing required argument"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := &papi.Mock{}
			test.init(client)
			var checkFuncs []resource.TestCheckFunc
			for k, v := range test.expectedAttributes {
				checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr("data.akamai_property.prop", k, v))
			}
			useClient(client, nil, func() {
				resource.Test(t, resource.TestCase{
					IsUnitTest:               true,
					ProtoV5ProviderFactories: testAccProviders,
					Steps: []resource.TestStep{{
						Config:      testutils.LoadFixtureString(t, fmt.Sprintf("testdata/TestDataProperty/%s", test.givenTF)),
						Check:       resource.ComposeAggregateTestCheckFunc(checkFuncs...),
						ExpectError: test.withError,
					},
					},
				})
			})
			client.AssertExpectations(t)
		})
	}
}

func testAccDataSourcePropertyBasic() string {
	return `
	provider "akamai" {
		papi_section = "papi"
	  }

data "akamai_property" "test" {
	name = "terraform-test-datasource"
	version = 1
}
`
}

func testAccCheckDataSourcePropertyDestroy(_ *terraform.State) error {
	log.Printf("[DEBUG] [Group] Searching for Property Delete skipped ")

	return nil
}
