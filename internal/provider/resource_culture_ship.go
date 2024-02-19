// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/matthewbaggett/terraform-provider-fun-names/internal/spaceships"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	mapplanmodifiers "github.com/matthewbaggett/terraform-provider-fun-names/internal/planmodifiers/map"
)

var _ resource.Resource = (*cultureShipResource)(nil)

func NewCultureShipResource() resource.Resource {
	return &cultureShipResource{}
}

type cultureShipResource struct{}

func (r *cultureShipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_culture_ship"
}

func (r *cultureShipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The resource `random_culture_ship` returns a name of a ship from the Culture Series by Ian M Banks\n" +
			"\n" +
			"It is much like the `random_pet` resource, but with a different name and a different set of default values.\n",
		Attributes: map[string]schema.Attribute{
			"keepers": schema.MapAttribute{
				Description: "Arbitrary map of values that, when changed, will trigger recreation of " +
					"resource. See [the main provider documentation](../index.html) for more information.",
				ElementType: types.StringType,
				Optional:    true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifiers.RequiresReplaceIfValuesNotNull(),
				},
			},
			"prefix": schema.StringAttribute{
				Description: "A string to prefix the name with.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"separator": schema.StringAttribute{
				Description: "The character to separate words in the ship name. Defaults to \"-\"",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("-"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The random ship name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *cultureShipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// This is necessary to ensure each call to petname is properly randomised:
	// the library uses `rand.Intn()` and does NOT seed `rand.Seed()` by default,
	// so this call takes care of that.
	spaceships.NonDeterministicMode()

	var plan cultureShipModelV0

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	separator := plan.Separator.ValueString()
	prefix := plan.Prefix.ValueString()

	ship := strings.ToLower(spaceships.Generate(separator))

	pn := cultureShipModelV0{
		Keepers:   plan.Keepers,
		Separator: types.StringValue(separator),
	}

	if prefix != "" {
		ship = fmt.Sprintf("%s%s%s", prefix, separator, ship)
		pn.Prefix = types.StringValue(prefix)
	} else {
		pn.Prefix = types.StringNull()
	}

	pn.ID = types.StringValue(ship)

	diags = resp.State.Set(ctx, pn)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *cultureShipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update ensures the plan value is copied to the state to complete the update.
func (r *cultureShipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model cultureShipModelV0

	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *cultureShipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

type cultureShipModelV0 struct {
	ID        types.String `tfsdk:"id"`
	Keepers   types.Map    `tfsdk:"keepers"`
	Prefix    types.String `tfsdk:"prefix"`
	Separator types.String `tfsdk:"separator"`
}
