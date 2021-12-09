package appsec

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v2/pkg/appsec"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/akamai"
	"github.com/akamai/terraform-provider-akamai/v2/pkg/tools"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// appsec v1
//
// https://developer.akamai.com/api/cloud_security/application_security/v1.html
func resourceIPGeo() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIPGeoCreate,
		ReadContext:   resourceIPGeoRead,
		UpdateContext: resourceIPGeoUpdate,
		DeleteContext: resourceIPGeoDelete,
		CustomizeDiff: customdiff.All(
			VerifyIDUnchanged,
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"config_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"security_policy_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					Allow,
					Block,
				}, false)),
			},
			"geo_network_lists": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ip_network_lists": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"exception_ip_network_lists": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}
func resourceIPGeoCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceIPGeoCreate")
	logger.Debugf("in resourceIPGeoCreate")

	configID, err := tools.GetIntValue("config_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	version := getModifiableConfigVersion(ctx, configID, "ipgeo", m)
	policyID, err := tools.GetStringValue("security_policy_id", d)
	if err != nil {
		return diag.FromErr(err)
	}
	mode, err := tools.GetStringValue("mode", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	blockedgeolists := tools.SetToStringSlice(d.Get("geo_network_lists").(*schema.Set))
	blockediplists := tools.SetToStringSlice(d.Get("ip_network_lists").(*schema.Set))
	exceptioniplists := tools.SetToStringSlice(d.Get("exception_ip_network_lists").(*schema.Set))

	createIPGeo := appsec.UpdateIPGeoRequest{
		ConfigID: configID,
		Version:  version,
		PolicyID: policyID,
	}
	if mode == Allow {
		createIPGeo.Block = "blockAllTrafficExceptAllowedIPs"
	}
	if mode == Block {
		createIPGeo.Block = "blockSpecificIPGeo"
	}
	createIPGeo.GeoControls.BlockedIPNetworkLists.NetworkList = blockedgeolists
	createIPGeo.IPControls.BlockedIPNetworkLists.NetworkList = blockediplists
	createIPGeo.IPControls.AllowedIPNetworkLists.NetworkList = exceptioniplists

	_, erru := client.UpdateIPGeo(ctx, createIPGeo)
	if erru != nil {
		logger.Errorf("calling 'createIPGeo': %s", erru.Error())
		return diag.FromErr(erru)
	}

	d.SetId(fmt.Sprintf("%d:%s", createIPGeo.ConfigID, createIPGeo.PolicyID))

	return resourceIPGeoRead(ctx, d, m)
}

func resourceIPGeoRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceIPGeoRead")
	logger.Debugf("in resourceIPGeoRead")

	idParts, err := splitID(d.Id(), 2, "configID:securityPolicyID")
	if err != nil {
		return diag.FromErr(err)
	}
	configID, err := strconv.Atoi(idParts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	version := getLatestConfigVersion(ctx, configID, m)
	policyID := idParts[1]

	getIPGeo := appsec.GetIPGeoRequest{
		ConfigID: configID,
		Version:  version,
		PolicyID: policyID,
	}

	ipgeo, err := client.GetIPGeo(ctx, getIPGeo)
	if err != nil {
		logger.Errorf("calling 'getIPGeo': %s", err.Error())
		return diag.FromErr(err)
	}

	if err := d.Set("config_id", getIPGeo.ConfigID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("security_policy_id", getIPGeo.PolicyID); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if ipgeo.Block == "blockAllTrafficExceptAllowedIPs" {
		if err := d.Set("mode", Allow); err != nil {
			return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
		}
	}
	if ipgeo.Block == "blockSpecificIPGeo" {
		if err := d.Set("mode", Block); err != nil {
			return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
		}
	}
	if err := d.Set("geo_network_lists", ipgeo.GeoControls.BlockedIPNetworkLists.NetworkList); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("ip_network_lists", ipgeo.IPControls.BlockedIPNetworkLists.NetworkList); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}
	if err := d.Set("exception_ip_network_lists", ipgeo.IPControls.AllowedIPNetworkLists.NetworkList); err != nil {
		return diag.Errorf("%s: %s", tools.ErrValueSet, err.Error())
	}

	return nil
}

func resourceIPGeoUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceIPGeoUpdate")
	logger.Debugf("in resourceIPGeoUpdate")

	idParts, err := splitID(d.Id(), 2, "configID:securityPolicyID")
	if err != nil {
		return diag.FromErr(err)
	}
	configID, err := strconv.Atoi(idParts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	version := getModifiableConfigVersion(ctx, configID, "ipgeo", m)
	policyID := idParts[1]
	mode, err := tools.GetStringValue("mode", d)
	if err != nil && !errors.Is(err, tools.ErrNotFound) {
		return diag.FromErr(err)
	}
	blockedgeolists := tools.SetToStringSlice(d.Get("geo_network_lists").(*schema.Set))
	blockediplists := tools.SetToStringSlice(d.Get("ip_network_lists").(*schema.Set))
	exceptioniplists := tools.SetToStringSlice(d.Get("exception_ip_network_lists").(*schema.Set))

	updateIPGeo := appsec.UpdateIPGeoRequest{
		ConfigID: configID,
		Version:  version,
		PolicyID: policyID,
	}
	if mode == Allow {
		updateIPGeo.Block = "blockAllTrafficExceptAllowedIPs"
	}
	if mode == Block {
		updateIPGeo.Block = "blockSpecificIPGeo"
	}

	updateIPGeo.GeoControls.BlockedIPNetworkLists.NetworkList = blockedgeolists
	updateIPGeo.IPControls.BlockedIPNetworkLists.NetworkList = blockediplists
	updateIPGeo.IPControls.AllowedIPNetworkLists.NetworkList = exceptioniplists

	_, erru := client.UpdateIPGeo(ctx, updateIPGeo)
	if erru != nil {
		logger.Errorf("calling 'updateIPGeo': %s", erru.Error())
		return diag.FromErr(erru)
	}

	return resourceIPGeoRead(ctx, d, m)
}

func resourceIPGeoDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := akamai.Meta(m)
	client := inst.Client(meta)
	logger := meta.Log("APPSEC", "resourceIPGeoDelete")
	logger.Debugf("in resourceIPGeoDelete")

	idParts, err := splitID(d.Id(), 2, "configID:securityPolicyID")
	if err != nil {
		return diag.FromErr(err)
	}
	configID, err := strconv.Atoi(idParts[0])
	if err != nil {
		return diag.FromErr(err)
	}
	version := getModifiableConfigVersion(ctx, configID, "ipgeo", m)
	policyID := idParts[1]

	getPolicyProtectionsRequest := appsec.GetPolicyProtectionsRequest{
		ConfigID: configID,
		Version:  version,
		PolicyID: policyID,
	}
	policyProtections, err := client.GetPolicyProtections(ctx, getPolicyProtectionsRequest)
	if err != nil {
		logger.Errorf("calling GetPolicyProtections: %s", err.Error())
		return diag.FromErr(err)
	}

	updatePolicyProtectionsRequest := appsec.UpdatePolicyProtectionsRequest{
		ConfigID:                      configID,
		Version:                       version,
		PolicyID:                      policyID,
		ApplyAPIConstraints:           policyProtections.ApplyAPIConstraints,
		ApplyApplicationLayerControls: policyProtections.ApplyApplicationLayerControls,
		ApplyBotmanControls:           policyProtections.ApplyBotmanControls,
		ApplyNetworkLayerControls:     false,
		ApplyRateControls:             policyProtections.ApplyRateControls,
		ApplyReputationControls:       policyProtections.ApplyReputationControls,
		ApplySlowPostControls:         policyProtections.ApplySlowPostControls,
	}
	_, err = client.UpdatePolicyProtections(ctx, updatePolicyProtectionsRequest)
	if err != nil {
		logger.Errorf("calling UpdatePolicyProtections: %s", err.Error())
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

// Definition of constant variables
const (
	Allow = "allow"
	Block = "block"
)