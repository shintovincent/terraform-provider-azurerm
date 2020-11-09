package example

import (
	"testing"
)

func TestAccAzureRMExample_basic(t *testing.T) {
	data := ResourceMetaData{}
	data.Decode(map[string]interface{}{
		"name":"tom",
		"list":[]interface{}{map[string]interface{}{"name":"test"}},
	})
}
