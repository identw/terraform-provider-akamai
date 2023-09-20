package edgeworkers

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/edgeworkers"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"
)

type testDataForEdgeWorkersActivation struct {
	EdgeWorkerID int
	Network      string
	Activations  []edgeworkers.Activation
}

var (
	expectReadEdgeWorkersActivation = func(t *testing.T, client *edgeworkers.Mock, data testDataForEdgeWorkersActivation, timesToRun int) {
		listActivationsReq := edgeworkers.ListActivationsRequest{
			EdgeWorkerID: data.EdgeWorkerID,
		}
		listActivationsRes := edgeworkers.ListActivationsResponse{
			Activations: data.Activations,
		}

		activations := sortActivationsByDate(data.Activations)
		latestActivation := &activations[0]

		listDeactivationsReq := edgeworkers.ListDeactivationsRequest{
			EdgeWorkerID: data.EdgeWorkerID,
			Version:      latestActivation.Version,
		}
		listDeactivationsRes := edgeworkers.ListDeactivationsResponse{}

		client.On("ListActivations", mock.Anything, listActivationsReq).Return(&listActivationsRes, nil).Times(timesToRun)
		client.On("ListDeactivations", mock.Anything, listDeactivationsReq).Return(&listDeactivationsRes, nil).Times(timesToRun)
	}

	expectReadEmptyEdgeWorkersActivation = func(t *testing.T, client *edgeworkers.Mock, data testDataForEdgeWorkersActivation, timesToRun int) {
		listActivationsReq := edgeworkers.ListActivationsRequest{
			EdgeWorkerID: data.EdgeWorkerID,
		}
		listActivationsRes := edgeworkers.ListActivationsResponse{
			Activations: data.Activations,
		}
		client.On("ListActivations", mock.Anything, listActivationsReq).Return(&listActivationsRes, nil).Times(timesToRun)
	}

	expectListActivationsError = func(t *testing.T, client *edgeworkers.Mock, errorMessage string) {
		listActivationsReq := edgeworkers.ListActivationsRequest{
			EdgeWorkerID: 1,
		}
		listActivationsRes := edgeworkers.ListActivationsResponse{}
		client.On("ListActivations", mock.Anything, listActivationsReq).Return(&listActivationsRes, fmt.Errorf(errorMessage)).Times(1)
	}

	oneActivationData = testDataForEdgeWorkersActivation{
		EdgeWorkerID: 1,
		Network:      stagingNetwork,
		Activations: []edgeworkers.Activation{
			{
				ActivationID: 11,
				Network:      stagingNetwork,
				Version:      "1.0",
				Status:       activationStatusComplete,
			},
		},
	}

	threeActivationsData = testDataForEdgeWorkersActivation{
		EdgeWorkerID: 2,
		Network:      productionNetwork,
		Activations: []edgeworkers.Activation{
			{
				ActivationID: 21,
				Network:      productionNetwork,
				Version:      "1.0",
				CreatedTime:  "2022-04-25T12:30:06Z",
				Status:       activationStatusComplete,
			},
			{
				ActivationID: 22,
				Network:      productionNetwork,
				Version:      "2.0",
				CreatedTime:  "2022-08-25T12:30:06Z",
				Status:       activationStatusComplete,
			},
			{
				ActivationID: 23,
				Network:      productionNetwork,
				Version:      "3.0",
				CreatedTime:  "2022-05-25T12:30:06Z",
				Status:       activationStatusComplete,
			},
		},
	}

	noActivationsData = testDataForEdgeWorkersActivation{
		EdgeWorkerID: 3,
		Network:      productionNetwork,
		Activations:  []edgeworkers.Activation{},
	}

	wrongStatusData = testDataForEdgeWorkersActivation{
		EdgeWorkerID: 4,
		Network:      stagingNetwork,
		Activations: []edgeworkers.Activation{
			{
				ActivationID: 21,
				Network:      stagingNetwork,
				Version:      "1.0",
				CreatedTime:  "2022-05-25T12:30:06Z",
				Status:       "ABORTED",
			},
			{
				ActivationID: 22,
				Network:      stagingNetwork,
				Version:      "2.0",
				CreatedTime:  "2022-07-25T12:30:06Z",
				Status:       "EXPIRED",
			},
		},
	}
)

func TestDataEdgeWorkersActivation(t *testing.T) {
	tests := map[string]struct {
		init       func(*testing.T, *edgeworkers.Mock, testDataForEdgeWorkersActivation)
		mockData   testDataForEdgeWorkersActivation
		configPath string
		error      *regexp.Regexp
	}{
		"happy path with one activation": {
			init: func(t *testing.T, m *edgeworkers.Mock, testData testDataForEdgeWorkersActivation) {
				expectReadEdgeWorkersActivation(t, m, testData, 5)
			},
			mockData:   oneActivationData,
			configPath: "testdata/TestDataEdgeWorkersActivation/one_activation.tf",
			error:      nil,
		},
		"happy path with three activations": {
			init: func(t *testing.T, m *edgeworkers.Mock, testData testDataForEdgeWorkersActivation) {
				expectReadEdgeWorkersActivation(t, m, testData, 5)
			},
			mockData:   threeActivationsData,
			configPath: "testdata/TestDataEdgeWorkersActivation/three_activations.tf",
			error:      nil,
		},
		"happy path with no activations": {
			init: func(t *testing.T, m *edgeworkers.Mock, testData testDataForEdgeWorkersActivation) {
				expectReadEmptyEdgeWorkersActivation(t, m, testData, 5)
			},
			mockData:   noActivationsData,
			configPath: "testdata/TestDataEdgeWorkersActivation/no_activations.tf",
			error:      nil,
		},
		"activation status not complete": {
			init: func(t *testing.T, m *edgeworkers.Mock, testData testDataForEdgeWorkersActivation) {
				expectReadEmptyEdgeWorkersActivation(t, m, testData, 5)
			},
			mockData:   wrongStatusData,
			configPath: "testdata/TestDataEdgeWorkersActivation/wrong_status.tf",
			error:      nil,
		},
		"could not list activations": {
			init: func(t *testing.T, m *edgeworkers.Mock, testData testDataForEdgeWorkersActivation) {
				expectListActivationsError(t, m, "could not fetch activations")
			},
			mockData:   testDataForEdgeWorkersActivation{},
			configPath: "testdata/TestDataEdgeWorkersActivation/one_activation.tf",
			error:      regexp.MustCompile("could not fetch activations"),
		},
		"edgeworker_id not provided": {
			init:       func(t *testing.T, m *edgeworkers.Mock, testData testDataForEdgeWorkersActivation) {},
			mockData:   testDataForEdgeWorkersActivation{},
			configPath: "testdata/TestDataEdgeWorkersActivation/no_edgeworker_id.tf",
			error:      regexp.MustCompile("Missing required argument"),
		},
		"network not provided": {
			init:       func(t *testing.T, m *edgeworkers.Mock, testData testDataForEdgeWorkersActivation) {},
			mockData:   testDataForEdgeWorkersActivation{},
			configPath: "testdata/TestDataEdgeWorkersActivation/no_network.tf",
			error:      regexp.MustCompile("Missing required argument"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := &edgeworkers.Mock{}
			test.init(t, client, test.mockData)
			useClient(client, func() {
				resource.UnitTest(t, resource.TestCase{
					ProtoV6ProviderFactories: testAccProviders,
					IsUnitTest:               true,
					Steps: []resource.TestStep{
						{
							Config:      testutils.LoadFixtureString(t, test.configPath),
							Check:       checkAttrsForEdgeWorkerActivation(test.mockData),
							ExpectError: test.error,
						},
					},
				})
			})
			client.AssertExpectations(t)
		})
	}
}

func checkAttrsForEdgeWorkerActivation(data testDataForEdgeWorkersActivation) resource.TestCheckFunc {
	if len(data.Activations) < 1 {
		return checkAttrsForEmptyEdgeWorkerActivation(data)
	}
	activations := sortActivationsByDate(filterActivationsByNetwork(data.Activations, data.Network))
	latestActivation := &activations[0]
	if latestActivation.Status != activationStatusComplete {
		return checkAttrsForEmptyEdgeWorkerActivation(data)
	}
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.akamai_edgeworker_activation.test", "edgeworker_id", strconv.Itoa(data.EdgeWorkerID)),
		resource.TestCheckResourceAttr("data.akamai_edgeworker_activation.test", "network", data.Network),
		resource.TestCheckResourceAttr("data.akamai_edgeworker_activation.test", "activation_id", strconv.Itoa(latestActivation.ActivationID)),
		resource.TestCheckResourceAttr("data.akamai_edgeworker_activation.test", "version", latestActivation.Version),
		resource.TestCheckResourceAttr("data.akamai_edgeworker_activation.test", "id", fmt.Sprintf("%d:%s", data.EdgeWorkerID, data.Network)),
	)
}

func checkAttrsForEmptyEdgeWorkerActivation(data testDataForEdgeWorkersActivation) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr("data.akamai_edgeworker_activation.test", "edgeworker_id", strconv.Itoa(data.EdgeWorkerID)),
		resource.TestCheckResourceAttr("data.akamai_edgeworker_activation.test", "network", data.Network),
		resource.TestCheckNoResourceAttr("data.akamai_edgeworker_activation.test", "activation_id"),
		resource.TestCheckNoResourceAttr("data.akamai_edgeworker_activation.test", "version"),
		resource.TestCheckResourceAttr("data.akamai_edgeworker_activation.test", "id", fmt.Sprintf("%d:%s", data.EdgeWorkerID, data.Network)),
	)
}
