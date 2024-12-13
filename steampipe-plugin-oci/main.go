package main

import (
	"github.com/opengovern/og-describer-oci/steampipe-plugin-oci/oci"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{PluginFunc: oci.Plugin})
}
