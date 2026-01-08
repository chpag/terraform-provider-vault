package actions

import (
	"context"

	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-vault/internal/framework/base"
	"github.com/hashicorp/terraform-provider-vault/internal/framework/client"
	"github.com/hashicorp/terraform-provider-vault/internal/framework/errutil"
	"github.com/hashicorp/vault/api"
)



var (
	_ action.Action = (*pluginReloadAction)(nil)
)

func NewPluginReloadAction() action.Action {
	return &pluginReloadAction{}
}

type pluginReloadAction struct{
	base.ActionWithConfigure
}

type pluginReloadModel struct {
	base.BaseModelLegacy
	
	Type  types.String  `tfsdk:"type"`
	Name  types.String `tfsdk:"name"`
	Scope types.String    `tfsdk:"scope"`
}

func (a *pluginReloadAction) Metadata(ctx context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pluginreload"
}

func (a *pluginReloadAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reload Vault PLugins using: https://developer.hashicorp.com/vault/api-docs/system/plugins-reload",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "The type of the plugin, as registered in the plugin catalog. One of 'auth', 'secret', 'database', or 'unknown'. If 'unknown', all plugin types with the provided name will be reloaded.",
				Required: true,
//				Validators: []validator.String{
//					ValidPluginReloadType(),
//				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the plugin to reload, as registered in the plugin catalog.",
				Required:    true,
			},
			"scope": schema.StringAttribute{
				Description: "The scope of the reload. If omitted, reloads the plugin or mounts on this Vault instance. If 'global', will begin reloading the plugin on all instances of a cluster.",
				Optional:    true,
			},
		},
	}
}

func (a *pluginReloadAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {

	var config pluginReloadModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}


	vaultClient, err := client.GetClient(ctx, a.Meta(), config.Namespace.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(errutil.ClientConfigureErr(err))
		return
	}

	var pluginType api.PluginType
	if config.Type.String() == "database" {
		pluginType = api.PluginTypeDatabase
	} else if config.Type.String() == "auth" {
		pluginType = api.PluginTypeCredential
	} else if config.Type.String() == "secret" {
		pluginType = api.PluginTypeSecrets
	} else if config.Type.String() == "unknown" {
		pluginType = api.PluginTypeUnknown
	} else {
		resp.Diagnostics.AddError("Error Unknwon PluginType", err.Error())
		return
	}


	if ! config.Namespace.IsNull() {
		resp.Diagnostics.AddError("PluginReload action can be only done on Root Namespace", err.Error())
		return
	}
	var pluginInput api.RootReloadPluginInput
	pluginInput.Plugin = config.Name.String()
	pluginInput.Type = pluginType
	pluginInput.Scope = config.Scope.String()

	reload_id, err := vaultClient.Sys().RootReloadPlugin(ctx,&pluginInput)
	if err != nil {
		resp.Diagnostics.AddError(errutil.VaultCreateErr(err))
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("\n\nPlugin Reload with id:%s", reload_id),
	})
}

func ValidPluginReloadType() {
	panic("unimplemented")
	return
}

