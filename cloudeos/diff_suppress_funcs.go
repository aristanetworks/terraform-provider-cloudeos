package cloudeos

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func suppressAttributeChange(attribute, old, new string, d *schema.ResourceData) bool {
	if old != "" && old != new {
		log.Fatalf("Attribute change not supported for %s, old value: %s, new value: %s",
			attribute, old, new)
		return true
	}
	return false
}
