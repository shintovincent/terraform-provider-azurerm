package appconfigurationkv

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/appconfiguration/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func dataSourceAppConfigurationKv() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAppConfigurationRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validate.AppConfigurationName,
			},

			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"value": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"label": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: false,
			},

			"resource_group_name": azure.SchemaResourceGroupNameForDataSource(),
		},
	}
}

func dataSourceAppConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppConfiguration.AppConfigurationsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return fmt.Errorf("App Configuration %q was not found in Resource Group %q", name, resourceGroup)
		}

		return fmt.Errorf("Error retrieving App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	resultPage, err := client.ListKeys(ctx, resourceGroup, name, "")
	if err != nil {
		return fmt.Errorf("Failed to receive access keys for App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	d.SetId(*resp.ID)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	skuName := ""
	if resp.Sku != nil && resp.Sku.Name != nil {
		skuName = *resp.Sku.Name
	}
	d.Set("sku", skuName)

	if props := resp.ConfigurationStoreProperties; props != nil {
		d.Set("endpoint", props.Endpoint)
	}

	accessKeys := flattenAppConfigurationAccessKeys(resultPage.Values())
	d.Set("primary_read_key", accessKeys.primaryReadKey)
	d.Set("primary_write_key", accessKeys.primaryWriteKey)
	d.Set("secondary_read_key", accessKeys.secondaryReadKey)
	d.Set("secondary_write_key", accessKeys.secondaryWriteKey)

	return tags.FlattenAndSet(d, resp.Tags)
}
