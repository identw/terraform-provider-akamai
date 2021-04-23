package appsec

import (
	"encoding/json"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/mock"
)

func TestAccAkamaiMatchTarget_res_basic(t *testing.T) {
	t.Run("match by MatchTarget ID", func(t *testing.T) {
		client := &mockappsec{}

		cu := appsec.UpdateMatchTargetResponse{}
		expectJSU := compactJSON(loadFixtureBytes("testdata/TestResMatchTarget/MatchTargetUpdated.json"))
		json.Unmarshal([]byte(expectJSU), &cu)

		cr := appsec.GetMatchTargetResponse{}
		expectJS := compactJSON(loadFixtureBytes("testdata/TestResMatchTarget/MatchTarget.json"))
		json.Unmarshal([]byte(expectJS), &cr)

		crmt := appsec.CreateMatchTargetResponse{}
		expectJSMT := compactJSON(loadFixtureBytes("testdata/TestResMatchTarget/MatchTargetCreated.json"))
		json.Unmarshal([]byte(expectJSMT), &crmt)

		rmmt := appsec.RemoveMatchTargetResponse{}
		expectJSRMT := compactJSON(loadFixtureBytes("testdata/TestResMatchTarget/MatchTargetCreated.json"))
		json.Unmarshal([]byte(expectJSRMT), &rmmt)

		config := appsec.GetConfigurationResponse{}
		expectConfigs := compactJSON(loadFixtureBytes("testdata/TestResConfiguration/LatestConfiguration.json"))
		json.Unmarshal([]byte(expectConfigs), &config)

		client.On("GetConfiguration",
			mock.Anything,
			appsec.GetConfigurationRequest{ConfigID: 43253},
		).Return(&config, nil)

		client.On("GetMatchTarget",
			mock.Anything, // ctx is irrelevant for this test
			appsec.GetMatchTargetRequest{ConfigID: 43253, ConfigVersion: 7, TargetID: 3008967},
		).Return(&cr, nil)

		client.On("CreateMatchTarget",
			mock.Anything, // ctx is irrelevant for this test
			appsec.CreateMatchTargetRequest{Type: "", ConfigID: 43253, ConfigVersion: 7, JsonPayloadRaw: json.RawMessage{0x20, 0x20, 0x20, 0x7b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x77, 0x65, 0x62, 0x73, 0x69, 0x74, 0x65, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x49, 0x64, 0x22, 0x3a, 0x20, 0x34, 0x33, 0x32, 0x35, 0x33, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3a, 0x20, 0x20, 0x31, 0x35, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x4e, 0x4f, 0x5f, 0x4d, 0x41, 0x54, 0x43, 0x48, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x7b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x42, 0x6f, 0x74, 0x6d, 0x61, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x52, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x52, 0x65, 0x70, 0x75, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x53, 0x6c, 0x6f, 0x77, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0xa, 0x20, 0x20, 0x20, 0x20, 0x7d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x66, 0x69, 0x6c, 0x65, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x3a, 0x20, 0x5b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x63, 0x61, 0x72, 0x62, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x70, 0x63, 0x74, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x70, 0x64, 0x66, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x73, 0x77, 0x66, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x63, 0x63, 0x74, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x6a, 0x70, 0x65, 0x67, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x6a, 0x73, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x77, 0x6d, 0x6c, 0x73, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x68, 0x64, 0x6d, 0x6c, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x70, 0x77, 0x73, 0x22, 0xa, 0x20, 0x20, 0x20, 0x20, 0x5d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x66, 0x69, 0x6c, 0x65, 0x50, 0x61, 0x74, 0x68, 0x73, 0x22, 0x3a, 0x20, 0x5b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x2f, 0x63, 0x61, 0x63, 0x68, 0x65, 0x2f, 0x61, 0x61, 0x61, 0x62, 0x62, 0x63, 0x2a, 0x22, 0xa, 0x20, 0x20, 0x20, 0x20, 0x5d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x3a, 0x20, 0x5b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x6d, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x77, 0x77, 0x77, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x6e, 0x65, 0x74, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x22, 0xa, 0x20, 0x20, 0x20, 0x20, 0x5d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x69, 0x73, 0x4e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x69, 0x73, 0x4e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x50, 0x61, 0x74, 0x68, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x22, 0x3a, 0x20, 0x7b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x49, 0x64, 0x22, 0x3a, 0x20, 0x22, 0x41, 0x41, 0x41, 0x41, 0x5f, 0x38, 0x31, 0x32, 0x33, 0x30, 0x22, 0xa, 0x20, 0x20, 0x20, 0x20, 0x7d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x3a, 0x20, 0x31, 0xa, 0x7d, 0xa}},
		).Return(&crmt, nil)

		client.On("UpdateMatchTarget",
			mock.Anything, // ctx is irrelevant for this test
			appsec.UpdateMatchTargetRequest{ConfigID: 43253, ConfigVersion: 7, JsonPayloadRaw: json.RawMessage{0x20, 0x20, 0x20, 0x7b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x74, 0x79, 0x70, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x77, 0x65, 0x62, 0x73, 0x69, 0x74, 0x65, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x49, 0x64, 0x22, 0x3a, 0x20, 0x34, 0x33, 0x32, 0x35, 0x33, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x3a, 0x20, 0x20, 0x31, 0x35, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x64, 0x65, 0x66, 0x61, 0x75, 0x6c, 0x74, 0x46, 0x69, 0x6c, 0x65, 0x22, 0x3a, 0x20, 0x22, 0x4e, 0x4f, 0x5f, 0x4d, 0x41, 0x54, 0x43, 0x48, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x53, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x7b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x41, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x42, 0x6f, 0x74, 0x6d, 0x61, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x4c, 0x61, 0x79, 0x65, 0x72, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x52, 0x61, 0x74, 0x65, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x52, 0x65, 0x70, 0x75, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x61, 0x70, 0x70, 0x6c, 0x79, 0x53, 0x6c, 0x6f, 0x77, 0x50, 0x6f, 0x73, 0x74, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x73, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0xa, 0x20, 0x20, 0x20, 0x20, 0x7d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x66, 0x69, 0x6c, 0x65, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x22, 0x3a, 0x20, 0x5b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x63, 0x61, 0x72, 0x62, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x70, 0x63, 0x74, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x70, 0x64, 0x66, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x73, 0x77, 0x66, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x63, 0x63, 0x74, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x6a, 0x70, 0x65, 0x67, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x6a, 0x73, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x77, 0x6d, 0x6c, 0x73, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x68, 0x64, 0x6d, 0x6c, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x70, 0x77, 0x73, 0x22, 0xa, 0x20, 0x20, 0x20, 0x20, 0x5d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x66, 0x69, 0x6c, 0x65, 0x50, 0x61, 0x74, 0x68, 0x73, 0x22, 0x3a, 0x20, 0x5b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x2f, 0x63, 0x61, 0x63, 0x68, 0x65, 0x2f, 0x61, 0x61, 0x61, 0x62, 0x62, 0x63, 0x2a, 0x22, 0xa, 0x20, 0x20, 0x20, 0x20, 0x5d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x3a, 0x20, 0x5b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x6d, 0x31, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x77, 0x77, 0x77, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x6e, 0x65, 0x74, 0x22, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x2e, 0x63, 0x6f, 0x6d, 0x22, 0xa, 0x20, 0x20, 0x20, 0x20, 0x5d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x69, 0x73, 0x4e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x46, 0x69, 0x6c, 0x65, 0x45, 0x78, 0x74, 0x65, 0x6e, 0x73, 0x69, 0x6f, 0x6e, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x69, 0x73, 0x4e, 0x65, 0x67, 0x61, 0x74, 0x69, 0x76, 0x65, 0x50, 0x61, 0x74, 0x68, 0x4d, 0x61, 0x74, 0x63, 0x68, 0x22, 0x3a, 0x20, 0x66, 0x61, 0x6c, 0x73, 0x65, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x22, 0x3a, 0x20, 0x7b, 0xa, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x22, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x49, 0x64, 0x22, 0x3a, 0x20, 0x22, 0x41, 0x41, 0x41, 0x41, 0x5f, 0x38, 0x31, 0x32, 0x33, 0x30, 0x22, 0xa, 0x20, 0x20, 0x20, 0x20, 0x7d, 0x2c, 0xa, 0x20, 0x20, 0x20, 0x20, 0x22, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x22, 0x3a, 0x20, 0x31, 0xa, 0x7d, 0xa}, TargetID: 3008967},
		).Return(&cu, nil)

		client.On("RemoveMatchTarget",
			mock.Anything, // ctx is irrelevant for this test
			appsec.RemoveMatchTargetRequest{ConfigID: 43253, ConfigVersion: 7, TargetID: 3008967},
		).Return(&rmmt, nil)

		useClient(client, func() {
			resource.Test(t, resource.TestCase{
				IsUnitTest: true,
				Providers:  testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: loadFixtureString("testdata/TestResMatchTarget/match_by_id.tf"),
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("akamai_appsec_match_target.test", "id", "43253:3008967"),
						),
						ExpectNonEmptyPlan: true,
					},
					{
						Config: loadFixtureString("testdata/TestResMatchTarget/update_by_id.tf"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("akamai_appsec_match_target.test", "id", "43253:3008967"),
						),
						ExpectNonEmptyPlan: true,
					},
				},
			})
		})

		client.AssertExpectations(t)
	})

}
