package apimanagement

import "github.com/Azure/azure-sdk-for-go/services/apimanagement/mgmt/2019-12-01/apimanagement"

const (
	// Terraform specifies that the user create request was sent by Terraform.
	Terraform apimanagement.AppType = "terraform"
)
