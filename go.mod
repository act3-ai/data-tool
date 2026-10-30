module github.com/act3-ai/data-tool

go 1.27.0

require (
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/act3-ai/bottle-schema v1.2.16
	github.com/act3-ai/data-telemetry/v3 v3.1.5
	github.com/act3-ai/go-common v0.0.0-20261022193825-2e565862f71d
	github.com/adrg/xdg v0.5.3
	github.com/djherbis/atime v1.1.0
	github.com/dustin/go-humanize v1.1.0
	github.com/fatih/color v1.19.0
	github.com/go-ozzo/ozzo-validation/v4 v4.5.0
	github.com/gosuri/uitable v0.0.4
	github.com/hashicorp/go-version v1.9.0
	github.com/klauspost/compress v1.20.3
	github.com/kopoli/go-terminal-size v0.0.0-20170219200355-5c97524c8b54
	github.com/notaryproject/notation-core-go v1.3.0
	github.com/notaryproject/notation-go v1.3.2
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.1
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	go.etcd.io/bbolt v1.5.0
	golang.org/x/exp v0.0.0-20261014143641-f0251eb4aa3e
	golang.org/x/net v0.62.0
	golang.org/x/sync v0.24.0
	golang.org/x/term v0.47.0
	golang.org/x/text v0.43.0
	k8s.io/apimachinery v0.37.2
	k8s.io/utils v0.0.0-20261023123148-ca04ba58ecab
	oras.land/oras-go/v2 v2.7.0
	sigs.k8s.io/yaml v1.6.0
)

//testing only dependencies
require (
	github.com/fortytw2/leaktest v1.3.0
	github.com/google/go-containerregistry v0.22.1
	github.com/stretchr/testify v1.12.1
)
require (
	cel.dev/expr v0.25.2 // indirect
	dario.cat/mergo v1.0.2 // indirect
	github.com/Azure/go-ntlmssp v0.1.1 // indirect
	github.com/MakeNowJust/heredoc/v2 v2.0.1 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.6.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/charmbracelet/x/ansi v0.11.9 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/elastic/crd-ref-docs v0.3.0 // indirect
	github.com/fxamacker/cbor/v2 v2.10.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.8 // indirect
	github.com/go-jose/go-jose/v4 v4.1.7 // indirect
	github.com/go-ldap/ldap/v3 v3.4.15 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v1.0.2 // indirect
	github.com/go-openapi/jsonreference v1.0.3 // indirect
	github.com/go-openapi/swag/jsonutils v0.29.2 // indirect
	github.com/gobuffalo/flect v1.0.3 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/gomarkdown/markdown v0.0.0-20261026184151-79e163b7b632 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/huandu/xstrings v1.6.3 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/jsonschema v0.14.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/mattn/go-colorable v0.1.16 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-runewidth v0.0.31 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/muhlemmer/gu v0.3.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/neilotoole/slogt v1.1.0 // indirect
	github.com/notaryproject/notation-plugin-framework-go v1.0.0 // indirect
	github.com/notaryproject/tspclient-go v1.0.0 // indirect
	github.com/pandatix/go-abnf v0.5.3 // indirect
	github.com/pb33f/go-yaml v0.1.1 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.2 // indirect
	github.com/prometheus/client_golang v1.25.0 // indirect
	github.com/prometheus/client_model v0.6.3 // indirect
	github.com/prometheus/common v0.73.1 // indirect
	github.com/prometheus/procfs v0.23.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/shopspring/decimal v1.6.0 // indirect
	github.com/sirupsen/logrus v1.10.2 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/veqryn/slog-context v0.9.0 // indirect
	github.com/veraison/go-cose v1.3.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zitadel/oidc/v3 v3.51.16 // indirect
	github.com/zitadel/schema v1.3.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.72.0 // indirect
	go.opentelemetry.io/otel v1.47.0 // indirect
	go.opentelemetry.io/otel/log v1.47.0 // indirect
	go.opentelemetry.io/otel/metric v1.47.0 // indirect
	go.opentelemetry.io/otel/trace v1.47.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/oauth2 v0.37.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/api v0.37.2 // indirect
	k8s.io/apiextensions-apiserver v0.37.2 // indirect
	k8s.io/code-generator v0.37.2 // indirect
	k8s.io/gengo/v2 v2.0.0-20260408192533-25e2208e0dc3 // indirect
	k8s.io/klog/v2 v2.140.0 // indirect
	k8s.io/kube-openapi v0.0.0-20261014044854-ca3452a79603 // indirect
	sigs.k8s.io/controller-tools v0.23.0 // indirect
	sigs.k8s.io/json v0.0.0-20260909141634-11ed52e25bc5 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.4.2 // indirect
	sigs.k8s.io/structured-merge-diff/v7 v7.0.0 // indirect
)

tool (
	github.com/elastic/crd-ref-docs
	sigs.k8s.io/controller-tools/cmd/controller-gen
)
