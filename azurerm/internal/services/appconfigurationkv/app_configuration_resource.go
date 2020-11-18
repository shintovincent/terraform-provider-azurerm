package appconfigurationkv

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/appconfiguration/mgmt/2019-10-01/appconfiguration"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	// "github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/appconfiguration/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/appconfiguration/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	azSchema "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/schema"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmAppConfigurationKv() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmAppConfigurationKvCreate,
		Read:   resourceArmAppConfigurationKvRead,
		Update: resourceArmAppConfigurationKvCreate,
		Delete: resourceArmAppConfigurationKvDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Importer: azSchema.ValidateResourceIDPriorToImport(func(id string) error {
			return nil
		}),

		Schema: map[string]*schema.Schema{
			"app_configuration_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
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
				ForceNew: true,
			},

			// the API changed and now returns the rg in lowercase
			// revert when https://github.com/Azure/azure-sdk-for-go/issues/6606 is fixed
			"resource_group_name": azure.SchemaResourceGroupNameDiffSuppress(),
		},
	}
}

func resourceArmAppConfigurationKvCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppConfiguration.AppConfigurationsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure ARM App Configuration creation.")

	name := d.Get("app_configuration_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("Error checking for presence of existing App Configuration %q (Resource Group %q): %s", name, resourceGroup, err)
		}
	}

	if existing.ID == nil && *existing.ID == "" {
		return fmt.Errorf("App Configuration not found")
	}

	resultPage, err := client.ListKeys(ctx, resourceGroup, name, "")
	if err != nil {
		return fmt.Errorf("Failed to receive access keys for App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	accessKeys := flattenAppConfigurationAccessKeys(resultPage.Values())
	flattenedSecret := accessKeys.primaryWriteKey[0].(map[string]interface{})

	parameters := appconfiguration.SetKeyValueParameters{
		Key:        utils.String(d.Get("key").(string)),
		Value:      utils.String(d.Get("value").(string)),
		Label:      utils.String(d.Get("label").(string)),
		Secret:     utils.String(flattenedSecret["secret"].(string)),
		Credential: utils.String(flattenedSecret["id"].(string)),
	}

	_, err = client.SetKeyValue(ctx, resourceGroup, name, parameters)
	if err != nil {
		return fmt.Errorf("Error creating App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	_, err = client.GetKeyValue(ctx, resourceGroup, name, parameters)
	if err != nil {
		return fmt.Errorf("Error retrieving App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	read, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		return fmt.Errorf("Error retrieving App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read App Configuration %s (resource Group %q) ID", name, resourceGroup)
	}

	log.Printf("-------------------------------------------------------------")
	log.Printf(*read.ID)
	d.SetId(d.Get("key").(string))

	return resourceArmAppConfigurationKvRead(d, meta)
}

func resourceArmAppConfigurationKvUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppConfiguration.AppConfigurationsClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure ARM App Configuration update.")
	id, err := parse.AppConfigurationID(d.Id())
	if err != nil {
		return err
	}

	parameters := appconfiguration.ConfigurationStoreUpdateParameters{
		Sku: &appconfiguration.Sku{
			Name: utils.String(d.Get("sku").(string)),
		},
		Tags: tags.Expand(d.Get("tags").(map[string]interface{})),
	}

	if d.HasChange("identity") {
		parameters.Identity = expandAppConfigurationIdentity(d.Get("identity").([]interface{}))
	}

	future, err := client.Update(ctx, id.ResourceGroup, id.Name, parameters)
	if err != nil {
		return fmt.Errorf("Error updating App Configuration %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("Error waiting for update of App Configuration %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}

	read, err := client.Get(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("Error retrieving App Configuration %q (Resource Group %q): %+v", id.Name, id.ResourceGroup, err)
	}
	if read.ID == nil {
		return fmt.Errorf("Cannot read App Configuration %s (resource Group %q) ID", id.Name, id.ResourceGroup)
	}

	d.SetId(*read.ID)

	return resourceArmAppConfigurationKvRead(d, meta)
}

func resourceArmAppConfigurationKvRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppConfiguration.AppConfigurationsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("app_configuration_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	resp, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] App Configuration %q was not found in Resource Group %q - removing from state!", name, resourceGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error making Read request on App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	resultPage, err := client.ListKeys(ctx, resourceGroup, name, "")
	if err != nil {
		return fmt.Errorf("Failed to receive access keys for App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	accessKeys := flattenAppConfigurationAccessKeys(resultPage.Values())
	flattenedSecret := accessKeys.primaryWriteKey[0].(map[string]interface{})

	parameters := appconfiguration.SetKeyValueParameters{
		Key:        utils.String(d.Get("key").(string)),
		Value:      utils.String(d.Get("value").(string)),
		Label:      utils.String(d.Get("label").(string)),
		Secret:     utils.String(flattenedSecret["secret"].(string)),
		Credential: utils.String(flattenedSecret["id"].(string)),
	}

	d.Set("app_configuration_name", resp.Name)
	d.Set("resource_group_name", resourceGroup)

	kvRead, err := client.GetKeyValue(ctx, resourceGroup, name, parameters)

	fmt.Println(kvRead.Key)

	return err
}

func resourceArmAppConfigurationKvDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppConfiguration.AppConfigurationsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for Azure ARM App Configuration deletion.")

	name := d.Get("app_configuration_name").(string)
	resourceGroup := d.Get("resource_group_name").(string)

	existing, err := client.Get(ctx, resourceGroup, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("Error checking for presence of existing App Configuration %q (Resource Group %q): %s", name, resourceGroup, err)
		}
	}

	if existing.ID == nil && *existing.ID == "" {
		return fmt.Errorf("App Configuration not found")
	}

	resultPage, err := client.ListKeys(ctx, resourceGroup, name, "")
	if err != nil {
		return fmt.Errorf("Failed to receive access keys for App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	accessKeys := flattenAppConfigurationAccessKeys(resultPage.Values())
	flattenedSecret := accessKeys.primaryWriteKey[0].(map[string]interface{})

	parameters := appconfiguration.SetKeyValueParameters{
		Key:        utils.String(d.Get("key").(string)),
		Label:      utils.String(d.Get("label").(string)),
		Secret:     utils.String(flattenedSecret["secret"].(string)),
		Credential: utils.String(flattenedSecret["id"].(string)),
	}

	_, err = client.DeleteKeyValue(ctx, resourceGroup, name, parameters)
	if err != nil {
		return fmt.Errorf("Error deleting App Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	return nil
}

type flattenedAccessKeys struct {
	primaryReadKey    []interface{}
	primaryWriteKey   []interface{}
	secondaryReadKey  []interface{}
	secondaryWriteKey []interface{}
}

func flattenAppConfigurationAccessKeys(values []appconfiguration.APIKey) flattenedAccessKeys {
	result := flattenedAccessKeys{
		primaryReadKey:    make([]interface{}, 0),
		primaryWriteKey:   make([]interface{}, 0),
		secondaryReadKey:  make([]interface{}, 0),
		secondaryWriteKey: make([]interface{}, 0),
	}

	for _, value := range values {
		if value.Name == nil || value.ReadOnly == nil {
			continue
		}

		accessKey := flattenAppConfigurationAccessKey(value)
		name := *value.Name
		readOnly := *value.ReadOnly

		if strings.HasPrefix(strings.ToLower(name), "primary") {
			if readOnly {
				result.primaryReadKey = accessKey
			} else {
				result.primaryWriteKey = accessKey
			}
		}

		if strings.HasPrefix(strings.ToLower(name), "secondary") {
			if readOnly {
				result.secondaryReadKey = accessKey
			} else {
				result.secondaryWriteKey = accessKey
			}
		}
	}

	return result
}

func flattenAppConfigurationAccessKey(input appconfiguration.APIKey) []interface{} {
	connectionString := ""

	if input.ConnectionString != nil {
		connectionString = *input.ConnectionString
	}

	id := ""
	if input.ID != nil {
		id = *input.ID
	}

	secret := ""
	if input.Value != nil {
		secret = *input.Value
	}

	return []interface{}{
		map[string]interface{}{
			"connection_string": connectionString,
			"id":                id,
			"secret":            secret,
		},
	}
}

func expandAppConfigurationIdentity(identities []interface{}) *appconfiguration.ResourceIdentity {
	if len(identities) == 0 {
		return &appconfiguration.ResourceIdentity{
			Type: appconfiguration.None,
		}
	}
	identity := identities[0].(map[string]interface{})
	identityType := appconfiguration.IdentityType(identity["type"].(string))
	return &appconfiguration.ResourceIdentity{
		Type: identityType,
	}
}

func flattenAppConfigurationIdentity(identity *appconfiguration.ResourceIdentity) []interface{} {
	if identity == nil || identity.Type == appconfiguration.None {
		return []interface{}{}
	}

	principalId := ""
	if identity.PrincipalID != nil {
		principalId = *identity.PrincipalID
	}

	tenantId := ""
	if identity.TenantID != nil {
		tenantId = *identity.TenantID
	}

	return []interface{}{
		map[string]interface{}{
			"type":         string(identity.Type),
			"principal_id": principalId,
			"tenant_id":    tenantId,
		},
	}
}
