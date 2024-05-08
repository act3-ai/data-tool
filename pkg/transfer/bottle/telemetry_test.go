package bottle

import (
	"context"
	"errors"
	"fmt"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote/auth"

	telemv1alpha1 "gitlab.com/act3-ai/asce/data/telemetry/pkg/apis/config.telemetry.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/tool/pkg/apis/config.dt.act3-ace.io/v1alpha1"
	"gitlab.com/act3-ai/asce/data/tool/pkg/conf"
	reg "gitlab.com/act3-ai/asce/data/tool/pkg/registry"
)

// If you change this function then file an issue with CSI to get that updated as well.

// This is how the CSI driver uses this library.
func ExamplePull() {
	ctx := context.Background()

	config := conf.NewConfiguration("example-agent")

	// define telemetry hosts
	userName := "testUser"
	hosts := []telemv1alpha1.Location{
		{
			Name: "ace-telemetry",
			URL:  "https://127.0.0.1:8100",
		},
	}

	// define required transfer configuration
	reference := "my.reg.com/some/bottle:v1"
	pullPath := "some/path"

	// optional bottle part selection
	// TODO: add example of TLS configuration
	registryConfig := v1alpha1.RegistryConfig{
		Configs: map[string]v1alpha1.Registry{
			"my.reg.com": {
				Endpoints: []string{"http://my.reg.com"}, // allow plain-http requests to my.reg.com
			},
		},
		EndpointConfig: map[string]v1alpha1.EndpointConfig{
			"my.reg.com": {
				ReferrersType: "auto",
			},
		},
	}

	credentialFn := reg.WithCredential(ctx, "my.reg.com", auth.Credential{Password: "my-secret"})

	rt := reg.NewRemoteTargeter(&registryConfig, "user-agent") // use custom registry configuration and/or user-agent
	targetFn := func(ctx context.Context, ref string) (oras.GraphTarget, error) {
		return rt.NewRemoteTarget(ctx, ref, credentialFn)
	}

	cachePath := "cache/path"
	labelSelectors := []string{"foo=bar", "x!=45"}
	partNames := []string{}
	artifacts := []string{}

	opts := []TransferOption{
		WithTelemetry(hosts, userName),
		WithNewGraphTargetFn(targetFn),
		WithCachePath(cachePath),
		WithPartSelection(labelSelectors, partNames, artifacts),
	}

	// build option set
	tc := NewTransferConfig(ctx, reference, pullPath, config, opts...)

	// also fails on send telemetry event failure
	_, err := Pull(ctx, *tc)
	if errors.Is(err, ErrTelemetrySend) {
		fmt.Println("Ignoring telemetry error") //nolint
	} else if err != nil {
		fmt.Println("Bottle pull failed") //nolint
	}
	fmt.Println("Success") //nolint
}
