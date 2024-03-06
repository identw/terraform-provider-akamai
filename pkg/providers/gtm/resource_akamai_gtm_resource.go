package gtm

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/gtm"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/session"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/tf"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/logger"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceGTMv1Resource() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceGTMv1ResourceCreate,
		ReadContext:   resourceGTMv1ResourceRead,
		UpdateContext: resourceGTMv1ResourceUpdate,
		DeleteContext: resourceGTMv1ResourceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGTMv1ResourceImport,
		},
		Schema: map[string]*schema.Schema{
			"domain": {
				Type:     schema.TypeString,
				Required: true,
			},
			"wait_on_complete": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host_header": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"aggregation_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"least_squares_decay": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"upper_bound": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"leader_string": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"constrained_property": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"load_imbalance_percentage": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"max_u_multiplicative_increment": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"decay_rate": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"resource_instance": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: resourceInstanceDiffSuppress,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"datacenter_id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"use_default_load_object": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"load_object": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"load_servers": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"load_object_port": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

// Create a new GTM Resource
func resourceGTMv1ResourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("Akamai GTM", "resourceGTMv1ResourceCreate")
	// create a context with logging for api calls
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)

	domain, err := tf.GetStringValue("domain", d)
	if err != nil {
		return diag.FromErr(err)
	}

	name, err := tf.GetStringValue("name", d)
	if err != nil {
		return diag.FromErr(err)
	}
	var diags diag.Diagnostics
	logger.Infof("Creating resource [%s] in domain [%s]", name, domain)
	newRsrc, err := populateNewResourceObject(ctx, meta, d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	logger.Debugf("Proposed New Resource: [%v]", newRsrc)
	cStatus, err := Client(meta).CreateResource(ctx, newRsrc, domain)
	if err != nil {
		logger.Errorf("Resource Create failed: %s", err.Error())
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Resource Create failed",
			Detail:   err.Error(),
		})
	}
	logger.Debugf("Resource Create status: %v", cStatus.Status)
	if cStatus.Status.PropagationStatus == "DENIED" {
		logger.Errorf(cStatus.Status.Message)
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  cStatus.Status.Message,
		})
	}

	waitOnComplete, err := tf.GetBoolValue("wait_on_complete", d)
	if err != nil {
		return diag.FromErr(err)
	}

	if waitOnComplete {
		done, err := waitForCompletion(ctx, domain, m)
		if done {
			logger.Infof("Resource Create completed")
		} else {
			if err == nil {
				logger.Infof("Resource Create pending")
			} else {
				logger.Errorf("Resource Create failed [%s]", err.Error())
				return append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Resource Create failed",
					Detail:   err.Error(),
				})
			}
		}
	}

	// Give terraform the ID. Format domain:resource
	resourceID := fmt.Sprintf("%s:%s", domain, cStatus.Resource.Name)
	logger.Debugf("Generated Resource. Resource ID: %s", resourceID)
	d.SetId(resourceID)
	return resourceGTMv1ResourceRead(ctx, d, m)

}

// read resource. updates state with entire API result configuration.
func resourceGTMv1ResourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("Akamai GTM", "resourceGTMv1ResourceRead")
	// create a context with logging for api calls
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)

	logger.Debugf("Reading Resource: %s", d.Id())
	var diags diag.Diagnostics
	// retrieve the property and domain
	domain, resource, err := parseResourceStringID(d.Id())
	if err != nil {
		logger.Errorf("Invalid resource Resource ID")
		return diag.FromErr(err)
	}
	rsrc, err := Client(meta).GetResource(ctx, resource, domain)
	if err != nil {
		logger.Errorf("Resource Read failed: %s", err.Error())
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Resource Read error",
			Detail:   err.Error(),
		})
	}
	populateTerraformResourceState(d, rsrc, m)
	logger.Debugf("READ %v", rsrc)
	return nil
}

// Update GTM Resource
func resourceGTMv1ResourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("Akamai GTM", "resourceGTMv1ResourceUpdate")
	// create a context with logging for api calls
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)

	logger.Infof("Updating Resource %s", d.Id())
	var diags diag.Diagnostics
	// pull domain and resource out of id
	domain, resource, err := parseResourceStringID(d.Id())
	if err != nil {
		logger.Errorf("Invalid resource ID")
		return diag.FromErr(err)
	}
	// Get existing property
	existRsrc, err := Client(meta).GetResource(ctx, resource, domain)
	if err != nil {
		logger.Errorf("Resource Update failed: %s", err.Error())
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Update Resource Read error",
			Detail:   err.Error(),
		})
	}
	logger.Debugf("Updating Resource BEFORE: %v", existRsrc)
	if err := populateResourceObject(ctx, d, existRsrc, m); err != nil {
		return diag.FromErr(err)
	}
	logger.Debugf("Updating Resource PROPOSED: %v", existRsrc)
	uStat, err := Client(meta).UpdateResource(ctx, existRsrc, domain)
	if err != nil {
		logger.Errorf("Resource Update failed: %s", err.Error())
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Resource Update error",
			Detail:   err.Error(),
		})
	}
	logger.Debugf("Resource Update status: %v", uStat)
	if uStat.PropagationStatus == "DENIED" {
		logger.Errorf(uStat.Message)
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  uStat.Message,
		})
	}

	waitOnComplete, err := tf.GetBoolValue("wait_on_complete", d)
	if err != nil {
		return diag.FromErr(err)
	}

	if waitOnComplete {
		done, err := waitForCompletion(ctx, domain, m)
		if done {
			logger.Infof("Resource update completed")
		} else {
			if err == nil {
				logger.Infof("Resource update pending")
			} else {
				logger.Errorf("Resource update failed [%s]", err.Error())
				return append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Resource Update failed",
					Detail:   err.Error(),
				})
			}
		}
	}

	return resourceGTMv1ResourceRead(ctx, d, m)
}

// Import GTM Resource.
func resourceGTMv1ResourceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	meta := meta.Must(m)
	logger := meta.Log("Akamai GTM", "resourceGTMv1ResourceImport")
	// create a context with logging for api calls
	ctx := context.Background()
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)

	logger.Infof("Resource [%s] Import", d.Id())
	// pull domain and resource out of resource id
	domain, resource, err := parseResourceStringID(d.Id())
	if err != nil {
		return []*schema.ResourceData{d}, err
	}
	rsrc, err := Client(meta).GetResource(ctx, resource, domain)
	if err != nil {
		return nil, err
	}
	_ = d.Set("domain", domain)
	_ = d.Set("wait_on_complete", true)
	populateTerraformResourceState(d, rsrc, m)

	// use same Id as passed in
	name, _ := tf.GetStringValue("name", d)
	logger.Infof("Resource [%s] [%s] Imported", d.Id(), name)
	return []*schema.ResourceData{d}, nil
}

// Delete GTM Resource.
func resourceGTMv1ResourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("Akamai GTM", "resourceGTMv1ResourceDelete")
	// create a context with logging for api calls
	ctx = session.ContextWithOptions(
		ctx,
		session.WithContextLog(logger),
	)

	logger.Debugf("Deleting Resource: %s", d.Id())
	var diags diag.Diagnostics
	// Get existing resource
	domain, resource, err := parseResourceStringID(d.Id())
	if err != nil {
		logger.Errorf("Invalid resource ID")
		return diag.FromErr(err)
	}
	existRsrc, err := Client(meta).GetResource(ctx, resource, domain)
	if err != nil {
		logger.Errorf("Resource Delete Read failed: %s", err.Error())
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Delete Resource Read error",
			Detail:   err.Error(),
		})
	}
	logger.Debugf("Deleting Resource: %v", existRsrc)
	uStat, err := Client(meta).DeleteResource(ctx, existRsrc, domain)
	if err != nil {
		logger.Errorf("Resource Delete failed: %s", err.Error())
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Delete Resource error",
			Detail:   err.Error(),
		})
	}
	logger.Debugf("Resource Delete status: %v", uStat)
	if uStat.PropagationStatus == "DENIED" {
		logger.Errorf(uStat.Message)
		return append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  uStat.Message,
		})
	}

	waitOnComplete, err := tf.GetBoolValue("wait_on_complete", d)
	if err != nil {
		return diag.FromErr(err)
	}

	if waitOnComplete {
		done, err := waitForCompletion(ctx, domain, m)
		if done {
			logger.Infof("Resource Delete completed")
		} else {
			if err == nil {
				logger.Infof("Resource Delete pending")
			} else {
				logger.Errorf("Resource Delete failed [%s]", err.Error())
				return append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Resource Delete failed",
					Detail:   err.Error(),
				})
			}
		}
	}

	// if successful ....
	d.SetId("")
	return nil
}

// Create and populate a new resource object from resource data
func populateNewResourceObject(ctx context.Context, meta meta.Meta, d *schema.ResourceData, m interface{}) (*gtm.Resource, error) {

	name, _ := tf.GetStringValue("name", d)
	rsrcObj := Client(meta).NewResource(ctx, name)
	rsrcObj.ResourceInstances = make([]*gtm.ResourceInstance, 0)
	err := populateResourceObject(ctx, d, rsrcObj, m)

	return rsrcObj, err

}

// nolint:gocyclo
// Populate existing resource object from resource data
func populateResourceObject(ctx context.Context, d *schema.ResourceData, rsrc *gtm.Resource, m interface{}) error {
	meta := meta.Must(m)
	logger := meta.Log("Akamai GTM", "resourceGTMv1ResourceDelete")

	vstr, err := tf.GetStringValue("name", d)
	if err == nil {
		rsrc.Name = vstr
	}
	vstr, err = tf.GetStringValue("type", d)
	if err == nil {
		rsrc.Type = vstr
	}
	vstr, err = tf.GetStringValue("host_header", d)
	if err == nil || d.HasChange("host_header") {
		rsrc.HostHeader = vstr
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() host_header failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	vfloat, err := tf.GetFloat64Value("least_squares_decay", d)
	if err == nil || d.HasChange("least_squares_decay") {
		rsrc.LeastSquaresDecay = vfloat
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() least_squares_decay failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	vint, err := tf.GetIntValue("upper_bound", d)
	if err == nil || d.HasChange("upper_bound") {
		rsrc.UpperBound = vint
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() upper_bound failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	vstr, err = tf.GetStringValue("description", d)
	if err == nil || d.HasChange("description") {
		rsrc.Description = vstr
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() description failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	vstr, err = tf.GetStringValue("leader_string", d)
	if err == nil || d.HasChange("leader_string") {
		rsrc.LeaderString = vstr
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() leader_string failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	vstr, err = tf.GetStringValue("constrained_property", d)
	if err == nil || d.HasChange("constrained_property") {
		rsrc.ConstrainedProperty = vstr
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() constrained_property failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	vstr, err = tf.GetStringValue("aggregation_type", d)
	if err == nil {
		rsrc.AggregationType = vstr
	}

	vfloat, err = tf.GetFloat64Value("load_imbalance_percentage", d)
	if err == nil || d.HasChange("load_imbalance_percentage") {
		rsrc.LoadImbalancePercentage = vfloat
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() load_imbalance_percentage failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	vfloat, err = tf.GetFloat64Value("max_u_multiplicative_increment", d)
	if err == nil || d.HasChange("max_u_multiplicative_increment") {
		rsrc.MaxUMultiplicativeIncrement = vfloat
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() max_u_multiplicative_increment failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	vfloat, err = tf.GetFloat64Value("decay_rate", d)
	if err == nil || d.HasChange("decay_rate") {
		rsrc.DecayRate = vfloat
	}
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		logger.Errorf("populateResourceObject() decay_rate failed: %v", err.Error())
		return fmt.Errorf("Resource Object could not be populated: %v", err.Error())
	}

	if _, ok := d.GetOk("resource_instance"); ok {
		populateResourceInstancesObject(ctx, meta, d, rsrc)
	} else if d.HasChange("resource_instance") {
		rsrc.ResourceInstances = make([]*gtm.ResourceInstance, 0)
	}

	return nil
}

// Populate Terraform state from provided Resource object
func populateTerraformResourceState(d *schema.ResourceData, rsrc *gtm.Resource, m interface{}) {
	meta := meta.Must(m)
	logger := meta.Log("Akamai GTM", "populateTerraformResourceState")

	logger.Debugf("Entering populateTerraformResourceState")
	// walk through all state elements
	for stateKey, stateValue := range map[string]interface{}{
		"name":                           rsrc.Name,
		"type":                           rsrc.Type,
		"host_header":                    rsrc.HostHeader,
		"least_squares_decay":            rsrc.LeastSquaresDecay,
		"description":                    rsrc.Description,
		"leader_string":                  rsrc.LeaderString,
		"constrained_property":           rsrc.ConstrainedProperty,
		"aggregation_type":               rsrc.AggregationType,
		"load_imbalance_percentage":      rsrc.LoadImbalancePercentage,
		"upper_bound":                    rsrc.UpperBound,
		"max_u_multiplicative_increment": rsrc.MaxUMultiplicativeIncrement,
		"decay_rate":                     rsrc.DecayRate,
	} {
		err := d.Set(stateKey, stateValue)
		if err != nil {
			logger.Errorf("populateTerraformResourceState failed: %s", err.Error())
		}
	}
	populateTerraformResourceInstancesState(d, rsrc, m)
}

// create and populate GTM Resource ResourceInstances object
func populateResourceInstancesObject(ctx context.Context, meta meta.Meta, d *schema.ResourceData, rsrc *gtm.Resource) {
	logger := meta.Log("Akamai GTM", "populateResourceInstancesObject")

	// pull apart List
	rsrcInstances, err := tf.GetListValue("resource_instance", d)
	if err == nil {
		rsrcInstanceObjList := make([]*gtm.ResourceInstance, len(rsrcInstances)) // create new object list
		for i, v := range rsrcInstances {
			riMap := v.(map[string]interface{})
			rsrcInstance := Client(meta).NewResourceInstance(ctx, rsrc, riMap["datacenter_id"].(int)) // create new object
			rsrcInstance.UseDefaultLoadObject = riMap["use_default_load_object"].(bool)
			if riMap["load_servers"] != nil {
				loadServers, ok := riMap["load_servers"].(*schema.Set)
				if !ok {
					logger.Warnf("wrong type conversion: expected *schema.Set, got %T", loadServers)
				}
				ls := make([]string, loadServers.Len())
				for i, sl := range loadServers.List() {
					ls[i] = sl.(string)
				}
				rsrcInstance.LoadServers = ls
			}
			rsrcInstance.LoadObject.LoadObject = riMap["load_object"].(string)
			rsrcInstance.LoadObjectPort = riMap["load_object_port"].(int)
			rsrcInstanceObjList[i] = rsrcInstance
		}
		rsrc.ResourceInstances = rsrcInstanceObjList
	}
}

// create and populate Terraform resource_instances schema
func populateTerraformResourceInstancesState(d *schema.ResourceData, rsrc *gtm.Resource, m interface{}) {
	meta := meta.Must(m)
	logger := meta.Log("Akamai GTM", "populateTerraformResourceInstancesState")

	riObjectInventory := make(map[int]*gtm.ResourceInstance, len(rsrc.ResourceInstances))
	if len(rsrc.ResourceInstances) > 0 {
		for _, riObj := range rsrc.ResourceInstances {
			riObjectInventory[riObj.DatacenterID] = riObj
		}
	}
	riStateList, _ := tf.GetInterfaceArrayValue("resource_instance", d)
	for _, riMap := range riStateList {
		ri := riMap.(map[string]interface{})
		objIndex := ri["datacenter_id"].(int)
		riObject := riObjectInventory[objIndex]
		if riObject == nil {
			logger.Warnf("Resource_instance %d NOT FOUND in returned GTM Object", ri["datacenter_id"])
			continue
		}
		ri["use_default_load_object"] = riObject.UseDefaultLoadObject
		ri["load_object"] = riObject.LoadObject.LoadObject
		ri["load_object_port"] = riObject.LoadObjectPort
		if ri["load_servers"] != nil {
			loadServers, ok := ri["load_servers"].(*schema.Set)
			if !ok {
				logger.Warnf("wrong type conversion: expected *schema.Set, got %T", loadServers)
			}
			ri["load_servers"] = reconcileTerraformLists(loadServers.List(), convertStringToInterfaceList(riObject.LoadServers, m), m)
		} else {
			ri["load_servers"] = riObject.LoadServers
		}
		// remove object
		delete(riObjectInventory, objIndex)
	}
	if len(riObjectInventory) > 0 {
		logger.Debugf("Resource_instance objects left...")
		// Objects not in the state yet. Add. Unfortunately, they'll likely not align with instance indices in the config
		for _, mriObj := range riObjectInventory {
			riNew := make(map[string]interface{})
			riNew["datacenter_id"] = mriObj.DatacenterID
			riNew["use_default_load_object"] = mriObj.UseDefaultLoadObject
			riNew["load_object"] = mriObj.LoadObject.LoadObject
			riNew["load_object_port"] = mriObj.LoadObjectPort
			riNew["load_servers"] = mriObj.LoadServers
			riStateList = append(riStateList, riNew)
		}
	}
	_ = d.Set("resource_instance", riStateList)

}

func resourceInstanceDiffSuppress(_, _, _ string, d *schema.ResourceData) bool {
	logger := logger.Get("Akamai GTM", "resourceInstanceDiffSuppress")
	oldRes, newRes := d.GetChange("resource_instance")

	oldList, ok := oldRes.([]interface{})
	if !ok {
		logger.Warnf("wrong type conversion: expected []interface{}, got %T", oldList)
	}
	newList, ok := newRes.([]interface{})
	if !ok {
		logger.Warnf("wrong type conversion: expected []interface{}, got %T", newList)
	}

	if len(oldList) != len(newList) {
		return false
	}

	sort.Slice(oldList, func(i, j int) bool {
		return oldList[i].(map[string]interface{})["datacenter_id"].(int) < oldList[j].(map[string]interface{})["datacenter_id"].(int)
	})
	sort.Slice(newList, func(i, j int) bool {
		return newList[i].(map[string]interface{})["datacenter_id"].(int) < newList[j].(map[string]interface{})["datacenter_id"].(int)
	})

	length := len(oldList)
	var oldServers, newServers map[string]interface{}
	for i := 0; i < length; i++ {
		oldServers, ok = oldList[i].(map[string]interface{})
		if !ok {
			logger.Warnf("wrong type conversion: expected map[string]interface{}, got %T", oldServers)
		}
		newServers, ok = newList[i].(map[string]interface{})
		if !ok {
			logger.Warnf("wrong type conversion: expected map[string]interface{}, got %T", newServers)
		}
		for k, v := range oldServers {
			if k == "load_servers" {
				if !loadServersEqual(oldServers["load_servers"], newServers["load_servers"]) {
					return false
				}
			} else {
				if newServers[k] != v {
					return false
				}
			}
		}
	}

	return true
}

// loadServersEqual checks whether load_servers are equal
func loadServersEqual(oldVal, newVal interface{}) bool {
	logger := logger.Get("Akamai GTM", "loadServersEqual")

	oldServers, ok := oldVal.(*schema.Set)
	if !ok {
		logger.Warnf("wrong type conversion: expected *schema.Set, got %T", oldServers)
		return false
	}

	newServers, ok := newVal.(*schema.Set)
	if !ok {
		logger.Warnf("wrong type conversion: expected *schema.Set, got %T", newServers)
		return false
	}

	if oldServers.Len() != newServers.Len() {
		return false
	}

	servers := make(map[string]bool, oldServers.Len())
	for _, server := range oldServers.List() {
		servers[server.(string)] = true
	}

	for _, server := range newServers.List() {
		_, ok = servers[server.(string)]
		if !ok {
			return false
		}
	}

	return true
}
