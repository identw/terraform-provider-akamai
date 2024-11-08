package property

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v8/pkg/papi"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v8/pkg/session"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/str"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/tf"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/timeouts"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/logger"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourcePropertyIncludeActivation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePropertyIncludeActivationCreate,
		ReadContext:   resourcePropertyIncludeActivationRead,
		UpdateContext: resourcePropertyIncludeActivationUpdate,
		DeleteContext: resourcePropertyIncludeActivationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourcePropertyIncludeActivationImport,
		},
		Schema: map[string]*schema.Schema{
			"include_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				StateFunc:   addPrefixToState("inc_"),
				Description: "The unique identifier of the include",
			},
			"contract_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				StateFunc:   addPrefixToState("ctr_"),
				Description: "The contract under which the include is activated",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				StateFunc:   addPrefixToState("grp_"),
				Description: "The group under which the include is activated",
			},
			"version": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The unique identifier of the include",
			},
			"network": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					string(papi.ActivationNetworkStaging), string(papi.ActivationNetworkProduction),
				}, false)),
				Description: "The network for which the activation will be performed",
			},
			"notify_emails": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "The list of email addresses to notify about an activation status",
			},
			"note": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "The note to assign to a log message of the activation request",
				Default:          "",
				DiffSuppressFunc: suppressNoteFieldForIncludeActivation,
			},
			"auto_acknowledge_rule_warnings": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Automatically acknowledge all rule warnings for activation and continue",
			},
			"validations": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The validation information in JSON format",
			},
			"compliance_record": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Provides an audit record when activating on a production network",
				Elem:        complianceRecordSchema,
			},
			"timeouts": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Enables to set timeout for processing",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: timeouts.ValidateDurationFormat,
						},
					},
				},
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Default: readTimeoutFromEnvOrDefault("AKAMAI_ACTIVATION_TIMEOUT", includeActivationTimeout),
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{{
			Version: 0,
			Type:    resourcePropertyIncludeActivationV0().CoreConfigSchema().ImpliedType(),
			Upgrade: timeouts.MigrateToExplicit(),
		}},
	}
}

func readTimeoutFromEnvOrDefault(name string, timeout time.Duration) *time.Duration {
	logger := logger.Get("readTimeoutFromEnvOrDefault")

	value := os.Getenv(name)
	if value != "" {
		n, err := strconv.Atoi(value)
		if err != nil {
			logger.Errorf("Provided timeout value %q is not a valid number: %s", n, err)
		} else {
			timeout = time.Minute * time.Duration(n)
		}
	}
	logger.Infof("using activation timeout value of %d minutes", timeout/time.Minute)
	return &timeout
}

var (
	activationPollInterval   = time.Minute
	includeActivationTimeout = time.Minute * 30
	getActivationInterval    = time.Second * 5
)

func resourcePropertyIncludeActivationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("PAPI", "resourcePropertyIncludeActivationCreate")
	ctx = session.ContextWithOptions(ctx, session.WithContextLog(logger))
	client := Client(meta)

	logger.Debug("Create property include activation")

	err := resourcePropertyIncludeActivationUpsert(ctx, d, client)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePropertyIncludeActivationRead(ctx, d, m)
}

func resourcePropertyIncludeActivationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("PAPI", "resourcePropertyIncludeActivationRead")
	ctx = session.ContextWithOptions(ctx, session.WithContextLog(logger))
	client := Client(meta)
	logger.Debug("Reading property include activation")

	rd, err := parsePropertyIncludeActivationResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	activation, err := getLatestActiveIncludeActivationResponseInNetwork(ctx, client, rd)
	if err != nil {
		return diag.FromErr(err)
	}

	if activation.Activation.ActivationType == papi.ActivationTypeDeactivate {
		logger.Info("include is deactivated, needs recreation")
		d.SetId("")
		return nil
	}

	var validations []byte
	if activation.Validations != nil {
		validations, err = json.Marshal(activation.Validations)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	attrs := make(map[string]interface{})
	attrs["include_id"] = activation.Activation.IncludeID
	attrs["contract_id"] = activation.ContractID
	attrs["group_id"] = activation.GroupID
	attrs["version"] = activation.Activation.IncludeVersion
	attrs["network"] = activation.Activation.Network
	attrs["notify_emails"] = activation.Activation.NotifyEmails
	attrs["note"] = activation.Activation.Note
	attrs["validations"] = string(validations)

	if len(strings.TrimSpace(activation.Activation.Note)) == 0 {
		attrs["note"] = ""
	}

	// it is impossible to fetch compliance_record and auto_acknowledge_rule_warnings attributes from server
	if err = tf.SetAttrs(d, attrs); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourcePropertyIncludeActivationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("PAPI", "resourcePropertyIncludeActivationUpdate")
	ctx = session.ContextWithOptions(ctx, session.WithContextLog(logger))
	client := Client(meta)
	logger.Debug("Updating property include activation")

	mutableAttrsHaveChanges := d.HasChanges("notify_emails", "auto_acknowledge_rule_warnings", "compliance_record")
	if mutableAttrsHaveChanges && !d.HasChanges("version") {
		return diag.FromErr(fmt.Errorf("attributes such as 'notify_emails', 'auto_acknowledge_rule_warnings', " +
			"'compliance_record' cannot be updated after resource creation without 'version' attribute modification"))
	}

	if !d.HasChangeExcept("timeouts") {
		logger.Debug("Only timeouts were updated, skipping")
		return nil
	}

	err := resourcePropertyIncludeActivationUpsert(ctx, d, client)
	if err != nil {
		return diag.FromErr(err)
	}
	return resourcePropertyIncludeActivationRead(ctx, d, m)
}

func resourcePropertyIncludeActivationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("PAPI", "resourcePropertyIncludeActivationDelete")
	ctx = session.ContextWithOptions(ctx, session.WithContextLog(logger))
	client := Client(meta)
	logger.Debug("Deactivating property include")

	activationResourceData := propertyIncludeActivationData{}
	if err := activationResourceData.populateFromResource(d); err != nil {
		return diag.FromErr(err)
	}

	logger.Debug("waiting for pending (de)activations")
	if err := waitUntilNoPendingActivationInNetwork(ctx, client, activationResourceData); err != nil {
		return diag.FromErr(err)
	}

	expectedIsActive, err := isLatestActiveExpectedDeactivated(ctx, client, activationResourceData)
	if err != nil {
		return diag.FromErr(err)
	}
	if expectedIsActive {
		// we are done here
		logger.Debug("include version already deactivated")
		return nil
	}

	logger.Debug("creating new deactivation")
	err = createNewDeactivation(ctx, client, activationResourceData)
	if err != nil {
		return diag.FromErr(err)
	}

	logger.Debug("waiting for pending deactivation")
	if err := waitUntilNoPendingActivationInNetwork(ctx, client, activationResourceData); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePropertyIncludeActivationImport(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	meta := meta.Must(m)
	logger := meta.Log("PAPI", "resourcePropertyIncludeActivationImport")
	logger.Debug("Importing property include activation")

	rd, err := parsePropertyIncludeActivationResourceID(d.Id())
	if err != nil {
		return nil, err
	}

	attrs := make(map[string]interface{})
	attrs["contract_id"] = rd.contractID
	attrs["group_id"] = rd.groupID
	attrs["include_id"] = rd.includeID
	attrs["network"] = rd.network

	// it is impossible to fetch auto_acknowledge_rule_warnings from server
	attrs["auto_acknowledge_rule_warnings"] = false

	if err := tf.SetAttrs(d, attrs); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourcePropertyIncludeActivationUpsert(ctx context.Context, d *schema.ResourceData, client papi.PAPI) error {
	logger := logger.Get("resourcePropertyIncludeActivationUpsert")

	activationResourceData := propertyIncludeActivationData{}
	if err := activationResourceData.populateFromResource(d); err != nil {
		return err
	}

	logger.Debug("waiting for pending activations")
	if err := waitUntilNoPendingActivationInNetwork(ctx, client, activationResourceData); err != nil {
		return err
	}

	logger.Debug("checking if include version is already active")
	expectedIsActive, err := isLatestActiveExpectedActivated(ctx, client, activationResourceData)
	if err != nil && !errors.Is(err, ErrNoLatestIncludeActivation) {
		return err
	}
	if expectedIsActive {
		// we are done here
		logger.Debug("include version already active")
		d.SetId(fmt.Sprintf("%s:%s:%s:%s", activationResourceData.contractID, activationResourceData.groupID, activationResourceData.includeID, activationResourceData.network))
		return nil
	}

	logger.Debug("creating new activation")
	err = createNewActivation(ctx, client, activationResourceData)
	if err != nil {
		return err
	}

	logger.Debug("waiting for pending activations")
	if err := waitUntilNoPendingActivationInNetwork(ctx, client, activationResourceData); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s:%s:%s", activationResourceData.contractID, activationResourceData.groupID, activationResourceData.includeID, activationResourceData.network))
	return nil
}

type propertyIncludeActivationData struct {
	includeID        string
	contractID       string
	groupID          string
	version          int
	network          string
	notifyEmails     []string
	note             string
	acknowledgement  bool
	complianceRecord []any
}

func (p *propertyIncludeActivationData) populateFromResource(d *schema.ResourceData) error {
	includeID, err := tf.GetStringValue("include_id", d)
	if err != nil {
		return err
	}
	p.includeID = str.AddPrefix(includeID, "inc_")

	contractID, err := tf.GetStringValue("contract_id", d)
	if err != nil {
		return err
	}
	p.contractID = str.AddPrefix(contractID, "ctr_")
	groupID, err := tf.GetStringValue("group_id", d)
	if err != nil {
		return err
	}
	p.groupID = str.AddPrefix(groupID, "grp_")
	p.network, err = tf.GetStringValue("network", d)
	if err != nil {
		return err
	}
	p.version, err = tf.GetIntValue("version", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return err
	}
	notifyEmailsSet, err := tf.GetSetValue("notify_emails", d)
	if err != nil {
		return err
	}
	p.notifyEmails = tf.SetToStringSlice(notifyEmailsSet)
	p.note, err = tf.GetStringValue("note", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return err
	}
	p.acknowledgement, err = tf.GetBoolValue("auto_acknowledge_rule_warnings", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return err
	}
	p.complianceRecord, err = tf.GetListValue("compliance_record", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return err
	}
	return nil
}

type propertyIncludeActivationID struct {
	contractID string
	groupID    string
	includeID  string
	network    string
}

func parsePropertyIncludeActivationResourceID(activationResourceID string) (*propertyIncludeActivationID, error) {
	id := strings.Split(activationResourceID, ":")
	if len(id) != 4 {
		return nil, fmt.Errorf("invalid include activation identifier: %s", activationResourceID)
	}
	contractID, groupID, includeID, network := id[0], id[1], id[2], id[3]
	return &propertyIncludeActivationID{
		contractID: contractID,
		groupID:    groupID,
		includeID:  includeID,
		network:    network,
	}, nil
}

func waitUntilNoPendingActivationInNetwork(ctx context.Context, client papi.PAPI, activationResourceData propertyIncludeActivationData) error {
	act, err := findLatestActivationInNetwork(ctx, client, &propertyIncludeActivationID{
		contractID: activationResourceData.contractID,
		groupID:    activationResourceData.groupID,
		includeID:  activationResourceData.includeID,
		network:    activationResourceData.network,
	})
	if errors.Is(err, ErrNoLatestIncludeActivation) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = waitForActivationCondition(ctx, client, activationResourceData.includeID, act.ActivationID,
		func(status papi.ActivationStatus) bool {
			return status == papi.ActivationStatusActive ||
				status == papi.ActivationStatusFailed ||
				status == papi.ActivationStatusAborted ||
				status == papi.ActivationStatusDeactivated
		})
	if err != nil {
		return err
	}

	return nil
}

func isLatestActiveExpectedWithActivationType(ctx context.Context, client papi.PAPI, activationResourceData propertyIncludeActivationData, expectedActivationType papi.ActivationType) (bool, error) {
	activation, err := getLatestActiveActivationInNetwork(ctx, client, &propertyIncludeActivationID{
		contractID: activationResourceData.contractID,
		groupID:    activationResourceData.groupID,
		includeID:  activationResourceData.includeID,
		network:    activationResourceData.network,
	})
	if errors.Is(err, ErrNoLatestIncludeActivation) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// expected conditions
	if activation.Status == papi.ActivationStatusActive &&
		activation.ActivationType == expectedActivationType &&
		activation.IncludeVersion == activationResourceData.version {
		return true, nil
	}
	return false, nil
}

func isLatestActiveExpectedDeactivated(ctx context.Context, client papi.PAPI, activationResourceData propertyIncludeActivationData) (bool, error) {
	return isLatestActiveExpectedWithActivationType(ctx, client, activationResourceData, papi.ActivationTypeDeactivate)
}

func isLatestActiveExpectedActivated(ctx context.Context, client papi.PAPI, activationResourceData propertyIncludeActivationData) (bool, error) {
	return isLatestActiveExpectedWithActivationType(ctx, client, activationResourceData, papi.ActivationTypeActivate)
}

func createNewActivation(ctx context.Context, client papi.PAPI, activationResourceData propertyIncludeActivationData) error {
	logger := logger.Get("createNewActivation")

	logger.Debug("preparing activation request")
	activateIncludeRequest := papi.ActivateIncludeRequest{
		IncludeID:              activationResourceData.includeID,
		Version:                activationResourceData.version,
		Network:                papi.ActivationNetwork(activationResourceData.network),
		Note:                   activationResourceData.note,
		NotifyEmails:           activationResourceData.notifyEmails,
		AcknowledgeAllWarnings: activationResourceData.acknowledgement,
	}

	activateIncludeRequest = papi.ActivateIncludeRequest(addComplianceRecord(activationResourceData.complianceRecord, papi.ActivateOrDeactivateIncludeRequest(activateIncludeRequest)))

	logger.Debug("sending include activation request")
	activationResponse, err := client.ActivateInclude(ctx, activateIncludeRequest)
	if err != nil {
		return err
	}

	logger.Debug("waiting for activation creation")
	// here is used temporary activationID
	if _, err := waitForActivationCreation(ctx, client, activationResourceData.includeID, activationResponse.ActivationID); err != nil {
		return err
	}
	return nil
}

func createNewDeactivation(ctx context.Context, client papi.PAPI, activationResourceData propertyIncludeActivationData) error {
	logger := logger.Get("createNewDeactivation")

	deactivateIncludeRequest := papi.DeactivateIncludeRequest{
		IncludeID:              activationResourceData.includeID,
		Version:                activationResourceData.version,
		Network:                papi.ActivationNetwork(activationResourceData.network),
		Note:                   activationResourceData.note,
		NotifyEmails:           activationResourceData.notifyEmails,
		AcknowledgeAllWarnings: activationResourceData.acknowledgement,
	}

	deactivateIncludeRequest = papi.DeactivateIncludeRequest(addComplianceRecord(activationResourceData.complianceRecord, papi.ActivateOrDeactivateIncludeRequest(deactivateIncludeRequest)))

	deactivation, err := client.DeactivateInclude(ctx, deactivateIncludeRequest)
	if err != nil {
		return err
	}

	logger.Info("waiting for creation of include deactivation")
	if _, err := waitForActivationCreation(ctx, client, activationResourceData.includeID, deactivation.ActivationID); err != nil {
		return err
	}

	return nil
}

func findLatestActivationWithCondition(ctx context.Context, client papi.PAPI, activationResourceID *propertyIncludeActivationID,
	cond func(papi.IncludeActivation) bool) (*papi.IncludeActivation, error) {
	versions, err := client.ListIncludeActivations(ctx, papi.ListIncludeActivationsRequest{
		ContractID: activationResourceID.contractID,
		GroupID:    activationResourceID.groupID,
		IncludeID:  activationResourceID.includeID,
	})
	if err != nil {
		return nil, err
	}
	activations := versions.Activations.Items
	if len(activations) == 0 {
		return nil, ErrNoLatestIncludeActivation
	}

	sort.Slice(activations, func(i, j int) bool {
		return activations[i].UpdateDate > activations[j].UpdateDate
	})

	for _, v := range activations {
		if cond(v) {
			return &v, nil
		}
	}
	return nil, ErrNoLatestIncludeActivation
}

func findLatestActivationInNetwork(ctx context.Context, client papi.PAPI, activationResourceID *propertyIncludeActivationID) (*papi.IncludeActivation, error) {
	return findLatestActivationWithCondition(ctx, client, activationResourceID,
		func(ia papi.IncludeActivation) bool {
			return ia.Network == papi.ActivationNetwork(activationResourceID.network)
		})
}

func getLatestActiveActivationInNetwork(ctx context.Context, client papi.PAPI, activationResourceID *propertyIncludeActivationID) (*papi.IncludeActivation, error) {
	act, err := findLatestActivationWithCondition(ctx, client, activationResourceID,
		func(ia papi.IncludeActivation) bool {
			return ia.Status == papi.ActivationStatusActive &&
				ia.Network == papi.ActivationNetwork(activationResourceID.network)
		})
	if err != nil {
		return nil, err
	}
	return act, nil
}

func getLatestActiveIncludeActivationResponseInNetwork(ctx context.Context, client papi.PAPI, activationResourceID *propertyIncludeActivationID) (*papi.GetIncludeActivationResponse, error) {
	act, err := getLatestActiveActivationInNetwork(ctx, client, activationResourceID)
	if err != nil {
		return nil, err
	}

	activation, err := client.GetIncludeActivation(ctx, papi.GetIncludeActivationRequest{
		IncludeID:    activationResourceID.includeID,
		ActivationID: act.ActivationID,
	})
	if err != nil {
		return nil, err
	}

	return activation, nil
}

func waitForActivationCreation(ctx context.Context, client papi.PAPI, includeID, activationID string) (*papi.GetIncludeActivationResponse, error) {
	for {
		activation, err := client.GetIncludeActivation(ctx, papi.GetIncludeActivationRequest{
			IncludeID:    includeID,
			ActivationID: activationID,
		})
		if err == nil {
			return activation, nil
		}

		if errors.Is(err, papi.ErrMissingComplianceRecord) {
			return nil, fmt.Errorf("for 'PRODUCTION' network, 'compliance_record' must be specified: %s", err)
		}
		if !errors.Is(err, papi.ErrNotFound) {
			// return in case we get unexpected error
			return nil, err
		}

		select {
		case <-time.After(getActivationInterval):
			// wait some time and check again
			continue
		}
	}
}

func waitForActivationCondition(ctx context.Context,
	client papi.PAPI,
	includeID, activationID string,
	cond func(papi.ActivationStatus) bool,
) (*papi.GetIncludeActivationResponse, error) {
	for {
		activation, err := client.GetIncludeActivation(ctx, papi.GetIncludeActivationRequest{
			IncludeID:    includeID,
			ActivationID: activationID,
		})
		if err != nil {
			return nil, err
		}

		actStatus := activation.Activation.Status
		if cond(actStatus) {
			return activation, nil
		}

		select {
		case <-time.After(activationPollInterval):
			continue
		case <-ctx.Done():
			return nil, terminateProcess(ctx, string(actStatus))
		}
	}
}

func terminateProcess(ctx context.Context, actStatus string) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("timeout waiting for activation status: current status: %s", actStatus)
	}
	if errors.Is(ctx.Err(), context.Canceled) {
		return fmt.Errorf("operation canceled while waiting for activation status, current status: %s", actStatus)
	}
	return fmt.Errorf("activation context terminated: %w", ctx.Err())
}

func addComplianceRecord(complianceRecord []interface{}, activateIncludeRequest papi.ActivateOrDeactivateIncludeRequest) papi.ActivateOrDeactivateIncludeRequest {
	if len(complianceRecord) == 0 {
		return activateIncludeRequest
	}

	crMap := complianceRecord[0].(map[string]interface{})

	if len(crMap["noncompliance_reason_none"].([]interface{})) != 0 {
		complianceRecordNone := &papi.ComplianceRecordNone{}
		if crMap["noncompliance_reason_none"].([]interface{})[0] != nil {
			crNoneMap := crMap["noncompliance_reason_none"].([]interface{})[0].(map[string]interface{})
			complianceRecordNone = &papi.ComplianceRecordNone{
				CustomerEmail:  crNoneMap["customer_email"].(string),
				PeerReviewedBy: crNoneMap["peer_reviewed_by"].(string),
				TicketID:       crNoneMap["ticket_id"].(string),
				UnitTested:     crNoneMap["unit_tested"].(bool),
			}
		}
		activateIncludeRequest.ComplianceRecord = complianceRecordNone
	} else if len(crMap["noncompliance_reason_other"].([]interface{})) != 0 {
		complianceRecordOther := &papi.ComplianceRecordOther{}
		if crMap["noncompliance_reason_other"].([]interface{})[0] != nil {
			crOtherMap := crMap["noncompliance_reason_other"].([]interface{})[0].(map[string]interface{})
			complianceRecordOther = &papi.ComplianceRecordOther{
				TicketID:                 crOtherMap["ticket_id"].(string),
				OtherNoncomplianceReason: crOtherMap["other_noncompliance_reason"].(string),
			}
		}
		activateIncludeRequest.ComplianceRecord = complianceRecordOther
	} else if len(crMap["noncompliance_reason_no_production_traffic"].([]interface{})) != 0 {
		complianceRecordNoProductionTraffic := &papi.ComplianceRecordNoProductionTraffic{}
		if crMap["noncompliance_reason_no_production_traffic"].([]interface{})[0] != nil {
			crNoProdTrafficMap := crMap["noncompliance_reason_no_production_traffic"].([]interface{})[0].(map[string]interface{})
			complianceRecordNoProductionTraffic = &papi.ComplianceRecordNoProductionTraffic{
				TicketID: crNoProdTrafficMap["ticket_id"].(string),
			}
		}
		activateIncludeRequest.ComplianceRecord = complianceRecordNoProductionTraffic
	} else if len(crMap["noncompliance_reason_emergency"].([]interface{})) != 0 {
		complianceRecordEmergency := &papi.ComplianceRecordEmergency{}
		if crMap["noncompliance_reason_emergency"].([]interface{})[0] != nil {
			crEmergencyMap := crMap["noncompliance_reason_emergency"].([]interface{})[0].(map[string]interface{})
			complianceRecordEmergency = &papi.ComplianceRecordEmergency{
				TicketID: crEmergencyMap["ticket_id"].(string),
			}
		}
		activateIncludeRequest.ComplianceRecord = complianceRecordEmergency
	}

	return activateIncludeRequest
}

func suppressNoteFieldForIncludeActivation(_, oldValue, newValue string, d *schema.ResourceData) bool {
	if oldValue != newValue && d.HasChanges("version", "network") {
		return false
	}
	return true
}
