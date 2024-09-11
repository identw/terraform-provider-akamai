package iam

import (
	"context"
	"fmt"
	"time"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v8/pkg/iam"
	"github.com/akamai/terraform-provider-akamai/v6/pkg/meta"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &cidrBlocksDataSource{}
	_ datasource.DataSourceWithConfigure = &cidrBlocksDataSource{}
)

type (
	cidrBlocksDataSource struct {
		meta meta.Meta
	}

	cidrBlocksSourceModel struct {
		CIDRBlocks []cidrBlockSourceModel `tfsdk:"cidr_blocks"`
	}
)

// NewCIDRBlocksDataSource returns a new iam CIDR block data source
func NewCIDRBlocksDataSource() datasource.DataSource {
	return &cidrBlocksDataSource{}
}

// Metadata configures data source's meta information
func (d *cidrBlocksDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "akamai_iam_cidr_blocks"
}

// Configure configures data source at the beginning of the lifecycle
func (d *cidrBlocksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			resp.Diagnostics.AddError(
				"Unexpected Data Source Configure Type",
				fmt.Sprintf("Expected meta.Meta, got: %T. Please report this issue to the provider developers.",
					req.ProviderData))
		}
	}()
	d.meta = meta.Must(req.ProviderData)
}

// Schema is used to define data source's terraform schema
func (d *cidrBlocksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Identity and Access Management CIDR block",
		Attributes: map[string]schema.Attribute{
			"cidr_blocks": schema.ListNestedAttribute{
				Description: "List of CIDR blocks on account's allowlist",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"actions": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Specifies activities available for the CIDR block.",
							Attributes: map[string]schema.Attribute{
								"delete": schema.BoolAttribute{
									Computed:    true,
									Description: "Whether you can delete this CIDR block.",
								},
								"edit": schema.BoolAttribute{
									Computed:    true,
									Description: "Whether you can edit this CIDR block.",
								},
							},
						},
						"cidr_block": schema.StringAttribute{
							Computed:    true,
							Description: "The value of an IP address or IP address range.",
						},
						"cidr_block_id": schema.Int64Attribute{
							Computed:    true,
							Description: "Unique identifier for each CIDR block.",
						},
						"comments": schema.StringAttribute{
							Computed:    true,
							Description: "Descriptive label you provide for the CIDR block.",
						},
						"created_by": schema.StringAttribute{
							Computed:    true,
							Description: "The user who created the CIDR block.",
						},
						"created_date": schema.StringAttribute{
							Computed:    true,
							Description: "ISO 8601 timestamp indicating when the CIDR block was created.",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the CIDR block is enabled.",
						},
						"modified_by": schema.StringAttribute{
							Computed:    true,
							Description: "The user who last edited the CIDR block.",
						},
						"modified_date": schema.StringAttribute{
							Computed:    true,
							Description: "ISO 8601 timestamp indicating when the CIDR block was last modified.",
						},
					},
				},
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state
func (d *cidrBlocksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "IAM CIDR Blocks DataSource Read")

	var data cidrBlocksSourceModel

	if resp.Diagnostics.Append(req.Config.Get(ctx, &data)...); resp.Diagnostics.HasError() {
		return
	}
	client := inst.Client(d.meta)

	cidrBlocks, err := client.ListCIDRBlocks(ctx, iam.ListCIDRBlocksRequest{
		Actions: true,
	})
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("%s:", ErrIAMListCIDRBlocks), err.Error())
		return
	}

	if resp.Diagnostics.Append(data.read(cidrBlocks)...); resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (d *cidrBlocksSourceModel) read(cidrBlocks *iam.ListCIDRBlocksResponse) diag.Diagnostics {
	for _, cidrBlock := range *cidrBlocks {
		block := cidrBlockSourceModel{
			CIDRBlock:    types.StringValue(cidrBlock.CIDRBlock),
			Enabled:      types.BoolValue(cidrBlock.Enabled),
			Comments:     types.StringValue(cidrBlock.Comments),
			CIDRBlockID:  types.Int64Value(cidrBlock.CIDRBlockID),
			CreatedBy:    types.StringValue(cidrBlock.CreatedBy),
			CreatedDate:  types.StringValue(cidrBlock.CreatedDate.Format(time.RFC3339Nano)),
			ModifiedBy:   types.StringValue(cidrBlock.ModifiedBy),
			ModifiedDate: types.StringValue(cidrBlock.ModifiedDate.Format(time.RFC3339Nano)),
		}
		if cidrBlock.Actions != nil {
			block.Actions = &actions{
				Delete: types.BoolValue(cidrBlock.Actions.Delete),
				Edit:   types.BoolValue(cidrBlock.Actions.Edit),
			}
		}

		d.CIDRBlocks = append(d.CIDRBlocks, block)
	}

	return nil
}