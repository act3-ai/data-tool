module gitlab.com/act3-ai/asce/data/tool

go 1.22.2

require (
	git.act3-ace.com/ace/data/schema v1.2.13
	git.act3-ace.com/ace/data/telemetry v0.20.1
	git.act3-ace.com/ace/go-common v0.0.0-20240319120227-e77a013aa92d
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/adrg/xdg v0.5.0
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.17.0
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/gosuri/uitable v0.0.4
	github.com/klauspost/compress v1.17.9
	github.com/kopoli/go-terminal-size v0.0.0-20170219200355-5c97524c8b54
	github.com/notaryproject/notation-core-go v1.0.3
	github.com/notaryproject/notation-go v1.1.1
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.1.0
	github.com/sourcegraph/conc v0.3.0
	github.com/spf13/cobra v1.8.1
	github.com/spf13/pflag v1.0.5
	go.etcd.io/bbolt v1.3.11
	golang.org/x/exp v0.0.0-20240719175910-8a7402abbf56
	golang.org/x/net v0.27.0
	golang.org/x/sync v0.8.0
	golang.org/x/term v0.22.0
	golang.org/x/text v0.17.0
	k8s.io/apimachinery v0.30.3
	k8s.io/utils v0.0.0-20240821151609-f90d01438635
	oras.land/oras-go/v2 v2.5.0
	sigs.k8s.io/yaml v1.4.0
)

// testing only dependencies
require (
	github.com/fortytw2/leaktest v1.3.0
	github.com/google/go-containerregistry v0.20.1
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/Azure/go-ntlmssp v0.0.0-20221128193559-754e69321358 // indirect
	github.com/MakeNowJust/heredoc/v2 v2.0.1 // indirect
	github.com/docker/cli v26.0.0+incompatible // indirect
	github.com/docker/docker v26.0.0+incompatible // indirect
	github.com/fxamacker/cbor/v2 v2.6.0 // indirect
	github.com/go-asn1-ber/asn1-ber v1.5.5 // indirect
	github.com/go-ldap/ldap/v3 v3.4.8 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.0 // indirect
	github.com/gomarkdown/markdown v0.0.0-20240419095408-642f0ee99ae2 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/neilotoole/slogt v1.1.0 // indirect
	github.com/notaryproject/notation-plugin-framework-go v1.0.0 // indirect
	github.com/veraison/go-cose v1.1.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.25.0 // indirect
	golang.org/x/mod v0.19.0 // indirect
)

require (
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.2.1 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/djherbis/atime v1.1.0
	github.com/docker/distribution v2.8.3+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.7.0
	github.com/huandu/xstrings v1.4.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
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
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/segmentio/ksuid v1.0.4 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/vbatts/tar-split v0.11.5 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.120.1 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.4.1 // indirect
)
