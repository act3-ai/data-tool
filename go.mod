module gitlab.com/act3-ai/asce/data/tool

go 1.23.1

toolchain go1.23.5

require (
	git.act3-ace.com/ace/data/schema v1.2.13
	git.act3-ace.com/ace/data/telemetry/v2 v2.1.0
	git.act3-ace.com/ace/go-auth v0.0.0-20241024134738-dfc3a5ec2ac3
	git.act3-ace.com/ace/go-common v0.0.0-20241029114748-ba3456041b7d
	github.com/Masterminds/sprig/v3 v3.3.0
	github.com/adrg/xdg v0.5.3
	github.com/djherbis/atime v1.1.0
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.18.0
	github.com/go-chi/chi/v5 v5.2.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gosuri/uitable v0.0.4
	github.com/klauspost/compress v1.17.11
	github.com/kopoli/go-terminal-size v0.0.0-20170219200355-5c97524c8b54
	github.com/notaryproject/notation-core-go v1.1.0
	github.com/notaryproject/notation-go v1.2.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	go.etcd.io/bbolt v1.3.11
	golang.org/x/exp v0.0.0-20250106191152-7588d65b2ba8
	golang.org/x/net v0.34.0
	golang.org/x/sync v0.10.0
	golang.org/x/term v0.28.0
	golang.org/x/text v0.21.0
	k8s.io/apimachinery v0.32.1
	k8s.io/utils v0.0.0-20241210054802-24370beab758
	oras.land/oras-go/v2 v2.5.0
	sigs.k8s.io/yaml v1.4.0
)

// testing only dependencies
require (
	github.com/fortytw2/leaktest v1.3.0
	github.com/google/go-containerregistry v0.20.3
	github.com/stretchr/testify v1.10.0
)

require (
	dario.cat/mergo v1.0.1 // indirect
	git.act3-ace.com/ace/hub/api/v6 v6.0.0-20241203160221-3a1e07ae7102 // indirect
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/MakeNowJust/heredoc/v2 v2.0.1 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.5 // indirect
	github.com/go-jose/go-jose/v4 v4.0.4 // indirect
	github.com/go-ldap/ldap/v3 v3.4.8 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/gomarkdown/markdown v0.0.0-20241205020045-f7e15b2f3e62 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/muhlemmer/gu v0.3.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/neilotoole/slogt v1.1.0 // indirect
	github.com/notaryproject/notation-plugin-framework-go v1.0.0 // indirect
	github.com/notaryproject/tspclient-go v0.2.0 // indirect
	github.com/veraison/go-cose v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/zitadel/logging v0.6.1 // indirect
	github.com/zitadel/oidc/v3 v3.31.0 // indirect
	github.com/zitadel/schema v1.3.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.33.0 // indirect
	go.opentelemetry.io/otel/metric v1.33.0 // indirect
	go.opentelemetry.io/otel/trace v1.33.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.32.0 // indirect
	golang.org/x/mod v0.22.0 // indirect
	golang.org/x/oauth2 v0.25.0 // indirect
	k8s.io/api v0.31.3 // indirect
	sigs.k8s.io/controller-runtime v0.19.0 // indirect
	sigs.k8s.io/gateway-api v1.2.1 // indirect
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.7.0
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/jsonschema v0.12.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.20.5 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.59.1 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/sys v0.29.0 // indirect
	google.golang.org/protobuf v1.36.3 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	sigs.k8s.io/json v0.0.0-20241010143419-9aa6b5e7a4b3 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.2 // indirect
)
