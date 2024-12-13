package steampipe

import (
	"github.com/opengovern/og-describer-oci/pkg/sdk/es"
)

var Map = map[string]string{
  "OCI::Image": "oci_image",
  "OCI::ImageTag": "oci_image_tag",
}

var DescriptionMap = map[string]interface{}{
  "OCI::Image": opengovernance.OCIImage{},
  "OCI::ImageTag": opengovernance.OCIImageTag{},
}

var ReverseMap = map[string]string{
  "oci_image": "OCI::Image",
  "oci_image_tag": "OCI::ImageTag",
}
