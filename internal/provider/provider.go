package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*B1fr0stProvider)(nil)

type B1fr0stProvider struct{ version string }

type ProviderModel struct {
	APIURL   types.String `tfsdk:"api_url"`
	APIToken types.String `tfsdk:"api_token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider { return &B1fr0stProvider{version: version} }
}

func (p *B1fr0stProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "b1fr0st"
	resp.Version = p.version
}

func (p *B1fr0stProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL for the b1fr0st.dev API. Defaults to the hosted service.",
			},
			"api_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "API token. Set directly or use B1FR0ST_API_TOKEN in the environment.",
			},
		},
	}
}

func (p *B1fr0stProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := strings.TrimSpace(config.APIToken.ValueString())
	if token == "" {
		token = strings.TrimSpace(os.Getenv("B1FR0ST_API_TOKEN"))
	}
	if token == "" {
		resp.Diagnostics.AddError("Missing API token", "Set api_token in the provider configuration or set B1FR0ST_API_TOKEN in the environment.")
		return
	}
	apiURL := strings.TrimRight(config.APIURL.ValueString(), "/")
	if apiURL == "" {
		apiURL = "https://cloud.b1fr0st.dev/api/v1"
	}
	client, err := NewAPIClient(apiURL, token)
	if err != nil {
		resp.Diagnostics.AddError("Invalid API URL", err.Error())
		return
	}
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *B1fr0stProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewInstanceResource,
		NewIngressResource,
	}
}

func (p *B1fr0stProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
