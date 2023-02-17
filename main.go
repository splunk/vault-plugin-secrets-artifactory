package main

import (
	"log"
	"os"

	"github.com/hashicorp/vault/api"

	"github.com/hashicorp/vault/sdk/plugin"
	artifactorysecrets "github.com/splunk/vault-plugin-secrets-artifactory/plugin"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	apiClientMeta := &api.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	_ = flags.Parse(os.Args[1:])

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := api.VaultPluginTLSProvider(tlsConfig)

	log.Printf("vault-artifactory-secrets-plugin %s, commit %s, built at %s\n", version, commit, date)
	if err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: artifactorysecrets.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	}); err != nil {
		log.Fatal(err)
	}
}
