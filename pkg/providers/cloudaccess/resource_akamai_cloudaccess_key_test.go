package cloudaccess

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v8/pkg/cloudaccess"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/ptr"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/common/testutils"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"
)

type (
	commonDataForAccessKey struct {
		accessKeyUID         int64
		accessKeyName        string
		authenticationMethod string
		contractID           string
		groupID              int64
		networkConfig        networkConfiguration
		credentialsA         credentials
		credentialsB         credentials
	}

	networkConfiguration struct {
		securityNetwork string
		additionalCDN   string
	}

	credentials struct {
		cloudAccessKeyID     string
		cloudSecretAccessKey string
		primaryKey           bool
	}

	commonDataForProperty struct {
		accessKeyUID      int64
		propertyName      string
		propertyID        string
		stagingVersion    int64
		productionVersion int64
	}

	commonDataForResource struct {
		accessKeyData []commonDataForAccessKey
		propertyData  *commonDataForProperty
	}
)

var (
	accessKeyMock = commonDataForAccessKey{
		accessKeyName:        "test_key_name",
		accessKeyUID:         12345,
		authenticationMethod: string(cloudaccess.AuthAWS),
		contractID:           "1-CTRACT",
		groupID:              12345,
		networkConfig: networkConfiguration{
			securityNetwork: string(cloudaccess.NetworkEnhanced),
			additionalCDN:   string(cloudaccess.ChinaCDN),
		},
		credentialsA: credentials{
			cloudAccessKeyID:     "test_key_id",
			cloudSecretAccessKey: "test_secret",
			primaryKey:           true,
		},
		credentialsB: credentials{
			cloudAccessKeyID:     "test_key_id_2",
			cloudSecretAccessKey: "test_secret_2",
			primaryKey:           false,
		},
	}
	secondKeyMock = commonDataForAccessKey{
		accessKeyName:        "test key name2",
		accessKeyUID:         5678,
		authenticationMethod: string(cloudaccess.AuthAWS),
		contractID:           "1-CTRACT",
		groupID:              12345,
		networkConfig: networkConfiguration{
			securityNetwork: string(cloudaccess.NetworkEnhanced),
			additionalCDN:   string(cloudaccess.ChinaCDN),
		},
		credentialsA: credentials{
			cloudAccessKeyID:     "2test_key_id",
			cloudSecretAccessKey: "2test_secret",
			primaryKey:           true,
		},
		credentialsB: credentials{
			cloudAccessKeyID:     "2test_key_id_2",
			cloudSecretAccessKey: "2test_secret_2",
			primaryKey:           false,
		},
	}

	onlyCredBMock = commonDataForAccessKey{
		accessKeyName:        "test_key_name",
		accessKeyUID:         12345,
		authenticationMethod: string(cloudaccess.AuthAWS),
		contractID:           "1-CTRACT",
		groupID:              12345,
		networkConfig: networkConfiguration{
			securityNetwork: string(cloudaccess.NetworkEnhanced),
			additionalCDN:   string(cloudaccess.ChinaCDN),
		},
		credentialsB: credentials{
			cloudAccessKeyID:     "test_key_id",
			cloudSecretAccessKey: "test_secret",
			primaryKey:           true,
		},
	}

	emptySecretMock = commonDataForAccessKey{
		accessKeyName:        "test_key_name",
		accessKeyUID:         12345,
		authenticationMethod: string(cloudaccess.AuthAWS),
		contractID:           "1-CTRACT",
		groupID:              12345,
		networkConfig: networkConfiguration{
			securityNetwork: string(cloudaccess.NetworkEnhanced),
			additionalCDN:   string(cloudaccess.ChinaCDN),
		},
		credentialsA: credentials{
			cloudAccessKeyID:     "test_key_id",
			cloudSecretAccessKey: "",
			primaryKey:           true,
		},
	}

	propertyMock = commonDataForProperty{
		accessKeyUID:      12345,
		propertyID:        "123123",
		propertyName:      "test_property_name",
		stagingVersion:    1,
		productionVersion: 1,
	}

	resourceMock = commonDataForResource{
		accessKeyData: []commonDataForAccessKey{accessKeyMock, secondKeyMock, onlyCredBMock, emptySecretMock},
		propertyData:  &propertyMock,
	}

	firstAccessKeyVersion  = int64(1)
	secondAccessKeyVersion = int64(2)
	thirdAccessKeyVersion  = int64(3)

	emptyVersionList       = 0
	oneElementVersionList  = 1
	twoElementsVersionList = 2
)

func TestAccessKeyResource(t *testing.T) {
	t.Parallel()
	pollingInterval = 1 * time.Millisecond
	deleteTimeout = 40 * time.Millisecond
	updateTimeout = 20 * time.Millisecond
	activationTimeout = 20 * time.Millisecond
	tests := map[string]struct {
		configPath string
		init       func(*testing.T, *cloudaccess.Mock, commonDataForResource)
		mockData   commonDataForResource
		steps      []resource.TestStep
		error      *regexp.Regexp
	}{
		"create access key one version": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith1Version(m, resourceData)
				mockReadAccessKeyWith1Version(m, resourceData)
				mockDeletionAccessKeyWith1Version(m, resourceData)
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"create access key two versions": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockDeletionAccessKeyWith2Versions(m, resourceData)
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"create access key only credentialsB": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				//creation
				mockCreationAccessKeyUsingCredB(m, resourceData)
				//read
				mockGetAccessKey(m, resourceData.accessKeyData[2]).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[2]).Once()
				//delete
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[2]).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
				m.On("DeleteAccessKeyVersion", mock.Anything, cloudaccess.DeleteAccessKeyVersionRequest{AccessKeyUID: resourceData.accessKeyData[2].accessKeyUID, Version: firstAccessKeyVersion}).
					Return(&cloudaccess.DeleteAccessKeyVersionResponse{
						AccessKeyUID:     resourceData.accessKeyData[2].accessKeyUID,
						CloudAccessKeyID: ptr.To(resourceData.accessKeyData[2].credentialsB.cloudAccessKeyID),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          firstAccessKeyVersion,
						VersionGUID:      "asde-efdr-reded",
					}, nil).Once()
				m.On("GetAccessKeyVersion", mock.Anything, cloudaccess.GetAccessKeyVersionRequest{AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID, Version: 1}).
					Return(&cloudaccess.GetAccessKeyVersionResponse{
						AccessKeyUID:     resourceData.accessKeyData[2].accessKeyUID,
						CloudAccessKeyID: ptr.To(resourceData.accessKeyData[2].credentialsB.cloudAccessKeyID),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          firstAccessKeyVersion,
						VersionGUID:      "asde-efdr-reded",
					}, nil).Once()
				mockDeleteAccessKey(m, resourceData.accessKeyData[2]).Once()
				var listOfKeysAfterDeletion []commonDataForAccessKey
				mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_using_credB.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"delete access key version only credentialsB": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				//create
				mockCreationAccessKeyUsingCredB(m, resourceData)
				//read
				mockGetAccessKey(m, resourceData.accessKeyData[2]).Times(2)
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[2]).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[2]).Once()
				//delete 1 version (credB)
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[2]).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
				m.On("DeleteAccessKeyVersion", mock.Anything, cloudaccess.DeleteAccessKeyVersionRequest{AccessKeyUID: resourceData.accessKeyData[2].accessKeyUID, Version: firstAccessKeyVersion}).
					Return(&cloudaccess.DeleteAccessKeyVersionResponse{
						AccessKeyUID:     resourceData.accessKeyData[2].accessKeyUID,
						CloudAccessKeyID: ptr.To(resourceData.accessKeyData[2].credentialsB.cloudAccessKeyID),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          firstAccessKeyVersion,
						VersionGUID:      "asde-efdr-reded",
					}, nil).Once()
				m.On("GetAccessKeyVersion", mock.Anything, cloudaccess.GetAccessKeyVersionRequest{AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID, Version: firstAccessKeyVersion}).
					Return(&cloudaccess.GetAccessKeyVersionResponse{
						AccessKeyUID:     resourceData.accessKeyData[2].accessKeyUID,
						CloudAccessKeyID: ptr.To(resourceData.accessKeyData[2].credentialsB.cloudAccessKeyID),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.PendingDeletion,
						Version:          firstAccessKeyVersion,
						VersionGUID:      "asde-efdr-reded",
					}, nil).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[2], emptyVersionList).Once()
				//read
				mockGetAccessKey(m, resourceData.accessKeyData[2]).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[2], emptyVersionList).Once()
				//delete key
				mockListAccessKeyVersions(m, resourceData.accessKeyData[2], emptyVersionList).Once()
				mockDeleteAccessKey(m, resourceData.accessKeyData[2]).Once()
				var listOfKeysAfterDeletion []commonDataForAccessKey
				mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_using_credB.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/creation_no_credentials.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", ""),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"basic name update": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith1Version(m, resourceData)
				mockReadAccessKeyWith1Version(m, resourceData)
				mockReadAccessKeyWith1Version(m, resourceData)
				mockUpdateAccessKey(m, resourceData.accessKeyData[0], "updated_key_name").Once()
				m.On("GetAccessKey", mock.Anything, cloudaccess.AccessKeyRequest{AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID}).
					Return(&cloudaccess.GetAccessKeyResponse{
						AccessKeyName:        "updated_key_name",
						AccessKeyUID:         resourceData.accessKeyData[0].accessKeyUID,
						AuthenticationMethod: resourceData.accessKeyData[0].authenticationMethod,
						NetworkConfiguration: &cloudaccess.SecureNetwork{
							SecurityNetwork: cloudaccess.NetworkType(resourceData.accessKeyData[0].networkConfig.securityNetwork),
							AdditionalCDN:   ptr.To(cloudaccess.CDNType(resourceData.accessKeyData[0].networkConfig.additionalCDN)),
						},
						LatestVersion: firstAccessKeyVersion,
						Groups: []cloudaccess.Group{
							{
								GroupID:     resourceData.accessKeyData[0].groupID,
								GroupName:   ptr.To("random group name"),
								ContractIDs: []string{resourceData.accessKeyData[0].contractID},
							},
						},
						CreatedBy:   "dev-user",
						CreatedTime: time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
					}, nil).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Twice()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Twice()
				mockDeletionAccessKeyWith1Version(m, resourceData)
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/updated_name.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "updated_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"single-credentials rotation": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				//delete version 1
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], firstAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, firstAccessKeyVersion).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], oneElementVersionList).Once()
				//create new version (no.3)
				m.On("CreateAccessKeyVersion", mock.Anything, cloudaccess.CreateAccessKeyVersionRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
					BodyParams: cloudaccess.CreateAccessKeyVersionBodyParams{
						CloudAccessKeyID:     "test_key_id_3",
						CloudSecretAccessKey: "test_secret_3",
					}}).Return(&cloudaccess.CreateAccessKeyVersionResponse{RequestID: 321321, RetryAfter: 1000}, nil).Once()
				m.On("GetAccessKeyVersionStatus", mock.Anything, cloudaccess.GetAccessKeyVersionStatusRequest{RequestID: 321321}).
					Return(&cloudaccess.GetAccessKeyVersionStatusResponse{
						AccessKeyVersion: &cloudaccess.KeyVersion{
							AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
							Version:      thirdAccessKeyVersion,
						},
						ProcessingStatus: cloudaccess.ProcessingDone,
						RequestDate:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						RequestedBy:      "dev-user",
					}, nil).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.Active, thirdAccessKeyVersion).Once()

				//read
				mockGetAccessKey(m, resourceData.accessKeyData[0]).Once()
				m.On("ListAccessKeyVersions", mock.Anything, cloudaccess.ListAccessKeyVersionsRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
				}).Return(&cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_3"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          thirdAccessKeyVersion,
						VersionGUID:      "ffff_eeee-ffffddd",
					},
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_2"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          secondAccessKeyVersion,
						VersionGUID:      "asdd-ads-dasdas",
					},
				},
				}, nil).Once()
				//Delete both credentials ( with version no.3 and no.2)
				m.On("ListAccessKeyVersions", mock.Anything, cloudaccess.ListAccessKeyVersionsRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
				}).Return(&cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_3"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          thirdAccessKeyVersion,
						VersionGUID:      "ffff_eeee-ffffddd",
					},
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_2"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          secondAccessKeyVersion,
						VersionGUID:      "asdd-ads-dasdas",
					},
				},
				}, nil).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, thirdAccessKeyVersion).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, secondAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], thirdAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, thirdAccessKeyVersion).Once()
				m.On("ListAccessKeyVersions", mock.Anything, cloudaccess.ListAccessKeyVersionsRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
				}).Return(&cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_3"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          thirdAccessKeyVersion,
						VersionGUID:      "ffff_eeee-ffffddd",
					},
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_2"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          secondAccessKeyVersion,
						VersionGUID:      "asdd-ads-dasdas",
					},
				},
				}, nil).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], oneElementVersionList).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, secondAccessKeyVersion).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], oneElementVersionList).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Once()
				mockDeleteAccessKey(m, resourceData.accessKeyData[0]).Once()
				var listOfKeysAfterDeletion []commonDataForAccessKey
				mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/single_credentials_rotation.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "ffff_eeee-ffffddd"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id_3"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret_3"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "3"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "ffff_eeee-ffffddd"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"cross-credentials rotation of access key": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				//delete version 2
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, secondAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, secondAccessKeyVersion).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()

				//create new version (no.3)
				m.On("CreateAccessKeyVersion", mock.Anything, cloudaccess.CreateAccessKeyVersionRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
					BodyParams: cloudaccess.CreateAccessKeyVersionBodyParams{
						CloudAccessKeyID:     "test_key_id_3",
						CloudSecretAccessKey: "test_secret_3",
					}}).Return(&cloudaccess.CreateAccessKeyVersionResponse{RequestID: 321321, RetryAfter: 1000}, nil).Once()
				m.On("GetAccessKeyVersionStatus", mock.Anything, cloudaccess.GetAccessKeyVersionStatusRequest{RequestID: 321321}).
					Return(&cloudaccess.GetAccessKeyVersionStatusResponse{
						AccessKeyVersion: &cloudaccess.KeyVersion{
							AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
							Version:      thirdAccessKeyVersion,
						},
						ProcessingStatus: cloudaccess.ProcessingDone,
						RequestDate:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						RequestedBy:      "dev-user",
					}, nil).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.Active, thirdAccessKeyVersion).Once()

				//read
				m.On("GetAccessKey", mock.Anything, cloudaccess.AccessKeyRequest{AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID}).
					Return(&cloudaccess.GetAccessKeyResponse{
						AccessKeyName:        resourceData.accessKeyData[0].accessKeyName,
						AccessKeyUID:         resourceData.accessKeyData[0].accessKeyUID,
						AuthenticationMethod: resourceData.accessKeyData[0].authenticationMethod,
						NetworkConfiguration: &cloudaccess.SecureNetwork{
							SecurityNetwork: cloudaccess.NetworkType(resourceData.accessKeyData[0].networkConfig.securityNetwork),
							AdditionalCDN:   ptr.To(cloudaccess.CDNType(resourceData.accessKeyData[0].networkConfig.additionalCDN)),
						},
						LatestVersion: thirdAccessKeyVersion,
						Groups: []cloudaccess.Group{
							{
								GroupID:     resourceData.accessKeyData[0].groupID,
								GroupName:   ptr.To("random group name"),
								ContractIDs: []string{resourceData.accessKeyData[0].contractID},
							},
						},
						CreatedBy:   "dev-user",
						CreatedTime: time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
					}, nil)
				m.On("ListAccessKeyVersions", mock.Anything, cloudaccess.ListAccessKeyVersionsRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
				}).Return(&cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_3"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          thirdAccessKeyVersion,
						VersionGUID:      "ffff_eeee-ffffddd",
					},
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          firstAccessKeyVersion,
						VersionGUID:      "asde-efdr-reded",
					},
				},
				}, nil).Once()
				//Delete both credentials ( with version no.3 and no.1)
				m.On("ListAccessKeyVersions", mock.Anything, cloudaccess.ListAccessKeyVersionsRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
				}).Return(&cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_3"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          thirdAccessKeyVersion,
						VersionGUID:      "ffff_eeee-ffffddd",
					},
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          firstAccessKeyVersion,
						VersionGUID:      "asde-efdr-reded",
					},
				},
				}, nil).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, thirdAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], thirdAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, thirdAccessKeyVersion).Once()
				m.On("ListAccessKeyVersions", mock.Anything, cloudaccess.ListAccessKeyVersionsRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
				}).Return(&cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id_3"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          thirdAccessKeyVersion,
						VersionGUID:      "ffff_eeee-ffffddd",
					},
					{
						AccessKeyUID:     resourceData.accessKeyData[0].accessKeyUID,
						CloudAccessKeyID: ptr.To("test_key_id"),
						CreatedBy:        "dev-user",
						CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
						DeploymentStatus: cloudaccess.Active,
						Version:          firstAccessKeyVersion,
						VersionGUID:      "asde-efdr-reded",
					},
				},
				}, nil).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], firstAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, firstAccessKeyVersion).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Once()
				mockDeleteAccessKey(m, resourceData.accessKeyData[0]).Once()
				var listOfKeysAfterDeletion []commonDataForAccessKey
				mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/cross_credentials_rotation.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "ffff_eeee-ffffddd"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_3"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_3"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "3"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "ffff_eeee-ffffddd"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"change primary flag": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				//only change of flag, and primary guid
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockDeletionAccessKeyWith2Versions(m, resourceData)
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/swap_primary_key.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"delete one version of access key": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				//delete 2nd version
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, secondAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, secondAccessKeyVersion).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()

				mockReadAccessKeyWith1Version(m, resourceData)
				// delete 1 version
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], firstAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, firstAccessKeyVersion).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Once()
				mockDeleteAccessKey(m, resourceData.accessKeyData[0]).Once()
				var listOfKeysAfterDeletion []commonDataForAccessKey
				mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"delete two versions of access key": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				//delete 2 versions
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, secondAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, secondAccessKeyVersion).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], firstAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, firstAccessKeyVersion).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Once()

				mockReadAccessKey(m, resourceData, emptyVersionList)
				// delete key with no versions
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Twice()
				mockDeleteAccessKey(m, resourceData.accessKeyData[0]).Once()
				var listOfKeysAfterDeletion []commonDataForAccessKey
				mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_no_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", ""),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"change order of credentials": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockDeletionAccessKeyWith2Versions(m, resourceData)
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/changed_order.tf"),
					ExpectError: regexp.MustCompile("cannot change order of `credentials_a` and `credentials_b`"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"change secret block": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockDeletionAccessKeyWith2Versions(m, resourceData)
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/changed_secret.tf"),
					ExpectError: regexp.MustCompile("\\s*cannot update cloud access secret without update of cloud access key id,\\s*expect in-place update of secret after import"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "changed_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"missing contract id": {
			steps: []resource.TestStep{{
				Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/missing_contract.tf"),
				ExpectError: regexp.MustCompile(`The argument "contract_id" is required, but no definition was found`),
			}},
		},
		"missing group id": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/missing_group.tf"),
					ExpectError: regexp.MustCompile(`The argument "group_id" is required, but no definition was found`),
				},
			},
		},
		"missing security network": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/missing_security_network.tf"),
					ExpectError: regexp.MustCompile(`Incorrect attribute value type`),
				},
			},
		},
		"missing primary key": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/missing_primary_key.tf"),
					ExpectError: regexp.MustCompile(`Incorrect attribute value type`),
				},
			},
		},
		"missing cloud access key": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/missing_cloud_access_key.tf"),
					ExpectError: regexp.MustCompile(`Incorrect attribute value type`),
				},
			},
		},
		"missing cloud access secret": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/missing_cloud_access_secret.tf"),
					ExpectError: regexp.MustCompile(`Incorrect attribute value type`),
				},
			},
		},
		"no credentials for creation": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/creation_no_credentials.tf"),
					ExpectError: regexp.MustCompile(`at least one credentials are required for creation`),
				},
			},
		},
		"timeout on creation": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreateAccessKey(m, resourceData.accessKeyData[0]).Once()
				mockGetAccessKeyStatus(m, 12345, resourceData.accessKeyData[0]).Once()
				//artificial sleep to trigger 20 ms timeout
				time.Sleep(21 * time.Millisecond)
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingActivation, firstAccessKeyVersion).Once() //timeout
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_with_timeout.tf"),
					ExpectError: regexp.MustCompile("reached activation timeout"),
				},
			},
		},
		"timeout on one version deletion": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				//delete 2nd version
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, secondAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()
				time.Sleep(50 * time.Millisecond)
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, secondAccessKeyVersion).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()

				// delete 1 version
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, secondAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, secondAccessKeyVersion).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], firstAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, firstAccessKeyVersion).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Twice()

				// delete key with no versions
				mockDeleteAccessKey(m, resourceData.accessKeyData[0]).Once()
				var listOfKeysAfterDeletion []commonDataForAccessKey
				mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions_with_timeouts.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_with_timeout.tf"),
					ExpectError: regexp.MustCompile("Error: deletion terminated"),
				},
			},
		},
		"fail of deletion - version assigned to property": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				//delete 2nd version - fail
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockLookupsProperties(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()

				//delete all versions
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, secondAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, secondAccessKeyVersion).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
				mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
				mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], firstAccessKeyVersion).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, firstAccessKeyVersion).Once()
				mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
				mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Twice()

				// delete key with no versions
				mockDeleteAccessKey(m, resourceData.accessKeyData[0]).Once()
				var listOfKeysAfterDeletion []commonDataForAccessKey
				mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()

			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create.tf"),
					ExpectError: regexp.MustCompile(fmt.Sprintf("cannot delete version: %d of access key %d assigned to property", secondAccessKeyVersion, 12345)),
				},
			},
		},
		"fail on creation - tainted resource": {
			init: func(t *testing.T, m *cloudaccess.Mock, resourceData commonDataForResource) {
				mockCreateAccessKey(m, resourceData.accessKeyData[0]).Once()
				mockGetAccessKeyStatus(m, 12345, resourceData.accessKeyData[0]).Once()
				mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.Active, firstAccessKeyVersion).Once()
				// fail and taint resource
				m.On("CreateAccessKeyVersion", mock.Anything, cloudaccess.CreateAccessKeyVersionRequest{
					AccessKeyUID: resourceData.accessKeyData[0].accessKeyUID,
					BodyParams: cloudaccess.CreateAccessKeyVersionBodyParams{
						CloudAccessKeyID:     resourceData.accessKeyData[0].credentialsB.cloudAccessKeyID,
						CloudSecretAccessKey: resourceData.accessKeyData[0].credentialsB.cloudSecretAccessKey,
					}}).Return(nil, cloudaccess.ErrCreateAccessKeyVersion).Once()
				//Delete before replace
				mockReadAccessKeyWith1Version(m, resourceData)
				mockDeletionAccessKeyWith1Version(m, resourceData)

				//Second successful creation attempt
				mockCreationAccessKeyWith2Versions(m, resourceData)
				mockReadAccessKey(m, resourceData, twoElementsVersionList)
				mockDeletionAccessKeyWith2Versions(m, resourceData)

			},
			mockData: resourceMock,
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					ExpectError: regexp.MustCompile("Error: create access key version failed"),
				},
				{
					Config: testutils.LoadFixtureString(t, "testdata/TestResAccessKey/create_2_versions.tf"),
					Taint:  []string{"akamai_cloudaccess_key.test"},
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_uid", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "access_key_name", "test_key_name"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "authentication_method", "AWS4_HMAC_SHA256"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "contract_id", "1-CTRACT"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "group_id", "12345"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "primary_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_access_key_id", "test_key_id"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.cloud_secret_access_key", "test_secret"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.primary_key", "true"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version", "1"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_a.version_guid", "asde-efdr-reded"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_access_key_id", "test_key_id_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.cloud_secret_access_key", "test_secret_2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.primary_key", "false"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version", "2"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "credentials_b.version_guid", "asdd-ads-dasdas"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.additional_cdn", "CHINA_CDN"),
						resource.TestCheckResourceAttr("akamai_cloudaccess_key.test", "network_configuration.security_network", "ENHANCED_TLS"),
					),
				},
			},
		},
		"fail on creation - not proper additional cdn": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/wrong_additional_cdn.tf"),
					ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
				},
			},
		},
		"fail on creation - not proper security network": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/wrong_security_network.tf"),
					ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
				},
			},
		},
		"fail on creation - not proper authentication method": {
			steps: []resource.TestStep{
				{
					Config:      testutils.LoadFixtureString(t, "testdata/TestResAccessKey/wrong_security_network.tf"),
					ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			client := &cloudaccess.Mock{}
			if test.init != nil {
				test.init(t, client, test.mockData)
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

func mockDeletionAccessKeyWith2Versions(m *cloudaccess.Mock, resourceData commonDataForResource) {
	mockListAccessKeyVersions(m, resourceData.accessKeyData[0], twoElementsVersionList).Once()
	mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
	mockLookupsPropertiesNoProperties(m, resourceData.propertyData, secondAccessKeyVersion).Once()
	mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], firstAccessKeyVersion).Once()
	mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, firstAccessKeyVersion).Once()
	mockListAccessKeyVersions(m, resourceData.accessKeyData[0], oneElementVersionList).Once()
	mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], secondAccessKeyVersion).Once()
	mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, secondAccessKeyVersion).Once()
	mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Once()
	mockDeleteAccessKey(m, resourceData.accessKeyData[0]).Once()
	var listOfKeysAfterDeletion []commonDataForAccessKey
	mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
}

func mockCreationAccessKeyWith2Versions(m *cloudaccess.Mock, resourceData commonDataForResource) {
	mockCreateAccessKey(m, resourceData.accessKeyData[0]).Once()
	mockGetAccessKeyStatus(m, 12345, resourceData.accessKeyData[0]).Once()
	mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.Active, firstAccessKeyVersion).Once()
	mockCreateAccessKeyVersion(m, resourceData.accessKeyData[0]).Once()
	mockGetAccessKeyVersionStatus(m, resourceData.accessKeyData[0], 124, secondAccessKeyVersion).Once()
	mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.Active, secondAccessKeyVersion).Once()
}

func mockDeletionAccessKeyWith1Version(m *cloudaccess.Mock, resourceData commonDataForResource) {
	mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
	mockLookupsPropertiesNoProperties(m, resourceData.propertyData, firstAccessKeyVersion).Once()
	mockDeleteAccessKeyVersion(m, resourceData.accessKeyData[0], firstAccessKeyVersion).Once()
	mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.PendingDeletion, firstAccessKeyVersion).Once()
	mockListAccessKeyVersions(m, resourceData.accessKeyData[0], emptyVersionList).Once()
	mockDeleteAccessKey(m, resourceData.accessKeyData[0]).Once()
	var listOfKeysAfterDeletion []commonDataForAccessKey
	mockListAccessKeys(m, append(listOfKeysAfterDeletion, resourceData.accessKeyData[1])).Once()
}

func mockReadAccessKey(m *cloudaccess.Mock, resourceData commonDataForResource, size int) {
	mockGetAccessKey(m, resourceData.accessKeyData[0]).Once()
	mockListAccessKeyVersions(m, resourceData.accessKeyData[0], size).Once()
}
func mockReadAccessKeyWith1Version(m *cloudaccess.Mock, resourceData commonDataForResource) {
	mockGetAccessKey(m, resourceData.accessKeyData[0]).Once()
	mockListAccessKeyVersionsOnly1Version(m, resourceData.accessKeyData[0]).Once()
}

func mockCreationAccessKeyWith1Version(m *cloudaccess.Mock, resourceData commonDataForResource) {
	mockCreateAccessKey(m, resourceData.accessKeyData[0]).Once()
	mockGetAccessKeyStatus(m, 12345, resourceData.accessKeyData[0]).Once()
	mockGetAccessKeyVersion(m, resourceData.accessKeyData[0], cloudaccess.Active, firstAccessKeyVersion).Once()
}
func mockCreationAccessKeyUsingCredB(m *cloudaccess.Mock, resourceData commonDataForResource) {
	mockCreateAccessKeyUsingCredB(m, resourceData.accessKeyData[2]).Once()
	mockGetAccessKeyStatus(m, 12345, resourceData.accessKeyData[2]).Once()
	m.On("GetAccessKeyVersion", mock.Anything, cloudaccess.GetAccessKeyVersionRequest{AccessKeyUID: resourceData.accessKeyData[2].accessKeyUID, Version: firstAccessKeyVersion}).
		Return(&cloudaccess.GetAccessKeyVersionResponse{
			AccessKeyUID:     resourceData.accessKeyData[2].accessKeyUID,
			CloudAccessKeyID: ptr.To(resourceData.accessKeyData[2].credentialsB.cloudAccessKeyID),
			CreatedBy:        "dev-user",
			CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			DeploymentStatus: cloudaccess.Active,
			Version:          firstAccessKeyVersion,
			VersionGUID:      "asde-efdr-reded",
		}, nil).Once()
}

func mockGetAccessKey(client *cloudaccess.Mock, testData commonDataForAccessKey) *mock.Call {
	return client.On("GetAccessKey", mock.Anything, cloudaccess.AccessKeyRequest{AccessKeyUID: testData.accessKeyUID}).
		Return(&cloudaccess.GetAccessKeyResponse{
			AccessKeyName:        testData.accessKeyName,
			AccessKeyUID:         testData.accessKeyUID,
			AuthenticationMethod: testData.authenticationMethod,
			NetworkConfiguration: &cloudaccess.SecureNetwork{
				SecurityNetwork: cloudaccess.NetworkType(testData.networkConfig.securityNetwork),
				AdditionalCDN:   ptr.To(cloudaccess.CDNType(testData.networkConfig.additionalCDN)),
			},
			LatestVersion: firstAccessKeyVersion,
			Groups: []cloudaccess.Group{
				{
					GroupID:     testData.groupID,
					GroupName:   ptr.To("random group name"),
					ContractIDs: []string{testData.contractID},
				},
			},
			CreatedBy:   "dev-user",
			CreatedTime: time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
		}, nil)
}

func mockGetAccessKeyStatus(client *cloudaccess.Mock, requestID int64, testData commonDataForAccessKey) *mock.Call {
	return client.On("GetAccessKeyStatus", mock.Anything, cloudaccess.GetAccessKeyStatusRequest{RequestID: requestID}).
		Return(&cloudaccess.GetAccessKeyStatusResponse{
			AccessKey: &cloudaccess.KeyLink{
				AccessKeyUID: testData.accessKeyUID,
			},
			AccessKeyVersion: &cloudaccess.KeyVersion{
				AccessKeyUID: testData.accessKeyUID,
				Version:      firstAccessKeyVersion,
			},
			ProcessingStatus: cloudaccess.ProcessingDone,
			Request: &cloudaccess.RequestInformation{
				AccessKeyName:        testData.accessKeyName,
				AuthenticationMethod: cloudaccess.AuthType(testData.authenticationMethod),
				ContractID:           testData.contractID,
				GroupID:              testData.groupID,
				NetworkConfiguration: &cloudaccess.SecureNetwork{
					SecurityNetwork: cloudaccess.NetworkType(testData.networkConfig.securityNetwork),
					AdditionalCDN:   ptr.To(cloudaccess.CDNType(testData.networkConfig.additionalCDN)),
				},
			},
			RequestDate: time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			RequestID:   requestID,
			RequestedBy: "dev-user",
		}, nil)
}

func mockCreateAccessKey(client *cloudaccess.Mock, testData commonDataForAccessKey) *mock.Call {
	return client.On("CreateAccessKey", mock.Anything, cloudaccess.CreateAccessKeyRequest{
		AccessKeyName:        testData.accessKeyName,
		AuthenticationMethod: testData.authenticationMethod,
		ContractID:           testData.contractID,
		GroupID:              testData.groupID,
		NetworkConfiguration: cloudaccess.SecureNetwork{
			SecurityNetwork: cloudaccess.NetworkType(testData.networkConfig.securityNetwork),
			AdditionalCDN:   ptr.To(cloudaccess.CDNType(testData.networkConfig.additionalCDN)),
		},
		Credentials: cloudaccess.Credentials{
			CloudAccessKeyID:     testData.credentialsA.cloudAccessKeyID,
			CloudSecretAccessKey: testData.credentialsA.cloudSecretAccessKey,
		},
	}).Return(&cloudaccess.CreateAccessKeyResponse{RequestID: 12345, RetryAfter: 1000}, nil)
}
func mockCreateAccessKeyUsingCredB(client *cloudaccess.Mock, testData commonDataForAccessKey) *mock.Call {
	return client.On("CreateAccessKey", mock.Anything, cloudaccess.CreateAccessKeyRequest{
		AccessKeyName:        testData.accessKeyName,
		AuthenticationMethod: testData.authenticationMethod,
		ContractID:           testData.contractID,
		GroupID:              testData.groupID,
		NetworkConfiguration: cloudaccess.SecureNetwork{
			SecurityNetwork: cloudaccess.NetworkType(testData.networkConfig.securityNetwork),
			AdditionalCDN:   ptr.To(cloudaccess.CDNType(testData.networkConfig.additionalCDN)),
		},
		Credentials: cloudaccess.Credentials{
			CloudAccessKeyID:     testData.credentialsB.cloudAccessKeyID,
			CloudSecretAccessKey: testData.credentialsB.cloudSecretAccessKey,
		},
	}).Return(&cloudaccess.CreateAccessKeyResponse{RequestID: 12345, RetryAfter: 1000}, nil)
}

func mockUpdateAccessKey(client *cloudaccess.Mock, testData commonDataForAccessKey, updatedName string) *mock.Call {
	return client.On("UpdateAccessKey", mock.Anything, cloudaccess.UpdateAccessKeyRequest{
		AccessKeyName: updatedName,
	}, cloudaccess.AccessKeyRequest{AccessKeyUID: testData.accessKeyUID},
	).Return(&cloudaccess.UpdateAccessKeyResponse{
		AccessKeyUID:  testData.accessKeyUID,
		AccessKeyName: updatedName,
	}, nil)
}
func mockDeleteAccessKey(client *cloudaccess.Mock, testData commonDataForAccessKey) *mock.Call {
	return client.On("DeleteAccessKey", mock.Anything, cloudaccess.AccessKeyRequest{
		AccessKeyUID: testData.accessKeyUID,
	},
	).Return(nil)
}

func mockListAccessKeys(client *cloudaccess.Mock, testData []commonDataForAccessKey) *mock.Call {
	var listResponse cloudaccess.ListAccessKeysResponse
	if len(testData) == 2 {
		listResponse = cloudaccess.ListAccessKeysResponse{
			AccessKeys: []cloudaccess.AccessKeyResponse{{
				AccessKeyName:        testData[0].accessKeyName,
				AccessKeyUID:         testData[0].accessKeyUID,
				AuthenticationMethod: testData[0].authenticationMethod,
				NetworkConfiguration: &cloudaccess.SecureNetwork{
					SecurityNetwork: cloudaccess.NetworkType(testData[0].networkConfig.securityNetwork),
					AdditionalCDN:   ptr.To(cloudaccess.CDNType(testData[0].networkConfig.additionalCDN)),
				},
				LatestVersion: firstAccessKeyVersion,
				Groups: []cloudaccess.Group{
					{
						GroupID:     testData[0].groupID,
						GroupName:   ptr.To("random_group_name"),
						ContractIDs: []string{testData[0].contractID},
					},
				},
				CreatedBy:   "dev-user",
				CreatedTime: time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			}, {
				AccessKeyName:        testData[1].accessKeyName,
				AccessKeyUID:         testData[1].accessKeyUID,
				AuthenticationMethod: testData[2].authenticationMethod,
				NetworkConfiguration: &cloudaccess.SecureNetwork{
					SecurityNetwork: cloudaccess.NetworkType(testData[1].networkConfig.securityNetwork),
					AdditionalCDN:   ptr.To(cloudaccess.CDNType(testData[1].networkConfig.additionalCDN)),
				},
				LatestVersion: firstAccessKeyVersion,
				Groups: []cloudaccess.Group{
					{
						GroupID:     testData[1].groupID,
						GroupName:   ptr.To("random_group_name"),
						ContractIDs: []string{testData[1].contractID},
					},
				},
				CreatedBy:   "dev-user",
				CreatedTime: time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			},
			},
		}
	}
	if len(testData) == 1 {
		listResponse = cloudaccess.ListAccessKeysResponse{
			AccessKeys: []cloudaccess.AccessKeyResponse{{
				AccessKeyName:        testData[0].accessKeyName,
				AccessKeyUID:         testData[0].accessKeyUID,
				AuthenticationMethod: testData[0].authenticationMethod,
				NetworkConfiguration: &cloudaccess.SecureNetwork{
					SecurityNetwork: cloudaccess.NetworkType(testData[0].networkConfig.securityNetwork),
					AdditionalCDN:   ptr.To(cloudaccess.CDNType(testData[0].networkConfig.additionalCDN)),
				},
				LatestVersion: firstAccessKeyVersion,
				Groups: []cloudaccess.Group{
					{
						GroupID:     testData[0].groupID,
						GroupName:   ptr.To("random_group_name"),
						ContractIDs: []string{testData[0].contractID},
					},
				},
				CreatedBy:   "dev-user",
				CreatedTime: time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			},
			},
		}
	}
	return client.On("ListAccessKeys", mock.Anything, cloudaccess.ListAccessKeysRequest{}).
		Return(&listResponse, nil)
}

func mockGetAccessKeyVersion(client *cloudaccess.Mock, testData commonDataForAccessKey, deploymentStatus cloudaccess.DeploymentStatus, version int64) *mock.Call {
	var cloudAccessKeyID, versionGUID string
	if version == firstAccessKeyVersion {
		cloudAccessKeyID = testData.credentialsA.cloudAccessKeyID
		versionGUID = "asde-efdr-reded"
	}
	if version == secondAccessKeyVersion {
		cloudAccessKeyID = testData.credentialsB.cloudAccessKeyID
		versionGUID = "asdd-ads-dasdas"
	}
	if version == thirdAccessKeyVersion {
		cloudAccessKeyID = "test_key_id_3"
		versionGUID = "ffff_eeee-ffffddd"
	}
	return client.On("GetAccessKeyVersion", mock.Anything, cloudaccess.GetAccessKeyVersionRequest{AccessKeyUID: testData.accessKeyUID, Version: version}).
		Return(&cloudaccess.GetAccessKeyVersionResponse{
			AccessKeyUID:     testData.accessKeyUID,
			CloudAccessKeyID: ptr.To(cloudAccessKeyID),
			CreatedBy:        "dev-user",
			CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			DeploymentStatus: deploymentStatus,
			Version:          version,
			VersionGUID:      versionGUID,
		}, nil)
}
func mockDeleteAccessKeyVersion(client *cloudaccess.Mock, testData commonDataForAccessKey, version int64) *mock.Call {
	var cloudAccessKeyID, versionGUID string
	if version == firstAccessKeyVersion {
		cloudAccessKeyID = testData.credentialsA.cloudAccessKeyID
		versionGUID = "asde-efdr-reded"
	}
	if version == secondAccessKeyVersion {
		cloudAccessKeyID = testData.credentialsB.cloudAccessKeyID
		versionGUID = "asdd-ads-dasdas"
	}
	return client.On("DeleteAccessKeyVersion", mock.Anything, cloudaccess.DeleteAccessKeyVersionRequest{AccessKeyUID: testData.accessKeyUID, Version: version}).
		Return(&cloudaccess.DeleteAccessKeyVersionResponse{
			AccessKeyUID:     testData.accessKeyUID,
			CloudAccessKeyID: ptr.To(cloudAccessKeyID),
			CreatedBy:        "dev-user",
			CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			DeploymentStatus: cloudaccess.Active,
			Version:          version,
			VersionGUID:      versionGUID,
		}, nil)
}

func mockGetAccessKeyVersionStatus(client *cloudaccess.Mock, testData commonDataForAccessKey, requestID int64, version int64) *mock.Call {
	return client.On("GetAccessKeyVersionStatus", mock.Anything, cloudaccess.GetAccessKeyVersionStatusRequest{RequestID: requestID}).
		Return(&cloudaccess.GetAccessKeyVersionStatusResponse{
			AccessKeyVersion: &cloudaccess.KeyVersion{
				AccessKeyUID: testData.accessKeyUID,
				Version:      version,
			},
			ProcessingStatus: cloudaccess.ProcessingDone,
			RequestDate:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			RequestedBy:      "dev-user",
		}, nil)
}

func mockCreateAccessKeyVersion(client *cloudaccess.Mock, testData commonDataForAccessKey) *mock.Call {
	return client.On("CreateAccessKeyVersion", mock.Anything, cloudaccess.CreateAccessKeyVersionRequest{
		AccessKeyUID: testData.accessKeyUID,
		BodyParams: cloudaccess.CreateAccessKeyVersionBodyParams{
			CloudAccessKeyID:     testData.credentialsB.cloudAccessKeyID,
			CloudSecretAccessKey: testData.credentialsB.cloudSecretAccessKey,
		}}).Return(&cloudaccess.CreateAccessKeyVersionResponse{RequestID: 124, RetryAfter: 1000}, nil)
}

func mockListAccessKeyVersions(client *cloudaccess.Mock, testData commonDataForAccessKey, size int) *mock.Call {
	var listAccessKeyVersionResp cloudaccess.ListAccessKeyVersionsResponse
	if size == twoElementsVersionList {
		listAccessKeyVersionResp = cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
			{
				AccessKeyUID:     testData.accessKeyUID,
				CloudAccessKeyID: ptr.To("test_key_id"),
				CreatedBy:        "dev-user",
				CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
				DeploymentStatus: cloudaccess.Active,
				Version:          firstAccessKeyVersion,
				VersionGUID:      "asde-efdr-reded",
			},
			{
				AccessKeyUID:     testData.accessKeyUID,
				CloudAccessKeyID: ptr.To("test_key_id_2"),
				CreatedBy:        "dev-user",
				CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
				DeploymentStatus: cloudaccess.Active,
				Version:          secondAccessKeyVersion,
				VersionGUID:      "asdd-ads-dasdas",
			},
		}}
	}
	if size == oneElementVersionList {
		listAccessKeyVersionResp = cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
			{
				AccessKeyUID:     testData.accessKeyUID,
				CloudAccessKeyID: ptr.To("test_key_id_2"),
				CreatedBy:        "dev-user",
				CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
				DeploymentStatus: cloudaccess.Active,
				Version:          secondAccessKeyVersion,
				VersionGUID:      "asdd-ads-dasdas",
			},
		},
		}
	}
	if size == emptyVersionList {
		listAccessKeyVersionResp = cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{}}
	}
	return client.On("ListAccessKeyVersions", mock.Anything, cloudaccess.ListAccessKeyVersionsRequest{
		AccessKeyUID: testData.accessKeyUID,
	}).Return(&listAccessKeyVersionResp, nil)
}
func mockListAccessKeyVersionsOnly1Version(client *cloudaccess.Mock, testData commonDataForAccessKey) *mock.Call {
	var listAccessKeyVersionResp = cloudaccess.ListAccessKeyVersionsResponse{AccessKeyVersions: []cloudaccess.AccessKeyVersion{
		{
			AccessKeyUID:     testData.accessKeyUID,
			CloudAccessKeyID: ptr.To("test_key_id"),
			CreatedBy:        "dev-user",
			CreatedTime:      time.Date(2024, 1, 10, 11, 9, 10, 67708, time.UTC),
			DeploymentStatus: cloudaccess.Active,
			Version:          firstAccessKeyVersion,
			VersionGUID:      "asde-efdr-reded",
		},
	},
	}
	return client.On("ListAccessKeyVersions", mock.Anything, cloudaccess.ListAccessKeyVersionsRequest{
		AccessKeyUID: testData.accessKeyUID,
	}).Return(&listAccessKeyVersionResp, nil)
}

func mockLookupsProperties(client *cloudaccess.Mock, testData commonDataForAccessKey, version int64) *mock.Call {
	return client.On("LookupProperties", mock.Anything, cloudaccess.LookupPropertiesRequest{
		AccessKeyUID: testData.accessKeyUID,
		Version:      version,
	}).Return(&cloudaccess.LookupPropertiesResponse{Properties: []cloudaccess.Property{
		{
			AccessKeyUID:      testData.accessKeyUID,
			PropertyID:        "123123",
			PropertyName:      "test_property_name",
			ProductionVersion: ptr.To(int64(1)),
			StagingVersion:    ptr.To(int64(1)),
		},
	}}, nil)
}

func mockLookupsPropertiesNoProperties(client *cloudaccess.Mock, testData *commonDataForProperty, version int64) *mock.Call {
	var lookupPropertiesRes cloudaccess.LookupPropertiesResponse
	if testData != nil {
		lookupPropertiesRes = cloudaccess.LookupPropertiesResponse{Properties: []cloudaccess.Property{}}
	} else {
		lookupPropertiesRes = cloudaccess.LookupPropertiesResponse{Properties: []cloudaccess.Property{}}
	}
	return client.On("LookupProperties", mock.Anything, cloudaccess.LookupPropertiesRequest{
		AccessKeyUID: testData.accessKeyUID,
		Version:      version,
	}).Return(&lookupPropertiesRes, nil)
}
