package botman

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v7/pkg/botman"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/hash"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/common/tf"
	"github.com/akamai/terraform-provider-akamai/v5/pkg/meta"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAkamaiDefinedBot() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAkamaiDefinedBotRead,
		Schema: map[string]*schema.Schema{
			"bot_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"json": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAkamaiDefinedBotRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	meta := meta.Must(m)
	logger := meta.Log("botman", "dataSourceAkamaiDefinedBotRead")

	botName, err := tf.GetStringValue("bot_name", d)
	if err != nil && !errors.Is(err, tf.ErrNotFound) {
		return diag.FromErr(err)
	}

	request := botman.GetAkamaiDefinedBotListRequest{
		BotName: botName,
	}

	response, err := getAkamaiDefinedBotList(ctx, request, m)
	if err != nil {
		logger.Errorf("calling 'GetAkamaiDefinedBotList': %s", err.Error())
		return diag.FromErr(err)
	}

	jsonBody, err := json.Marshal(response)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("json", string(jsonBody)); err != nil {
		return diag.Errorf("%s: %s", tf.ErrValueSet, err.Error())
	}

	d.SetId(hash.GetSHAString(string(jsonBody)))
	return nil
}
