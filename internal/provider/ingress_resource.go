package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*IngressResource)(nil)
var _ resource.ResourceWithConfigure = (*IngressResource)(nil)

type IngressResource struct{ client *APIClient }
type IngressModel struct {
	ID           types.String `tfsdk:"id"`
	TenantID     types.String `tfsdk:"tenant_id"`
	Name         types.String `tfsdk:"name"`
	DomainID     types.String `tfsdk:"domain_id"`
	Subdomain    types.String `tfsdk:"subdomain"`
	SiteID       types.Int64  `tfsdk:"site_id"`
	InstanceName types.String `tfsdk:"instance_name"`
	Port         types.Int64  `tfsdk:"port"`
	Hostname     types.String `tfsdk:"hostname"`
	Certificate  types.String `tfsdk:"certificate"`
	Target       types.String `tfsdk:"target"`
}

func NewIngressResource() resource.Resource { return &IngressResource{} }
func (r *IngressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ingress"
}
func (r *IngressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":            schema.StringAttribute{Computed: true, Description: "Managed Pangolin resource identifier."},
		"tenant_id":     schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Tenant UUID that owns this ingress."},
		"name":          schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Display name for the Pangolin resource."},
		"domain_id":     schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Pangolin domain identifier assigned to the tenant."},
		"subdomain":     schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "DNS label to create under the selected domain."},
		"site_id":       schema.Int64Attribute{Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}, Description: "Online Newt site identifier."},
		"instance_name": schema.StringAttribute{Optional: true, PlanModifiers: replace, Description: "Running tenant instance. Omit only for the administrator system ingress."},
		"port":          schema.Int64Attribute{Required: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}, Description: "Service port on the instance; use 8080 for the system console ingress."},
		"hostname":      schema.StringAttribute{Computed: true, Description: "Public hostname provisioned in Pangolin."},
		"certificate":   schema.StringAttribute{Computed: true, Description: "Certificate status reported by the platform."},
		"target":        schema.StringAttribute{Computed: true, Description: "Resolved upstream target."},
	}}
}
func (r *IngressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}
func (r *IngressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var p IngressModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := map[string]any{"tenantId": p.TenantID.ValueString(), "name": p.Name.ValueString(), "domainId": p.DomainID.ValueString(), "subdomain": p.Subdomain.ValueString(), "siteId": p.SiteID.ValueInt64(), "port": p.Port.ValueInt64()}
	if !p.InstanceName.IsNull() && !p.InstanceName.IsUnknown() {
		input["instanceName"] = p.InstanceName.ValueString()
	}
	var out struct {
		ResourceID  string `json:"resourceId"`
		Hostname    string `json:"hostname"`
		Certificate string `json:"certificate"`
		Target      string `json:"target"`
	}
	_, err := r.client.Request(ctx, "POST", "/pangolin/resources", input, &out)
	if err != nil {
		resp.Diagnostics.AddError("Create ingress failed", err.Error())
		return
	}
	p.ID = types.StringValue(out.ResourceID)
	p.Hostname = types.StringValue(out.Hostname)
	p.Certificate = types.StringValue(out.Certificate)
	p.Target = types.StringValue(out.Target)
	resp.Diagnostics.Append(resp.State.Set(ctx, &p)...)
}
func (r *IngressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var s IngressModel
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out map[string]any
	status, err := r.client.Request(ctx, "GET", "/pangolin", nil, &out)
	if status == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read ingress failed", err.Error())
		return
	}
	found := findResourceID(out, s.ID.ValueString())
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &s)...)
}
func findResourceID(v any, id string) bool {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if (k == "resourceId" || k == "id") && fmt.Sprint(val) == id {
				return true
			}
			if findResourceID(val, id) {
				return true
			}
		}
	case []any:
		for _, val := range x {
			if findResourceID(val, id) {
				return true
			}
		}
	}
	return false
}
func (r *IngressResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {}
func (r *IngressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var s IngressModel
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Request(ctx, "DELETE", "/pangolin/resources/"+pathID(s.ID.ValueString()), nil, nil)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Delete ingress failed", err.Error())
	}
}
