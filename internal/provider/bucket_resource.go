package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = (*BucketResource)(nil)
var _ resource.ResourceWithConfigure = (*BucketResource)(nil)
var _ resource.ResourceWithImportState = (*BucketResource)(nil)

type BucketResource struct{ client *APIClient }
type BucketModel struct {
	ID              types.String `tfsdk:"id"`
	ProjectID       types.String `tfsdk:"project_id"`
	Name            types.String `tfsdk:"name"`
	QuotaGiB        types.Int64  `tfsdk:"quota_gib"`
	DomainID        types.String `tfsdk:"domain_id"`
	Subdomain       types.String `tfsdk:"subdomain"`
	SiteID          types.Int64  `tfsdk:"site_id"`
	Hostname        types.String `tfsdk:"hostname"`
	S3Endpoint      types.String `tfsdk:"s3_endpoint"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	UsedBytes       types.Int64  `tfsdk:"used_bytes"`
	Objects         types.Int64  `tfsdk:"objects"`
}

func NewBucketResource() resource.Resource { return &BucketResource{} }
func (r *BucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}
func (r *BucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{Attributes: map[string]schema.Attribute{
		"id":                schema.StringAttribute{Computed: true, Description: "Platform bucket identifier."},
		"project_id":        schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Project UUID that owns the bucket."},
		"name":              schema.StringAttribute{Required: true, PlanModifiers: replace, Description: "Globally unique S3 bucket name."},
		"quota_gib":         schema.Int64Attribute{Required: true, Description: "Maximum bucket size in GiB. Quotas share the project's storage budget."},
		"domain_id":         schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: replace, Description: "Optional Pangolin domain assigned to the tenant."},
		"subdomain":         schema.StringAttribute{Optional: true, Computed: true, PlanModifiers: replace, Description: "Optional public hostname label. Pangolin provisions TLS after DNS is configured."},
		"site_id":           schema.Int64Attribute{Optional: true, Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}, Description: "Online Newt site used for a custom domain route."},
		"hostname":          schema.StringAttribute{Computed: true, Description: "Public bucket endpoint hostname when a domain is configured."},
		"s3_endpoint":       schema.StringAttribute{Computed: true, Description: "S3-compatible endpoint."},
		"access_key_id":     schema.StringAttribute{Computed: true, Description: "Dedicated Garage access key identifier."},
		"secret_access_key": schema.StringAttribute{Computed: true, Sensitive: true, Description: "Generated once on creation. Store Terraform state securely."},
		"used_bytes":        schema.Int64Attribute{Computed: true, Description: "Current Garage object data size in bytes."},
		"objects":           schema.Int64Attribute{Computed: true, Description: "Current number of stored objects."},
	}}
}
func (r *BucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *BucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
func (r *BucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input := map[string]any{"projectId": plan.ProjectID.ValueString(), "name": plan.Name.ValueString(), "quotaGiB": plan.QuotaGiB.ValueInt64()}
	if !plan.DomainID.IsNull() && plan.DomainID.ValueString() != "" {
		input["domainId"] = plan.DomainID.ValueString()
		input["subdomain"] = plan.Subdomain.ValueString()
		input["siteId"] = plan.SiteID.ValueInt64()
	}
	var out struct {
		Bucket map[string]any `json:"bucket"`
	}
	_, err := r.client.Request(ctx, "POST", "/buckets", input, &out)
	if err != nil {
		resp.Diagnostics.AddError("Create bucket failed", err.Error())
		return
	}
	bucketModelFromAPI(&plan, out.Bucket, true)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Bucket map[string]any `json:"bucket"`
	}
	status, err := r.client.Request(ctx, "GET", "/buckets/"+pathID(state.ID.ValueString()), nil, &out)
	if status == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Read bucket failed", err.Error())
		return
	}
	bucketModelFromAPI(&state, out.Bucket, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *BucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BucketModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out map[string]any
	_, err := r.client.Request(ctx, "PUT", "/buckets/"+pathID(plan.ID.ValueString()), map[string]any{"quotaGiB": plan.QuotaGiB.ValueInt64()}, &out)
	if err != nil {
		resp.Diagnostics.AddError("Update bucket quota failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BucketModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Request(ctx, "DELETE", "/buckets/"+pathID(state.ID.ValueString()), nil, nil)
	if err != nil && status != 404 {
		resp.Diagnostics.AddError("Delete bucket failed", err.Error())
	}
}
func bucketModelFromAPI(model *BucketModel, raw map[string]any, created bool) {
	if v := anyString(raw["id"]); v != "" {
		model.ID = types.StringValue(v)
	}
	if v := anyString(raw["projectId"]); v != "" {
		model.ProjectID = types.StringValue(v)
	}
	if v := anyString(raw["name"]); v != "" {
		model.Name = types.StringValue(v)
	}
	if v, ok := raw["quotaGiB"]; ok {
		model.QuotaGiB = types.Int64Value(int64Value(v))
	}
	model.DomainID = optionalString(raw["domainId"])
	model.Subdomain = optionalString(raw["subdomain"])
	model.Hostname = types.StringValue(anyString(raw["hostname"]))
	model.S3Endpoint = types.StringValue(anyString(raw["s3Endpoint"]))
	model.AccessKeyID = types.StringValue(anyString(raw["accessKeyId"]))
	if created {
		model.SecretAccessKey = types.StringValue(anyString(raw["secretAccessKey"]))
	}
	model.UsedBytes = types.Int64Value(int64Value(raw["usedBytes"]))
	model.Objects = types.Int64Value(int64Value(raw["objects"]))
	if v, ok := raw["siteId"]; ok {
		model.SiteID = types.Int64Value(int64Value(v))
	}
}
func optionalString(value any) types.String {
	if value == nil || anyString(value) == "" {
		return types.StringNull()
	}
	return types.StringValue(anyString(value))
}
