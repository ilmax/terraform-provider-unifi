package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/ilmax/terraform-provider-unifi/internal/legacyprovider"
)

var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to enable debug")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/ilmax/unifilegacy",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), legacyprovider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
