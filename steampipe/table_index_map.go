package steampipe

import (
	"github.com/opengovern/og-describer-oci/pkg/sdk/es"
)

var Map = map[string]string{
  "OCI::Artifact": "oci_artifact",
}

var DescriptionMap = map[string]interface{}{
  "OCI::Artifact": opengovernance.OCIArtifact{},
}

var ReverseMap = map[string]string{
  "oci_artifact": "OCI::Artifact",
}
