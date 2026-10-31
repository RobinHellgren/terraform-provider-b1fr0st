package main

import (
	"context"
	"log"

	"github.com/RobinHellgren/terraform-provider-b1fr0st/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "dev"

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/robinhellgren/b1fr0st",
	})
	if err != nil {
		log.Fatal(err)
	}
}
