# Changelog

All notable changes to this project will be documented in this file.

## [1.15.21] - 2025-04-15

### 💼 Other

- Update brews config

## [1.15.20] - 2025-04-15

### 💼 Other

- Archive binaries

## [1.15.19] - 2025-04-15

### 💼 Other

- Add checksums file

## [1.15.18] - 2025-04-14

### 🗡️ Dagger

- *(release)* Minor fixes to goreleaser config and usage

## [1.15.17] - 2025-04-14

### 🐛 Bug Fixes (goreleaser)

- *(goreleaser)* Remove goreleaser by @nathan-joslin

### 🐛 Bug Fixes (release)

- *(release)* Update release script for homebrew tap by @nathan-joslin

### 🗡️ Dagger

- Build with goreleaser
- Release with goreleaser
- Update publish step with goreleaser

## [1.15.16] - 2025-04-11

### 🐛 Bug Fixes (goreleaser)

- *(goreleaser)* Another attempt

## [1.15.15] - 2025-04-11

### 🐛 Bug Fixes (goreleaser)

- *(goreleaser)* Further refinement

## [1.15.14] - 2025-04-10

### 🐛 Bug Fixes (goreleaser)

- *(goreleaser)* Disable archives in favor of binaries

## [1.15.13] - 2025-04-10

### 🐛 Bug Fixes (goreleaser)

- *(goreleaser)* Typo

## [1.15.12] - 2025-04-10

### 🐛 Bug Fixes (goreleaser)

- *(goreleaser)* Skip builds

## [1.15.11] - 2025-04-10

### 🐛 Bug Fixes (tap)

- *(tap)* Add goreleaser for homebrew tap updates

### 🗡️ Dagger

- *(image)* Fix image source annotation to connect image to repo in github

## [1.15.10] - 2025-04-10

### 🐛 Bug Fixes (test,git)

- *(test,git)* Ensure commit signing is not used when creating tags in test repository by @nathan-joslin

### 💼 Other

- Remove .gitlab-ci.yml, add github actions workflow by @nathan-joslin
- *(go)* Update golangci-lint config with github local prefixes by @nathan-joslin
- *(markdown)* Remove gitlab issue templates from markdownlint config ignore list by @nathan-joslin
- Convert release script to use github references by @nathan-joslin
- Move binary builds to publish step in release script to avoid dirty builds by @nathan-joslin
- *(go)* Update remaining gitlab references by @nathan-joslin
- *(go)* Appease golangci-lint by @nathan-joslin
- Update release.sh to enable image scaning by @nathan-joslin

### 📦 Dependencies

- Bump go-common digest by @nathan-joslin
- Bump golang.org/x/exp digest to 7e4ce0ab07d0 by @nathan-joslin

### 🗡️ Dagger

- Bump to v0.18.2 by @nathan-joslin
- *(deps)* Bump golangci-lint module to v0.9.0 by @nathan-joslin
- *(deps)* Bump registry-config module to v0.8.0 by @nathan-joslin
- Convert to github references and gh cli by @nathan-joslin
- Remove unused const for gitlab cli by @nathan-joslin
- *(test,functional)* Minor updates to handling of services by @nathan-joslin
- *(release)* Fix release notes images tag by @nathan-joslin
- *(release)* Fix order of release notes

## [1.15.9] - 2025-04-01

### 🐛 Bug Fixes (dagger,build)

- *(dagger,build)* Add build tags to reduce binary size and versioning

### 📦 Dependencies

- *(secret)* Remove internal/secret in favor of go-common/pkg/secret

### 🗡️ Dagger

- *(release)* Update git-cliff changelong generation to prepend only
- Bump to v0.18.0
- Add table of images to release notes

### 🧪 Testing

- *(dagger)* Add standalone golangci-lint function
- *(markdownlint)* Ignore linting auto-generated release docs
- *(integration)* Add missing telemetry response validation

## [1.15.8] - 2025-03-31

### 🚀 Features

- *(dagger)* Move pipeline to dagger, new release script.

## [1.15.7](https://git.act3-ace.com/ace/data/tool/compare/v1.15.6...v1.15.7) (2025-03-26)


### Bug Fixes

* **deps:** update gitlab.com/act3-ai/asce/go-common digest to 734a59d ([21aa4ee](https://git.act3-ace.com/ace/data/tool/commit/21aa4ee031dfa7e592cbd30a899c002a9ac29b55))
* **deps:** update k8s.io/utils digest to 1f6e0b7 ([e72fae2](https://git.act3-ace.com/ace/data/tool/commit/e72fae28cd315f87e2d6e4afca4e99596141914c))
* **scan:** handling of grype v0.90.0 checksum parsing ([27e6180](https://git.act3-ace.com/ace/data/tool/commit/27e618094ee73e1d22519724defab6846e0f4cff))

## [1.15.6](https://git.act3-ace.com/ace/data/tool/compare/v1.15.5...v1.15.6) (2025-03-10)


### Bug Fixes

* **scan:** handle and display grype warnings ([4cd1594](https://git.act3-ace.com/ace/data/tool/commit/4cd15941683e90224ae23b9b5eb2df1c6c85c71e))

## [1.15.5](https://git.act3-ace.com/ace/data/tool/compare/v1.15.4...v1.15.5) (2025-03-07)


### Bug Fixes

* **deps:** update dependency devsecops/cicd/pipeline to v21.0.4 ([3cccb5e](https://git.act3-ace.com/ace/data/tool/commit/3cccb5ea1d5f3657a509702df2b27fe7a755d402))
* **deps:** update gitlab.com/act3-ai/asce/go-common digest to 59817e7 ([c544501](https://git.act3-ace.com/ace/data/tool/commit/c544501b2a9a35abc315cfe000eb6cef385f3044))
* **deps:** update golang.org/x/exp digest to 054e65f ([373815e](https://git.act3-ace.com/ace/data/tool/commit/373815ea239d432d7e5616185208cb035344cb7f))
* **deps:** update golang.org/x/exp digest to dead583 ([e560ab8](https://git.act3-ace.com/ace/data/tool/commit/e560ab81aed73fb388a012c58d48819b7e11e17f))
* **deps:** update module github.com/golangci/golangci-lint to v1.64.6 ([6d33d3f](https://git.act3-ace.com/ace/data/tool/commit/6d33d3fa4cc61a0e50d4b3dc0d5fabc5b8fa631a))
* **deps:** update module github.com/notaryproject/notation-go to v1.3.1 ([adfd3d9](https://git.act3-ace.com/ace/data/tool/commit/adfd3d9c17d0e31e4f7cb7e1ac593d7955d3de8d))
* **deps:** update module github.com/opencontainers/image-spec to v1.1.1 ([52bfc96](https://git.act3-ace.com/ace/data/tool/commit/52bfc964cafffd6348ed7cbaadb52deaddc4c0de))
* **deps:** update module golang.org/x/net to v0.37.0 ([3ebe208](https://git.act3-ace.com/ace/data/tool/commit/3ebe208522d33981138d8a09f1ba7d4a513fdbe8))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.1.12 ([cb5ea88](https://git.act3-ace.com/ace/data/tool/commit/cb5ea88551ce1c32a79966dcbf0a6fb40e75545c))
* **scan:** grype db checksum parsing ([940f7e6](https://git.act3-ace.com/ace/data/tool/commit/940f7e62c8af20723c77587e23e905e40e20b660))

## [1.15.4](https://git.act3-ace.com/ace/data/tool/compare/v1.15.3...v1.15.4) (2025-02-20)


### Bug Fixes

* **deps:** remove go-chi from pypi ([f6a7bcb](https://git.act3-ace.com/ace/data/tool/commit/f6a7bcb982b7c4e876d48dbe4a338a0efaeae3ce))
* **deps:** switch git.act3-ace.com/ace/data/schema to gitlab.com/act3-ai/asce/data/schema ([0fe0b2e](https://git.act3-ace.com/ace/data/tool/commit/0fe0b2eea84a8cb7ccfdb2e405e5a2be415d1759))
* **deps:** update dependency devsecops/cicd/pipeline to v20.2.13 ([29f82b7](https://git.act3-ace.com/ace/data/tool/commit/29f82b7ba637a74a24cdd1f36d6435f10363d808))
* **deps:** update dependency devsecops/cicd/pipeline to v21 ([1889f33](https://git.act3-ace.com/ace/data/tool/commit/1889f33b898da6cb4024cd541d21f4797598ee4c))
* **deps:** update gitlab.com/act3-ai/asce/go-common digest to c9e7dd3 ([71958ee](https://git.act3-ace.com/ace/data/tool/commit/71958ee15b09b7c41598a81afbca5cc575d87737))
* **deps:** update golang.org/x/exp digest to aa4b98e ([0fdfaa9](https://git.act3-ace.com/ace/data/tool/commit/0fdfaa9e1a6b4adfb5bff70af6e6497ba3903213))
* **deps:** update module git.act3-ace.com/ace/data/telemetry/v3 to v3.0.2 ([60e8d66](https://git.act3-ace.com/ace/data/tool/commit/60e8d66e71b3e722e0e673630e8205497e055de8))
* **deps:** update module github.com/golangci/golangci-lint to v1.64.5 ([f231e54](https://git.act3-ace.com/ace/data/tool/commit/f231e54e46ee5d2fdac76f16c58f814047fdaf1a))
* **deps:** update module github.com/klauspost/compress to v1.18.0 ([11cc1e6](https://git.act3-ace.com/ace/data/tool/commit/11cc1e69f273873f9b2742d302834f23b3e27b5a))
* **deps:** update module github.com/spf13/cobra to v1.9.1 ([ec67674](https://git.act3-ace.com/ace/data/tool/commit/ec6767437418cc5fb7a010ad05c0197279d84dc1))
* **deps:** update module go.etcd.io/bbolt to v1.4.0 ([df06442](https://git.act3-ace.com/ace/data/tool/commit/df06442f52fe80367fe4b377ed022136bfc6e818))
* **deps:** update module k8s.io/apimachinery to v0.32.2 ([25b9f84](https://git.act3-ace.com/ace/data/tool/commit/25b9f84bba3cb8575b9af340af00f01d070580cc))
* **deps:** update module sigs.k8s.io/controller-tools to v0.17.2 ([6a445cf](https://git.act3-ace.com/ace/data/tool/commit/6a445cf18fb1a4d263ed5caf49182139c601f307))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.1.11 ([84c0a41](https://git.act3-ace.com/ace/data/tool/commit/84c0a41288ba11729842f375c39d8582367c91bd))
* **mirror:** Deserialize does not push referrers ([95ee472](https://git.act3-ace.com/ace/data/tool/commit/95ee472e0cd6e1d15d41567df38cec67447b9492))
* **renovate:** increase renovate MR limit ([644a80c](https://git.act3-ace.com/ace/data/tool/commit/644a80c5b7e11e89e2c133df68964be6bb33eb73))

## [1.15.3](https://git.act3-ace.com/ace/data/tool/compare/v1.15.2...v1.15.3) (2025-02-04)


### Bug Fixes

* **deps:** convert git.act3-ace.com/ace/go-common to gitlab.com/act3-ai/asce/go-common ([247a42a](https://git.act3-ace.com/ace/data/tool/commit/247a42a198b5968ad504065d1b71cc6ccf74e498))
* **deps:** remove go-chi from bottle gui ([9b0bd16](https://git.act3-ace.com/ace/data/tool/commit/9b0bd164895e9cb4f279805b4df4b643f0eac5c2))
* **deps:** update dependency devsecops/cicd/pipeline to v20.2.8 ([80ec5b7](https://git.act3-ace.com/ace/data/tool/commit/80ec5b7cee3c6ae725a9f21c2ae339e5197dd1aa))
* **deps:** update dependency devsecops/cicd/pipeline to v20.2.9 ([aab227b](https://git.act3-ace.com/ace/data/tool/commit/aab227ba2420e35f6f67a74654ee77432feb22fd))
* **deps:** update git.act3-ace.com/ace/go-auth digest to a991d7d ([2c1c541](https://git.act3-ace.com/ace/data/tool/commit/2c1c54152a7c0b1a9c0369647a3e2ef2ee846f5c))
* **deps:** update gitlab.com/act3-ai/asce/go-common digest to c79ed3c ([fc62fee](https://git.act3-ace.com/ace/data/tool/commit/fc62fee5724122340b8a44ebe6042e69f3790a92))
* **deps:** update golang.org/x/exp digest to e0ece0d ([92ade0e](https://git.act3-ace.com/ace/data/tool/commit/92ade0eb667862e4705ff8b1111e8d275653b761))
* **deps:** update module git.act3-ace.com/ace/data/telemetry/v2 to v3 ([8eb776a](https://git.act3-ace.com/ace/data/tool/commit/8eb776a0ef2b1140ed3420983caeefa615a63982))
* **deps:** update module github.com/google/go-containerregistry to v0.20.3 ([d0a8301](https://git.act3-ace.com/ace/data/tool/commit/d0a8301f4010f964038f5edbdc45beed01d77dce))
* **deps:** update module github.com/notaryproject/notation-core-go to v1.2.0 ([a265a2d](https://git.act3-ace.com/ace/data/tool/commit/a265a2d247e68e34979ffc529cec6dd90ea87fc0))
* **deps:** update module github.com/notaryproject/notation-go to v1.3.0 ([6c715cc](https://git.act3-ace.com/ace/data/tool/commit/6c715cca8be0b172d44d1075ea085172e1929d17))
* **deps:** update module github.com/spf13/pflag to v1.0.6 ([d3ca3d7](https://git.act3-ace.com/ace/data/tool/commit/d3ca3d7b936f764bc1aaad19c2f7447853ceef58))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.1.8 ([8d26f52](https://git.act3-ace.com/ace/data/tool/commit/8d26f52e859b9ab4ce5c8b4f65aa2301df8e33fe))
* **docs:** Add tutorial for batch-serialize ([32ca155](https://git.act3-ace.com/ace/data/tool/commit/32ca155167c650cf72a63598c96c373734ea0f40))
* restore windows compatibility ([a5e7541](https://git.act3-ace.com/ace/data/tool/commit/a5e75416b30203ad935c365f34c5071b0c0fe456))
* **test:** fsutil test ([b9fc007](https://git.act3-ace.com/ace/data/tool/commit/b9fc00766ff85b5bcaf7585b4a14428b43a83dc5))

## [1.15.2](https://git.act3-ace.com/ace/data/tool/compare/v1.15.1...v1.15.2) (2025-01-21)


### Bug Fixes

* **deps:** update dependency devsecops/cicd/pipeline to v20.2.7 ([eb49130](https://git.act3-ace.com/ace/data/tool/commit/eb4913062b9a051a5e45bcfb2fc7d5a66626e4fd))
* **deps:** update dependency go to v1.23.5 ([d5b462f](https://git.act3-ace.com/ace/data/tool/commit/d5b462f125cacebfe52fc9a06d1e9c0f5c410f49))
* **deps:** update module k8s.io/apimachinery to v0.32.1 ([1914043](https://git.act3-ace.com/ace/data/tool/commit/191404313061d8ed64bd3fd7a747cbfd5e60a47d))
* **deps:** update module sigs.k8s.io/controller-tools to v0.17.1 ([aa96050](https://git.act3-ace.com/ace/data/tool/commit/aa96050474e11c1517d5af759311e878edd376d4))
* **git:** improve git version extraction ([1df8781](https://git.act3-ace.com/ace/data/tool/commit/1df87816ae8a0fe5f7b19394aba9aac7906a62aa))
* **login:** CLI secrets now work on mac os x ([9de6d08](https://git.act3-ace.com/ace/data/tool/commit/9de6d086b6ae79fcd7834eea059e7ba66372c7d2))

## [1.15.1](https://git.act3-ace.com/ace/data/tool/compare/v1.15.0...v1.15.1) (2025-01-08)


### Bug Fixes

* **deps:** update devsecops/cicd/pipeline to v20.2.5 ([d71e13c](https://git.act3-ace.com/ace/data/tool/commit/d71e13c749a1e2a9b21e8a73b8dcf87949fa0928))
* **deps:** update golang.org/x/exp digest to 4a55095 ([3d4ff1f](https://git.act3-ace.com/ace/data/tool/commit/3d4ff1f291a6171cc6b770190d94f131b5dbd159))
* **deps:** update golang.org/x/exp digest to 7588d65 ([71ee233](https://git.act3-ace.com/ace/data/tool/commit/71ee233f73bd29a75bc9dbdfb93fde60c67f14a4))
* **deps:** update golang.org/x/exp digest to 7d7fa50 ([69da008](https://git.act3-ace.com/ace/data/tool/commit/69da00823d502a7e6b9e0d37e9fe5d94a9e230c3))
* **deps:** update k8s.io/utils digest to 24370be ([edcab6d](https://git.act3-ace.com/ace/data/tool/commit/edcab6d60ca32469084a8802f96341e830bae6d6))
* **deps:** update module git.act3-ace.com/ace/data/telemetry/v2 to v2.1.0 ([80a2ab0](https://git.act3-ace.com/ace/data/tool/commit/80a2ab040b23162aa38eaa023b3e6e8421a4dff0))
* **deps:** update module github.com/go-chi/chi/v5 to v5.2.0 ([0cc3ac5](https://git.act3-ace.com/ace/data/tool/commit/0cc3ac5a4b073b8a00ae7291cdf037dc3ce31734))
* **deps:** update module github.com/golangci/golangci-lint to v1.63.4 ([cc3e36a](https://git.act3-ace.com/ace/data/tool/commit/cc3e36a309c04c9e9c4bd4386d16328ea27b709f))
* **deps:** update module golang.org/x/net to v0.34.0 ([c62cb75](https://git.act3-ace.com/ace/data/tool/commit/c62cb7515c6fdc3fc4da9fca22c5cc9ffbc903d0))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.1.5 ([f39c864](https://git.act3-ace.com/ace/data/tool/commit/f39c864735edfb83779a1e1d4ec90231ba1afb3d))
* **git:** handling of rewritten git history ([f7a6bd3](https://git.act3-ace.com/ace/data/tool/commit/f7a6bd372bf637535c978ca3506f77635d35f664))
* **git:** LFS fails when provided a relative cache path ([36d5a85](https://git.act3-ace.com/ace/data/tool/commit/36d5a8574d7c779de00801f4ed9bac0a7fed3bf9))
* **mirror,diff:** Add support for extra manifests annotations and new output types ([1652cbf](https://git.act3-ace.com/ace/data/tool/commit/1652cbf8d6d91949a512492426f440669b89f204))
* **mirror:** add blob caching to gather ([95477b6](https://git.act3-ace.com/ace/data/tool/commit/95477b692b2431228ac6f54f98f42a5b4b0c9edd))

# [1.15.0](https://git.act3-ace.com/ace/data/tool/compare/v1.14.1...v1.15.0) (2024-12-11)


### Bug Fixes

* **deps:** update golang.org/x/exp digest to 43b7b7c ([393d18b](https://git.act3-ace.com/ace/data/tool/commit/393d18b6ee40497b67512c6a96bdaf9c42846d72))
* **deps:** update module golang.org/x/net to v0.32.0 ([f2ad564](https://git.act3-ace.com/ace/data/tool/commit/f2ad5645ed221b0dea8b332ecf20f1a0cccc666b))
* **login:** remove support for insecure passwords ([5f2a39d](https://git.act3-ace.com/ace/data/tool/commit/5f2a39d03479182719f6dc77b9914aafa5a496ac))
* **pypi:** failed requirements output missing semicolon if no constraints ([e1a07df](https://git.act3-ace.com/ace/data/tool/commit/e1a07dfa331a075346e082c8b6e0e08fadc56c41))


### Features

* **mirror:** Add continue flag to clone ([d39e5ca](https://git.act3-ace.com/ace/data/tool/commit/d39e5caaca61595498a714f123c6c8f4b2cf1aa5))

## [1.14.1](https://git.act3-ace.com/ace/data/tool/compare/v1.14.0...v1.14.1) (2024-12-06)


### Bug Fixes

* **bottle:** failure to push cached signatures ([549263f](https://git.act3-ace.com/ace/data/tool/commit/549263f5a560aa820f866fdc930d128efe62edd2))
* **login:** add more secure methods for providing passwords ([67b36c3](https://git.act3-ace.com/ace/data/tool/commit/67b36c3df48199e303f740ecffe3d5d7e988f6a9))

# [1.14.0](https://git.act3-ace.com/ace/data/tool/compare/v1.13.2...v1.14.0) (2024-12-04)


### Bug Fixes

* **bottle:** fail during caching of existing notation signature ([7a297bc](https://git.act3-ace.com/ace/data/tool/commit/7a297bc5f28df2b7bf9abf5a878b578fbfff1a5e))
* **bottle:** fail early when pulling into existing bottles or sub-directories ([0eacdeb](https://git.act3-ace.com/ace/data/tool/commit/0eacdeb51cdbc3f8ca915c5121baa2af13594af0))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.35 ([756577e](https://git.act3-ace.com/ace/data/tool/commit/756577ea1e568a2ec5e578dce8162edf748148cf))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.36 ([3879aa3](https://git.act3-ace.com/ace/data/tool/commit/3879aa39bdc9817f92b227871ba89d70b965fffe))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.37 ([e316b51](https://git.act3-ace.com/ace/data/tool/commit/e316b51103f392f1a2dc58948515240ed3feef86))
* **deps:** update dependency go to v1.23.4 ([e586b7e](https://git.act3-ace.com/ace/data/tool/commit/e586b7e86559f2bf5a05d5f15b4b12ab5e88f854))
* **deps:** update git.act3-ace.com/ace/go-common digest to ba34560 ([6d6ae5b](https://git.act3-ace.com/ace/data/tool/commit/6d6ae5b689bf66dbb0f44c0d46693d4b440f598a))
* **deps:** update golang.org/x/exp digest to 2d47ceb ([d2e5643](https://git.act3-ace.com/ace/data/tool/commit/d2e5643ff91fc7793595464aa281c339c9429661))
* **deps:** update golang.org/x/exp digest to f66d83c ([e8b9f07](https://git.act3-ace.com/ace/data/tool/commit/e8b9f071cf7e90117dedc7e9e83ad96735955521))
* **deps:** update k8s.io/utils digest to 6fe5fd8 ([7548832](https://git.act3-ace.com/ace/data/tool/commit/754883256f061d94b41ac8bce296f29a5f03c59b))
* **deps:** update module github.com/adrg/xdg to v0.5.2 ([c9e0486](https://git.act3-ace.com/ace/data/tool/commit/c9e04868d23d7f074d250bee52f9be1f2254ad87))
* **deps:** update module github.com/adrg/xdg to v0.5.3 ([4845080](https://git.act3-ace.com/ace/data/tool/commit/48450803213c007119475c12adab2ea69a619e13))
* **deps:** update module github.com/fatih/color to v1.18.0 ([7a19135](https://git.act3-ace.com/ace/data/tool/commit/7a191356824fb108742da1ecc6cbde0213b0680c))
* **deps:** update module github.com/golangci/golangci-lint to v1.61.0 ([a141b68](https://git.act3-ace.com/ace/data/tool/commit/a141b68f4281a76fc0228f3de0dea4a15ace063d))
* **deps:** update module github.com/golangci/golangci-lint to v1.62.0 ([697effa](https://git.act3-ace.com/ace/data/tool/commit/697effa53b2395da256ecb3b3daafef1f0a35b0a))
* **deps:** update module github.com/golangci/golangci-lint to v1.62.2 ([3f4a19f](https://git.act3-ace.com/ace/data/tool/commit/3f4a19fb7914c5cfdf37cc9e51c72c060d951161))
* **deps:** update module github.com/google/ko to v0.17.1 ([26b0a1b](https://git.act3-ace.com/ace/data/tool/commit/26b0a1b89a3a0976f24a89b386e9866d5b32173f))
* **deps:** update module github.com/klauspost/compress to v1.17.11 ([5151cf8](https://git.act3-ace.com/ace/data/tool/commit/5151cf8217336aefabe2330f60d0532e7449bde3))
* **deps:** update module github.com/stretchr/testify to v1.10.0 ([14b33f1](https://git.act3-ace.com/ace/data/tool/commit/14b33f151331bdab5e4dee89dbc1769bc9c52c06))
* **deps:** update module golang.org/x/net to v0.31.0 ([9e1980b](https://git.act3-ace.com/ace/data/tool/commit/9e1980b5a9116265680b1cd84dd6626db571ed85))
* **deps:** update module golang.org/x/sync to v0.10.0 ([1a880d0](https://git.act3-ace.com/ace/data/tool/commit/1a880d02629162742a30061ee056eaf3c8fbcf3f))
* **deps:** update module golang.org/x/term to v0.27.0 ([ada329b](https://git.act3-ace.com/ace/data/tool/commit/ada329b01e6c6f045f2d14eed8554e727d34550a))
* **deps:** update module k8s.io/apimachinery to v0.31.2 ([d221037](https://git.act3-ace.com/ace/data/tool/commit/d2210371cb40736d4cf02ce3b170d45e5c3f1994))
* **deps:** update module k8s.io/apimachinery to v0.31.3 ([2541739](https://git.act3-ace.com/ace/data/tool/commit/2541739aa90a26ad521b7d7f1d97d8cc84c63ddf))
* **deps:** update module sigs.k8s.io/controller-tools to v0.16.5 ([152aca4](https://git.act3-ace.com/ace/data/tool/commit/152aca43dc17fa475c9ed1c9072787b9fc92a5f1))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.61 ([c47f0a1](https://git.act3-ace.com/ace/data/tool/commit/c47f0a1f0bc30ac9f96f37aa099190b609d3cd71))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.62 ([02375ae](https://git.act3-ace.com/ace/data/tool/commit/02375aefb09e3fef88d9ea1b369c26571a432f49))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.1.3 ([b225a4d](https://git.act3-ace.com/ace/data/tool/commit/b225a4d717eb220d8c3aaed44c832442d67d0039))
* **docs:** bottle init missing --bottle-dir flag ([a2dc207](https://git.act3-ace.com/ace/data/tool/commit/a2dc2073158c550278272e7bbddd738631b37dad))
* **gather:** add full digest reference to output ([93b6afc](https://git.act3-ace.com/ace/data/tool/commit/93b6afc400d33cf800c3cfa9d58929d8a565f1ea))
* **mirror,scan:** security scan fix source file repo creation ([05c50f8](https://git.act3-ace.com/ace/data/tool/commit/05c50f8ee14592ce29e0665c4823c57fac33552d))
* **mirror:** panic when provided an invalid mapper ([2f58f8d](https://git.act3-ace.com/ace/data/tool/commit/2f58f8d9987bcfc642883eb022eebcb88cb87fdc))
* **renovate:** add auto-run go mod tidy to configuration ([123f9e2](https://git.act3-ace.com/ace/data/tool/commit/123f9e2ca51e33342a1e217f00ffe8459bcddc2f))


### Features

* add oauth support for telemetry ([927bb74](https://git.act3-ace.com/ace/data/tool/commit/927bb74e3124e5db0d46a9c1b6a773712d46d64e))
* **mirror:** add mirror diff command ([e9b5d67](https://git.act3-ace.com/ace/data/tool/commit/e9b5d67932b1ea73dc2d4036e183fc05f2a99e54))
* refactor security scan with enhancements and SBOM command additions ([b43e777](https://git.act3-ace.com/ace/data/tool/commit/b43e7779a730c983787fe2617040f78d58519334))

## [1.13.2](https://git.act3-ace.com/ace/data/tool/compare/v1.13.1...v1.13.2) (2024-10-11)


### Bug Fixes

* **batch-serialize:** missing descriptors regression ([dd12179](https://git.act3-ace.com/ace/data/tool/commit/dd1217928d6f6d5ab2c77de00d0ccde4481b3f5b))
* **bottle:** deprecation and part update handling ([3d88e85](https://git.act3-ace.com/ace/data/tool/commit/3d88e856a5ef3a5c6b0b06a7cdad1e8a3ef5f8aa))
* **deps:** update dependency go to v1.23.2 ([9d26fbd](https://git.act3-ace.com/ace/data/tool/commit/9d26fbd6e423cbc3c67ec94cbae73ff0c2708e7f))
* **deps:** update golang.org/x/exp digest to 225e2ab ([7b09c65](https://git.act3-ace.com/ace/data/tool/commit/7b09c655716b327ae4ae6e9b695de5cb0a610aa6))
* resolving alternate registry endpoints ([3ef5d73](https://git.act3-ace.com/ace/data/tool/commit/3ef5d7336e7f6c71ddf9d9e6fa0f9acf71b890e2))
* **scan:** fail if missing syft exececutable ([ab17a70](https://git.act3-ace.com/ace/data/tool/commit/ab17a70c80656b9376a1c9f70437d63d121cf41e))
* **scan:** prevent grype database updates ([615ce97](https://git.act3-ace.com/ace/data/tool/commit/615ce978b4570246705dffbb9e12bf2822a44eaf))

## [1.13.1](https://git.act3-ace.com/ace/data/tool/compare/v1.13.0...v1.13.1) (2024-10-03)


### Bug Fixes

* **ci:** use the correct values for MM channel and MM username ([731d570](https://git.act3-ace.com/ace/data/tool/commit/731d570a9278fb598e4fbd7cfa7bb13b53407692))
* **mirror:** Non-deterministic gather index ([f1d6add](https://git.act3-ace.com/ace/data/tool/commit/f1d6adddb5d35243166042f562cf70a048f6139f))

# [1.13.0](https://git.act3-ace.com/ace/data/tool/compare/v1.12.1...v1.13.0) (2024-09-11)


### Bug Fixes

* **ci:** add GOMEMLIMIT and GOMAXPROCS to gitlab-ci.yml ([0cce36b](https://git.act3-ace.com/ace/data/tool/commit/0cce36b8a0d3efdf2a98e2c90435c22845b3c521))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.29 ([cdafcff](https://git.act3-ace.com/ace/data/tool/commit/cdafcffadde45ddbecd2f050b5426f836ecdc864))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.32 ([ce621ed](https://git.act3-ace.com/ace/data/tool/commit/ce621ed89ad4011e0ae58e1d29c6b9527adb5149))
* **deps:** update dependency go to v1.23.1 ([edbecc3](https://git.act3-ace.com/ace/data/tool/commit/edbecc32fd238583b4b81e60e36682a325baad4d))
* **deps:** update golang.org/x/exp digest to 701f63a ([191c647](https://git.act3-ace.com/ace/data/tool/commit/191c647d4daebb5ebff5cf9080dc893b1f3e64fd))
* **deps:** update golang.org/x/exp digest to 9b4947d ([2105a4c](https://git.act3-ace.com/ace/data/tool/commit/2105a4c7831a38a57d2d6d00f62d5bc332b68311))
* **deps:** update golang.org/x/exp digest to e7e105d ([a832417](https://git.act3-ace.com/ace/data/tool/commit/a8324174e684b459ed69407848d846f3c600a2ba))
* **deps:** update k8s.io/utils digest to 702e33f ([c70390f](https://git.act3-ace.com/ace/data/tool/commit/c70390fd92c3e6fec1d26f704154a9ac70dd41ab))
* **deps:** update k8s.io/utils digest to f90d014 ([1dfa180](https://git.act3-ace.com/ace/data/tool/commit/1dfa1804848c56ff6ea73bf959213869b479431b))
* **deps:** update module git.act3-ace.com/ace/data/telemetry to v1 ([7b09ddc](https://git.act3-ace.com/ace/data/tool/commit/7b09ddc0c175136fe0bc9b2433a52cb02663b89c))
* **deps:** update module git.act3-ace.com/ace/data/telemetry to v1.0.1 ([cdde48c](https://git.act3-ace.com/ace/data/tool/commit/cdde48c31e982aa8ca538fa403beae581cbb9e6b))
* **deps:** update module github.com/google/go-containerregistry to v0.20.2 ([54282fa](https://git.act3-ace.com/ace/data/tool/commit/54282fab2178a3f06341541768ee12cd7e9b3e11))
* **deps:** update module github.com/masterminds/sprig/v3 to v3.3.0 ([7f332ac](https://git.act3-ace.com/ace/data/tool/commit/7f332ac9b6dd825a6568ce1fc007a3093d65c25c))
* **deps:** update module github.com/notaryproject/notation-core-go to v1.1.0 ([46b454d](https://git.act3-ace.com/ace/data/tool/commit/46b454d23f67483f5b7690ab3900b9be128ade0f))
* **deps:** update module github.com/notaryproject/notation-go to v1.2.0 ([dd4de3a](https://git.act3-ace.com/ace/data/tool/commit/dd4de3a02ad6c01ecb547218767c8e857e35b108))
* **deps:** update module github.com/notaryproject/notation-go to v1.2.1 ([e5a7cae](https://git.act3-ace.com/ace/data/tool/commit/e5a7cae444786f0b993765cf72b50c6e379494a2))
* **deps:** update module go.etcd.io/bbolt to v1.3.11 ([ed92fc9](https://git.act3-ace.com/ace/data/tool/commit/ed92fc977868ae6c009c093c9fecba6108bf8048))
* **deps:** update module golang.org/x/net to v0.28.0 ([e9c2260](https://git.act3-ace.com/ace/data/tool/commit/e9c22605a6917407b2c584964d97592b39262315))
* **deps:** update module golang.org/x/sync to v0.8.0 ([d33e4d4](https://git.act3-ace.com/ace/data/tool/commit/d33e4d47ad15c67800d9a5b5f71cf6ccfbf26975))
* **deps:** update module golang.org/x/term to v0.23.0 ([3f8910d](https://git.act3-ace.com/ace/data/tool/commit/3f8910d5b9b4bb600f671451ab1d0ca5ddf16104))
* **deps:** update module golang.org/x/text to v0.17.0 ([dde95de](https://git.act3-ace.com/ace/data/tool/commit/dde95de826bbc0f847b23982f2ea685f9569b254))
* **deps:** update module k8s.io/apimachinery to v0.31.0 ([6fb1bda](https://git.act3-ace.com/ace/data/tool/commit/6fb1bdabd7db934b4170f53daf7531631d190b4a))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.56 ([5df0fad](https://git.act3-ace.com/ace/data/tool/commit/5df0fad3129ba50949067a1f7646d86ce925b730))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.57 ([cef5d31](https://git.act3-ace.com/ace/data/tool/commit/cef5d3111dd934b786e8be6affcf83c17c703902))
* **make:** bump controller-gen to v0.16.2 ([d19cb09](https://git.act3-ace.com/ace/data/tool/commit/d19cb091cb380b207cb9b177512e4613e5a1cc21))
* **make:** bump crd-ref-docs to v0.1.0 ([99ca461](https://git.act3-ace.com/ace/data/tool/commit/99ca4611ef4f79ff206aa3b8b589b5d9f11ea3e7))
* **make:** bump golangci-lint to v1.60.3 ([40efb2d](https://git.act3-ace.com/ace/data/tool/commit/40efb2dfa4aa91c6ab3bae7809590ac2da1de91d))
* **make:** bump ko to v0.16.0 ([e380519](https://git.act3-ace.com/ace/data/tool/commit/e38051924421de956bb2add9d85aff873fc061b8))
* **security:** scan does not wait for results of all images ([be3f93f](https://git.act3-ace.com/ace/data/tool/commit/be3f93f988d88608308eeb99aeb8274ae94be3b2))


### Features

* **mirror:** Add 'org.opencontainers.image.ref.name' annotation to gather index to support Personal RKE2 image loading ([36f793a](https://git.act3-ace.com/ace/data/tool/commit/36f793a64cc34ef254013513f2857d1afb1f5994))
* **mirror:** Add compression option for serialize ([9f652a5](https://git.act3-ace.com/ace/data/tool/commit/9f652a55f2b839d4ee5cf7bea04f4e16dc5a8f0c))
* **mirror:** make reference/tag requirement optional for ace-dt mirror (un)archive ([f730f61](https://git.act3-ace.com/ace/data/tool/commit/f730f614ef672d2f25c207b8c4c4994d99345561))
* **mirror:** Support RKE2 OCI layout in mirror tarballs ([dda854b](https://git.act3-ace.com/ace/data/tool/commit/dda854b64557e16fee005bc98637040399656061))

## [1.12.1](https://git.act3-ace.com/ace/data/tool/compare/v1.12.0...v1.12.1) (2024-08-01)


### Bug Fixes

* bump golangci-lint version in Makefile to 1.59.0 ([b466abb](https://git.act3-ace.com/ace/data/tool/commit/b466abb4a25193ecfb2619bc85d3b33af6dbdb52))
* **config:** add missing registries field in example configuration file ([6e0dbff](https://git.act3-ace.com/ace/data/tool/commit/6e0dbff2cc347b25acc7e6ccc3d2c53e51e4e369))
* **deps:** bump golangci-lint to 1.59.1 in makefile ([c86d978](https://git.act3-ace.com/ace/data/tool/commit/c86d97843e93fcad59ae3b42326d82c4c820ee3b))
* **deps:** bump pipeline to v19.0.22 ([ab0d0cf](https://git.act3-ace.com/ace/data/tool/commit/ab0d0cfd5609e86310f065fab5a9082e6bc4aa4e))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.18 ([3ca0fa4](https://git.act3-ace.com/ace/data/tool/commit/3ca0fa4ac459892de56f1e54d048d334497f0253))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.20 ([bd756ab](https://git.act3-ace.com/ace/data/tool/commit/bd756ab0e34740a7860b27db01334e6f3a477289))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.23 ([968e71e](https://git.act3-ace.com/ace/data/tool/commit/968e71e26aa08ce00ab87debbb7a8830c83ac7b1))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.24 ([97ef257](https://git.act3-ace.com/ace/data/tool/commit/97ef2576efa21907866cabfc34b49613fb89174e))
* **deps:** update golang.org/x/exp digest to 46b0784 ([d9892da](https://git.act3-ace.com/ace/data/tool/commit/d9892da2a41a944182165ecc4be1d5f4aa43bb91))
* **deps:** update golang.org/x/exp digest to 7f521ea ([f442b97](https://git.act3-ace.com/ace/data/tool/commit/f442b97072433537c633246244e6809b22460eb4))
* **deps:** update golang.org/x/exp digest to 8a7402a ([7f705e0](https://git.act3-ace.com/ace/data/tool/commit/7f705e0d1276c93b79bedef4531b9e72cded86f1))
* **deps:** update golang.org/x/exp digest to fc45aab ([da5c077](https://git.act3-ace.com/ace/data/tool/commit/da5c0770b6a997e25927a4c7571566d0bba352b9))
* **deps:** update k8s.io/utils digest to 18e509b ([b70054c](https://git.act3-ace.com/ace/data/tool/commit/b70054c887201f32d31841d39896b7b0d029fb76))
* **deps:** update module git.act3-ace.com/ace/data/telemetry to v0.20.1 ([e112bb8](https://git.act3-ace.com/ace/data/tool/commit/e112bb828b4967f09f13386fd31336a539e20e82))
* **deps:** update module github.com/adrg/xdg to v0.5.0 ([1ca8b8d](https://git.act3-ace.com/ace/data/tool/commit/1ca8b8d90e95143c2b95dfce2a63cf4f1a08db1c))
* **deps:** update module github.com/go-chi/chi/v5 to v5.0.13 ([445f78f](https://git.act3-ace.com/ace/data/tool/commit/445f78f02092529fe08d708bb62b0fcaced53ab1))
* **deps:** update module github.com/go-chi/chi/v5 to v5.0.14 ([b4751c0](https://git.act3-ace.com/ace/data/tool/commit/b4751c0de15069521f4e517230d644d74a8ca615))
* **deps:** update module github.com/go-chi/chi/v5 to v5.1.0 ([9b7b225](https://git.act3-ace.com/ace/data/tool/commit/9b7b225df6306c7771824a3e258187c186ec2bad))
* **deps:** update module github.com/google/go-containerregistry to v0.19.2 ([6b25fc0](https://git.act3-ace.com/ace/data/tool/commit/6b25fc0014cacf93dd41a98e4582bf849c1b5473))
* **deps:** update module github.com/google/go-containerregistry to v0.20.0 ([e97fdf6](https://git.act3-ace.com/ace/data/tool/commit/e97fdf6b2a7412c9a7d1fa7d9b76be789d16cccb))
* **deps:** update module github.com/google/go-containerregistry to v0.20.1 ([880e289](https://git.act3-ace.com/ace/data/tool/commit/880e289027b43940339f80322495cc16ff91f85a))
* **deps:** update module github.com/hashicorp/go-version to v1.7.0 ([1411268](https://git.act3-ace.com/ace/data/tool/commit/1411268976259ece5906fd44545ec962d3a76ca8))
* **deps:** update module github.com/klauspost/compress to v1.17.9 ([dd8e7ed](https://git.act3-ace.com/ace/data/tool/commit/dd8e7ed997b37171edffa9b2027dbfa5689f8bfc))
* **deps:** update module github.com/notaryproject/notation-go to v1.1.1 ([4f17936](https://git.act3-ace.com/ace/data/tool/commit/4f179364b9b0d894c9d954557da7af2352a8c539))
* **deps:** update module github.com/spf13/cobra to v1.8.1 ([1059866](https://git.act3-ace.com/ace/data/tool/commit/1059866e8df0e228b7799125e6a534bc8779660c))
* **deps:** update module golang.org/x/net to v0.26.0 ([6dfc159](https://git.act3-ace.com/ace/data/tool/commit/6dfc159eb04cbf5bb5d7c43c6cad3f1766b99c51))
* **deps:** update module golang.org/x/net to v0.27.0 ([80fcb02](https://git.act3-ace.com/ace/data/tool/commit/80fcb0205c4c8ba63d5d9e2ee79d8190344c4a19))
* **deps:** update module golang.org/x/text to v0.16.0 ([edf6d76](https://git.act3-ace.com/ace/data/tool/commit/edf6d762955be621fc1500a8ddd5fea176db9ab8))
* **deps:** update module k8s.io/apimachinery to v0.30.2 ([34c52f5](https://git.act3-ace.com/ace/data/tool/commit/34c52f54819af9e77634c6cb152fd914d40472ce))
* **deps:** update module k8s.io/apimachinery to v0.30.3 ([093d72e](https://git.act3-ace.com/ace/data/tool/commit/093d72e9484dcf8bb20ea32628373e0d5c47321d))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.50 ([29864d1](https://git.act3-ace.com/ace/data/tool/commit/29864d1071c337527cea2da785d243cd1e692d66))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.52 ([fd081ca](https://git.act3-ace.com/ace/data/tool/commit/fd081caf400b1f040826bf9c58a3a7a62521f860))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.53 ([7598a20](https://git.act3-ace.com/ace/data/tool/commit/7598a2014f094e52dfdf8e0b154dc638144f38bc))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.54 ([16fa266](https://git.act3-ace.com/ace/data/tool/commit/16fa266152b6acb5a3ec5a1b92384ed84fab7728))
* **docs:** add brief telemetry description to bottle creator guide ([2c2d2e7](https://git.act3-ace.com/ace/data/tool/commit/2c2d2e7799cb02279780b9651e6fc65bcbb5ff00))
* **docs:** update mirror.md with proper annotation ([a88f619](https://git.act3-ace.com/ace/data/tool/commit/a88f61980c6467486b2f5900cc18a84cd1283952))
* **git:** support context canceling ([3871988](https://git.act3-ace.com/ace/data/tool/commit/3871988aed4b362232742ba595421405465fbf61))
* handling of TLS certificates when configuring registries ([7703a79](https://git.act3-ace.com/ace/data/tool/commit/7703a7903f7c3bb1b816a395241f33009cb7af60))
* HTTP Retry Copy Failure ([597b14c](https://git.act3-ace.com/ace/data/tool/commit/597b14c3b85e9bcae33755147d49d9dbab3b4a2d))
* **mirror:** Deserialize/Unarchive reports missing blobs due to mismatched mediatypes ([e878bb8](https://git.act3-ace.com/ace/data/tool/commit/e878bb8e0189897b656d64f57b9bc8fc0f7366e6))
* registry cache initialization skipped if config file is invalid ([b9e68f9](https://git.act3-ace.com/ace/data/tool/commit/b9e68f9fb9d048fee084b06bfffb82ee15eb3bac))
* remove oras hack docs ([d2084db](https://git.act3-ace.com/ace/data/tool/commit/d2084dbde87165abf6f1b39b1796f51567cea351))
* **tree:** remove the arrows ([1b20310](https://git.act3-ace.com/ace/data/tool/commit/1b20310e586bd6b45b9fe92082e3dc3132d4bd10))

# [1.12.0](https://git.act3-ace.com/ace/data/tool/compare/v1.11.0...v1.12.0) (2024-05-24)


### Bug Fixes

* add command grouping for the main command ([b007012](https://git.act3-ace.com/ace/data/tool/commit/b007012b5c09c24a9c29af5493bd6288338de8f9))
* add context to Prune() ([80ba2d6](https://git.act3-ace.com/ace/data/tool/commit/80ba2d6e82c75703b87d181510f6de860ddb59b9))
* added context linters ([e0321a5](https://git.act3-ace.com/ace/data/tool/commit/e0321a5c18c1750dbc1bdfed63219497f326823f))
* added gorelease ([e17cc27](https://git.act3-ace.com/ace/data/tool/commit/e17cc27a0f4c17932fb8e9534006300206f73758))
* added renovate for private repos ([52bbb31](https://git.act3-ace.com/ace/data/tool/commit/52bbb31a08caabe9be72271146ad8a7b5e0121f2))
* allow gorelease update ([0985f3a](https://git.act3-ace.com/ace/data/tool/commit/0985f3a767e81e8b0b4955fac68f5efd33bda2ce))
* allow minor breaking changes ([c484bb3](https://git.act3-ace.com/ace/data/tool/commit/c484bb3fd57391cfe847a02a758a3919a79b319d))
* allowing breaking changes to obsolete logging methods ([53c3e00](https://git.act3-ace.com/ace/data/tool/commit/53c3e005cc633712a5e0e85855fd9f893d00a583))
* call Fetch when we ultimately want the data ([691de3e](https://git.act3-ace.com/ace/data/tool/commit/691de3ecca5440a0005667ff6b071ec4e55335b9))
* **ci:** bump CI pipeline ([76bf11e](https://git.act3-ace.com/ace/data/tool/commit/76bf11eb4e2773b54621b2781ff5be2bea351117))
* **ci:** bump pipeline ([17e29b3](https://git.act3-ace.com/ace/data/tool/commit/17e29b34dadde45609cf588579f06d52dfb25da5))
* **ci:** bump pipeline ([b9c489e](https://git.act3-ace.com/ace/data/tool/commit/b9c489ebe6f52d0514b006eb8c4ea995de471b61))
* **ci:** bump pipeline ([8ac0ba5](https://git.act3-ace.com/ace/data/tool/commit/8ac0ba520a6ef772a22ab5150ea940385b73a528))
* **ci:** bump pipeline again to fix govulncheck ([e84978b](https://git.act3-ace.com/ace/data/tool/commit/e84978b86d1d9cca04143d3d577fcc64ffec3986))
* **ci:** bump pipeline and use default distroless ([f9f6c82](https://git.act3-ace.com/ace/data/tool/commit/f9f6c82227d5d088b5fc4e48b201f815c09ed850))
* **ci:** bump pipeline to v18.0.2 ([ad0310e](https://git.act3-ace.com/ace/data/tool/commit/ad0310ecf6f8d2b808d117c19fff0d735ff317fc))
* **ci:** bump the pipeline again to fix govulncheck ([e96d6e3](https://git.act3-ace.com/ace/data/tool/commit/e96d6e3e08e3f464e109c48c3bae7e2f36ba69cc))
* **ci:** bump the pipeline version ([8612a66](https://git.act3-ace.com/ace/data/tool/commit/8612a667fdcc76c5287cc70599bc0ec509b6842d))
* **ci:** try again ([bf3496b](https://git.act3-ace.com/ace/data/tool/commit/bf3496b7944551220e138badd788b142d9eed18b))
* **ci:** try again with linux/arm64 ([3a3c2b4](https://git.act3-ace.com/ace/data/tool/commit/3a3c2b4a4159cd07057036e932038ac2de5d2445))
* **ci:** update job name ([895ae60](https://git.act3-ace.com/ace/data/tool/commit/895ae60909ff52332aededf13f13efd882146d34))
* **ci:** updated go-container-registry ([1604b34](https://git.act3-ace.com/ace/data/tool/commit/1604b34d158d822736d9e84a8b01d19b92b505cc))
* **ci:** updated to v18 of the pipelines ([092af7c](https://git.act3-ace.com/ace/data/tool/commit/092af7c71b144d389c8fc3f2ff5724ede068a08e))
* crash when checking git-lfs version if git-lfs is not installed ([2971072](https://git.act3-ace.com/ace/data/tool/commit/2971072f69ac5e6d085898286f49dcc57abebb19))
* creation of empty LFS manifests when using --lfs on non-lfs enabled repositories. ([9db6b4e](https://git.act3-ace.com/ace/data/tool/commit/9db6b4e6eefe99e286f9f185d84b6d879f44d1bb))
* **deps:** bump dependencies ([ff40643](https://git.act3-ace.com/ace/data/tool/commit/ff40643089204a40f763f9ad00500ecd183ed230))
* **deps:** bump golang.org/x/net to latest (v0.23.0) ([399e40f](https://git.act3-ace.com/ace/data/tool/commit/399e40ffa1fe89a99a642ccaae55f9371e1cd088))
* **deps:** bump oras to the latest - v2.5.0 ([620445d](https://git.act3-ace.com/ace/data/tool/commit/620445d62eeb2d901677d78262490d17abf56dd5))
* **deps:** bumped versions of tools ([a4ad4fc](https://git.act3-ace.com/ace/data/tool/commit/a4ad4fc530e49af2756f106aae1f6eb29d7da880))
* **deps:** minor updates on all other components ([5cb0c38](https://git.act3-ace.com/ace/data/tool/commit/5cb0c38e9ee58c8d44b2a71717321fa27c5430fc))
* **deps:** temporarily return go toolchain version to 1.22.2 to enable pipeline to progress ([66c6a58](https://git.act3-ace.com/ace/data/tool/commit/66c6a586a9b914520cbc930d54bec69834f5b7b4))
* **deps:** update dependencies ([5b4f8c0](https://git.act3-ace.com/ace/data/tool/commit/5b4f8c0469c73f50f5025409f816dfa16c09208f))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.12 ([e7850c3](https://git.act3-ace.com/ace/data/tool/commit/e7850c39ba0b41e0b2829636f14032f7e38d0b8b))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.16 ([d127636](https://git.act3-ace.com/ace/data/tool/commit/d127636e650168c6992c36e40686aabe78642934))
* **deps:** update dependency devsecops/cicd/pipeline to v19.0.17 ([173ebb9](https://git.act3-ace.com/ace/data/tool/commit/173ebb95b8ef25a13d242ce274699f50d69ab43e))
* **deps:** update dependency go to v1.22.2 ([3bcf2d4](https://git.act3-ace.com/ace/data/tool/commit/3bcf2d40ccf894a9c775b71e7abf003e351c7100))
* **deps:** update docker/cli as well to match ([1481579](https://git.act3-ace.com/ace/data/tool/commit/14815797892eb79ecdafa0ed45adf3a079814735))
* **deps:** update go common to mar13 2024 commit ([c855a24](https://git.act3-ace.com/ace/data/tool/commit/c855a2436161537632eba07d958c292427fef986))
* **deps:** update go dependencies ([c485e76](https://git.act3-ace.com/ace/data/tool/commit/c485e76d40c025ca0d497c3814293e2cf32d63cb))
* **deps:** update golang.org/x/exp digest to fe59bbe ([ff30afd](https://git.act3-ace.com/ace/data/tool/commit/ff30afddb80d761583f6610faf24e7af3e8b3e37))
* **deps:** update golang.org/x/text from v0.14.0 to v0.15.0 ([3d813d0](https://git.act3-ace.com/ace/data/tool/commit/3d813d04bdd26ee117b3b610b3abc28018665866))
* **deps:** update k8s.io/utils digest to 0849a56 ([b9ff8fe](https://git.act3-ace.com/ace/data/tool/commit/b9ff8feeb3686a4147720787f8d847cc10fb0935))
* **deps:** update k8s.io/utils digest to fe8a2dd ([93928ad](https://git.act3-ace.com/ace/data/tool/commit/93928ad8755269eddffc9c4af8f2e0e6d996d057))
* **deps:** update module git.act3-ace.com/ace/data/telemetry to v0.19.2 ([2e6e72b](https://git.act3-ace.com/ace/data/tool/commit/2e6e72bb6360dcd13a5f3bf04217e8dfea7e6528))
* **deps:** update module github.com/fatih/color to v1.17.0 ([0cedc43](https://git.act3-ace.com/ace/data/tool/commit/0cedc43049d573db0e6935b0f1f6d5f82098b5a5))
* **deps:** update module go.etcd.io/bbolt to v1.3.10 ([5af1b49](https://git.act3-ace.com/ace/data/tool/commit/5af1b490faf8b647fe2018fb4a5b516ae18ab2ce))
* **deps:** update module golang.org/x/net to v0.25.0 ([6d4c138](https://git.act3-ace.com/ace/data/tool/commit/6d4c13866cfbdad55dde3b5e979c2614a6420b0a))
* **deps:** update module golang.org/x/term to v0.20.0 ([b7bffdb](https://git.act3-ace.com/ace/data/tool/commit/b7bffdb1211d747e6f2a785e37b5b8591319b706))
* **deps:** update module k8s.io/apimachinery to v0.30.1 ([e9e4090](https://git.act3-ace.com/ace/data/tool/commit/e9e409083de291b2f71e7de428351bc5239beecf))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.45 ([52818f5](https://git.act3-ace.com/ace/data/tool/commit/52818f5e07b77942425e86f84a2ce927b28ef3e3))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.46 ([1a158e7](https://git.act3-ace.com/ace/data/tool/commit/1a158e7c481a8b941caee815fba40efdd691b1f0))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.47 ([cde8b76](https://git.act3-ace.com/ace/data/tool/commit/cde8b765aded247dd56d235c0157f91703223078))
* **deps:** update reg.git.act3-ace.com/devsecops/cicd/images/ci-utils docker tag to v1.0.49 ([f846e99](https://git.act3-ace.com/ace/data/tool/commit/f846e99145ed154a0975162c512922d062053f03))
* **deps:** update telemetry and go-common ([6ad9b03](https://git.act3-ace.com/ace/data/tool/commit/6ad9b03acf599554794bc37e9319de8616b11df4))
* **deps:** update to go 1.22.3 for vulnerability fix ([5cd32c9](https://git.act3-ace.com/ace/data/tool/commit/5cd32c9cf66a6802efa3e62ca0c9bf2c13193546))
* **deps:** update x/exp to digest 9bf2c3d13842 ([a09c0c8](https://git.act3-ace.com/ace/data/tool/commit/a09c0c88ed3fbe7920856fdd0bd658fd22d71265))
* **dep:** update vulnerable dependency ([b80865b](https://git.act3-ace.com/ace/data/tool/commit/b80865b27a5b11a9ba8f91ccee027180dfc334fb))
* disable go display to get the pipeline building ([151d5e1](https://git.act3-ace.com/ace/data/tool/commit/151d5e17c7139905850e3fb56ea0289bc20abb80))
* disable verify (again) ([54a746a](https://git.act3-ace.com/ace/data/tool/commit/54a746a5a127d24bbf78c43812e381c4535d434a))
* **docs:** edits responding to feedback ([0444a3a](https://git.act3-ace.com/ace/data/tool/commit/0444a3a07cdbc96653c1e9c6235de3932cc257af))
* **docs:** fix relative link ([b830596](https://git.act3-ace.com/ace/data/tool/commit/b830596d8fb7a3183ad46e7c46e40c9cc1b080d1))
* **docs:** remove additional resources section in anticipation of forthcoming mkdocs site that will make that section unnecessary on individual pages ([e20852f](https://git.act3-ace.com/ace/data/tool/commit/e20852f0b4e0171c0affde8ec02fcce8e346b526))
* **docs:** resolve TODOs ([83c66ae](https://git.act3-ace.com/ace/data/tool/commit/83c66ae24bc1ace378126bbd527e95d041227942))
* **docs:** update language for code blocks ([6b9c838](https://git.act3-ace.com/ace/data/tool/commit/6b9c83864e908bfb07f91b593aeb65cae59b2153))
* **docs:** update MAINTAINERS file ([cf9631b](https://git.act3-ace.com/ace/data/tool/commit/cf9631be3ebcf0a31f1fb7219bd0c90c7e8c0340))
* **docs:** update registry:2 to be fully qualified ([860e3ab](https://git.act3-ace.com/ace/data/tool/commit/860e3ab2d037614ed7abe0a42f180ff76ea2b4fa))
* **docs:** update the public bottle reference ([60b874e](https://git.act3-ace.com/ace/data/tool/commit/60b874e75b01e4e5f3805b1ea8ee29b3ca9a4634))
* embed reference to missing file ([8c7d30c](https://git.act3-ace.com/ace/data/tool/commit/8c7d30c03f6e68a9acbcd193024831f91b62d448))
* embeded docs ([80d885e](https://git.act3-ace.com/ace/data/tool/commit/80d885e24e1ffba93e5cec069913b2611f64119b))
* expose credential.Store ([9483598](https://git.act3-ace.com/ace/data/tool/commit/94835982372571b79104b1452c6e305dd0aa6b8b))
* from-oci pushing to non-existant destination and general pushing of lfs files ([c226e7f](https://git.act3-ace.com/ace/data/tool/commit/c226e7f1b00612d2502a83f7aeae9fd7901694bd))
* golanci-lint issues (misspellings) ([577196f](https://git.act3-ace.com/ace/data/tool/commit/577196fd95ce6fadae1376240a684d7e9d76defc))
* **go:** upgrade to 1.22.1 ([971da44](https://git.act3-ace.com/ace/data/tool/commit/971da447d9f4be7cf69b04c19b49e09d2e90e676))
* handling of deleted lfs files ([4659e30](https://git.act3-ace.com/ace/data/tool/commit/4659e304a7d2abea2264d70afae2faad8e7ef461))
* ignore testdata dirs for renovate ([9584994](https://git.act3-ace.com/ace/data/tool/commit/9584994a2f6aead9a63b92ac946870b4f78445ad))
* improved the loggingTransport for HTTP ([75f6ded](https://git.act3-ace.com/ace/data/tool/commit/75f6ded17f2d3510c4592e72e88bbb2f51126a42))
* improved the loggingTransport for HTTP ([837c2ee](https://git.act3-ace.com/ace/data/tool/commit/837c2ee161df1880c289f9215b06d37d4cae026c))
* lint issue ([d3b52b2](https://git.act3-ace.com/ace/data/tool/commit/d3b52b2a5032e979c0b370ace46022c9d74bbe4a))
* lint issue ([8f7696d](https://git.act3-ace.com/ace/data/tool/commit/8f7696dba4a0b45f4083ec1b8cb067b43ce75b15))
* lint issues and turn on more linters ([5e00cda](https://git.act3-ace.com/ace/data/tool/commit/5e00cdaf39aba1548dad0bfc61d57c1c570c3f97))
* **lint:** yaml lint issues ([2069627](https://git.act3-ace.com/ace/data/tool/commit/2069627c9389a581998a5e1e8b8dcc354e03aa83))
* make existence cache concurrency safe ([da7d214](https://git.act3-ace.com/ace/data/tool/commit/da7d2141ff209a03109e6833d6a726f3fef8072b))
* **mirror:** added comment support back to sources.list (CSV file) ([025185f](https://git.act3-ace.com/ace/data/tool/commit/025185f94822f8c3a07bbac1f2da6f7af9fa4116))
* **mirror:** cleanup of naming ([adc5348](https://git.act3-ace.com/ace/data/tool/commit/adc53486ba2a30bc7398a246145d850631d868da))
* **mirror:** only set label annotations when we have them ([41b2681](https://git.act3-ace.com/ace/data/tool/commit/41b2681f9c7805d0e07442bf465e877ff76efa02))
* missing import ([b4a2a74](https://git.act3-ace.com/ace/data/tool/commit/b4a2a7476c03d8a89c1cb25fe10ab20cce70ae16))
* more filepath.Join() fixes ([6013335](https://git.act3-ace.com/ace/data/tool/commit/601333521f64c6704ef7eaa3f6c59868a29be9d3))
* moved logging utils to go-common ([d55b5e1](https://git.act3-ace.com/ace/data/tool/commit/d55b5e140fba93149732aad8481452e21e7e1a77))
* **pages:** fix path bug in pages job ([f94d348](https://git.act3-ace.com/ace/data/tool/commit/f94d3489518f32200af560ec8c137f6d46171fdc))
* pass context into logging calls ([144a4dd](https://git.act3-ace.com/ace/data/tool/commit/144a4dd8e248af1560628260d15de5b181eacbf5))
* redact the location URL as well in the logging transport ([178b4ff](https://git.act3-ace.com/ace/data/tool/commit/178b4ffa90de66061d14f5b2bf2f2cae5693468d))
* remove dependence on version.Get() ([71fef3e](https://git.act3-ace.com/ace/data/tool/commit/71fef3e7e1d6acbbd074c503933d601e96e93421))
* remove discard handler since it is not used ([92f6c94](https://git.act3-ace.com/ace/data/tool/commit/92f6c94cee04faead29442bbb7f841d3ce1e4428))
* remove formatting (root.Info instead of root.Infof) from output on telemetry bottle push url notification ([2f9b7b0](https://git.act3-ace.com/ace/data/tool/commit/2f9b7b0300b34a9f881dac63284fdcc3998839eb))
* remove interface indirection and use SlogFrontend structure pointer directly ([9cfbdad](https://git.act3-ace.com/ace/data/tool/commit/9cfbdadda2d1d3c7e63818e5f39c7a56753eda19))
* remove no longer needed gorelease allowed deviation ([c47a733](https://git.act3-ace.com/ace/data/tool/commit/c47a733731d5e59b867c9a5c62c2ab877d27ffb6))
* remove old registry references ([8e26781](https://git.act3-ace.com/ace/data/tool/commit/8e2678109ac348894b32d90cb2ec4db1dc53d1c7))
* remove unreliable-head-requests hack ([ed4c66a](https://git.act3-ace.com/ace/data/tool/commit/ed4c66a38b10daa693719cea03eb064c3bdadcf8))
* remove unused credential add functions ([92fb296](https://git.act3-ace.com/ace/data/tool/commit/92fb29661c0d3dce3d1ff4c7453fe5e162f332b7))
* removed cached storage in favor of being more explicit ([8b66bde](https://git.act3-ace.com/ace/data/tool/commit/8b66bdea0519720d6d436d4c66a0af550239a100))
* removed logr ([8140c90](https://git.act3-ace.com/ace/data/tool/commit/8140c902ddd11deb036c964811c3c8a8108b68d7))
* rename TEST_REGISTRY_OCI to TEST_REGISTRY ([a7df02d](https://git.act3-ace.com/ace/data/tool/commit/a7df02d7b9fe69bf6a9439dd635ce52c28c04a32))
* rendered project blueprints (checkpoint commit made by act3-pt) ([73ad020](https://git.act3-ace.com/ace/data/tool/commit/73ad020118a98502143e31e2503a4e4d83b4c1e0))
* rendered project blueprints (checkpoint commit made by act3-pt) ([a0446db](https://git.act3-ace.com/ace/data/tool/commit/a0446db28b9b6710e045904568c01fcdee5d8f15))
* replace occurances of logr with logging.LogFrontend ([81a0280](https://git.act3-ace.com/ace/data/tool/commit/81a0280efa0af5d98fc5a5ceaa447dfd73333443))
* rest of the lint issues from sloglint ([264f7f7](https://git.act3-ace.com/ace/data/tool/commit/264f7f7be2178b51369db20fab92c30866a8acc1))
* rework log biasing to use handler interface versus slog override ([f07032e](https://git.act3-ace.com/ace/data/tool/commit/f07032ee5327f21f60dfc42f2ee51f16e3b74789))
* second functional test failed because we did not set TEST_REGISTRY_OCI ([ed444ca](https://git.act3-ace.com/ace/data/tool/commit/ed444ca3cac626138d43ccce60544e7d30c5d9ca))
* set the DB location in the telemetry server that we stand up ([99d328f](https://git.act3-ace.com/ace/data/tool/commit/99d328f90ca8b208287dc2154a234e5a033119e9))
* some lint complaints ([2bd1e5b](https://git.act3-ace.com/ace/data/tool/commit/2bd1e5b1e22dbe4b56f142483e429538fc88395d))
* subject printing ([7217104](https://git.act3-ace.com/ace/data/tool/commit/7217104e31f0240ec6bbf1b4ce4ddb828bfef110))
* switch releaserc to yaml ([2bac414](https://git.act3-ace.com/ace/data/tool/commit/2bac41475af7b46bc3f7df12128ec9db5757b341))
* switch to a string for the version passed to cli.NewToolCmd() ([16bb547](https://git.act3-ace.com/ace/data/tool/commit/16bb54771b8abeacb78b4b9795467802bce57021))
* **tap:** move to github tap ([54e9743](https://git.act3-ace.com/ace/data/tool/commit/54e97430d692d26fe8b9331f4f89b0f7667fd695))
* trim the interface down for Pull() ([42fb22e](https://git.act3-ace.com/ace/data/tool/commit/42fb22e5b0aa543e3a4768db9fcc95e69d3e17ec))
* turn gorelease into a warning for now ([85e21d6](https://git.act3-ace.com/ace/data/tool/commit/85e21d6547047b79a9a5c484c573091765f60a92))
* update CRD_REF_DOCS_VERSION ([dbbc9ee](https://git.act3-ace.com/ace/data/tool/commit/dbbc9ee5de52224ebf0524d834dadcb2d37d9d3d))
* update golangci lint ([ec8fc29](https://git.act3-ace.com/ace/data/tool/commit/ec8fc299207797fb0fe657b89be0b538bc0f1e05))
* update verify to allow overrides ([34831d0](https://git.act3-ace.com/ace/data/tool/commit/34831d0d3e4167c4e86f99c246a2287df6977525))
* update verify.sh ([33fe201](https://git.act3-ace.com/ace/data/tool/commit/33fe2011cb4d9d03f27f9dd972891feb62496635))
* use concat ([9f574a2](https://git.act3-ace.com/ace/data/tool/commit/9f574a25c3e437a0103b74e317afb99c8c21e8e0))
* use correct notary signature mediatype for telemetry signature updates, and update telemetry dependency version ([3bdf555](https://git.act3-ace.com/ace/data/tool/commit/3bdf5550b723d58192abbb388d499f52bd8d7ec5))
* works with artifactory now ([84de3d7](https://git.act3-ace.com/ace/data/tool/commit/84de3d7ae8970faef65d54c915d73f69611ac307))


### Features

* add bottle media type to ArtifactType field in bottle manifests ([0785bb0](https://git.act3-ace.com/ace/data/tool/commit/0785bb081f9d3413182159b19f9fa053f816206c))
* **oci-tree:** implemented --show-blobs ([e43f704](https://git.act3-ace.com/ace/data/tool/commit/e43f704387f4da47cdf7ae100ad5106cdbb22426))
* replace logr with extended slog ([cfa6c58](https://git.act3-ace.com/ace/data/tool/commit/cfa6c5899f893ddf6175082b375eac49f4b9f6c3))
* Support entire repository handling in git feature ([e8bfcb2](https://git.act3-ace.com/ace/data/tool/commit/e8bfcb25b25163ce176c2efdefe367420ea7c886))

# [1.11.0](https://git.act3-ace.com/ace/data/tool/compare/v1.10.0...v1.11.0) (2024-1-23)


### Bug Fixes

* **deps:** updated oras back to the main branch ([01450d0](https://git.act3-ace.com/ace/data/tool/commit/01450d0b8b82c30d6d96d00fd29eb0972a4bb3eb))
* **docs:** fix lint ([4ccf7a2](https://git.act3-ace.com/ace/data/tool/commit/4ccf7a2517acab1482bbb604667935f3ce363665))
* finished up the login action ([d879ae1](https://git.act3-ace.com/ace/data/tool/commit/d879ae1e591f367515d3384abc626e725a72b4f7))
* lint issue ([d19fb95](https://git.act3-ace.com/ace/data/tool/commit/d19fb952def213f253db3e86c24805aa8381d094))
* parallelize more of "pypi to-oci" ([dc22286](https://git.act3-ace.com/ace/data/tool/commit/dc2228614a3ed1d94ba1800462e69ca6e8ce3987))
* **pypi:** added os and arch derived labels ([a264346](https://git.act3-ace.com/ace/data/tool/commit/a2643467cecb582c3a16a7269bfcda54332bf936))
* **pypi:** bail out early if an error occurs ([d3de749](https://git.act3-ace.com/ace/data/tool/commit/d3de749cff26e3673c0485a7c222392741791a95))
* **pypi:** better handling of duplicate distribution files ([ed583ef](https://git.act3-ace.com/ace/data/tool/commit/ed583efe4bd39cacdc641a1290e9e85acbc8d753))
* **pypi:** fix string formatting of python requirements ([0b88c82](https://git.act3-ace.com/ace/data/tool/commit/0b88c8210cf19ebee65cf32f207ba3f247ef7bef))
* restore blobinfocache update calls in transfer code ([cc0d639](https://git.act3-ace.com/ace/data/tool/commit/cc0d639ba21cc534df830fb1ad9f37ccdd0e7ce7))
* switch login/logout to oras ([6bba37b](https://git.act3-ace.com/ace/data/tool/commit/6bba37b919facf78b4a1f3caa8ddce3f5a89014d))


### Features

* add repository caching to ace-dt git ([a343d26](https://git.act3-ace.com/ace/data/tool/commit/a343d266b21de0de183799b25929aecb48c77171))
* added a "ace-dt pyi labels" command to inspect labels and test selectors ([4960a6b](https://git.act3-ace.com/ace/data/tool/commit/4960a6b1683dbc98126d7e5b0e8ea0d553176510))
* **pypi:** get auth credentials from the docker credential store ([b57ba2b](https://git.act3-ace.com/ace/data/tool/commit/b57ba2be5c7bcdadb76b8f65640acd5612bed69c)), closes [#430](https://git.act3-ace.com/ace/data/tool/issues/430)

# [1.10.0](https://git.act3-ace.com/ace/data/tool/compare/v1.9.0...v1.10.0) (2024-1-9)


### Bug Fixes

* added a "--failed-requirements" flag to "ace-dt pypi to-oci" ([14a7da8](https://git.act3-ace.com/ace/data/tool/commit/14a7da8615e8ffb78166fa9e12eebfdcb25672b7)), closes [#443](https://git.act3-ace.com/ace/data/tool/issues/443)
* avoid uploading unchanged manifests in "ace-dt pypi to-oci" ([eb2b9cc](https://git.act3-ace.com/ace/data/tool/commit/eb2b9cce8bc4682a33768bd89acf9a1d4143f071)), closes [#443](https://git.act3-ace.com/ace/data/tool/issues/443)
* **deps:** update go-common ([8887fdf](https://git.act3-ace.com/ace/data/tool/commit/8887fdfe402508df0241a1a90f54f393a9bec841))
* **deps:** update go-common ([1240d34](https://git.act3-ace.com/ace/data/tool/commit/1240d34844353a610eea10cdf53e280cd9bfc9a8))
* **deps:** update oras-go ([5fd77ef](https://git.act3-ace.com/ace/data/tool/commit/5fd77ef41d1873963ba0f65264f325c2d7afc503))
* **deps:** update oras-go and use there auth cache ([b197581](https://git.act3-ace.com/ace/data/tool/commit/b197581881a71d8d6ff41c5d8374c0a2a7a3585c))
* **deps:** update pipeline to v17 ([f60c14a](https://git.act3-ace.com/ace/data/tool/commit/f60c14a1f56e0b7f775a91151bcec732c9de2e15))
* **docs:** updated docs ([0eab2f8](https://git.act3-ace.com/ace/data/tool/commit/0eab2f8ce3f5f63c60b4013ca1d4acb4ec78840a))
* **lint:** ignore some lint warnings ([5cfd5ba](https://git.act3-ace.com/ace/data/tool/commit/5cfd5ba1ee1eab82e9423f83071762b9a90997a4))
* **lint:** lint errors fixed ([16ff860](https://git.act3-ace.com/ace/data/tool/commit/16ff860669ba95cc0870840c4c38a7ded215eb3c))
* minor cleanup to pypi ([5c86639](https://git.act3-ace.com/ace/data/tool/commit/5c8663968de26ffc1c206a721f5823ae1bb16eea))
* switch to the new way of mounting from ORAS ([4011527](https://git.act3-ace.com/ace/data/tool/commit/4011527bf9595d82154737e76c6efec2ece5a784))


### Features

* added concurrency to pypi to-oci ([9aa638a](https://git.act3-ace.com/ace/data/tool/commit/9aa638a7b4f7afd9562732f90c946f73410358f6)), closes [#443](https://git.act3-ace.com/ace/data/tool/issues/443)

# [1.9.0](https://git.act3-ace.com/ace/data/tool/compare/v1.8.1...v1.9.0) (2023-12-21)


### Bug Fixes

* fix errored path in embedded schemas ([adbc9cd](https://git.act3-ace.com/ace/data/tool/commit/adbc9cd0a10f115d5009dc7d785f70ac489e2b3a))
* lint issues ([f9ad2b2](https://git.act3-ace.com/ace/data/tool/commit/f9ad2b2496a71122749b56b036c8ea3e4dcd48c2))


### Features

* new documentation features for generating static documentation and viewing docs in your terminal ([62bd83a](https://git.act3-ace.com/ace/data/tool/commit/62bd83acb2f28a5f7c89b1571cd3e14aced0c3f7))

## [1.8.1](https://git.act3-ace.com/ace/data/tool/compare/v1.8.0...v1.8.1) (2023-12-21)


### Bug Fixes

* add documentation for "ace-dt pypi to-pypi" ([985dea7](https://git.act3-ace.com/ace/data/tool/commit/985dea736d1ce00446680231cf4cbe0216480afc))
* **ci:** bump CI pipeline to v16 ([8ea5eb1](https://git.act3-ace.com/ace/data/tool/commit/8ea5eb19c98c611f58c7b67915c4ad65ef5cbebb))
* **deps:** update dependencies ([08e316d](https://git.act3-ace.com/ace/data/tool/commit/08e316d229e2effc6218f1d44592b906c26ce360))
* disable host-based auth cache ([24cfdfb](https://git.act3-ace.com/ace/data/tool/commit/24cfdfb9468bde7597690979b7ddd41ac04a8043))
* filter out tags that look like digests ([49313d3](https://git.act3-ace.com/ace/data/tool/commit/49313d376f83f158d9bf8291a87c872c39e98c00))

# [1.8.0](https://git.act3-ace.com/ace/data/tool/compare/v1.7.1...v1.8.0) (2023-12-04)


### Bug Fixes

* added support for extra packages in pypi ([b0e8960](https://git.act3-ace.com/ace/data/tool/commit/b0e896038ecc8db1cbfee9cf17bcd9a961ce209d))
* **ci:** remove trigger for fips ([7d7f533](https://git.act3-ace.com/ace/data/tool/commit/7d7f5339e9caeadfb8470a354d330d0c2e000fa0))
* improved support for non-compliant registries ([7732b8b](https://git.act3-ace.com/ace/data/tool/commit/7732b8b9794fae840b76e275189c5f91b84bbd7f))
* missing python project should to be an error ([76dc5e7](https://git.act3-ace.com/ace/data/tool/commit/76dc5e71426804355832fabd8a664505457fd894))
* **pypi:** revert back to using Blobs().Resolve() to get the full descriptor of a blob given the digest ([e698e04](https://git.act3-ace.com/ace/data/tool/commit/e698e04a215e839b2844c12006f80a0eda5b9d18))
* store the python metadata as a layer instead of in the config ([57f2fca](https://git.act3-ace.com/ace/data/tool/commit/57f2fca96ff9f107878721a2512c37f8b1b650a7))
* update schemas ([1189900](https://git.act3-ace.com/ace/data/tool/commit/118990075dab7c64e6c34f22f41fc2a0ef3475a3))


### Features

* added basic version support for pypi. ([0f7eab3](https://git.act3-ace.com/ace/data/tool/commit/0f7eab39c759cb2dbe96cbd54fb8605fcbd135b9))

## [1.7.1](https://git.act3-ace.com/ace/data/tool/compare/v1.7.0...v1.7.1) (2023-11-27)


### Bug Fixes

* dry run for to-pypi and add retry client ([d2b757d](https://git.act3-ace.com/ace/data/tool/commit/d2b757dc89a19264be3620e1749d64227b1cd86d))

# [1.7.0](https://git.act3-ace.com/ace/data/tool/compare/v1.6.1...v1.7.0) (2023-11-23)


### Bug Fixes

* added an artifact type to the gathered image ([0ef119c](https://git.act3-ace.com/ace/data/tool/commit/0ef119c83b17e533b0afcf7b4b66902a94ef03b2))
* **deps:** bump oras to the lastest ([05ba71e](https://git.act3-ace.com/ace/data/tool/commit/05ba71ea07164425407d9467599decf8a82bdb91))
* **docs:** comment out GitLab issue by email ([9d125a5](https://git.act3-ace.com/ace/data/tool/commit/9d125a5821acf7422600a75145ff941186c5bd41))
* improved the requirements parser ([5f62569](https://git.act3-ace.com/ace/data/tool/commit/5f625699b75267297ca759791c79bf62e51bc9b6))
* lint ([85b38e9](https://git.act3-ace.com/ace/data/tool/commit/85b38e9a0cdfa2d0c3e62f24b1e686549e4089b0))
* minor cleanup to the git code ([3bc1a78](https://git.act3-ace.com/ace/data/tool/commit/3bc1a78ea750bb1fef89db02380695a00295327a))
* minor nit pick ([aed5e32](https://git.act3-ace.com/ace/data/tool/commit/aed5e325b37dace2dc60224e5bf4c79463745803))


### Features

* added index-url and extra-index-url support in requirements files ([d237a35](https://git.act3-ace.com/ace/data/tool/commit/d237a356a4634f7ff04aa488c943921b793fdcf1))

## [1.6.1](https://git.act3-ace.com/ace/data/tool/compare/v1.6.0...v1.6.1) (2023-11-13)


### Bug Fixes

* all prefix mapping is now tested ([1bcf349](https://git.act3-ace.com/ace/data/tool/commit/1bcf349a4bd2b3e9cf6868c3431e1290c26e6897))
* allow images to be the same and add correct by counting ([65eee77](https://git.act3-ace.com/ace/data/tool/commit/65eee7764c86a109c89544176b1b8f2c1de883d1))
* bump go.mod to to 1.21.4 ([03174bb](https://git.act3-ace.com/ace/data/tool/commit/03174bb1a55cc1c99d481222c4c1b37937840bff))
* **ci:** bump CI pipeline ([8c26c45](https://git.act3-ace.com/ace/data/tool/commit/8c26c4527a158ff47e0bb2388c229d23c8d9bbf8))
* **deps:** bump dependencies ([bf7d282](https://git.act3-ace.com/ace/data/tool/commit/bf7d28253c3da3bb1740f4d5b8449ce2bcd5a839))
* do not write blobs out to archive more than once ([71d074e](https://git.act3-ace.com/ace/data/tool/commit/71d074e297145b7ac516e3821e6b92014e08e521))
* gather was not tagging properly ([42d6516](https://git.act3-ace.com/ace/data/tool/commit/42d6516d6373a13590daa874e850a0ca63d78734))
* manifests are now always included in the tar archive ([2405dc0](https://git.act3-ace.com/ace/data/tool/commit/2405dc0489b453252b44c5e178f95c9cbf186a24))
* removed "Buffer Fill" it is not easy to disable ([ae8658e](https://git.act3-ace.com/ace/data/tool/commit/ae8658e50f7f2c34316dae28cd183b744b8791a9))
* when processing existing images use all successors ([97e0126](https://git.act3-ace.com/ace/data/tool/commit/97e0126678f96074e6a9a09ed04c96882f0aedc2))

# [1.6.0](https://git.act3-ace.com/ace/data/tool/compare/v1.5.6...v1.6.0) (2023-11-06)


### Bug Fixes

* allow empty lines in image list provided to gather ([dac5169](https://git.act3-ace.com/ace/data/tool/commit/dac5169638b4d4cecac5a9ab5399c17789911a88))
* bump ORAS fork ([793af33](https://git.act3-ace.com/ace/data/tool/commit/793af33c11ce7fc2d044157a9a1afe3166145954))
* byte formatting ([114526b](https://git.act3-ace.com/ace/data/tool/commit/114526b2c961a538090c4a0c5dc650c771eb22bc))
* byte reporting was reporting a pointer address instead ([d1d538a](https://git.act3-ace.com/ace/data/tool/commit/d1d538a29f8fe40968c8f1fb6bb630e7c5c9b5b0))
* **ci:** bump pipeline ([8e916f2](https://git.act3-ace.com/ace/data/tool/commit/8e916f2fcc20e451667ffe9a4812f76a6dec3802))
* clarify instruction to run ACT3 Login script from local workstation ([9a3cd83](https://git.act3-ace.com/ace/data/tool/commit/9a3cd83d661a600fad9fe0259f41f9edea6c37da))
* Copy should not tag. ([1cd05ab](https://git.act3-ace.com/ace/data/tool/commit/1cd05ab41308aaba1432cdeacc9b194ccd413b96))
* **deps:** update bottle schema module ([d9f4ac4](https://git.act3-ace.com/ace/data/tool/commit/d9f4ac4bd30764745b62165729b902bd3b91c13e))
* **docs:** add default lang to ace-dt in ace hub ([8e08a20](https://git.act3-ace.com/ace/data/tool/commit/8e08a2091085a67f6dbbed089c3823c2e62f19d7))
* **docs:** link login repo;adjust lang for clarity ([f1ff74b](https://git.act3-ace.com/ace/data/tool/commit/f1ff74b7af45c9eace4d11515012d71f76f7005d))
* **docs:** replace link ([1ba2afe](https://git.act3-ace.com/ace/data/tool/commit/1ba2afe580445dec79ed733a01b9fe044d49cf2f))
* **docs:** standardize macOS capitalization ([8044859](https://git.act3-ace.com/ace/data/tool/commit/804485962ae60d22baf96202dc24490b2b3c24c3))
* **docs:** update QSG language edits ([9c098f5](https://git.act3-ace.com/ace/data/tool/commit/9c098f57bc37ae245081e2853900d5da8982685f))
* eliminate repetition ([7723ac7](https://git.act3-ace.com/ace/data/tool/commit/7723ac7d0092d2b31f07fabbdf0c57e1e5ae30fb))
* inject the slog and logr into the context ([d6e471c](https://git.act3-ace.com/ace/data/tool/commit/d6e471c88cb2a2280ee71783c9d9a8da1be1dc83))
* **log:**  getting this to build by using adaptors ([a2b15ba](https://git.act3-ace.com/ace/data/tool/commit/a2b15ba78eef629c3dee4b80e96f704107240cfc))
* minor changes to formatting of an error ([ed9d1b6](https://git.act3-ace.com/ace/data/tool/commit/ed9d1b61ec0f037fbb7cdd3b784b6b4d51bb25c0))
* minor tree formatting nitpick ([8b8de89](https://git.act3-ace.com/ace/data/tool/commit/8b8de8922c145477aa47e8ac4ca96351f3d2ee6c))
* referrers API now works with registry.k8s.io ([291a385](https://git.act3-ace.com/ace/data/tool/commit/291a3857cd6181549690e81c9ef0a4f12c706d96))
* **schema validators:** add config file validation in IDEs by generating JSON Schema ([3d17530](https://git.act3-ace.com/ace/data/tool/commit/3d175303a8a0ef7da1e1b1e2ffc24e2dfd13408f))
* support subjects as successors of images ([91ade8a](https://git.act3-ace.com/ace/data/tool/commit/91ade8aa1390ffcfd89cd05bf05fb4c7d2e40f3a))


### Features

* added a hack to handle unreliable HEAD requestsion ([66af668](https://git.act3-ace.com/ace/data/tool/commit/66af668da1bf07887f2ccef8f9de29a654ab0e26))
* added an optional logging transport for ORAS HTTP requests ([c6beaf9](https://git.act3-ace.com/ace/data/tool/commit/c6beaf9ad812d855a150d7c3f425b0f8a8f780dd))
* during deserialize, only push blob if it does not exist ([f8c7902](https://git.act3-ace.com/ace/data/tool/commit/f8c79027b4c194d4bdbc5a62c7139f7b327e22db))

## [1.5.6](https://git.act3-ace.com/ace/data/tool/compare/v1.5.5...v1.5.6) (2023-10-30)


### Bug Fixes

* **pypi:** metadata broke the actual file downloads ([7633168](https://git.act3-ace.com/ace/data/tool/commit/7633168ecaaf7e0462680c5facaf244cffbb9395)), closes [#417](https://git.act3-ace.com/ace/data/tool/issues/417)

## [1.5.5](https://git.act3-ace.com/ace/data/tool/compare/v1.5.4...v1.5.5) (2023-10-17)


### Bug Fixes

* bump pipeline to v15.0.6 ([ce77f5a](https://git.act3-ace.com/ace/data/tool/commit/ce77f5a9080dcb2cf82a397125dc09928f9b2a46))

## [1.5.4](https://git.act3-ace.com/ace/data/tool/compare/v1.5.3...v1.5.4) (2023-10-17)


### Bug Fixes

* pushdir layer tarball structure ([71a8afb](https://git.act3-ace.com/ace/data/tool/commit/71a8afb8ca3bad1e1ea25eed443bf58c9ee8b7ec))

## [1.5.3](https://git.act3-ace.com/ace/data/tool/compare/v1.5.2...v1.5.3) (2023-10-17)


### Bug Fixes

* minor update to pushdir ([5a2c015](https://git.act3-ace.com/ace/data/tool/commit/5a2c015edf3ad55649389be4749f9ac06e889357))

## [1.5.2](https://git.act3-ace.com/ace/data/tool/compare/v1.5.1...v1.5.2) (2023-10-17)


### Bug Fixes

* cleanup gather messages ([325a2f1](https://git.act3-ace.com/ace/data/tool/commit/325a2f1c265ec98655c3ea920dabdcddfeb947af))
* status updates with the simple UI ([340554e](https://git.act3-ace.com/ace/data/tool/commit/340554e698ed6855a738631b88cf5965ca1fc174))

## [1.5.1](https://git.act3-ace.com/ace/data/tool/compare/v1.5.0...v1.5.1) (2023-10-16)


### Bug Fixes

* disable referrers by default ([3c1fa8d](https://git.act3-ace.com/ace/data/tool/commit/3c1fa8dacea3c59d8eea212f40b32282956c61e1))

# [1.5.0](https://git.act3-ace.com/ace/data/tool/compare/v1.4.0...v1.5.0) (2023-10-16)


### Bug Fixes

* added lint to tests ([5757906](https://git.act3-ace.com/ace/data/tool/commit/5757906cd91fbc7cfedf4653d9a8a344c7c7b992))
* annotation names and added serialization format ([3958059](https://git.act3-ace.com/ace/data/tool/commit/395805945a87f9f44d94a655f362cf9d2538dfb3))
* bump pipeline to v15.0.3 ([16f09aa](https://git.act3-ace.com/ace/data/tool/commit/16f09aae3841b621c8f38404ae3689a37531baeb))
* disable project tool job ([278a578](https://git.act3-ace.com/ace/data/tool/commit/278a578948da779205c53f109d6ffb65e2e7c805))
* **docs:** move go troubleshooting step into go install section ([2ab92b5](https://git.act3-ace.com/ace/data/tool/commit/2ab92b56e7d77b2f77eed02ce42e3e0a3f4bacfb))
* filesystem comparison in tests ([6055aa3](https://git.act3-ace.com/ace/data/tool/commit/6055aa35c2e4274af65091952515106ec0c7b561))
* lint ([c99d76d](https://git.act3-ace.com/ace/data/tool/commit/c99d76d11a8b86e146de76d42cfc6ccc1639112d))
* lint errors in tests ([69b2efc](https://git.act3-ace.com/ace/data/tool/commit/69b2efc1fe7a9b294f31cc173206cfabcb61420d))
* lint issues in test ([66f127a](https://git.act3-ace.com/ace/data/tool/commit/66f127aa3cc35cc826392b15d93f1d91227f596c))
* more lint issues ([2f5ba4c](https://git.act3-ace.com/ace/data/tool/commit/2f5ba4cd1c14b78e7063064aa3fc22f6f36df494))
* only include extra-manifests descriptor if we have any to include ([84dfefe](https://git.act3-ace.com/ace/data/tool/commit/84dfefe8332253c0231bf72db93dd3c6491997e8))
* switch back to static binaries ([002ff29](https://git.act3-ace.com/ace/data/tool/commit/002ff29926c7ce51ab0cfd8b57fe5d5582c35301))
* vulnerability in github.com/x/net ([af70f36](https://git.act3-ace.com/ace/data/tool/commit/af70f36c02b4b0fd9f5c577b94976ec358b13074))


### Features

* added predecessor support to serialize ([72f2d57](https://git.act3-ace.com/ace/data/tool/commit/72f2d5736365fab34e9e844866501e44e9585b1c))

# [1.4.0](https://git.act3-ace.com/ace/data/tool/compare/v1.3.8...v1.4.0) (2023-09-28)


### Bug Fixes

* **ci:** bump ci pipeline ([bbb1485](https://git.act3-ace.com/ace/data/tool/commit/bbb148538b1a14a6abb8df156af984cf22f20a5b))
* **deps:** upgraded dependencies ([59cbd90](https://git.act3-ace.com/ace/data/tool/commit/59cbd9097ccd98e45e42abab4db5e10a49ef824c))
* deserialize test compatibility to OCI tree ([e5239d4](https://git.act3-ace.com/ace/data/tool/commit/e5239d41b0da209541ae9ef93a3c937893618e00))
* ignore commented out lines image list for gather ([95a377e](https://git.act3-ace.com/ace/data/tool/commit/95a377eaa77df6c1b7c6de3c14d9f3713f99b8d4))
* lint and compiling errors ([8ea28a7](https://git.act3-ace.com/ace/data/tool/commit/8ea28a7d4eba2d90b51e1cc88479f6785822b489))
* scatter formatting ([b7bc741](https://git.act3-ace.com/ace/data/tool/commit/b7bc74113ad5d12161b48261fa16d89ccff016a5))
* scatter indentation ([27687ec](https://git.act3-ace.com/ace/data/tool/commit/27687ec17ff59f421285c4a778ad158d3145d162))
* task completion in gather ([c13f2a9](https://git.act3-ace.com/ace/data/tool/commit/c13f2a9af877e1fca47b0a7f35c2cf6bf65486f9))


### Features

* added a nest mapper ([7c35463](https://git.act3-ace.com/ace/data/tool/commit/7c354637bbb32445f8f7723004489369b3181133))

## [1.3.8](https://git.act3-ace.com/ace/data/tool/compare/v1.3.7...v1.3.8) (2023-09-12)


### Bug Fixes

* **ci:** bump pipeline to ge to go 1.21.1 ([9fafecb](https://git.act3-ace.com/ace/data/tool/commit/9fafecb59e6432b9700ef48447d0167d2974f18a))
* **ci:** bump the pipeline to v13.1 to support golang 1.21 ([4b37b5a](https://git.act3-ace.com/ace/data/tool/commit/4b37b5a12d6c52d730255a250eed38069c7d7b5f))
* **ci:** update pipeline ([9279d04](https://git.act3-ace.com/ace/data/tool/commit/9279d04c5e2901a6df0703718c5d37704579d7cc))
* **ci:** upgrade pipeline again ([090cedd](https://git.act3-ace.com/ace/data/tool/commit/090cedd250db321e657ec0c45f30fe41638df132))
* depgaurd and updated tools ([06ac41f](https://git.act3-ace.com/ace/data/tool/commit/06ac41f58e322471bd63730d316008c0308a8efd))
* use fixed width formatting for percentages ([0228da1](https://git.act3-ace.com/ace/data/tool/commit/0228da14f7d10eb1641f6394bc40d1ee00f121e7))

## [1.3.7](https://git.act3-ace.com/ace/data/tool/compare/v1.3.6...v1.3.7) (2023-08-09)


### Bug Fixes

* update deps to fix a vulnerability in golang.com/x/net ([8d4154f](https://git.act3-ace.com/ace/data/tool/commit/8d4154f9791908e81505323d25f6546785bcac8f))

## [1.3.6](https://git.act3-ace.com/ace/data/tool/compare/v1.3.5...v1.3.6) (2023-06-26)


### Bug Fixes

* disable cache in serialize ([1f02d46](https://git.act3-ace.com/ace/data/tool/commit/1f02d4693bbfa9b73816b25599683c03cb79db8c)), closes [#380](https://git.act3-ace.com/ace/data/tool/issues/380)

## [1.3.5](https://git.act3-ace.com/ace/data/tool/compare/v1.3.4...v1.3.5) (2023-06-16)


### Bug Fixes

* go-common fix for testing equality of directory FileInfo entries. ([b7ee11c](https://git.act3-ace.com/ace/data/tool/commit/b7ee11c11dae70779fd41f4ca0aa2e597db9fdb7))
* sort pypi packages and distributions when rendering HTML ([3a9b431](https://git.act3-ace.com/ace/data/tool/commit/3a9b4319f04ea6d124cb223f82504e41df6115b9))

## [1.3.4](https://git.act3-ace.com/ace/data/tool/compare/v1.3.3...v1.3.4) (2023-06-14)


### Bug Fixes

* add comparison opts changes from go-common ([93269a4](https://git.act3-ace.com/ace/data/tool/commit/93269a47a85c6893f8885dabeffd1d2458eb1ded))
* bump CI pipeline ([3d15b55](https://git.act3-ace.com/ace/data/tool/commit/3d15b5569ef1946c2fbc6c33df4f00abceabb5ef))
* bump dependencies ([d7e1698](https://git.act3-ace.com/ace/data/tool/commit/d7e16988452ecc12dd6c9ef18a18e764d8a7b84d))
* bump go-common ([ff430d0](https://git.act3-ace.com/ace/data/tool/commit/ff430d08f5c5c4e37097c8540e80697f19f56069))
* bump go-common, implement changes ([2f5816a](https://git.act3-ace.com/ace/data/tool/commit/2f5816a0d4d47e971fc8ab1f48913ea6d20491aa))
* close the loop for cancellation for transfer queues, fixing a goroutine leak ([34802fe](https://git.act3-ace.com/ace/data/tool/commit/34802fed507251cacb553a6020e4669557510763))
* **doc:** update link to cli reference ([68be0a1](https://git.act3-ace.com/ace/data/tool/commit/68be0a1df51dcad584b4bbcf51c29d63fb03a90e))
* merge conflicts from main ([9a9b842](https://git.act3-ace.com/ace/data/tool/commit/9a9b842d1312fbd79a1dae42773cacadcb30ec85))
* mklint issues ([b30fd8b](https://git.act3-ace.com/ace/data/tool/commit/b30fd8b0dddecc012df4f36663a45366e6f51ba3))
* more deps updates ([e4d1f6e](https://git.act3-ace.com/ace/data/tool/commit/e4d1f6e3ffd59ce9c3b88ed379e14160566366b6))
* pypi now properly ignores extras ([4cd5ea2](https://git.act3-ace.com/ace/data/tool/commit/4cd5ea2702ae1b9ca3a2dd21a4c920ee145d3c4f)), closes [#372](https://git.act3-ace.com/ace/data/tool/issues/372)
* relative paths for fsutil ([7dd09c9](https://git.act3-ace.com/ace/data/tool/commit/7dd09c92e3cd9556404d84c0eb1a4ffe251de80c))
* remove queue bucketing in transfer worker, use transfer worker schedule task instead of direct queue enqueue ([b979d5a](https://git.act3-ace.com/ace/data/tool/commit/b979d5a17cc0aead7bbb3cf41dd49297ee2c24b9))
* replace with go-common fsutil ([8974dd7](https://git.act3-ace.com/ace/data/tool/commit/8974dd7676146cbf566b55b9820c71954d6b9ebc))
* resolve merge conflicts with main ([2bbbeaa](https://git.act3-ace.com/ace/data/tool/commit/2bbbeaa85ee9afa36b748b65edec1b2d7f4cdfca))
* set worker for transfer settings clone ([5c658c3](https://git.act3-ace.com/ace/data/tool/commit/5c658c3cc70b87d2ca2088567f35e7ce333ec93e))
* switch to undeprecated API from docker for username and password ([acf283c](https://git.act3-ace.com/ace/data/tool/commit/acf283cba4bf98bdfea7a787e6b3eb1674449b30))
* use source in fsutil ([c430e9b](https://git.act3-ace.com/ace/data/tool/commit/c430e9bea550272d5e0e2ad1cd9326cb254ea2cd))

## [1.3.3](https://git.act3-ace.com/ace/data/tool/compare/v1.3.2...v1.3.3) (2023-05-17)


### Bug Fixes

* **ci:** upgrade pipeline again ([1c7782b](https://git.act3-ace.com/ace/data/tool/commit/1c7782be65c066adb68666c96faa6f2538ce696b))

## [1.3.2](https://git.act3-ace.com/ace/data/tool/compare/v1.3.1...v1.3.2) (2023-05-17)


### Bug Fixes

* **ci:** update CI pipeline ([76321cf](https://git.act3-ace.com/ace/data/tool/commit/76321cfeaba5e581134aa025b2ec4266c6dbeaf0))

## [1.3.1](https://git.act3-ace.com/ace/data/tool/compare/v1.3.0...v1.3.1) (2023-04-17)


### Bug Fixes

* docs in release ([7e55a67](https://git.act3-ace.com/ace/data/tool/commit/7e55a67cf29b2c23afc00078460642d9243a09f7))

# [1.3.0](https://git.act3-ace.com/ace/data/tool/compare/v1.2.0...v1.3.0) (2023-04-17)


### Bug Fixes

* a lint issue in pypi functionality ([af49a34](https://git.act3-ace.com/ace/data/tool/commit/af49a349438dcb33017f7b75f8ba5f25113ae52c))
* add test cases to blockbuf ([acaa707](https://git.act3-ace.com/ace/data/tool/commit/acaa7073e7b2f3ace4c45aa2287bfec0edb99275))
* blockbuf is more fixed ([0b763cd](https://git.act3-ace.com/ace/data/tool/commit/0b763cd891bff3a8a728d1b338f94d34d85d8daa))
* bump CI version so we can use go 1.20 ([b898ed1](https://git.act3-ace.com/ace/data/tool/commit/b898ed161864f48cde37373ca83a6bfa5ec6098a))
* insecure handling in some mirror subcommands ([a1cfdb5](https://git.act3-ace.com/ace/data/tool/commit/a1cfdb593713229144ce1321caabe4245c50141f))
* lint issues ([c55b203](https://git.act3-ace.com/ace/data/tool/commit/c55b2030468ba5891b214073d03dab6a88c20776))
* logging levels ([02e1593](https://git.act3-ace.com/ace/data/tool/commit/02e1593bcb6a9b867a5f777f600f2ceffbeb5cf4))
* moved release.sh work to gitlab-CI as a different job ([337e28d](https://git.act3-ace.com/ace/data/tool/commit/337e28df7333118e621c84278819b7cd48d70845))
* output tweaks ([94a37cb](https://git.act3-ace.com/ace/data/tool/commit/94a37cb2d0eba339753e3686ce7cdbdbdbfa34aa))
* prune needed newline ([af0cc80](https://git.act3-ace.com/ace/data/tool/commit/af0cc804371f72a841b4a8127027c23c1d33a895))
* rewind payload readers during PUT authentication ([7327f5e](https://git.act3-ace.com/ace/data/tool/commit/7327f5e62bca03489012bc29813a552dbbdb8fba))


### Features

* bump pipeline version ([fe3d343](https://git.act3-ace.com/ace/data/tool/commit/fe3d3437420c0b3311aa61f16ada555c1710b46d))

# [1.2.0](https://git.act3-ace.com/ace/data/tool/compare/v1.1.0...v1.2.0) (2023-03-31)


### Bug Fixes

* add telemetry test with push and deprecates ([dfdacd4](https://git.act3-ace.com/ace/data/tool/commit/dfdacd41b775cb2ee9b0a53aab900eb89f59617f))
* add test for no-deprecates flag ([b47820a](https://git.act3-ace.com/ace/data/tool/commit/b47820a7c667109d19d68a93f124c38e6f358360))
* add visualize command for testing UI ([8d29f32](https://git.act3-ace.com/ace/data/tool/commit/8d29f327d1ae44c8b55de5027f0fc3266cb96d74))
* added completion time to non-progress tassk in the UI ([68926e9](https://git.act3-ace.com/ace/data/tool/commit/68926e97643d10670e3ffd7e7aacf5541fe2f86b))
* added test and added insecure flag ([10ed3ad](https://git.act3-ace.com/ace/data/tool/commit/10ed3ad7b2cf45b572d2f228345ac70d6e300489))
* blockbuf was using twice as much memory as necessary ([a6c7120](https://git.act3-ace.com/ace/data/tool/commit/a6c71202586873fe18b2ff916c2200ad4ded4b8c))
* bottle push error ([633c49b](https://git.act3-ace.com/ace/data/tool/commit/633c49bd8ba569d01145309d9bbe0bdb482dd23d))
* change when manifests are added to taskmap ([080c227](https://git.act3-ace.com/ace/data/tool/commit/080c227e30126be615fb9eb3bdf09146b90dfb1a))
* Comment should end in a period (godot) ([06f8248](https://git.act3-ace.com/ace/data/tool/commit/06f8248fdb582d5f4b640c767c36aaaf23f23700))
* copy any old work from ui-redux ([a51949d](https://git.act3-ace.com/ace/data/tool/commit/a51949df5415c8d590ea4c11df7fa7a9a607b8d7))
* correct defer statement ordering ([207475d](https://git.act3-ace.com/ace/data/tool/commit/207475d5f47249a26ee6351bcd67f2a91d1daac8))
* correct pointers and declutter output ([e2cef46](https://git.act3-ace.com/ace/data/tool/commit/e2cef4662ef80f27e5880916d8503f90f4375a1a))
* deprecate pulled bottleID if version migration ([09afe2d](https://git.act3-ace.com/ace/data/tool/commit/09afe2d694ba86e7376732b3b2d374e960c826ee))
* doc and new flag for no-deprecate ([964a2ea](https://git.act3-ace.com/ace/data/tool/commit/964a2eacb111bdff2804fc660b6b876944e5c189))
* errors that need to manage counterMap ([f6eeb04](https://git.act3-ace.com/ace/data/tool/commit/f6eeb04f8ef5613ce7f8c42226b3c86037d61385))
* group renamed to task ([4809de6](https://git.act3-ace.com/ace/data/tool/commit/4809de6d51e40ab6a7be1801062cd33a10781a98))
* implement MR comment requests ([5ef9ef9](https://git.act3-ace.com/ace/data/tool/commit/5ef9ef9b5698244fd374d64e1293126196541193))
* JFrog did not like our accept headers for pushing manifests ([0f3528a](https://git.act3-ace.com/ace/data/tool/commit/0f3528acc676f4c27c69796f74129484caf62d15))
* lint ([6bc270c](https://git.act3-ace.com/ace/data/tool/commit/6bc270c2f91485cfffd5ce0968787cd97ef12cfe))
* lint and removed extra logs ([b329a2e](https://git.act3-ace.com/ace/data/tool/commit/b329a2e736fedfb1da95a6ada4409d9ea3b9a4e2))
* lint errors ([1f35931](https://git.act3-ace.com/ace/data/tool/commit/1f359317530cbf290588ad63f4cfea38d71573ff))
* lock stack ([09bdf02](https://git.act3-ace.com/ace/data/tool/commit/09bdf02dc397197955cd9d97f0eb15202d479490))
* merge in changes from main ([b04ca19](https://git.act3-ace.com/ace/data/tool/commit/b04ca19e4c5695443e8b9316b53d98d5292e6ab9))
* merge in changes with main ([6f73502](https://git.act3-ace.com/ace/data/tool/commit/6f7350260fad24884b90e52e68cfe842c07e1537))
* merge in upstream changes ([4d07bae](https://git.act3-ace.com/ace/data/tool/commit/4d07bae8eb9334f1cee6e1b98140c5dde61f2f3c))
* merge with upstream changes ([2b0df45](https://git.act3-ace.com/ace/data/tool/commit/2b0df459592dd08a0b9e5c954eda4224d3c79a6e))
* merge work, add multicolor progress ([8f43b60](https://git.act3-ace.com/ace/data/tool/commit/8f43b607570dbcaa36f3c6db3ca22e6a66138ace))
* more visual clutter cleanup and pointer checking ([04d231a](https://git.act3-ace.com/ace/data/tool/commit/04d231a77d3fd051a5e716a0572bd1e6da190534))
* move from map to struct ([a58ba94](https://git.act3-ace.com/ace/data/tool/commit/a58ba94034e3f5073ea526c7fb236b4b10a4e8cf))
* move started and completed to logs, remove root logger from printing the tracker ([17088c1](https://git.act3-ace.com/ace/data/tool/commit/17088c139a303d5aa8b6197ccfdd9f60d7e6bba4))
* MR resolutions, output formatting ([fc686a1](https://git.act3-ace.com/ace/data/tool/commit/fc686a144b82aaaa963c4bb1b2b3ee1d0b7e3914))
* only complete rootUI once in runner.go ([8f1cd3b](https://git.act3-ace.com/ace/data/tool/commit/8f1cd3b1b85fef0c8374fbf86403cabf6849e8f0))
* output for deprecate ([2c5d49b](https://git.act3-ace.com/ace/data/tool/commit/2c5d49babc5c169a2f08388dbf0a873521059d36))
* progress needs to also send task update ([9a78c22](https://git.act3-ace.com/ace/data/tool/commit/9a78c22e0fe17b447c24cd8b11289968eab601fc))
* put task pointer in ctx instead of func definition ([1443587](https://git.act3-ace.com/ace/data/tool/commit/1443587465f9926dee9d266d12bc3bd158de0f73))
* re-implement counters for ace-dt commands ([e39818c](https://git.act3-ace.com/ace/data/tool/commit/e39818ca1f2a4d6119628aa5e9c1fd1235958c65))
* re-implement counting ([3b15031](https://git.act3-ace.com/ace/data/tool/commit/3b15031b750bfd0ffcfbaa10d5edef5ef293b712))
* readability and closing more ui elements ([575af52](https://git.act3-ace.com/ace/data/tool/commit/575af52a3179384f9d60bb1bb5206b603d4b2943))
* reduce complexity of commit ([b7b434a](https://git.act3-ace.com/ace/data/tool/commit/b7b434a86115e2a218d0acb357f9bc44cf434296))
* redundant verbage ([bad987e](https://git.act3-ace.com/ace/data/tool/commit/bad987e512af97f21f326727e9cdac1293d168ef))
* refactor legacy progress into byte tracker ([2005080](https://git.act3-ace.com/ace/data/tool/commit/20050807278e546efe492b06810d4576b89a37c8))
* remove shorthand for no-deprecates flag ([0c337a6](https://git.act3-ace.com/ace/data/tool/commit/0c337a6fd2655c8707eb5370dc4f0245a267a1ee))
* scrap old changes, add reference flag for bottle references ([97bd159](https://git.act3-ace.com/ace/data/tool/commit/97bd15942330198e69336976fe199f5976e61279))
* slices are always pass by reference ([641f707](https://git.act3-ace.com/ace/data/tool/commit/641f70716088ac4bce2496b38fb13af95f5c857b))
* start looking at terminal size, color change different progress groups ([2c2894d](https://git.act3-ace.com/ace/data/tool/commit/2c2894de22eb687080b261be27b6a0636a1e48eb))
* start looking into data races ([316395b](https://git.act3-ace.com/ace/data/tool/commit/316395b2211ce1cd22ee53c0ffe5cc5c07631fd6))
* symlink test ([cc4f8cd](https://git.act3-ace.com/ace/data/tool/commit/cc4f8cdb2f0514aa6e8a1c58a11633861ae58289))
* terminal size support ([05db094](https://git.act3-ace.com/ace/data/tool/commit/05db0940d7727ea02ffc8753d2620807d8e38737))
* try ui stack solution to counters ([46be9cc](https://git.act3-ace.com/ace/data/tool/commit/46be9cc2d47078183716e8530dfe13de9f17c04b))
* UI cleanup and improved editing ([2d86538](https://git.act3-ace.com/ace/data/tool/commit/2d86538a36b8b77723f8bf061fc85cd3b93689b7))
* update transfer for new UI / progress ([d9b80c2](https://git.act3-ace.com/ace/data/tool/commit/d9b80c27b2def70da7a06dec522dbf5bab6516c4))
* use layer digest instead of content digest ([73f1434](https://git.act3-ace.com/ace/data/tool/commit/73f1434646d522c542f8cb0533dc0f36cb9b1ef1))


### Features

* add deprecation support for bottle commit command ([31a2b7e](https://git.act3-ace.com/ace/data/tool/commit/31a2b7e1836f93b7bce347f4aafd6f00044313a9))
* added an alpha beta filter based byte tracker ([0bb0d09](https://git.act3-ace.com/ace/data/tool/commit/0bb0d09d62eb921424da63dd4a81006159bff60d))
* Use transport based authentication (roundtripper) instead of retry based ([e88e864](https://git.act3-ace.com/ace/data/tool/commit/e88e86469a6f49eb8a75873c98bdcd140e1c2e8c))

# [1.1.0](https://git.act3-ace.com/ace/data/tool/compare/v1.0.7...v1.1.0) (2023-02-06)


### Bug Fixes

* report unknown size instead of 0 size for uncommitted parts in bottle show ([75ac68f](https://git.act3-ace.com/ace/data/tool/commit/75ac68f7164c280eb7109e6570c850bf309e311b))
* unit tests ([5d7a1e1](https://git.act3-ace.com/ace/data/tool/commit/5d7a1e1aed686c8e35a27eb982a2a0617215a26a))


### Features

* added OCI directory support to tree ([a9215c7](https://git.act3-ace.com/ace/data/tool/commit/a9215c7eec253a2cf69d25f658f74320901bd407))

## [1.0.7](https://git.act3-ace.com/ace/data/tool/compare/v1.0.6...v1.0.7) (2023-01-23)


### Bug Fixes

* added a sample dockerfile ([1afb37b](https://git.act3-ace.com/ace/data/tool/commit/1afb37bf8ecbeb659e87b8859808ced77b546feb))
* added back in the content size when doing a reset ([2604c84](https://git.act3-ace.com/ace/data/tool/commit/2604c84e96d888a87f778aa1aed888e80bcf479c))
* added command grouping for the bottle subcommand ([3176b84](https://git.act3-ace.com/ace/data/tool/commit/3176b845d16c4107898a4408680e2f3d35a2357e))
* bottle pull now preserves bottle ID ([034d633](https://git.act3-ace.com/ace/data/tool/commit/034d633cbb4261f64c52fd162899d04ec80210c2))
* bump schema ([e46c66e](https://git.act3-ace.com/ace/data/tool/commit/e46c66e8ff6a511eb48f6d0a8ed01f0fddbc2933))
* bumped dependencies ([9e138d6](https://git.act3-ace.com/ace/data/tool/commit/9e138d65e2eddb622d5e912cf3de94d8fcc28c2d))
* bumped the ci pipeline ([74e0b10](https://git.act3-ace.com/ace/data/tool/commit/74e0b10777e36f568b91297d043acb5e56fd6c66))
* found a minor bug with trailing slashes ([9f59572](https://git.act3-ace.com/ace/data/tool/commit/9f59572252c149bc4d3e2574fcfcec089950989a))
* leave the shell to handle "~" if it supports it ([4a34f03](https://git.act3-ace.com/ace/data/tool/commit/4a34f03b7ae52ebca1da3c3abedc7e201f09928f))
* no longer require absolute path for "ace-dt bottle source add" ([b6c76f0](https://git.act3-ace.com/ace/data/tool/commit/b6c76f02366e22eb0366ba4533b2351c6c1833c5))
* race condition in qtransfer ([2b9847e](https://git.act3-ace.com/ace/data/tool/commit/2b9847e85fd25d4d338cc30d383bb1babb7caa5c))
* removed one more context ([91c251f](https://git.act3-ace.com/ace/data/tool/commit/91c251ffdac2298ead0a0df3d85caa0f897ed869))
* report error to user if unable to load configuration in config command ([6ed7cf1](https://git.act3-ace.com/ace/data/tool/commit/6ed7cf1d0f5021daaa7a0b63b3b0079bf56e1e74))
* telemetry error handling was incorrect ([e388440](https://git.act3-ace.com/ace/data/tool/commit/e38844090d0ff140f57ed70e5f4cdbc082c7f74e))
* typo in sample config ([804c5a5](https://git.act3-ace.com/ace/data/tool/commit/804c5a56d04e75a8b8efaf0567d4a4f4f1810620))
* unified bottle relative file path handling in artifact commands ([47e137f](https://git.act3-ace.com/ace/data/tool/commit/47e137febd86ed07f277e08657dbcdf77702250f))
* upgrade logger verbosity when provided to telemetry ([48577ba](https://git.act3-ace.com/ace/data/tool/commit/48577ba624b5fb7a6bbed431edca5528ff3aa82c))
* we now add parts before processing labels ([d88e29d](https://git.act3-ace.com/ace/data/tool/commit/d88e29d80b51509d2285bf7dcb907bf589de6940))

## [1.0.6](https://git.act3-ace.com/ace/data/tool/compare/v1.0.5...v1.0.6) (2022-12-22)


### Bug Fixes

* formatting in "ace-dt bottle show" when parts do not have labels ([ffa9cac](https://git.act3-ace.com/ace/data/tool/commit/ffa9cacc5b5fc781b5ff486bb78721abc259fbf8))
* support for legacy bottles ([1321a03](https://git.act3-ace.com/ace/data/tool/commit/1321a03095d1a474aaf89fcab05be83df419f2b9))
* suppress logging one level while loading config ([89a345d](https://git.act3-ace.com/ace/data/tool/commit/89a345d1e351f4b39d7d40d7e23b8ccf61a626bd))
* updated help to show how to output the sample configuration ([278eba0](https://git.act3-ace.com/ace/data/tool/commit/278eba0a384bd7405fabe175e162721872e4f5b3))
* virtual parts ([2b3ecd8](https://git.act3-ace.com/ace/data/tool/commit/2b3ecd8c0a98e6abd7fef6afafe3c1123d34f88f))

## [1.0.5](https://git.act3-ace.com/ace/data/tool/compare/v1.0.4...v1.0.5) (2022-12-19)


### Bug Fixes

* caching while pulling ([a1c1573](https://git.act3-ace.com/ace/data/tool/commit/a1c1573d9de0846937671754599c52e7e28d7108))
* correct some additional bugs with handling directory parts, and matching from existing labels.yaml ([6fa47cd](https://git.act3-ace.com/ace/data/tool/commit/6fa47cd75fd5c3ce0e2e3144ba472b01fe8e78dd))
* do not create spurious entry.yaml file ([0f9312c](https://git.act3-ace.com/ace/data/tool/commit/0f9312ca86321c71c86e6a340809177608edfcad))
* pipeline unit test reporting ([b9ed98b](https://git.act3-ace.com/ace/data/tool/commit/b9ed98bcc403df69b4ac2b368c1b7c14be2dac5f))
* return actual error instead of "bottle not found" ([934efd6](https://git.act3-ace.com/ace/data/tool/commit/934efd651dcdffd081f5e671342a3659584fe84c))

## [1.0.4](https://git.act3-ace.com/ace/data/tool/compare/v1.0.3...v1.0.4) (2022-12-12)


### Bug Fixes

* bottle part label not working for directories, and regex parsing fix ([dd9cb58](https://git.act3-ace.com/ace/data/tool/commit/dd9cb584a3de5df48a0fc774f168adb4438ea132))
* small lint complaint and slight clarity correction for part label help text ([cf7e829](https://git.act3-ace.com/ace/data/tool/commit/cf7e8297176749ec9e05d4f67f3e00d42337c8dc))

## [1.0.3](https://git.act3-ace.com/ace/data/tool/compare/v1.0.2...v1.0.3) (2022-12-09)


### Bug Fixes

* **ci:** bump to fix packaging for homebrew ([c0b5948](https://git.act3-ace.com/ace/data/tool/commit/c0b59484c02ad242c59d05a749a8198c49df6f5a))
* double check and fix expected args and handle edge cases ([299650d](https://git.act3-ace.com/ace/data/tool/commit/299650d6d40d9c0f87c7c4f21c62677b422fc0f0))
* triggers ([10476d4](https://git.act3-ace.com/ace/data/tool/commit/10476d44d5b7f002a738e3f8fce8c90c99720de4))

## [1.0.2](https://git.act3-ace.com/ace/data/tool/compare/v1.0.1...v1.0.2) (2022-12-09)


### Bug Fixes

*  pulling by bottle ID failed prematurely ([6a41f75](https://git.act3-ace.com/ace/data/tool/commit/6a41f75daa6876a41960cb5af9926b7f03a37eec))
* **ci:** bump pipeline to fix cobertura ([bc96bbd](https://git.act3-ace.com/ace/data/tool/commit/bc96bbdbca306437f80b5f9fa9ab3e316983aa8b))
* update the pipeline to v8.2.22 ([b69f11d](https://git.act3-ace.com/ace/data/tool/commit/b69f11da524504c7d21a1015959e222a26435009))

## [1.0.1](https://git.act3-ace.com/ace/data/tool/compare/v1.0.0...v1.0.1) (2022-12-08)


### Bug Fixes

* bottle pulls that do not hit the telemetry server failed without returning an error ([e1fc2d2](https://git.act3-ace.com/ace/data/tool/commit/e1fc2d2066c80085050b9dab16a0bc521e5d30fe))
* resolve Location url references to allow relative urls during push ([f9035ee](https://git.act3-ace.com/ace/data/tool/commit/f9035ee33045cdf8ab93755f12e249e3217657b5))

# [1.0.0](https://git.act3-ace.com/ace/data/tool/compare/v0.25.4...v1.0.0) (2022-12-06)


### Bug Fixes

* a few more function comments added ([cd368ea](https://git.act3-ace.com/ace/data/tool/commit/cd368ea248949133d0bae0078c48aabe10c7ed22))
* account for empty telemetry Locations ([16fd435](https://git.act3-ace.com/ace/data/tool/commit/16fd4350785da4099c5ebccae48b119087c2744e))
* actually prune the cache with PruneCache() ([5593647](https://git.act3-ace.com/ace/data/tool/commit/5593647acdeaa4da28e238178864682206e1721b))
* Add app name to ko build step in ci ([7e5d3b4](https://git.act3-ace.com/ace/data/tool/commit/7e5d3b4dd333a51f711eea0c15e5e758925760e7))
* add authentication for repository index retrieval ([45538a7](https://git.act3-ace.com/ace/data/tool/commit/45538a713a3306d9cdf428c0da70c4ab2c059591))
* add back in pull tests ([030b611](https://git.act3-ace.com/ace/data/tool/commit/030b611299c1e561a64010c2d9d7786d5b27b7bc))
* add bottle source path on pull for telemetry use, and remove duplicated manifest parse code ([217acee](https://git.act3-ace.com/ace/data/tool/commit/217aceed29b6741d3b0fc8ecae7ccfbe1ffee417))
* add call to telemetry in bottle pull action ([ef8e9db](https://git.act3-ace.com/ace/data/tool/commit/ef8e9dba576a085796550d42cd04a46761c8c09f))
* add check for symlink in filewalker ([b31cae7](https://git.act3-ace.com/ace/data/tool/commit/b31cae79113e38f75b94e783668846e133496cd7))
* add config commands ([fb53525](https://git.act3-ace.com/ace/data/tool/commit/fb53525b20822ac0e15a5014696ff0c24caf2363))
* add context to cmd level funcs ([993fecd](https://git.act3-ace.com/ace/data/tool/commit/993fecd1d62beca4858ca23540442047b144475f))
* add context to commands. move logging logic to cmd pkg ([b103c31](https://git.act3-ace.com/ace/data/tool/commit/b103c319ea0b54315a8549fbfa7f5ed843fd0d77))
* add crude debug mode for no eta and implement speed decay with pull speed only no eta ([270ada9](https://git.act3-ace.com/ace/data/tool/commit/270ada95af464dc86f33fb7d2b21cdc6f8317e0a))
* add dir tests ([e8f714e](https://git.act3-ace.com/ace/data/tool/commit/e8f714eb8b373c4dd2aab544883a1fef3fc9811a))
* add docker.io default registry mapping to config ([4cd0e68](https://git.act3-ace.com/ace/data/tool/commit/4cd0e684fbf4a5203c2f8ed44e9df1f18fbdb376))
* add documentation for AuthState handling functions ([9c5ae98](https://git.act3-ace.com/ace/data/tool/commit/9c5ae98ef92a4d83fd1ef76cb939c84ae05fc1a5))
* add documentation for state observer declarations ([3083293](https://git.act3-ace.com/ace/data/tool/commit/30832932cfd8ab76c60c0e4ec1b88dea4b0b2de1))
* add documentation for TrackingStorageProvider interface support in LocalOCIStorageProvider ([aff50ea](https://git.act3-ace.com/ace/data/tool/commit/aff50ea81f0bda42eab1e841251b3ef5c4f84197))
* add expiration of 180 seconds to scoped authorizers ([7e00dc4](https://git.act3-ace.com/ace/data/tool/commit/7e00dc42c6206279528d6218536931d301a918fa))
* add extra logging to manifest transfer at level 2+ ([bd75040](https://git.act3-ace.com/ace/data/tool/commit/bd750408e0409490c3de5b910ac70a83ecdbe2f2))
* add in meta tests ([99e87ae](https://git.act3-ace.com/ace/data/tool/commit/99e87ae38cd3f1836c3eb011c4fdde4b6ebdfda3))
* add logic to allow symlinks for init, commit, and push ([b56ceee](https://git.act3-ace.com/ace/data/tool/commit/b56ceeee2393a3dc9a69a2e2fac5b9d32fe46fe2))
* add mirror tests back in ([f08491f](https://git.act3-ace.com/ace/data/tool/commit/f08491f9ae798825fe97aca052c55aa090b547d2))
* add missing documentation for functions in datastore and bottle, and errors ([089071d](https://git.act3-ace.com/ace/data/tool/commit/089071dda7c69f32ea71d14c38a8a683007b5486))
* add more of the basic func tests back ([467701e](https://git.act3-ace.com/ace/data/tool/commit/467701ef07de6ceb4627c0031a70600f94aecf31))
* add more tests ([34d9606](https://git.act3-ace.com/ace/data/tool/commit/34d9606da0f0afed6480b101b937ea8e2b5a007a))
* add progress to commit genmeta (digesting) ([337a113](https://git.act3-ace.com/ace/data/tool/commit/337a113938998cfdb00c69da8937d4d9decf9423))
* add retryable http function documentation ([1ffd78a](https://git.act3-ace.com/ace/data/tool/commit/1ffd78a49acb12f1cf85b4e731338c07f88f9c67))
* add scheme if specified to telemetry repository references ([e029bed](https://git.act3-ace.com/ace/data/tool/commit/e029bed7caf5487ea8ad9be0a287ee638f476ed1))
* add size to part track and oci.fileinfo ([f937b01](https://git.act3-ace.com/ace/data/tool/commit/f937b01c949835ee6f4ce21c3ad877a4f44bc383))
* add start of meta commands ([868cb86](https://git.act3-ace.com/ace/data/tool/commit/868cb86c6131fb38fc79540495b3a392b97bd11a))
* add tests ([030bbc5](https://git.act3-ace.com/ace/data/tool/commit/030bbc52955a845f08f82c834dfe67d40895f186))
* add transfer progress and object count tracking, and improved push failure reporting ([7bb6849](https://git.act3-ace.com/ace/data/tool/commit/7bb6849082d61b33aef5c0defe1aff063d0234d2))
* add variadic argument forwarding to log wrapper in retryable http wrapper ([a226b1f](https://git.act3-ace.com/ace/data/tool/commit/a226b1f045827856885654defa1de5bcc41afefa))
* added artifacts to the integration tests ([1d063c1](https://git.act3-ace.com/ace/data/tool/commit/1d063c1dd8d22ed89f41d513d964fb1cc7f0f363))
* added bottle path to path used in statusDirVisitor ([c949e8a](https://git.act3-ace.com/ace/data/tool/commit/c949e8a037abac57f06d7532515f301fdfeca875))
* added bottle serialization HACK ([0441a5e](https://git.act3-ace.com/ace/data/tool/commit/0441a5e08025568d34fe3116a32740f98d2b87f1))
* added data race detection ([0502a4b](https://git.act3-ace.com/ace/data/tool/commit/0502a4b7b828f6347cd4d03442b09da7405f6967))
* added hard link test ([39508b3](https://git.act3-ace.com/ace/data/tool/commit/39508b3182fbaa4eb23ba269cbeae4aa16afe386))
* added mediatype for gitlab compatibility ([b49686c](https://git.act3-ace.com/ace/data/tool/commit/b49686ceac6af0f8259c86637edfba4a3b31d6f3))
* added missing deps ([25e0e29](https://git.act3-ace.com/ace/data/tool/commit/25e0e299726de7b5f066a5a867fb9b2e974890ff))
* added mroe to go.sum file ([a35b832](https://git.act3-ace.com/ace/data/tool/commit/a35b832d3d8b3dc6851faed5d97e3fd6e577860b))
* added PreConfigHandler and refactored again ([f6fc289](https://git.act3-ace.com/ace/data/tool/commit/f6fc28962305b59e53f8d41a3861cae08aa5a11b))
* added support for pulling old bottles ([29d1d75](https://git.act3-ace.com/ace/data/tool/commit/29d1d75cee037b15923d1aee62e6da1ce6b70716))
* added the platform information into the image config ([e96b08f](https://git.act3-ace.com/ace/data/tool/commit/e96b08f2f00d5b974005190d96d8e114531e84e6))
* additional error passing where previously string passing ([e4315f1](https://git.act3-ace.com/ace/data/tool/commit/e4315f11ba53856b4934f85dd849b225aaab09e4))
* address race condition in TestTransferTask::Transfer ([7795b4c](https://git.act3-ace.com/ace/data/tool/commit/7795b4c545a043dda592985fe2089a7da78caefa))
* adjust log levels to more closely match manifesto guidelines ([50361b6](https://git.act3-ace.com/ace/data/tool/commit/50361b6dc0099b8b07b7b7c5cd5489f32209ad76))
* allow pull with re-ordered fields in schema ([fea0c45](https://git.act3-ace.com/ace/data/tool/commit/fea0c4560b0edcdb04fa5a6315ae3ce34b69b838)), closes [#281](https://git.act3-ace.com/ace/data/tool/issues/281)
* an issue with periodic failures in Test_Functional_Pull_Versions unit test ([2a2ba76](https://git.act3-ace.com/ace/data/tool/commit/2a2ba760594652d460fd8e35d4c7e2d0ff4b35eb)), closes [#315](https://git.act3-ace.com/ace/data/tool/issues/315)
* another data race squashed by Jack and I ([a1671a2](https://git.act3-ace.com/ace/data/tool/commit/a1671a27c017393e36d25065f828330f26d2aec3))
* better error handling in the transfer go routine ([d44803f](https://git.act3-ace.com/ace/data/tool/commit/d44803f7c7395cd7da9680e15155ce806e910e41))
* blob correct blob ledger tracking for manifest lists, and properly count existing items on pull for progress ([66a1945](https://git.act3-ace.com/ace/data/tool/commit/66a19451d88935dcce99156f68f1c2ec1e5790ab))
* bottle pull with auth ([2363931](https://git.act3-ace.com/ace/data/tool/commit/2363931c7fdbb29323dab565233a8de05d7302ce))
* bump gitlab-ci ver. ([82e1410](https://git.act3-ace.com/ace/data/tool/commit/82e1410345f41379dac2d42d5685ba6567b7a38e))
* bump schema to always omitempty fields in bottle.json ([b768fbb](https://git.act3-ace.com/ace/data/tool/commit/b768fbb2d3c322849d82feadadcacc88e05f91f1))
* bump schema to always omitempty fields in bottle.json ([a320370](https://git.act3-ace.com/ace/data/tool/commit/a320370112570793a8f276adca1cedb8a42edaa1))
* bump telemetry and schema ([df86948](https://git.act3-ace.com/ace/data/tool/commit/df8694806b53c47a316e8eaf34357e22b190a39d))
* bumped telemetry ([4996aed](https://git.act3-ace.com/ace/data/tool/commit/4996aed09b0f5f85b6cafd7ac86933b8fa48f6b2))
* cache auth to reduce authentication retries, manifest list fixes, and reduction on transfer worker threads ([4f0a493](https://git.act3-ace.com/ace/data/tool/commit/4f0a49308d755b276f8364ab43c03f91e10e4148))
* cache local if /tmp not available ([6f39309](https://git.act3-ace.com/ace/data/tool/commit/6f393099f6dd887605d4d7ea499cda8886a53e3d))
* calculate manifest digest from previous manifest data on telemetry pull events ([7f32c19](https://git.act3-ace.com/ace/data/tool/commit/7f32c19857321ea32dde6d4579ef384b95e36147))
* capture default registry mapping in functional test ([4158ba5](https://git.act3-ace.com/ace/data/tool/commit/4158ba59f97089bd25a38107042de97053207a67))
* catch more close errors ([65a3ad8](https://git.act3-ace.com/ace/data/tool/commit/65a3ad8e79568029ac6563fdaf96362fae0fb76d))
* change how tests are closed ([5fc3a3a](https://git.act3-ace.com/ace/data/tool/commit/5fc3a3af20fe5a42c9361627dcd7dc5890a04020))
* changed the command instead of entrypoint to fix functional tests hopefully. ([4aa5f75](https://git.act3-ace.com/ace/data/tool/commit/4aa5f750eb0d0af72b2008024af10aa98f91e59e))
* check return err pass 1 (done) ([e3525df](https://git.act3-ace.com/ace/data/tool/commit/e3525dfcef8507863d2c58e095c5a9d77d1d3ab4))
* check return err pass 1 (not done) ([adb16e1](https://git.act3-ace.com/ace/data/tool/commit/adb16e1d395b9cc7102bc416bcdd9048811028c1))
* CI lint issue ([170c9f1](https://git.act3-ace.com/ace/data/tool/commit/170c9f1ec8b6fe4af9b37fe09b3ee3cde0b269d6))
* **ci:** bump to v8.2.0 ([618c60c](https://git.act3-ace.com/ace/data/tool/commit/618c60c49692af03e1c4c777835fd97de7d6be81))
* **ci:** update to gitlab 15 ([8d23e34](https://git.act3-ace.com/ace/data/tool/commit/8d23e349a23cc1b877ca2006672551029355992b))
* clean up command flags ([64f6d1f](https://git.act3-ace.com/ace/data/tool/commit/64f6d1f0fcbb0884ec578d542bb5951d37840b34))
* cleanup ([974f00a](https://git.act3-ace.com/ace/data/tool/commit/974f00aa3a52cf878ee548e05175b6e0e327b6f2))
* cleanup ([af37184](https://git.act3-ace.com/ace/data/tool/commit/af37184beca79c415dc1d38285479536d2561633))
* cleanup ([1f0d0d6](https://git.act3-ace.com/ace/data/tool/commit/1f0d0d665e59c6b0afd1f20e57a0ebb48e7c8433))
* cleanup ([7728508](https://git.act3-ace.com/ace/data/tool/commit/772850843616e4f3945bd50fe1fc7a1738a9a33f))
* cleanup ([88873b9](https://git.act3-ace.com/ace/data/tool/commit/88873b9e23c467c5d9bb46b3c2c566b42f4ba2d6))
* cleanup ([ced8e14](https://git.act3-ace.com/ace/data/tool/commit/ced8e1458e73511afa2cf374a775c024de55efd6))
* cleanup ([f78590c](https://git.act3-ace.com/ace/data/tool/commit/f78590c0f2f8d180d35eaa99234de91d416cc954))
* cleanup and add to mirror push ([4581322](https://git.act3-ace.com/ace/data/tool/commit/45813229fd6ea2d0819986ca565b1738e4582716))
* cleanup and de-complexify ([270870f](https://git.act3-ace.com/ace/data/tool/commit/270870f7e813122b65ee82a4148fdcf7ec5fc722))
* cleanup and standardize ([5aa4c98](https://git.act3-ace.com/ace/data/tool/commit/5aa4c984f9ee2c1f4fce4a9a46fa2a7f5a196c26))
* clear default CheckRetry function when calling RetryableHTTPClient, to enable set-once behavior ([a8b9302](https://git.act3-ace.com/ace/data/tool/commit/a8b93024a67ed5931780126852a720cd51208f6e))
* collect errors from pull ([bdcfde2](https://git.act3-ace.com/ace/data/tool/commit/bdcfde24dddc1bb5b5e6934510f32bad15d3a0e3))
* comment out test.Helper() ([dc762ca](https://git.act3-ace.com/ace/data/tool/commit/dc762cac18184affd40366ded339e19297cb486d))
* commit not saving label consistency fixes and correct adding multiple lables in one command line ([c9adb6d](https://git.act3-ace.com/ace/data/tool/commit/c9adb6dc8e28b4da8ee48bc75b6bea200639c96c))
* commit tests now work ([9dd24c0](https://git.act3-ace.com/ace/data/tool/commit/9dd24c0df3d5469805e51d5892de09a3de69c0b7))
* concurrency for OCIIndex is now set properly ([38fb247](https://git.act3-ace.com/ace/data/tool/commit/38fb247b114ce2f6a4cd0712f3ad53302b001f7e))
* correct another issue with trailing slashes for labels, and missing legacy media type ([320b2db](https://git.act3-ace.com/ace/data/tool/commit/320b2db0cbd8bc7d3ead1a49f818975838eb777d))
* correct broken labeling for multiple labels on glob matched paths ([f5c17c1](https://git.act3-ace.com/ace/data/tool/commit/f5c17c1e51c324ed4fc022b13512189dee663b34))
* correct delete, show, and pull functional test failures ([4813b41](https://git.act3-ace.com/ace/data/tool/commit/4813b41336d9f5140fd6a900059eec5d8e30d098))
* correct error handling for adding lable to part that already has that label ([f33e824](https://git.act3-ace.com/ace/data/tool/commit/f33e8245c72ccb9d134936b9dfe38b4fab8c8fe0))
* correct golangci file and change phrasing for pipe errors ([1c866b7](https://git.act3-ace.com/ace/data/tool/commit/1c866b7ac535ddc40256690e3c3e5a391e67ac05))
* correct inverted error detection logic that caused local config to be skipped ([dfad373](https://git.act3-ace.com/ace/data/tool/commit/dfad373752e6fe3f69bfe4a6d2f56412f2f70d82))
* correct misinterpretation of certain file characters as regular expression matches during labeling ([4821a73](https://git.act3-ace.com/ace/data/tool/commit/4821a73ad624684a95afe3f8606706d932ca284f))
* correct temp config path ([79dc8e2](https://git.act3-ace.com/ace/data/tool/commit/79dc8e2f552b60ce8bcbfa686bf228f0aa117118))
* correctly send pull event rather than push event for telemetry on pull ([ddc438d](https://git.act3-ace.com/ace/data/tool/commit/ddc438dc278d073f3165819205fca8aeef225633))
* correctly set length and capacity with make() ([a9c616c](https://git.act3-ace.com/ace/data/tool/commit/a9c616c6d2b050c817a194b0714412c1c1c974c1))
* counter related crash and mismatched counts on mirror ([7000427](https://git.act3-ace.com/ace/data/tool/commit/700042717e4a376d62b37bab40ed4cc0005db62a))
* cranking up some logging ([28a554d](https://git.act3-ace.com/ace/data/tool/commit/28a554d1713dc667fc2bd3d8eb3e845ffc189891))
* custom closer didnt want to close twice ([9c3cdb9](https://git.act3-ace.com/ace/data/tool/commit/9c3cdb954d329728243f245b4542423197888909))
* custom closers closing twice errors ([da5dc84](https://git.act3-ace.com/ace/data/tool/commit/da5dc849ed627367fff7b43798bb24a54746d073))
* define compression ratio properly ([e72af09](https://git.act3-ace.com/ace/data/tool/commit/e72af091159b8eb344a4d1a844f4a6ddb3307cc1))
* delete and show test ([f9ca530](https://git.act3-ace.com/ace/data/tool/commit/f9ca530ad9f60c2321b86275114df9666b9250d4))
* delete failure ([d4c2680](https://git.act3-ace.com/ace/data/tool/commit/d4c26803278a0bd6926c25cc1b5aca3de44af6e2))
* **deps:** update golang to 1.19 ([0d8fe81](https://git.act3-ace.com/ace/data/tool/commit/0d8fe81b034883060b93d712f781f6eedf79e547))
* **deps:** update pipeline to v7.0.17 ([237c425](https://git.act3-ace.com/ace/data/tool/commit/237c42578c82fc80e377e1de8513348cbb577ef4))
* digest data race ([f5bd0ef](https://git.act3-ace.com/ace/data/tool/commit/f5bd0efd5e1592e374318abb651368ab79784dfd))
* disable cache ([046e9c7](https://git.act3-ace.com/ace/data/tool/commit/046e9c77e97f255a49fad2e80c90fb7686268dfd))
* discard old context references ([adc6640](https://git.act3-ace.com/ace/data/tool/commit/adc664076b100a662e09c225a32a517a94938312))
* display failure stop if fail, more cleanup ([edc1636](https://git.act3-ace.com/ace/data/tool/commit/edc16362d2fd79a763750aa58c2fe260e15f319f))
* docs cleanup ([c143923](https://git.act3-ace.com/ace/data/tool/commit/c143923ca4d9d230f034ac01705833614dd6267b))
* docs updates and lint fixes ([4e7619d](https://git.act3-ace.com/ace/data/tool/commit/4e7619d8ec179569c4f7402087e6fb973e4483cf))
* don't wrap nil ([0adec6d](https://git.act3-ace.com/ace/data/tool/commit/0adec6df15dce7f61737c4326db4197452aa6b9d))
* enabled client and proxy caching for the files ([e4a5ba0](https://git.act3-ace.com/ace/data/tool/commit/e4a5ba061b4e59e572f72054825e08178aea43cb))
* extra-index-url flag was wrong ([13a510c](https://git.act3-ace.com/ace/data/tool/commit/13a510cd7eedeb6efc9e83d3e858abf7da117093))
* fallback to original pathspec if glob fails to find results (testing fix) ([f10f9ad](https://git.act3-ace.com/ace/data/tool/commit/f10f9adc8ebb3b473628aebcea534f1ce2b37e5f))
* final updates and cleanup ([2b131ce](https://git.act3-ace.com/ace/data/tool/commit/2b131ce69cd775a710c94fc7fab1f7d9cb7747c7))
* finalize testhelper rename ([0b59bc3](https://git.act3-ace.com/ace/data/tool/commit/0b59bc3f93c886d94678ddd90784dda3acc0ab14))
* first comments ([90134dc](https://git.act3-ace.com/ace/data/tool/commit/90134dcea37ee46539981b8ca5ccb97620770929))
* first draft at getting csv data from progress ([f0c5ca6](https://git.act3-ace.com/ace/data/tool/commit/f0c5ca63c2eb8560b34f5c3f58813a4b60f04350))
* fix tests ([6c8e420](https://git.act3-ace.com/ace/data/tool/commit/6c8e420c2f89aff7dfece47f4f84a382d08d01dd))
* fix version labels ([b9ac268](https://git.act3-ace.com/ace/data/tool/commit/b9ac268bfdcb84458e9eb89c82eff0287b917dc8))
* fixed a bug that impacted harbor registries ([9416dad](https://git.act3-ace.com/ace/data/tool/commit/9416dad8608c53fadfbfabf87cf07e7429334ded))
* fixed a unit test for archiving ([03c94bb](https://git.act3-ace.com/ace/data/tool/commit/03c94bbf01d2cda228620a4b259d8f8d3205f249))
* fixed label validation ([0174bcf](https://git.act3-ace.com/ace/data/tool/commit/0174bcff0c366cfe3dba43e7c671b39cc451d793))
* fmt pass ([1d5dfea](https://git.act3-ace.com/ace/data/tool/commit/1d5dfea3babf521f41bdd62bfeac18605d6ddf88))
* follow symlinks correctly (recursively) ([5271a49](https://git.act3-ace.com/ace/data/tool/commit/5271a49173cb8ea82dcaba223ab1077e7e08c90d))
* forward individual manifest destiation refs to cloned transfersettings ([115bb0c](https://git.act3-ace.com/ace/data/tool/commit/115bb0c928d3f24b2e30438f0759f66aeb953444))
* func tests ([124f70e](https://git.act3-ace.com/ace/data/tool/commit/124f70e5a88add79b1ef043166bf4d080d7441ff))
* functional tests had the old flag name ([c407346](https://git.act3-ace.com/ace/data/tool/commit/c4073463b386e407b03c08910c4ec02be61b0f63))
* get rid of an extra space in the readme ([3d58194](https://git.act3-ace.com/ace/data/tool/commit/3d58194fe00a43f65f997908465814fcdcd4f289))
* handle missing algorithms when matching bottle ID more gracefully ([6f4fcf2](https://git.act3-ace.com/ace/data/tool/commit/6f4fcf2bd0bf00bf82ba107e2e939be016fb4c81))
* ignore cognitive limit for printing bottle info ([7aaea45](https://git.act3-ace.com/ace/data/tool/commit/7aaea4545c78dbf7462e98d8e682b0839d9b8346))
* ignore missing telemetry URLs ([9f44a02](https://git.act3-ace.com/ace/data/tool/commit/9f44a02ffc14db108a7961f8b6d097328a1730c3))
* image building in CI ([356e67b](https://git.act3-ace.com/ace/data/tool/commit/356e67bdbe2b9aeb5a95f96982d05a79e8ab7321))
* incorrect root path determination on windows ([f67affe](https://git.act3-ace.com/ace/data/tool/commit/f67affe95260399c4c65717c5960d5feb1b1e60e))
* increase test limits ([8e2d75c](https://git.act3-ace.com/ace/data/tool/commit/8e2d75c192900de84c1b816096242db32df51147))
* increase verbosity ([f6023bc](https://git.act3-ace.com/ace/data/tool/commit/f6023bc6a7de8fc66e64b46008557fa8915b4faf))
* increase verbosity on functional tests output in gitlabci ([6a903d6](https://git.act3-ace.com/ace/data/tool/commit/6a903d61ac652ba9cadcf94d87f1021f9dc5f399))
* increased the memory of the jobs ([6cf072c](https://git.act3-ace.com/ace/data/tool/commit/6cf072cbdda6990ad3789da27211a7e00b939075))
* indexing in unit test for authors ([eb121db](https://git.act3-ace.com/ace/data/tool/commit/eb121db4abeb164a86a1facb5ff4145ccd257954))
* just return EOF not err, also lint ([aa05df6](https://git.act3-ace.com/ace/data/tool/commit/aa05df66a052895268b014486f771b82cc98a63e))
* ko build version ([7278a20](https://git.act3-ace.com/ace/data/tool/commit/7278a206c991d8c5a752b2124bf9a185ab7974d7))
* link up logger passing correctly ([509dbd0](https://git.act3-ace.com/ace/data/tool/commit/509dbd091ea60043bbc0acc1d841de8625d63138))
* lint ([de3e4b1](https://git.act3-ace.com/ace/data/tool/commit/de3e4b13a83d9e9a62c0aa5588a6019fc97acf1f))
* lint ([7802ddf](https://git.act3-ace.com/ace/data/tool/commit/7802ddf04ade60709ee9c1e12d59c64c427606dd))
* lint error for ignored error handling ([c3c33ab](https://git.act3-ace.com/ace/data/tool/commit/c3c33ab887a94b5e7ef31dfac6c02f51d2370883))
* lint errors ([a2acb8a](https://git.act3-ace.com/ace/data/tool/commit/a2acb8a3b6f5b40c603390c3af6696a70eb8d840))
* lint errors ([9595ff0](https://git.act3-ace.com/ace/data/tool/commit/9595ff079d5780bed6d57c92154aa8595a907723))
* lint errors ([c8026c8](https://git.act3-ace.com/ace/data/tool/commit/c8026c8f485dc39a4f46d1e673b39e06833c3b29))
* lint errors ([c0318fa](https://git.act3-ace.com/ace/data/tool/commit/c0318fa3050d5474337ea4e29defb458ee562d65))
* lint errors ([238e539](https://git.act3-ace.com/ace/data/tool/commit/238e53969f615c747d9cdb34db7a8dcb2e7fc3f8))
* lint errors ([2615f8b](https://git.act3-ace.com/ace/data/tool/commit/2615f8b84071502011a21d888901fa1d009bedc3))
* lint errors and fix comments ([1100e1e](https://git.act3-ace.com/ace/data/tool/commit/1100e1e6bcfb110c9938e1511adb92326b7f078b))
* lint issue ([2922145](https://git.act3-ace.com/ace/data/tool/commit/2922145fa67e5c40e60a5d7806c4175fcf125358))
* lint wrap ([9fc2217](https://git.act3-ace.com/ace/data/tool/commit/9fc22174bbe838d13afb7fbe851ab0ee5f356eae))
* linter errors ([87cd0fd](https://git.act3-ace.com/ace/data/tool/commit/87cd0fd0fa9b59405773c67e980c5ca20f657481))
* linting and bump CI pipeline ([e976712](https://git.act3-ace.com/ace/data/tool/commit/e976712c39815f2c48a999cadef22928b2b691cc))
* linting and error handling issues ([8a3d0c4](https://git.act3-ace.com/ace/data/tool/commit/8a3d0c432a2f4582689d165dd4c93cbbf5a68e6b))
* log.error pass ([5151dd1](https://git.act3-ace.com/ace/data/tool/commit/5151dd1bad4f4583cb7dcec255d49aa761b04562))
* log.info pass ([db86764](https://git.act3-ace.com/ace/data/tool/commit/db8676414b2365d2025419c45a7f8f142a99046f))
* logging level for auth ([e856eae](https://git.act3-ace.com/ace/data/tool/commit/e856eaefbb32f98088072d5f15baff3aa6224714))
* logging levels pass1 ([4592fd4](https://git.act3-ace.com/ace/data/tool/commit/4592fd47000e5eed639b9ac4918ebfe08e70945c))
* made the maps in auth.go ([b5b4b4e](https://git.act3-ace.com/ace/data/tool/commit/b5b4b4eb4d895cb861ff33b858747ecc4cff70cb))
* major fix for functional tests ([eb6cc8e](https://git.act3-ace.com/ace/data/tool/commit/eb6cc8eee35da546fa9a2ecbc0c795d6df472da6))
* make telemetry listen on more than the container's localhost ([abba135](https://git.act3-ace.com/ace/data/tool/commit/abba13537f970288a29cda1dbc8d81aed2f63baf))
* merge in changes from master ([2bb5f84](https://git.act3-ace.com/ace/data/tool/commit/2bb5f846a536e77ce1d2ae430fd72869a53118df))
* merge in changes from master ([2dc38ac](https://git.act3-ace.com/ace/data/tool/commit/2dc38aca925aec52eb75e326054b944582bf3484))
* merge in httpretrable auth fix ([3684ac4](https://git.act3-ace.com/ace/data/tool/commit/3684ac4aff7539bb499380102e6ccdcef678378d))
* merge in master changes ([4d7b903](https://git.act3-ace.com/ace/data/tool/commit/4d7b9030fdadf77d3cbb652db7799649b3957108))
* merge in master changes ([6131a67](https://git.act3-ace.com/ace/data/tool/commit/6131a67d6afa7964fc1e81ce8fb1c48503c84505))
* merge with cmd to action ([30f35b5](https://git.act3-ace.com/ace/data/tool/commit/30f35b552ad3414946daf5879d642a6fdce7db85))
* merge with master ([850f68e](https://git.act3-ace.com/ace/data/tool/commit/850f68e12d938bd5c0dead85e4913169d4282e9d))
* merge with master ([215b3bc](https://git.act3-ace.com/ace/data/tool/commit/215b3bcd138c69fd6061389ad50b8dba496c1b74))
* merge with master ([ac293da](https://git.act3-ace.com/ace/data/tool/commit/ac293dae74fe493ad46d4b6562982984431fcc42))
* merge with master ([8049649](https://git.act3-ace.com/ace/data/tool/commit/8049649fbacdbbbb912c7f832198467eb291f94c))
* merge with master ([3290d94](https://git.act3-ace.com/ace/data/tool/commit/3290d94efb6cdf27b509e5ef00bfe5b8d77ba4fa))
* merge with master ([171c2a4](https://git.act3-ace.com/ace/data/tool/commit/171c2a431a80b9142c49e2f19286c87c22ed02dc))
* merge with master ([b9a77a2](https://git.act3-ace.com/ace/data/tool/commit/b9a77a27994f79ba2f1b3c943f9db54013319d1e))
* merge with master ([b76b1b5](https://git.act3-ace.com/ace/data/tool/commit/b76b1b53085d34444e799f2134d54766f1036234))
* merge with master ([e7b2d01](https://git.act3-ace.com/ace/data/tool/commit/e7b2d01ae135ee4f042f9e1b88bd0eb17e6cadce))
* merge with master ([0d011a2](https://git.act3-ace.com/ace/data/tool/commit/0d011a21623d85774349dbf582179ca45c5d31b4))
* merge with master ([fbf308b](https://git.act3-ace.com/ace/data/tool/commit/fbf308bbfffe7d33d75c94d2645616458b5a917c))
* merge with master ([5aaf108](https://git.act3-ace.com/ace/data/tool/commit/5aaf1089143f4dd5c4b0b6e671e92333cb368101))
* merge with master ([b5257b9](https://git.act3-ace.com/ace/data/tool/commit/b5257b99487300afee1e6bda4d79afa0a42c9d0d))
* merge with master ([6b5bea1](https://git.act3-ace.com/ace/data/tool/commit/6b5bea14584be04eed7011d3c578be0dd4c2cf6a))
* merge with master ([60b4049](https://git.act3-ace.com/ace/data/tool/commit/60b404954de4640d337d1ade149bcbca780e9d8c))
* merge with master ([866a3dd](https://git.act3-ace.com/ace/data/tool/commit/866a3dd7ea3a17401a19b698cbc18cd3572fe838))
* merge with master ([0a2301f](https://git.act3-ace.com/ace/data/tool/commit/0a2301f883d8f9f8d946b35022937e244a45e1be))
* merge with master ([5bcfc4b](https://git.act3-ace.com/ace/data/tool/commit/5bcfc4b909a46c868ea8d475008b03eb6ac241c8))
* merge with master ([9591820](https://git.act3-ace.com/ace/data/tool/commit/95918201dcc00168df64b2f3b9d92a75bd7d4e12))
* merge with master ([d360e51](https://git.act3-ace.com/ace/data/tool/commit/d360e519c5f0e526ab7e86ecf6fadb07c7917303))
* merge with master + cleanup ([0a435d1](https://git.act3-ace.com/ace/data/tool/commit/0a435d1a79b64b4038eff0842c684ff7424a6984))
* merge with master, fix merge conflicts ([3436cbf](https://git.act3-ace.com/ace/data/tool/commit/3436cbf9d55bc71d9461a15605ba0f3dd12fcacf))
* merge with new func tests ([c39320c](https://git.act3-ace.com/ace/data/tool/commit/c39320ca7bf28a98feae4f3aa9eaf68d2d653459))
* missed a config ([d234440](https://git.act3-ace.com/ace/data/tool/commit/d234440b158fb0f22e189a13790f3d6668ffa0b9))
* missing file size on file reader during mirror push ([ec9d01b](https://git.act3-ace.com/ace/data/tool/commit/ec9d01b0bcd962cf7eb59a4a2f1e4446dd04096e))
* more error handling fixes relating to the use of log.Error ([3414d03](https://git.act3-ace.com/ace/data/tool/commit/3414d0327709a3a6f408aa429800701e94ca5214))
* more log.Error fixes ([7872005](https://git.act3-ace.com/ace/data/tool/commit/7872005933146b979017c5de01a8c74b7097cdb5))
* move progress debug mode to flag instead of config ([f636ec9](https://git.act3-ace.com/ace/data/tool/commit/f636ec93f2eef7889d5381115ebd1775c456e1df))
* move some functionality over to functesting pkg ([bca9262](https://git.act3-ace.com/ace/data/tool/commit/bca92625cdbe59cf35c43bf77ece6664420efc96))
* move to one test func per test case ([d089761](https://git.act3-ace.com/ace/data/tool/commit/d0897610fc389ee6676e564614c058028a947c8c))
* moved assertions into the bottle helper and command helper ([42f5671](https://git.act3-ace.com/ace/data/tool/commit/42f56717d0da46804a3440c01fc6c60869f05134))
* MR comments ([121e919](https://git.act3-ace.com/ace/data/tool/commit/121e919cc8a7715269e3abd0f32201cbd32ef177))
* multiple pypis argument parsing is fixed ([f8556dc](https://git.act3-ace.com/ace/data/tool/commit/f8556dce79efb288ad2f37b7bbda4bd06f220861))
* name normalization ([297c0bd](https://git.act3-ace.com/ace/data/tool/commit/297c0bdd39cd37aee86fa8d2d9c75a1108f9e012))
* need to check if virtualPartTracker exists ([786130f](https://git.act3-ace.com/ace/data/tool/commit/786130f4cc1e85f750a537268a6296949f3c41ac))
* no longer use git config for user name and email when creating a bottle ([85f2c9f](https://git.act3-ace.com/ace/data/tool/commit/85f2c9f7962db4907b89fed7bb29a1ce249089e4))
* output http body rather than processed manifest data for debugging error ([65ec880](https://git.act3-ace.com/ace/data/tool/commit/65ec880a75bd7e6965c0e49c2ad1ca9d0a462736))
* output the name command to stdout ([a7cbe1e](https://git.act3-ace.com/ace/data/tool/commit/a7cbe1eadd0bd980f7b37db7f7aff1f7a181b3a6))
* panic on invalid narrow auth after timeout ([5d149bf](https://git.act3-ace.com/ace/data/tool/commit/5d149bf5a0db8194f5a20b06965c3d838fdce96a))
* pass progress debug mode through the context not explicit settings for operations ([5ba4b2a](https://git.act3-ace.com/ace/data/tool/commit/5ba4b2a569b402ba8814c922ce51e8da52df5dd8))
* pre-write config file or config tests ([27466ea](https://git.act3-ace.com/ace/data/tool/commit/27466ea6a16b1c77cb0af93ae9374efaffcf6cbd))
* preserve old config and manifest through a schema upgrade for telemetry association ([c95fc14](https://git.act3-ace.com/ace/data/tool/commit/c95fc14c57f24288786425283b30c21f528167c0))
* printing out instead of logging for bottle artifact message ([163ada0](https://git.act3-ace.com/ace/data/tool/commit/163ada0ca69ab9a845c396c44957da11a99c7eae))
* properly build write-bottle-id file ([fc7ec72](https://git.act3-ace.com/ace/data/tool/commit/fc7ec72aeeab9d651c7ead54b0fe55cf302fe244))
* pull was not honoring telemetry failure properly ([af9628a](https://git.act3-ace.com/ace/data/tool/commit/af9628a99f0eac85ca5afa02c7fcd3eeec3e0a84))
* punting on reducing the cognitive complexity of these functions ([3780a2b](https://git.act3-ace.com/ace/data/tool/commit/3780a2b079e424fda551e0c11d5832f7ca0d5469))
* push and pull now round trip with directories ([413604d](https://git.act3-ace.com/ace/data/tool/commit/413604dd7c93f79312ea464264347b983a813283))
* push test forgot to load bottle ([6eb573d](https://git.act3-ace.com/ace/data/tool/commit/6eb573d394a7ad98c5799113f39e725605aeb657))
* push test with telemetry ([e25f7e0](https://git.act3-ace.com/ace/data/tool/commit/e25f7e07394ce7fd4ddd78998cffb40dcec641d0))
* put the error on the error channel in PullContentTransfer.Transfer ([9fec1e9](https://git.act3-ace.com/ace/data/tool/commit/9fec1e9c08f7469b28a0170edacf6fb52756f1fe))
* read needs to return EOF ([c3b572f](https://git.act3-ace.com/ace/data/tool/commit/c3b572f60a6a0d2197cdc5bc54a8cb8f0c12ed80))
* redo lost fixes ([2e0cd24](https://git.act3-ace.com/ace/data/tool/commit/2e0cd247d0e6831cc1b44c95a30ff8c341ae0101))
* redo symlink test config fix ([deaaab9](https://git.act3-ace.com/ace/data/tool/commit/deaaab910a638c8df9af52e9d899ef171034ba67))
* reduce test amount ([2e99109](https://git.act3-ace.com/ace/data/tool/commit/2e99109b9235f5b2723a4c2b00bb68bd07142e41))
* reduce test size ([b1a91e6](https://git.act3-ace.com/ace/data/tool/commit/b1a91e65a35fb687653e25a9ef06e34a5dcc9028))
* refactor of progresstracker ([76c4c25](https://git.act3-ace.com/ace/data/tool/commit/76c4c2530733f3b3cb9ed2a3a55fd6ff7e78af4e))
* refactor progress display ([76888b7](https://git.act3-ace.com/ace/data/tool/commit/76888b71ec214b2c6b25b36ef682ae6020e2ad74))
* refactor speed and eta ([56f4447](https://git.act3-ace.com/ace/data/tool/commit/56f44477ac7a3bfa4215b9bb8645c99ecd601914))
* remove authorizer from collection ([faf786c](https://git.act3-ace.com/ace/data/tool/commit/faf786cfcf035765a8ba90e686a792a225115c36))
* remove dest exists data race ([8fc9e34](https://git.act3-ace.com/ace/data/tool/commit/8fc9e34241e24f3931c369e966046fcb7750a403))
* remove dir part generation ([894e69d](https://git.act3-ace.com/ace/data/tool/commit/894e69db0d3e4e856d79adc36260f937df2403c7))
* remove duplicated manifest list handler code and redirect to oci package ([c4c4270](https://git.act3-ace.com/ace/data/tool/commit/c4c4270beb79b2ca875ceb9931ef287859f2d3a6))
* remove editor stuff from .gitignore ([bb94523](https://git.act3-ace.com/ace/data/tool/commit/bb9452377d338621a2b4ee07af588d8682f3ccdd))
* remove empty telemetry config ([ee6b087](https://git.act3-ace.com/ace/data/tool/commit/ee6b087fc62c1fe63b2e95ad97c050e7f78182fe))
* remove error return for nil error when getting auth credentials ([80e1644](https://git.act3-ace.com/ace/data/tool/commit/80e1644722a6826f6622eee93f1ec67fb24f276d))
* remove logging pkg ([1bc0439](https://git.act3-ace.com/ace/data/tool/commit/1bc0439a4b9a31cf48de4ed2898824af0793c45c))
* Remove mention of homebrew tap trigger ([cd57f05](https://git.act3-ace.com/ace/data/tool/commit/cd57f0597db8f1db14ac27a7059b0b68a05cfa77))
* remove more un needed changes ([7eeb65c](https://git.act3-ace.com/ace/data/tool/commit/7eeb65c6fb228b48e5afd3aa4890a825350c1464))
* remove not needed changes ([528b2b6](https://git.act3-ace.com/ace/data/tool/commit/528b2b6c4dad002ec68ff3c09dfbd5dcd2f3c8a4))
* remove old config stuff ([150f5fb](https://git.act3-ace.com/ace/data/tool/commit/150f5fb2bd775e22177a4db6320bb59fe559481f))
* remove redundant tests and set a test cache + scratch path ([22d5885](https://git.act3-ace.com/ace/data/tool/commit/22d58858ac80f15a35881d8d69f154fdeb1acae9))
* remove regHelper ([4dadbe2](https://git.act3-ace.com/ace/data/tool/commit/4dadbe2e707d874a3e09d91d201e5254ff6fcca7))
* remove some TODOs and go mod tidy ([fdb3b07](https://git.act3-ace.com/ace/data/tool/commit/fdb3b0795b6ff312baaf766d87ccb342cf0831d2))
* remove usage on functional test failure. fix context for functional tests ([d2b2019](https://git.act3-ace.com/ace/data/tool/commit/d2b2019d581147941ec2d582a0a0de73bfd53ce0))
* remove WithName() misuse ([87aaa4c](https://git.act3-ace.com/ace/data/tool/commit/87aaa4caaeb45563b01124ec4ffcd42d5c9ea656))
* removed the extraneous help menu in every functional test ([e05e26b](https://git.act3-ace.com/ace/data/tool/commit/e05e26b68d7863a01e5b7c50f4602019fab6f0ad))
* removed the need for go 1.19 ([9a14ede](https://git.act3-ace.com/ace/data/tool/commit/9a14edee42205394f744004e7863d56b172aff1b))
* removed unnecessary job modification in CI ([94c4b57](https://git.act3-ace.com/ace/data/tool/commit/94c4b57983f915e1c59bcb7f742659b4443ce41a))
* rename functest to testhelper ([f14efad](https://git.act3-ace.com/ace/data/tool/commit/f14efad34f854477b200a31aa19aa46d45f591ba))
* rename old helper to new tracker ([aec0413](https://git.act3-ace.com/ace/data/tool/commit/aec04135d5072b48449a77f7690f7ffa0e057eec))
* replace os.IsNotExist() with errors.Is() ([00f08ab](https://git.act3-ace.com/ace/data/tool/commit/00f08abb3e9a8aefd3ee7115ec53625661aca520))
* replace require with assert ([ac4bdd8](https://git.act3-ace.com/ace/data/tool/commit/ac4bdd8aff0fac7ce64b239bd6571a6fdc5336d6))
* resolve http retry auth func data race ([75cea50](https://git.act3-ace.com/ace/data/tool/commit/75cea50d264083a61c2483eeec40544abb3cceec))
* resolved telemetry issue ([9346271](https://git.act3-ace.com/ace/data/tool/commit/9346271a279aa404fac39e1c59abdc3e0092fe98))
* restore authentication for mirror feature, add context to some missing areas ([01418e0](https://git.act3-ace.com/ace/data/tool/commit/01418e0bf770a48b0cfdd5fef4b4e4016b41458b))
* restore bottle schema based pull ([f5acbdc](https://git.act3-ace.com/ace/data/tool/commit/f5acbdc81051aa94216d1cbe605a3929471cd896))
* restructure command building ([de308ca](https://git.act3-ace.com/ace/data/tool/commit/de308ca8bd51657f9d507a8a1e9ec7b7674c5cb9))
* return correct content descriptor as part of content pull result ([cb6c78e](https://git.act3-ace.com/ace/data/tool/commit/cb6c78e0c111b5be5aea7bda0e8524a495a92b55))
* return to original delete test ([9734b0c](https://git.act3-ace.com/ace/data/tool/commit/9734b0c4766969bd0a6c65d1f0616ad76cdd8179))
* revert gitlabci pipeline ([2843f8f](https://git.act3-ace.com/ace/data/tool/commit/2843f8f672423d2313704bc047acb6b6f11f11b4))
* satisfy lint ([afb9a63](https://git.act3-ace.com/ace/data/tool/commit/afb9a630644ff30d3e6ebe859347e51f3bec8709))
* seconds in the formatting. ([87ff127](https://git.act3-ace.com/ace/data/tool/commit/87ff1271ae6d0ac93f97ff35e1f40b3c3168547c))
* separate environment variables from scheme definition ([55501d6](https://git.act3-ace.com/ace/data/tool/commit/55501d61d72853d12925639f9d8b21c4abb33cba))
* some data race fixes ([502e890](https://git.act3-ace.com/ace/data/tool/commit/502e89069b3179aed035e5ce32b69e88b1ef923f))
* start working on display debug mode ([dda77c9](https://git.act3-ace.com/ace/data/tool/commit/dda77c9ce09eebf170b192af1217e2e0e9763903))
* switched to persistent flags ([55ffe35](https://git.act3-ace.com/ace/data/tool/commit/55ffe352d74df6686cdd89e585bec632ef95d3bb))
* switched to the correct apis doc generator ([c8c2be0](https://git.act3-ace.com/ace/data/tool/commit/c8c2be010c992c8344ff34798d44a50526367ff1))
* symlink test use correct cache ([73fcbac](https://git.act3-ace.com/ace/data/tool/commit/73fcbac67111f42d122f5baab0e3f73427f25717))
* symlink to directory ([1d90ec1](https://git.act3-ace.com/ace/data/tool/commit/1d90ec18672560eb0944731240e29c210b0a7452))
* telemetry errors preventing bottle retrieval from completing ([da5be72](https://git.act3-ace.com/ace/data/tool/commit/da5be72c72483e1852de5171423cfdf403decf49))
* telemetry username usage in CI ([5a6c113](https://git.act3-ace.com/ace/data/tool/commit/5a6c113f0628475003c3afd179ac5321f03497ef))
* telemetry was not starting ([2ea1d2f](https://git.act3-ace.com/ace/data/tool/commit/2ea1d2f30a730045ad5bf548b59ec0023646fb45))
* test using wrong function name ([143240c](https://git.act3-ace.com/ace/data/tool/commit/143240ce59b3759759253a8faa2c919c24bb1045))
* testing apply env ([22f16d5](https://git.act3-ace.com/ace/data/tool/commit/22f16d57228a5c9ae67c0819f96b10a024bb6b33))
* try only using temp config on write ([14fca31](https://git.act3-ace.com/ace/data/tool/commit/14fca3161dd733dd919e44a2602688c7559acfd1))
* try remove all ([58e9dc8](https://git.act3-ace.com/ace/data/tool/commit/58e9dc88f246f4d2e17ce7765f49cf4e94615ec3))
* uncomment t.helper() ([6031238](https://git.act3-ace.com/ace/data/tool/commit/6031238a385c6e9cce14a88674f8871dc4c9b7b0))
* update /pkg/cache step1 ([e248b02](https://git.act3-ace.com/ace/data/tool/commit/e248b02149c8729aa74c92c7c2daa4517d6a07ca))
* update config tests to use explicit test config ([eb73f88](https://git.act3-ace.com/ace/data/tool/commit/eb73f8885ab9e1c92df9157ee8378dc58b35dbea))
* update dir tests ([fbbeab9](https://git.act3-ace.com/ace/data/tool/commit/fbbeab916b47ce888c25932fe6e356163b6dc51a))
* update gitlab ref to v8.0.6 ([d1ce61d](https://git.act3-ace.com/ace/data/tool/commit/d1ce61d714667c5f10e7650a8381596b633d9553))
* update gitlabci ([e26ebd9](https://git.act3-ace.com/ace/data/tool/commit/e26ebd943b894d444365c3683a5893ea82648cde))
* update how to start and end tests ([4a1d69c](https://git.act3-ace.com/ace/data/tool/commit/4a1d69c0ac7788e7b21ad60f10a7b5f4c8849b99))
* Update integration test to use a bottle with the latest format ([5be734d](https://git.act3-ace.com/ace/data/tool/commit/5be734d34c8c032dcb8cfa77939dabe7a5d887ad))
* update internal ([6112fae](https://git.act3-ace.com/ace/data/tool/commit/6112fae9642d8659477e46deba7f202e4423df7b))
* update ko application name DO_RELEASE ([244104b](https://git.act3-ace.com/ace/data/tool/commit/244104bad2274bf7c948fc3d3cbabddeaf529762))
* update logger names ([d86ebbb](https://git.act3-ace.com/ace/data/tool/commit/d86ebbb3c91fea2bfa5c2b67f9e2ea727cb58243))
* update logs in telemetry to not report errors, but log warnings ([3b10171](https://git.act3-ace.com/ace/data/tool/commit/3b10171b63a202c6d198184fd34c41a88734fc19))
* update mnist bottle bottle version for integration round trip test. ([28ce258](https://git.act3-ace.com/ace/data/tool/commit/28ce2588338001cbf50a39d173375e1ed8465290))
* Update notes regarding release building. ([051f2b9](https://git.act3-ace.com/ace/data/tool/commit/051f2b9e4843e25dcc88556249f75bfe6701548d))
* update pipeline to v8.0.19 ([d4f280b](https://git.act3-ace.com/ace/data/tool/commit/d4f280bb13fef766c1dac86009d573f082b8399f))
* update pipeline version DO_RELEASE ([1f45cca](https://git.act3-ace.com/ace/data/tool/commit/1f45ccaee14d9587b8f950f165b40555ce94cab3))
* update pkg/archive ([842b838](https://git.act3-ace.com/ace/data/tool/commit/842b838c007ec01ff7c45b85b2ba48f767637b5b))
* update pkg/bottle ([5f46e80](https://git.act3-ace.com/ace/data/tool/commit/5f46e807a3f5bb332a707ca93eea033c06bd1fb5))
* update pkg/bottleref ([4a735a8](https://git.act3-ace.com/ace/data/tool/commit/4a735a8590ccaba5f8154c6ede7c805d2e23c92d))
* update pkg/extract ([f7d9272](https://git.act3-ace.com/ace/data/tool/commit/f7d9272e3ac6dddefccc2e66e92543743b179be1))
* update pkg/fetch ([1e7929a](https://git.act3-ace.com/ace/data/tool/commit/1e7929aa85372e820a50393f53e7cb6c10e7606f))
* update pkg/genmeta ([0644f20](https://git.act3-ace.com/ace/data/tool/commit/0644f20e17f89d0a96bb869848dbb6a93959dc71))
* update pkg/label ([0592eab](https://git.act3-ace.com/ace/data/tool/commit/0592eabb25f838db00518d21ee671ad86e6f217d))
* update pkg/mirror ([3232fcf](https://git.act3-ace.com/ace/data/tool/commit/3232fcfab470efbf7c06ab5ac16e0ed81b445c92))
* update pkg/oci ([d53f654](https://git.act3-ace.com/ace/data/tool/commit/d53f654cfe6d317e69dac820565a7f82df8b54a9))
* update pkg/ociauth/resolver ([8a29b11](https://git.act3-ace.com/ace/data/tool/commit/8a29b113ced0b87976584a7310786b3c6589357f))
* update pkg/pull, pkg/push ([c65870d](https://git.act3-ace.com/ace/data/tool/commit/c65870d2baab8c98c670b70f7bf1240bbb260888))
* update pkg/status ([a037c0e](https://git.act3-ace.com/ace/data/tool/commit/a037c0ef741d83aeca50c76cd448b00ccc5453bd))
* update pkg/telemetry ([7a916b7](https://git.act3-ace.com/ace/data/tool/commit/7a916b7201e3a9c8e2d4b13ef38d1fba8a8e9023))
* update rest of code, only thing left is std libs ([bccc1a4](https://git.act3-ace.com/ace/data/tool/commit/bccc1a46eeab6d9b33f934ad579b2f9031a1974c))
* update test and fix bug with helper ([fc4c2f6](https://git.act3-ace.com/ace/data/tool/commit/fc4c2f672e18a6cda3880cbbf9bff328ba0b72e1))
* update tests for symlink ([91b8063](https://git.act3-ace.com/ace/data/tool/commit/91b80634c8118d96a8c3413111f7991758f259e8))
* update to project tool go-cli template ([52be8df](https://git.act3-ace.com/ace/data/tool/commit/52be8dfce4f9623af3f2794cbd0c56a954b0cebe))
* update with master ([24d0cde](https://git.act3-ace.com/ace/data/tool/commit/24d0cde74395d323b924c1a1f35405d5a9a8c0dd))
* update yml to yaml for consistancy ([5b29317](https://git.act3-ace.com/ace/data/tool/commit/5b293173f555dd8c0eb886f4503f2f5887e9a453))
* updated first init test, works :-) ([b40e86a](https://git.act3-ace.com/ace/data/tool/commit/b40e86a9da3a46401043476853fc51a0f386a70c))
* updated telemetry (without git LFS files) ([f1ca855](https://git.act3-ace.com/ace/data/tool/commit/f1ca85523b1b98852f9c7b425cd1d11fc2f3a12f))
* updated the docs to change -p to -d ([a768c51](https://git.act3-ace.com/ace/data/tool/commit/a768c513f3034ddc7ac35429bf5248192a061cb3))
* updated to work with schema@master ([41c67a7](https://git.act3-ace.com/ace/data/tool/commit/41c67a7b13045d06a43cbd05d3b60b6a3bf90de9))
* upgrade old bottles for labels yaml file name change, and ensure paths are created for files in subdirs ([41469a2](https://git.act3-ace.com/ace/data/tool/commit/41469a2870d3c1d30b5e335955cac7658b9b5cf4))
* upgraded to schema v1.0.0 ([467fe91](https://git.act3-ace.com/ace/data/tool/commit/467fe91af3d77264326a1d1efe522204a722c641))
* url's are not file paths ([67d1991](https://git.act3-ace.com/ace/data/tool/commit/67d19912333d8a8f3b58c6a9c1e1d42e1bc2f810))
* use in memory DB for functional test in CI ([6d05de7](https://git.act3-ace.com/ace/data/tool/commit/6d05de78328a9057ac010c318e3e11fbc362b35b))
* use k8s selectors (really the LabelSelectorSet from telemetry) ([3ffa18f](https://git.act3-ace.com/ace/data/tool/commit/3ffa18fc8166d68f33f23ba243fcd6d3da5097c0))
* use reference for generating the netrc ([6375f6d](https://git.act3-ace.com/ace/data/tool/commit/6375f6df033e1355be6f0317af55fea8604accb6))
* use telemetry's v1alpha1.Location ([f7af16b](https://git.act3-ace.com/ace/data/tool/commit/f7af16b6db94f52ac33c23e320f30cf5b4d18a86))
* use watchProgressTracker versus watch speed in ociPullImages ([419f953](https://git.act3-ace.com/ace/data/tool/commit/419f9533955d8e791846e18c50ee6d20204b8dda))
* verbose run ([41b70d8](https://git.act3-ace.com/ace/data/tool/commit/41b70d841406d7c870ca9614e3d82cf9af3a5dea))
* verbosity ([39a30f7](https://git.act3-ace.com/ace/data/tool/commit/39a30f714c3118149f6125c83923c3be97e07fa7))
* version info and relative paths ([4c97a49](https://git.act3-ace.com/ace/data/tool/commit/4c97a4965ff6d7156cb76dbd1af0c6895a1aca68))
* vet=all issue w.r.t. json tags ([0ff5819](https://git.act3-ace.com/ace/data/tool/commit/0ff5819830d24d81260e4cf3d484159ad6b458d2))
* we no longer use fixed ports in the makefile (to avoid conflicts) ([61a2fd2](https://git.act3-ace.com/ace/data/tool/commit/61a2fd20f5c052759996aaa774b05dcc9daac104))
* windows fixes for functional tests ([b3cd0d7](https://git.act3-ace.com/ace/data/tool/commit/b3cd0d747d86486233cb1d8c9d498afc0045d834))
* windows related unit test failures ([4b64d26](https://git.act3-ace.com/ace/data/tool/commit/4b64d26e16424196c98d12b977af59a47d25f13a))
* work on standardization and de-randomization ([2f76ee7](https://git.act3-ace.com/ace/data/tool/commit/2f76ee7ed0b01ef51522f5ca560f03755cf64c18))
* working on the integration tests ([edd3844](https://git.act3-ace.com/ace/data/tool/commit/edd3844df73102d6cbbc221d8f144e7cd990ade1))
* write test logs to file ([258d8ad](https://git.act3-ace.com/ace/data/tool/commit/258d8addd11f450fa4ad9cda4d27f2a8d434dec1))
* wrong fileExist check ([118053b](https://git.act3-ace.com/ace/data/tool/commit/118053bb9dc50fc999e3e05371b43d48a97262be))


### Features

* absolute paths to bottles in source add now resolve to bottleID ([743647a](https://git.act3-ace.com/ace/data/tool/commit/743647a01cf72fee822f1034d2bb3348e81233d7))
* added --dry-run ([02c1c08](https://git.act3-ace.com/ace/data/tool/commit/02c1c08148ab5a92460d0189127e33d006a8a1df))
* added a filter subcommand to make the filters baked into the executable ([3a54d30](https://git.act3-ace.com/ace/data/tool/commit/3a54d30313d9b1cbf843bf40a60918566e896fe4))
* added a timeout to SendEvent in the telemetry adaptor ([c3de580](https://git.act3-ace.com/ace/data/tool/commit/c3de580d0e7f7c8d88dc82fdf6043881a9b1ee8b))
* added auth support ([daeb370](https://git.act3-ace.com/ace/data/tool/commit/daeb370fb0877ec330992247bb21206796d895e3))
* added content digest checking at sync time ([2993080](https://git.act3-ace.com/ace/data/tool/commit/2993080b75730e6b31c0603e5598f523bdb33962))
* added multiple telemetry support ([6887650](https://git.act3-ace.com/ace/data/tool/commit/6887650877f0dc89c1a94f5261813822eaeada25))
* added our modified commands from crane to support alt install of ace/env ([d6a43ef](https://git.act3-ace.com/ace/data/tool/commit/d6a43ef8702f1bf1055f28aa0ad787bfc0104c67))
* added outputting (from Jack's code) ([34d9c8d](https://git.act3-ace.com/ace/data/tool/commit/34d9c8dc5232883c4752953c999cf7d78197742a))
* added requirements processing ([c0cee30](https://git.act3-ace.com/ace/data/tool/commit/c0cee305e08cb5cbe4226f936a3a0debf2d516ed))
* added source distribution support ([8a13f09](https://git.act3-ace.com/ace/data/tool/commit/8a13f0992896615e7885b4d0a82ddc544efac1ad))
* added support for multiple python package indices ([15438cd](https://git.act3-ace.com/ace/data/tool/commit/15438cd1a57dd63cb8b0ea19220e637344ac30d7))
* added support for multiple requirement files ([280847c](https://git.act3-ace.com/ace/data/tool/commit/280847c0aa15351e6a942be0c9085f7e148bfd5e))
* added the abilty to specify package names on the command line ([767b6ee](https://git.act3-ace.com/ace/data/tool/commit/767b6eeceb029ee3c640c59a4d34860236b13dbb))
* added yanked support ([7841dc5](https://git.act3-ace.com/ace/data/tool/commit/7841dc5ca434570deab1834499c0b54a56466c9e))
* check if manifest exists with expected digest at a given location, short circuiting the push if already there ([267ccab](https://git.act3-ace.com/ace/data/tool/commit/267ccab5dceab5355dff25350710934acff1740d))
* output bottle config json to config directory ([4685410](https://git.act3-ace.com/ace/data/tool/commit/46854100afb355ecbc4b102ce2b6953e4ef683d0))
* print out bottle manifest to .dt dir ([100b4e2](https://git.act3-ace.com/ace/data/tool/commit/100b4e20c923e3a08812ce7f314cfb4cc1d34f98))
* upgraded schema to include the new v1 bottle config ([16dc4b3](https://git.act3-ace.com/ace/data/tool/commit/16dc4b3cef7905f1486d391b752485981807938d))


### BREAKING CHANGES

* directory archive part names are now appended with a /, this is a transmogrithing security concern update which breaks backward compatibility,

## [0.25.4](https://git.act3-ace.com/ace/data/tool/compare/v0.25.3...v0.25.4) (2022-03-24)


### Bug Fixes

* remove debug image build from pipeline.  DO_RELEASE=true ([dbcfff7](https://git.act3-ace.com/ace/data/tool/commit/dbcfff739dfd3c9ebaf88fc52697bbaa4da8761b))

## [0.25.3](https://git.act3-ace.com/ace/data/tool/compare/v0.25.2...v0.25.3) (2022-03-24)


### Bug Fixes

*  add a  fix and a readme mention of the fix to get semantic release's attention.  DO_RELEASE=true ([114f795](https://git.act3-ace.com/ace/data/tool/commit/114f795a1328a89319342ba9351530c8318bb63b))

## [0.25.2](https://git.act3-ace.com/ace/data/tool/compare/v0.25.1...v0.25.2) (2022-03-23)


### Bug Fixes

* correct selector set to selector list matching logic ([41286e7](https://git.act3-ace.com/ace/data/tool/commit/41286e72311b5bfb48e3acb91a77b2cd5b2f6ff9))
* unit tests for failing funcs ([4dae836](https://git.act3-ace.com/ace/data/tool/commit/4dae836ada360772f9c98560f2274b8d3731eb56))

## [0.25.1](https://git.act3-ace.com/ace/data/tool/compare/v0.25.0...v0.25.1) (2022-03-22)


### Bug Fixes

* apply labels to parts within archive parts, deconstructing archived part as necessary ([cd00dce](https://git.act3-ace.com/ace/data/tool/commit/cd00dceed86c0ba0590f8a9b724bfcda91f5c92e))
* authorizer errors during pull ([dec6a97](https://git.act3-ace.com/ace/data/tool/commit/dec6a97677eb1afb4bdfd124c170a35d26350b03))
* config sample ([22d46c4](https://git.act3-ace.com/ace/data/tool/commit/22d46c4d46d07f294c1ab769ca3615012a78317b))
* correct error message string for part not found, to match test ([088ba57](https://git.act3-ace.com/ace/data/tool/commit/088ba57397590e5fb29554bedba040f5f89b7461))
* explicitly ignore response body close errors in push/layers ([1dc9fe7](https://git.act3-ace.com/ace/data/tool/commit/1dc9fe74ce4802f187637fb59e2362f8a4e52ba8))
* lint fix ([1468581](https://git.act3-ace.com/ace/data/tool/commit/14685813e81f2ccded81bb07786aa3038deaab22))
* made ErrTelemetry public ([cdfba2c](https://git.act3-ace.com/ace/data/tool/commit/cdfba2c9f23c758a3a19c5142b5fc2541cb9551a))
* mismatched key/value pairs for logging ([48cb430](https://git.act3-ace.com/ace/data/tool/commit/48cb43069bcd53f05301b01074fc34a218488004))
* missing comment documentation for new WithAuthCreds function ([b33f00e](https://git.act3-ace.com/ace/data/tool/commit/b33f00e211a89d78fd0427072171653ce432cf6a))
* remove deprecated ace-dt commands ([92a399c](https://git.act3-ace.com/ace/data/tool/commit/92a399cfb8a010657e7dc41ebc61b1e0a94d5e7f))
* remove deprecated NoExpand NoDigest pull flags ([283ccef](https://git.act3-ace.com/ace/data/tool/commit/283ccefbc01a4c77552e93d0945af180031deacf))
* remove more nodigest noexpand logic ([420018a](https://git.act3-ace.com/ace/data/tool/commit/420018ad5f830354640965d84ecdec27402fd4af))
* remove old + unused repolist logic ([aec9fca](https://git.act3-ace.com/ace/data/tool/commit/aec9fca8fd61b8f576fa3a821e243b84b21226f3))
* remove old functions used in removed ace-dt commands ([3c56dba](https://git.act3-ace.com/ace/data/tool/commit/3c56dba14c53e70d3cfba9af8d68ce6859beb7fe))
* remove repo pkg, move contents to bottleref pkg ([4bfe14b](https://git.act3-ace.com/ace/data/tool/commit/4bfe14b544cad49874043e446deae10bb0227386))
* remove suite from ace-dt ([c2d8a8f](https://git.act3-ace.com/ace/data/tool/commit/c2d8a8f121733bbd95ffb2a7d29a39ae75362d10))
* remove unused gzip pipe stream reader function ([1e3d446](https://git.act3-ace.com/ace/data/tool/commit/1e3d446bca4e4ed9243975e64854681c7f784d65))
* upgrade to GO 1.18 ([2978594](https://git.act3-ace.com/ace/data/tool/commit/2978594ebaf5ac4c4da862e78af1aac434b48aaf))
* validator + test ([5ef8f85](https://git.act3-ace.com/ace/data/tool/commit/5ef8f859bdae9a1f8c8bdbb9f9830f80041d2de6))

# [0.25.0](https://git.act3-ace.com/ace/data/tool/compare/v0.24.3...v0.25.0) (2022-03-17)


### Bug Fixes

* bad func call ([3e0e9d6](https://git.act3-ace.com/ace/data/tool/commit/3e0e9d672c567ea9ac26125de10b0baf262b5f74))
* bad len() expression ([3b20271](https://git.act3-ace.com/ace/data/tool/commit/3b20271866ffcb3e82d3c458f8d860ffc614401d))
* combine special pull / normal pull. fix logging error ([c1ef1bc](https://git.act3-ace.com/ace/data/tool/commit/c1ef1bc9d9ce7cf42703fbba4f12a08024e1f7f0))
* correct panic when archiving virtual parts, and a few thread synchronization issues ([c8df585](https://git.act3-ace.com/ace/data/tool/commit/c8df585e56702ddad02076b1e12b9359f61c09e1))
* go mod tidy ([21bb6bc](https://git.act3-ace.com/ace/data/tool/commit/21bb6bc1547f550b77162bf503dd5616b31409d3))
* lint issues + docs ([9702c0b](https://git.act3-ace.com/ace/data/tool/commit/9702c0b7bdfa71ee1213186d55af728eaca8d968))
* mark bottle dirty when virtual parts are present, to ensure the virtual part info is saved ([8fddb73](https://git.act3-ace.com/ace/data/tool/commit/8fddb73123ce7dc69068fddf69495ec39f7c32ed))
* merge with master ([60b6fb1](https://git.act3-ace.com/ace/data/tool/commit/60b6fb1dc7bd3b9930f38e50bcfb7eb06dfd3728))
* merge with master ([23cbfd1](https://git.act3-ace.com/ace/data/tool/commit/23cbfd1f3f4e2aafb9ee606b99aec855819a9e86))
* merge with master ([595ab8d](https://git.act3-ace.com/ace/data/tool/commit/595ab8d6f92c46ccada3dc97d0126f54372f8a5a))
* merge with master ([845aa5e](https://git.act3-ace.com/ace/data/tool/commit/845aa5e74a5aed0aea19fe50d83ee2162ac4287c))
* move bottle validation out of LoadAndUpgrade and before push ([2b66b52](https://git.act3-ace.com/ace/data/tool/commit/2b66b52bfdf8deb7a7ed705337ce0b102dfaf1dc))
* move pull cmd func out of cmd pkg ([b3f1bb0](https://git.act3-ace.com/ace/data/tool/commit/b3f1bb0cbe413ebcd6d303600a1ff7c6a355beb3))
* moved config to pkg level, cleaned up bottle pull params, refactored localPullOpts ([719899c](https://git.act3-ace.com/ace/data/tool/commit/719899c2a6b778bc3bb41be4903f1137d4824dad))
* properly track telem flag changes + changed log.error to log.info on ok telem bottle search failure ([eac3f8f](https://git.act3-ace.com/ace/data/tool/commit/eac3f8fc4024b67f5229e343f0d2eb73cf968f71))
* protect against nil pointer dereference ([91b0673](https://git.act3-ace.com/ace/data/tool/commit/91b0673ba53b8023648206cfd0693c7b15703f0f))
* re-add write-bottle-id flag to bottle pull subcommand ([d611b27](https://git.act3-ace.com/ace/data/tool/commit/d611b27eaafafdf42f66e232ac1bf0fbb9ce9e5a))
* replace viper with aconfig start ([7626ecc](https://git.act3-ace.com/ace/data/tool/commit/7626ecc5a2f0de34b43737df51b5f967565b91ce))
* standardize docs, add more examples ([bbaef31](https://git.act3-ace.com/ace/data/tool/commit/bbaef318066cac174c03ef3a7ab80ea3fdca64ad))
* todos ([e4654d4](https://git.act3-ace.com/ace/data/tool/commit/e4654d476905cb374ed72d38c84ad92ae36d684c))
* tolerate missing parts.json files after migrating older bottles, and re-fix migration version result check ([4117f44](https://git.act3-ace.com/ace/data/tool/commit/4117f44ccf00b5b44fa9a88bd362259ef573fbe4))
* update + standardize docs ([488f493](https://git.act3-ace.com/ace/data/tool/commit/488f493875089947f06104da4e70fb947ab88e39))
* update csi-driver pull function ([c103473](https://git.act3-ace.com/ace/data/tool/commit/c10347305181708647e611a29735cf6728192ec9))
* update CsiPull to expect csi-driver to set log and config ([e8ddb82](https://git.act3-ace.com/ace/data/tool/commit/e8ddb820fe49137390909f29f7a8aea4318b1c44))
* update examples for new flags, edit link to usage.md ([67e7fba](https://git.act3-ace.com/ace/data/tool/commit/67e7fbae05608e44b43b168156ccf5863e0a7aec))
* update telemetry logic in pull func ([d03b1d2](https://git.act3-ace.com/ace/data/tool/commit/d03b1d2eeb8aa7ed4a4c242967dab7b0a036218f))
* update usage.md ([5d5f050](https://git.act3-ace.com/ace/data/tool/commit/5d5f0500f81d17e624270603e9f1acb5cc8e7837))
* use config definition in pull struct ([9626a10](https://git.act3-ace.com/ace/data/tool/commit/9626a10e5cf1194534a76328df00ca8480a78aa2))
* use schema validation to validate bottle data instead of custom validator code ([e0f0f11](https://git.act3-ace.com/ace/data/tool/commit/e0f0f110e930d999c766b8299634f12436102722))
* we leaked file descriptors ([b4c8b6a](https://git.act3-ace.com/ace/data/tool/commit/b4c8b6a8cfb2943b23bd5f9451f69048bbaff691))


### Features

* enable virtual parts to be pulled from a remote host if not found in local cache ([d93b90b](https://git.act3-ace.com/ace/data/tool/commit/d93b90be84d35a5501838b2b051d412dfa26a5eb))

## [0.24.3](https://git.act3-ace.com/ace/data/tool/compare/v0.24.2...v0.24.3) (2022-02-24)


### Bug Fixes

* added main.go to releaserc ([10d8b88](https://git.act3-ace.com/ace/data/tool/commit/10d8b8834261d21d47d384ffcacebbf36a47cd2c))
* baked in the version ([d45d708](https://git.act3-ace.com/ace/data/tool/commit/d45d708cdcdd51cbd4097703c74799e5ce1dd5b2))

## [0.24.2](https://git.act3-ace.com/ace/data/tool/compare/v0.24.1...v0.24.2) (2022-02-24)


### Bug Fixes

* removed "replace" from go.mod and updated deps ([fb3d6b4](https://git.act3-ace.com/ace/data/tool/commit/fb3d6b4a0fd76db947be968e30f587df3906e1eb))

## [0.24.1](https://git.act3-ace.com/ace/data/tool/compare/v0.24.0...v0.24.1) (2022-02-24)


### Bug Fixes

* adding a fix to trick semantic release into processing the release.  DO_RELEASE=true ([3968ddd](https://git.act3-ace.com/ace/data/tool/commit/3968ddd2460af0cb836ea1b0f4fb6e7af23f256e))

# [0.24.0](https://git.act3-ace.com/ace/data/tool/compare/v0.23.0...v0.24.0) (2022-02-23)


### Bug Fixes

* add cmd name to function test log file ([a813995](https://git.act3-ace.com/ace/data/tool/commit/a813995b0ce29a4df65aa09a9f4dee263d937079))
* bandaid case mismatches during yaml validation, and use ToDocumentedYAML to save yaml config ([08b236c](https://git.act3-ace.com/ace/data/tool/commit/08b236cfeb31ed00925523cc0d22551403d9c017))
* convert yaml to json prior to performing migration ([163d813](https://git.act3-ace.com/ace/data/tool/commit/163d813a659cab3ac53e51db0910a8d70277aef5))
* crash when accessing empty digest for encoded form ([e1344c1](https://git.act3-ace.com/ace/data/tool/commit/e1344c1248216e14bd3f03f07509e6be33035353))
* read returned bottle API Version from migration versus version string.  DO_RELEASE=true ([2c1260d](https://git.act3-ace.com/ace/data/tool/commit/2c1260d36dc05df5121a37d171109464fddefc6d))
* rmove local parts information to ace-dt instead of schema ([8414dc2](https://git.act3-ace.com/ace/data/tool/commit/8414dc26f8fdeb915f68355376f0bb7a0e6693ac))
* sweeping update due to schema digest and migration changes ([12dc628](https://git.act3-ace.com/ace/data/tool/commit/12dc62821715b2e07951dfead0635984fb951869))
* switch to using sigs.k8s.io yaml parser for reading bottle def from yaml ([fe8f030](https://git.act3-ace.com/ace/data/tool/commit/fe8f030f98aa13148b4a36c8d0fe902f717e497d))
* update to schema 0.2.10 ([cfa1fa6](https://git.act3-ace.com/ace/data/tool/commit/cfa1fa6f6fd229fc291e33167adbe4f2ebbccbf8))


### Features

* **mirror-telemetry:** add telemetry to mirror push if an object is a bottle ([a995de0](https://git.act3-ace.com/ace/data/tool/commit/a995de0885cb2cd6955958e56dca5732acaf4314))

# [0.23.0](https://git.act3-ace.com/ace/data/tool/compare/v0.22.0...v0.23.0) (2022-01-31)


### Bug Fixes

* stop trying to pull bottle after success when using telemetry query for location ([8d96447](https://git.act3-ace.com/ace/data/tool/commit/8d96447b682c2a068eb72fed60c3fc83cc92a1e4))


### Features

* **fragment-selectors:** allow uri fragments to specify selectors on bottle... ([074ebbf](https://git.act3-ace.com/ace/data/tool/commit/074ebbf85ee7837b46b9cf144b240612fa8107c6))

# [0.22.0](https://git.act3-ace.com/ace/data/tool/compare/v0.21.0...v0.22.0) (2022-01-26)


### Bug Fixes

* added a check to the integration test ([3c322db](https://git.act3-ace.com/ace/data/tool/commit/3c322dbf228d39ac568c815c3eda5bcbc4805c3c)), closes [#158](https://git.act3-ace.com/ace/data/tool/issues/158)
* added new func for exported (other project) pull functionalily ([2277a5b](https://git.act3-ace.com/ace/data/tool/commit/2277a5b5f47c5ae75327dd0a68ac6173a168ee99))
* all easy lint issues ([1f1ddfc](https://git.act3-ace.com/ace/data/tool/commit/1f1ddfc3e9438f5a926e02696ca26929d7e3098e))
* allow insecure for local testing ([06cb24c](https://git.act3-ace.com/ace/data/tool/commit/06cb24c690ffa90aee5aade0b99a849afcc7531a))
* calculate manifest digest when raw manifest data is set in bottle ([6d26001](https://git.act3-ace.com/ace/data/tool/commit/6d26001192a9c4d82c893c2ef982545432805037))
* curl exit code ([2912305](https://git.act3-ace.com/ace/data/tool/commit/2912305e656bf4ada704d675692b78e797b8a6d2))
* event field names ([b1609f3](https://git.act3-ace.com/ace/data/tool/commit/b1609f3ea4aa039d5fd6573fbd29241f129fdcca))
* exit code on failed pull with bottle id was incorrect ([ab8e4ce](https://git.act3-ace.com/ace/data/tool/commit/ab8e4ce9fb53a86a121e0d1c22aa3f766eb86cfc))
* first draft at pullStruct ([b849f95](https://git.act3-ace.com/ace/data/tool/commit/b849f9517a63d58788c656432b53db7f5dcb2ffa))
* **insecure-reg:** Add --allow-insecure flag to pull command to enable tls downgrade for insecure registries ([606eabc](https://git.act3-ace.com/ace/data/tool/commit/606eabcf0e8bc934a46770349d11e3755935ad77))
* merge with master, fix merge conflicts ([50fce80](https://git.act3-ace.com/ace/data/tool/commit/50fce803cb244b8f1cd306994e5d2c8cbdb8c6f3))
* **migration:** rework version migration to be independent of current local bottle version ([c080843](https://git.act3-ace.com/ace/data/tool/commit/c080843a9056da5901d791c20e7458e7bde8fe7d))
* no progress for exported func ([a44735f](https://git.act3-ace.com/ace/data/tool/commit/a44735fd208f672b6718d178aae335e844ea42cd))
* rewind manifest data buffer for authorization retry ([1e4bd67](https://git.act3-ace.com/ace/data/tool/commit/1e4bd67dc05d3b97189826647a513c50a94ae18d))
* satisfy my golint errors ([fd8a1ad](https://git.act3-ace.com/ace/data/tool/commit/fd8a1ad50ca07632523381844b6a61d7f4df5760))
* seg fault in telemetry ([ec73059](https://git.act3-ace.com/ace/data/tool/commit/ec730594f06ebd0ee0b0d421b0decad948e4687b))
* set to initial v0.1.0 version for ace bottle schema ([2fb1bd9](https://git.act3-ace.com/ace/data/tool/commit/2fb1bd9dc6a68efeeecbc7af83a0fa8922138344))
* un-export unnecessary exported functions ([5cdab67](https://git.act3-ace.com/ace/data/tool/commit/5cdab67b18049c2abedabc3781ffe9e7568c9627))
* update debug logging ([4eca13d](https://git.act3-ace.com/ace/data/tool/commit/4eca13dcd42e90f790a030907f869f499ca448e0))
* update find bottle with telemtry server block, pull out logic from NewPullWithEvent ([4725c6d](https://git.act3-ace.com/ace/data/tool/commit/4725c6d44191850c45be918a61cfc9864f58369a))
* update max_http_connections ([d9256ec](https://git.act3-ace.com/ace/data/tool/commit/d9256ec6197bded5816fadab39be27b17b881dda))
* use the correct selectors, not a global var ([89be5c2](https://git.act3-ace.com/ace/data/tool/commit/89be5c23c6aa5a4e821380e590e466d160c671ac))
* **validate:** allow empty values in URL fields to pass validation ([37506d7](https://git.act3-ace.com/ace/data/tool/commit/37506d78e5c7b44f5e25f1aeb0a16aff3980ef1d))
* when the the value for "ace-dt meta set" contains an "=" ([3777f2d](https://git.act3-ace.com/ace/data/tool/commit/3777f2dd67e4e29c77f1fc548947ba775e45c989))


### Features

* add support for configurable external pull requests ([e148181](https://git.act3-ace.com/ace/data/tool/commit/e1481814d69826d3128a0639a6e762dc7e43b6d1))
* **check-bottleid:** add flag to bottle pull to return an error if recieved bottle config data does not match given bottleID ([dc65ddf](https://git.act3-ace.com/ace/data/tool/commit/dc65ddfffd7cccc0650c66bf8f18a7a7166aee83))
* **login:** add the ability to specify username and password on the command line ([d6307a0](https://git.act3-ace.com/ace/data/tool/commit/d6307a028cd534db613fa4afa38b3f26daa560b1))
* **part-attribute:** add part attributes and allow regular expressions for... ([5ccac7c](https://git.act3-ace.com/ace/data/tool/commit/5ccac7cf32f4d47a33ab999ed46ba8b916bd69f7))
* **tel-fail-opt:** set --report-telem-failure on bottle commands to report telemetry failures instead of silently proceeding ([ab9c54a](https://git.act3-ace.com/ace/data/tool/commit/ab9c54a61ea5a311bb88d0284169083e3e8ec5bf))

# [0.21.0](https://git.act3-ace.com/ace/data/tool/compare/v0.20.1...v0.21.0) (2021-12-17)


### Bug Fixes

* add timestamp + username to event. Fix manifestDisgest in event. ([804da8d](https://git.act3-ace.com/ace/data/tool/commit/804da8dafc92435d4c85306b522dc43e6ef58cc3))
* address lint errors ([cad7bf4](https://git.act3-ace.com/ace/data/tool/commit/cad7bf4a24a4c31df9a683860a84095744320eab))
* address lint errors2 ([248d8b0](https://git.act3-ace.com/ace/data/tool/commit/248d8b0f083c9cd81bb0a00832187fbef902737e))
* better labels, and repo host on event ([e6d975e](https://git.act3-ace.com/ace/data/tool/commit/e6d975e2e27666a91fe51292b7e639a17157318c))
* chain layer transfer notification callback to support both blobinfocache and external callbacks during QueueLayers ([0c62f1d](https://git.act3-ace.com/ace/data/tool/commit/0c62f1d9f3d9b701c17a1db571811b831e4bb531))
* change repo to avoid malformed bottle path ([2307449](https://git.act3-ace.com/ace/data/tool/commit/23074492fbe76a96f889376f54ca37128d5364c5))
* change yaml struct tags to json for part track structure ([2f67de1](https://git.act3-ace.com/ace/data/tool/commit/2f67de1de996a99fc6f833bc5129594205fc4aeb))
* correct manifest digest format in delete command ([3006ad0](https://git.act3-ace.com/ace/data/tool/commit/3006ad0249302a9b5b20251e35bf10537de8011f))
* correct newest schema version to v1beta1 ([6b4f338](https://git.act3-ace.com/ace/data/tool/commit/6b4f338d5d64ef68af731de63118b47417bb213c))
* doc fix ([4c2fad1](https://git.act3-ace.com/ace/data/tool/commit/4c2fad1389ec6f20bf37ccbd4c0b612a6e1333b1))
* doc fix again ([ba2e373](https://git.act3-ace.com/ace/data/tool/commit/ba2e3732080a0967d1499bbd84b1690a6989c82d))
* golint errors ([9b73fe4](https://git.act3-ace.com/ace/data/tool/commit/9b73fe4c7385a7f266ee7eff55babf847e0fd22d))
* golint errors2 ([da4844a](https://git.act3-ace.com/ace/data/tool/commit/da4844a80cae39d6d0f190f358339668f5ff2d15))
* move --write-bottle-id parameter to the bottle command group, enabling the output for all bottle commands ([ff9429b](https://git.act3-ace.com/ace/data/tool/commit/ff9429b4e98c6b74d4888c69b8cf07afa2968a52))
* move meta commands to bottle. add label command to bottle to directly edit labels if in bottle root dir ([bc345d1](https://git.act3-ace.com/ace/data/tool/commit/bc345d1a5f5037b028686f4f391e1e474008a865))
* refactor and remove remaining oras dependencies from push processes ([72b61dc](https://git.act3-ace.com/ace/data/tool/commit/72b61dcc701903964f82a9ad48433f8b43cb6ef7))
* remove unused arg from mapsetter ([8f98cd2](https://git.act3-ace.com/ace/data/tool/commit/8f98cd2ca703faa6406d370bc45bd0688dbf9d6e))
* satisfy lint errors ([b74ba74](https://git.act3-ace.com/ace/data/tool/commit/b74ba742d740720b89c245c40a86cf2f673faf83))
* test linter ([ad70c1a](https://git.act3-ace.com/ace/data/tool/commit/ad70c1a0efcaa391da4e0822355b678948d472b9))
* update username setting, fail if no username ([d76fc6f](https://git.act3-ace.com/ace/data/tool/commit/d76fc6fbf5a4d15e09f734801a78e32684206fb5))
* updated docs ([bd3935b](https://git.act3-ace.com/ace/data/tool/commit/bd3935be6ae11e2e38bc1ed358619903d9abceaa))
* **yaml-cleanup:** remove parts from bottle yaml and track in a separate parts.json file ([c8eb703](https://git.act3-ace.com/ace/data/tool/commit/c8eb7031e55bf23a0449a956f1c598f4352478dd))


### Features

* **blobinfocache:** Implement blobinfocache from containers, using boltdb local blob location tracking ([20a4141](https://git.act3-ace.com/ace/data/tool/commit/20a4141552af51372f2454f065f064f7caf42d31))
* **bottleID-write:** add a command line argument --write-bottle-id that accepts a filename to save a bottleid after push ([1ec67da](https://git.act3-ace.com/ace/data/tool/commit/1ec67dabc0b6139decbf364890aaceda9b9531af))
* **local-show:** add the ability to produce bottle show output for a local bottle directory ([c1196ed](https://git.act3-ace.com/ace/data/tool/commit/c1196ededb0edb401889572ad7271938ebfe27e4))

## [0.20.1](https://git.act3-ace.com/ace/data/tool/compare/v0.20.0...v0.20.1) (2021-12-02)


### Bug Fixes

* add env vars to test jobs ([37c685e](https://git.act3-ace.com/ace/data/tool/commit/37c685ec39495162ce095f27f0d013573fdf57f1))
* added schemes to author info, updated to use latest scheme, updates sources seedData ([c415c27](https://git.act3-ace.com/ace/data/tool/commit/c415c27f1a8477e1be220614dc09670d56b3bab5))
* change regex for GOFLAGS test env var ([bc2b7b8](https://git.act3-ace.com/ace/data/tool/commit/bc2b7b884fc30d9699f68c71f2e0bcf8e41387f2))
* gitlabci issues ([180befa](https://git.act3-ace.com/ace/data/tool/commit/180befa4e93d76acbf83506b597953e7670aabfb))
* gitlabci issues 2 ([e326533](https://git.act3-ace.com/ace/data/tool/commit/e326533349b2474504902171e7cc8b0fe748747e))
* **log:** set log level to full verbosity for log files, separate from console log level ([9cf013f](https://git.act3-ace.com/ace/data/tool/commit/9cf013ff0bd30f88451ef2e447fe64d22c126cf0))
* merge with master ([0c20890](https://git.act3-ace.com/ace/data/tool/commit/0c208904ec7d6d59a86b9024bd2168acc94e5adb))
* **mirrorfetch:** Update manifest digest instead of adding new if source/decl match but digest has changed ([8a4310a](https://git.act3-ace.com/ace/data/tool/commit/8a4310a03c922ae565330617cf36e865abf87bc4))
* removed gitattributes ([6efbbd3](https://git.act3-ace.com/ace/data/tool/commit/6efbbd37602586af2c3bebd865ce1fe0c6668420))
* rename tests to fit regex matching ([1bed169](https://git.act3-ace.com/ace/data/tool/commit/1bed169cf3507176e578d34cbeaabadb084bd371))
* return functional test to original ./... ([7579129](https://git.act3-ace.com/ace/data/tool/commit/7579129756ff686bc6eb57b9da682c54a384a933))
* update functional test definition to use latest golang conventions ([2938744](https://git.act3-ace.com/ace/data/tool/commit/2938744dee5b837cdbbda6a4c160f9ee83d2aa47))
* update new tests to include _Unit_ ([4836725](https://git.act3-ace.com/ace/data/tool/commit/4836725b1487261f8e8e5f0ebe3009aa7fae807d))
* where did the tests go from ci display? ([83260ce](https://git.act3-ace.com/ace/data/tool/commit/83260ce21583248b4aa907d4fe44b2d757904e30))

# [0.20.0](https://git.act3-ace.com/ace/data/tool/compare/v0.19.0...v0.20.0) (2021-11-22)


### Bug Fixes

* add comment to exported LinkBottles ([b914eb7](https://git.act3-ace.com/ace/data/tool/commit/b914eb745be3b5f037e80fbc358835271e9d11ac))
* added more comments ([1190505](https://git.act3-ace.com/ace/data/tool/commit/119050512fb079def56b30bdf7f35c59f6041f63))
* **fetch:** add download resume when pulling large data blobs ([e7ed228](https://git.act3-ace.com/ace/data/tool/commit/e7ed228a3d350f8823d3a59ce87e9b8c670e674a))
* final clean up ([76b0f7e](https://git.act3-ace.com/ace/data/tool/commit/76b0f7e6d0784e53681c4f1a94c872992778628b))
* pull fetching artifacts into a shell scripts, genbottle now checks if supplied folder exists ([b9b705a](https://git.act3-ace.com/ace/data/tool/commit/b9b705aa3fbdab9115b98d363079fedcf6d21709))
* push fetch artifact script ([6601a1c](https://git.act3-ace.com/ace/data/tool/commit/6601a1ccce53fd5264c21a37ef4a1756a5cd3f0d))
* removed empty code file, updated gitignore, addressed lint errors, preshufflings files prior to bottle addition ([3e2c864](https://git.act3-ace.com/ace/data/tool/commit/3e2c864e208d2351457dee8c3e39490358b85c8a))


### Features

* added ability to add samples artificats ([907eb82](https://git.act3-ace.com/ace/data/tool/commit/907eb820e795ba4dc2363dec2df3b0c0924ae6b5))
* added capability to link bottles at generation in DAG fashion ([206dd8a](https://git.act3-ace.com/ace/data/tool/commit/206dd8a1ff77963f0e7bafba5e20cb6fae79e55d))
* **bottlelocate:** add ability to query telemetry for bottle based on bottleID ([22b2dfa](https://git.act3-ace.com/ace/data/tool/commit/22b2dfadf906d88e175b9d740a759f46606c87a7))
* generated and linked documentation of commands ([1af04b0](https://git.act3-ace.com/ace/data/tool/commit/1af04b08cbf5dcdbd750b0da2bd50f9c3ddc7be3))

# [0.19.0](https://git.act3-ace.com/ace/data/tool/compare/v0.18.4...v0.19.0) (2021-11-10)


### Bug Fixes

* add a few fixes for generating bottles ([051e49a](https://git.act3-ace.com/ace/data/tool/commit/051e49ae71cece064fd5443ebb4766174c16ac13))
* add back nil check for time recorder ([e6fea79](https://git.act3-ace.com/ace/data/tool/commit/e6fea79f86ec4eef394b33c996b090d971913534))
* add better metadata for bottles ([5290c14](https://git.act3-ace.com/ace/data/tool/commit/5290c14b23a536ceeb7bad97fd02156ac1a6a70b))
* add digest to public artifact ([b856df6](https://git.act3-ace.com/ace/data/tool/commit/b856df6182d0236eba6a5fd18304836ba0736a7f))
* add flag for csv seed data ([2ad0e8a](https://git.act3-ace.com/ace/data/tool/commit/2ad0e8a60d8b2ce7b6ee8a51d89ddc4dfe2c7279))
* add manifest raw data to event, remove bottleID, and other API adjustments ([4829cb9](https://git.act3-ace.com/ace/data/tool/commit/4829cb9cb7537ae530e6eac5843c7d8b9dd6fe59))
* add manifest raw data to event, remove bottleID, and other API adjustments ([c3f5acd](https://git.act3-ace.com/ace/data/tool/commit/c3f5acdc0c4d82042b07c8c405f2f5448901b6ef))
* add time durations and archive file data to telemetry events ([31fa5ce](https://git.act3-ace.com/ace/data/tool/commit/31fa5ce17ba0093f7dd76fae3568fd37388ee1d3))
* add time durations and archive file data to telemetry events ([ee0b676](https://git.act3-ace.com/ace/data/tool/commit/ee0b6767fab65a91196e0544f38f76a2fb229a63))
* added marker to skip linting ([b01cfce](https://git.act3-ace.com/ace/data/tool/commit/b01cfcec78fb8fdd47ff50a37db84696f2c58c73))
* address most golang-ci lint issues ([2c78a7d](https://git.act3-ace.com/ace/data/tool/commit/2c78a7debd610f9702c94f843728b85c0c0b2284))
* don't assume bottles have authors ([645d738](https://git.act3-ace.com/ace/data/tool/commit/645d7384919106a5c81a2de2a368fe46b66da9e5))
* fix labels to match k8s labels. Fix typos ([e8f0de9](https://git.act3-ace.com/ace/data/tool/commit/e8f0de926ba6a290d28b7895fd0f28efe5483906))
* include testvals.csv to project ([3d9c5a4](https://git.act3-ace.com/ace/data/tool/commit/3d9c5a4696db6f7963d028480fa31d6c4bf6fd50))
* query construction improvements in SendEvent ([32578e9](https://git.act3-ace.com/ace/data/tool/commit/32578e97ff50ca183998eac4586c42c82cc0f202))
* query construction improvements in SendEvent ([d9bcf4f](https://git.act3-ace.com/ace/data/tool/commit/d9bcf4f9a3fef1b350910295b90c39b7013880e4))
* remove keywords, replace with annotations ([15a9686](https://git.act3-ace.com/ace/data/tool/commit/15a9686be7b4265fb0534366ff1643cc9b1273af))
* remove output from command ([6326260](https://git.act3-ace.com/ace/data/tool/commit/6326260a6fba6c2963fe2f0058f839a7a8bd5122))
* removed empty branch from if stmt ([e52a5b7](https://git.act3-ace.com/ace/data/tool/commit/e52a5b76fa761b8fd7379ae335a5c0ddbc63e6af))
* removed more instances of keywords in tests and logic ([d16da18](https://git.act3-ace.com/ace/data/tool/commit/d16da1862f8b37ae14e640cdc31f3e35c18ae1ae))
* resolve cache during genbottles init. ([8feb464](https://git.act3-ace.com/ace/data/tool/commit/8feb464b3be54fd8eeda9ebeab004319148b779b))
* resolve merge conflicts with master ([c6d5691](https://git.act3-ace.com/ace/data/tool/commit/c6d5691468239c8c5bfa4c4dde534ceaec4d60c8))
* revert fully to previous function ([a1bed1b](https://git.act3-ace.com/ace/data/tool/commit/a1bed1bb35413f4193039a02c1f5f3a63b0f2935))
* reverted too much ([f851983](https://git.act3-ace.com/ace/data/tool/commit/f8519835bee8ecc8d694d7423a749b76c0376301))
* small changes and fixes for integrating with testing telemetry server ([37d3de0](https://git.act3-ace.com/ace/data/tool/commit/37d3de0a777edb064fd3f0d77a04b02428cd597d))
* small changes and fixes for integrating with testing telemetry server ([17629d1](https://git.act3-ace.com/ace/data/tool/commit/17629d1768331c2f1112db61626358f417753e9b))
* supress usage if the command actually runs and returns an error ([a252e65](https://git.act3-ace.com/ace/data/tool/commit/a252e6562c261c53eaa8db341bdf75344770b4f4))
* **telemetry:** refactor telemetry to match expected api ([996f36b](https://git.act3-ace.com/ace/data/tool/commit/996f36b814195fae5a234f7ca5a9633d3426ba87))
* **telemetry:** refactor telemetry to match expected api ([6a50a97](https://git.act3-ace.com/ace/data/tool/commit/6a50a97abe5bfb4a27a1045d3608b2b0c7fae876))
* update help documentation for config cmd changes ([9e3f3de](https://git.act3-ace.com/ace/data/tool/commit/9e3f3deb82197ace49f08aac26e108574c91a9e6))
* update help documentation for config cmd changes ([48dbb0c](https://git.act3-ace.com/ace/data/tool/commit/48dbb0cbb5b0e5e2e3247c39fe9dfb0d60dbb769))
* updated contact information to use example.com domain name. ([2e251ac](https://git.act3-ace.com/ace/data/tool/commit/2e251ace58dce6de5688f848d30bbfb3e641d6c6))
* updated event method ([70fc104](https://git.act3-ace.com/ace/data/tool/commit/70fc10435034f9189e33536259f07919d7861303))
* use url parser for telemetry SendEvent vs silly manual parsing ([7f2e5cf](https://git.act3-ace.com/ace/data/tool/commit/7f2e5cfdba0f4d4c98ed359f306453270bc9cc80))
* use url parser for telemetry SendEvent vs silly manual parsing ([ba0f585](https://git.act3-ace.com/ace/data/tool/commit/ba0f5857150a6a3989fde6e33c45a7d048c0c1ac))
* write json file to genbottles ([708584d](https://git.act3-ace.com/ace/data/tool/commit/708584daffd777ea3c24fc9bef9625283735f062))


### Features

* add genBottle command to util group ([8496baf](https://git.act3-ace.com/ace/data/tool/commit/8496baf60fac3be7fe38d6631d817b4ca8d913a6))
* added function that sends events assuming that telemtry needs all items ([7735070](https://git.act3-ace.com/ace/data/tool/commit/77350707fbbc1ecc95e8b72c1ce8c289b380095f))
* **config_mgmt:** Add config management; set persistent config, view current settings, save settings, and reset to defaults ([08cae90](https://git.act3-ace.com/ace/data/tool/commit/08cae90586f6d915d6e7fd150c64b80b7744d05d))
* **config_mgmt:** Add config management; set persistent config, view current settings, save settings, and reset to defaults ([1fb6b14](https://git.act3-ace.com/ace/data/tool/commit/1fb6b14c13405408632807d8c436aad9db997f6d))
* **telemetry:** add basic telemetry reporting for push and pull events ([17b6f23](https://git.act3-ace.com/ace/data/tool/commit/17b6f23cf32866a7d76eecebe916fedd047d402b))
* **telemetry:** add basic telemetry reporting for push and pull events ([a894862](https://git.act3-ace.com/ace/data/tool/commit/a894862d2f266fea9a1df2093286f38bf139258d))

## [0.18.4](https://git.act3-ace.com/ace/data/tool/compare/v0.18.3...v0.18.4) (2021-10-15)


### Bug Fixes

* **mirror-arch:** Add manifest list handling, choosing an appropriate image based on architecture selectors.  DO_RELEASE=true ([120006e](https://git.act3-ace.com/ace/data/tool/commit/120006e341de2ad665fd5717952386fea7770747))

## [0.18.3](https://git.act3-ace.com/ace/data/tool/compare/v0.18.2...v0.18.3) (2021-10-12)


### Bug Fixes

* change parts size to unit64 ([88941bd](https://git.act3-ace.com/ace/data/tool/commit/88941bd236c80149ac836d898487225e4836f50c))
* **ci:** bumped ci pipeline version ([ce2e802](https://git.act3-ace.com/ace/data/tool/commit/ce2e8027645af93c450dc7faa9d9fb16ec95b0e7))
* Correct parsing for container refs that have both tag and digest.  DO_RELEASE=true ([be5cd80](https://git.act3-ace.com/ace/data/tool/commit/be5cd8061028f380c1d564b542f5099b39166856))
* refactor validation a bit to simplify and clean it up a bit, and update gitlab.ci to latest dependency version ([8dbecbf](https://git.act3-ace.com/ace/data/tool/commit/8dbecbf2d392ddc0af592b97c848ad6fe2820548))
* remove uint64. Return to int64 ([f9d69f3](https://git.act3-ace.com/ace/data/tool/commit/f9d69f3423bd388aec954d10cea3e55cb4535432))
* satisfy linting errors ([c10e3cb](https://git.act3-ace.com/ace/data/tool/commit/c10e3cb5a07a28b427e73fc71d9a6705ade86651))
* update bottle labels to follow kubernetes label syntax ([4f3ee8b](https://git.act3-ace.com/ace/data/tool/commit/4f3ee8b7174a68db1f85837c14f2158022bbedf1))
* update labels docs ([4d10295](https://git.act3-ace.com/ace/data/tool/commit/4d10295bb744a2fabd0b89ddde8c19f3a86363f6))
* update public artifacts to follow new format and expect certain types ([8d0c2ba](https://git.act3-ace.com/ace/data/tool/commit/8d0c2ba5afb217305fe7065035e60c29e1518715))

## [0.18.1](https://git.act3-ace.com/ace/data/tool/compare/v0.18.0...v0.18.1) (2021-09-29)


### Bug Fixes

* **docs:** added links to docs ([4735762](https://git.act3-ace.com/ace/data/tool/commit/4735762fc8ed0a3297ba5f31cddcd58001bcd534))
* add check for repo skip on default destination mapping case.  DO_RELEASE=true ([b863d18](https://git.act3-ace.com/ace/data/tool/commit/b863d18f987014dc5239be6590446cc6583d684d))
* **mirror-destmap:** fix destination map handling multiple destinations and tag wildcards ([a94e378](https://git.act3-ace.com/ace/data/tool/commit/a94e378c5953440a4da86c95722c247c45188a6a))
* add test and edit expected errors ([2a2b317](https://git.act3-ace.com/ace/data/tool/commit/2a2b3174db4a4558aa3d93992fe76099944eb675))
* allow setting of new elem in slice fields of schema ([332512d](https://git.act3-ace.com/ace/data/tool/commit/332512d06389f8c108368f68cf87b9ed07420a8c))
* correct migration of old versions ([29543ea](https://git.act3-ace.com/ace/data/tool/commit/29543eafa4166e31e991734b12cd6248ba1790b4))
* delete comments and reorg logic statements ([08f68af](https://git.act3-ace.com/ace/data/tool/commit/08f68af118189d6633dcc57cac0f67e069da024e))
* extra tests and output fixes ([1d8ff55](https://git.act3-ace.com/ace/data/tool/commit/1d8ff55b6b33c7762488c9ac352bf3dd0fcf16ae))
* remove set helper that Titled input. just causes bugs ([44df325](https://git.act3-ace.com/ace/data/tool/commit/44df3257fd56c8c4fae04b97fa11e442b40787e1))
* update vermigrators tests ([223db4b](https://git.act3-ace.com/ace/data/tool/commit/223db4bde6dc85d60e2eeb0b7a6a2801f4b4d2f9))
* vermigrators need to transfer everything, not just the different/new things ([03d4c1e](https://git.act3-ace.com/ace/data/tool/commit/03d4c1eaa4156751e8abf9c34157c054b332df85))

## [0.18.1](https://git.act3-ace.com/ace/data/tool/compare/v0.18.0...v0.18.1) (2021-09-28)


### Bug Fixes

* add check for repo skip on default destination mapping case.  DO_RELEASE=true ([b863d18](https://git.act3-ace.com/ace/data/tool/commit/b863d18f987014dc5239be6590446cc6583d684d))
* **mirror-destmap:** fix destination map handling multiple destinations and tag wildcards ([a94e378](https://git.act3-ace.com/ace/data/tool/commit/a94e378c5953440a4da86c95722c247c45188a6a))
* add test and edit expected errors ([2a2b317](https://git.act3-ace.com/ace/data/tool/commit/2a2b3174db4a4558aa3d93992fe76099944eb675))
* allow setting of new elem in slice fields of schema ([332512d](https://git.act3-ace.com/ace/data/tool/commit/332512d06389f8c108368f68cf87b9ed07420a8c))
* correct migration of old versions ([29543ea](https://git.act3-ace.com/ace/data/tool/commit/29543eafa4166e31e991734b12cd6248ba1790b4))
* delete comments and reorg logic statements ([08f68af](https://git.act3-ace.com/ace/data/tool/commit/08f68af118189d6633dcc57cac0f67e069da024e))
* extra tests and output fixes ([1d8ff55](https://git.act3-ace.com/ace/data/tool/commit/1d8ff55b6b33c7762488c9ac352bf3dd0fcf16ae))
* remove set helper that Titled input. just causes bugs ([44df325](https://git.act3-ace.com/ace/data/tool/commit/44df3257fd56c8c4fae04b97fa11e442b40787e1))
* update vermigrators tests ([223db4b](https://git.act3-ace.com/ace/data/tool/commit/223db4bde6dc85d60e2eeb0b7a6a2801f4b4d2f9))
* vermigrators need to transfer everything, not just the different/new things ([03d4c1e](https://git.act3-ace.com/ace/data/tool/commit/03d4c1eaa4156751e8abf9c34157c054b332df85))

## [0.18.1](https://git.act3-ace.com/ace/data/tool/compare/v0.18.0...v0.18.1) (2021-09-21)


### Bug Fixes

* **mirror-destmap:** fix destination map handling multiple destinations and tag wildcards ([b7415ca](https://git.act3-ace.com/ace/data/tool/commit/b7415ca47cb92cec9ca9da4c8adb21e5dc8dac90))

# [0.18.0](https://git.act3-ace.com/ace/data/tool/compare/v0.17.0...v0.18.0) (2021-09-20)


### Bug Fixes

* add write mutex around output writer ([41476c6](https://git.act3-ace.com/ace/data/tool/commit/41476c6c98945eaf5f6f07c20a7dedb151260a31))
* better pull cmd output ([e97228c](https://git.act3-ace.com/ace/data/tool/commit/e97228cabf1542ed844002d8a7f6605a7564880b))
* change name of output write method ([5cd4237](https://git.act3-ace.com/ace/data/tool/commit/5cd4237366ca2e906644cabc8089b1759323b5a0))
* changed a few remaining fmt.print to either logs or command outputs ([0bcfd56](https://git.act3-ace.com/ace/data/tool/commit/0bcfd56bd2c2b773c8c84f70134879bfb67f7033))
* correct crash by instantiating output writer in progress bar ([c91e450](https://git.act3-ace.com/ace/data/tool/commit/c91e4504df65bf766d0d44fe135cf2f8cca1c1ad))
* correct missing argument for NewBasicTransferQueue in qtransfer tests ([d376342](https://git.act3-ace.com/ace/data/tool/commit/d37634221197fb42614f69d275f4c5ba6f578906))
* correct references to either encoded part of digest vs actual digest ([ec7956f](https://git.act3-ace.com/ace/data/tool/commit/ec7956ff2e1ac83dca02e83958aa3670868c1d31))
* correct some lint complaints, and set default-signifies-exhaustive in lint settings ([01be519](https://git.act3-ace.com/ace/data/tool/commit/01be519ee30b313d01fcc4c16f3acb7b4c0a7e26))
* correct unit tests and include mod depends + mod tidy ([158097d](https://git.act3-ace.com/ace/data/tool/commit/158097da8d98dd9d31c393d284d5b75d6f748983))
* correctly parse digests ([b00c0a8](https://git.act3-ace.com/ace/data/tool/commit/b00c0a891fc9aa702075961725beab264088703f))
* document Outputter ([798c551](https://git.act3-ace.com/ace/data/tool/commit/798c551f8df34b88df253278f4b7603c8d9617aa))
* functional test failure due to non-allocated progress tracker ([3e3cbf5](https://git.act3-ace.com/ace/data/tool/commit/3e3cbf512aeb5d53020d44403f7eac344c9ca6bd))
* functional tests ([53c49bc](https://git.act3-ace.com/ace/data/tool/commit/53c49bcd6f22f9891cf769b748c95d0e235364f6))
* functional tests ([7560532](https://git.act3-ace.com/ace/data/tool/commit/7560532bd9fc7e55c4df968d97527a7605ebfd1d))
* include legacy MediaTypes and add logic to handle pulling legacy MediaTypes ([cc97fb4](https://git.act3-ace.com/ace/data/tool/commit/cc97fb4343da09d2699581f3794c5550bec47b18))
* integrate new progress bar with transferworker ([3393e88](https://git.act3-ace.com/ace/data/tool/commit/3393e8820077f8cf87ab194efbb4d85ebe60eecb))
* log parsing fixes ([a3b1822](https://git.act3-ace.com/ace/data/tool/commit/a3b1822fc8c060652d2ea9abe3ec68815e09c1b0))
* merge with master ([a518c68](https://git.act3-ace.com/ace/data/tool/commit/a518c6836a8f61bf2d02eabf801814524e493e79))
* merge with next ([936da50](https://git.act3-ace.com/ace/data/tool/commit/936da5005078242cf6893f35c5fc0d53ce78a866))
* merge with next ([a48dcf5](https://git.act3-ace.com/ace/data/tool/commit/a48dcf51f9912a7abd05a8ca64b9218b7f033e18))
* merge with next again ([13bae67](https://git.act3-ace.com/ace/data/tool/commit/13bae67c90b168d695396acc33ccdf896f898dbc))
* merge with next. move config.go command to proper location ([2560501](https://git.act3-ace.com/ace/data/tool/commit/2560501c0a6747370ee9bf0c0f492e03a55c85ed))
* merge with output cleanup ([70fa1f3](https://git.act3-ace.com/ace/data/tool/commit/70fa1f31636ce2fa168de3b811f5a2623be0671b))
* merge with output normalization ([dbd05a6](https://git.act3-ace.com/ace/data/tool/commit/dbd05a62187fa3f8b53bac887902bc3db2136e44))
* move bottle expiration to manifest during push ([6bf7e81](https://git.act3-ace.com/ace/data/tool/commit/6bf7e81489f5c2d5e38e35c1d6e074a95cbda32e))
* no author is set during gitlab-runner functional tests. pivot to different set test field ([ec0b1bb](https://git.act3-ace.com/ace/data/tool/commit/ec0b1bb29c1041e9ac52d32a125eca7f6fb43eae))
* parse logs for errors and parse meta output for changes ([25eea45](https://git.act3-ace.com/ace/data/tool/commit/25eea45c353e10f198892f220db60f42dfaa7519))
* refactor output writing to use a struct similar to the logger ([03d60a8](https://git.act3-ace.com/ace/data/tool/commit/03d60a8d86d2ee5a2eb5dac9d0db1f2f6f768c99))
* remove digest map, add digest string ([ff0da72](https://git.act3-ace.com/ace/data/tool/commit/ff0da724bf9490aef80eaad13b4de2f20c59da39))
* satisfy golint errors ([79722d6](https://git.act3-ace.com/ace/data/tool/commit/79722d655d7048314c24b94a07249eacc247ed9e))
* satisfy linting errors ([167547f](https://git.act3-ace.com/ace/data/tool/commit/167547fc5f5b12fb9b1f836cbba865829c9b567b))
* set executable name ([2f10c6d](https://git.act3-ace.com/ace/data/tool/commit/2f10c6df4f504a2c9cabf75edb4967f93d128c0a))
* set executable name ([9930072](https://git.act3-ace.com/ace/data/tool/commit/99300720efc1a0bb527575076195990758ad7a17))
* solve digest man removal. name log files as .txt so they show up nice in gitlab artifacts ([e5dc104](https://git.act3-ace.com/ace/data/tool/commit/e5dc104761e5a4ab7c27fbf5e5e874c722233d37))
* specify correct path to functional tests ([2313e28](https://git.act3-ace.com/ace/data/tool/commit/2313e280d05094cae4e2e305f1b488cbb442cc2d))
* specify correct path to functional tests ([c8d4dcf](https://git.act3-ace.com/ace/data/tool/commit/c8d4dcf9d0b27b8df51a62e95ffa4bf7a05cde2f))
* start changing schema ([1d22d56](https://git.act3-ace.com/ace/data/tool/commit/1d22d569f620466a0b5972187601c5b6b2c982cd))
* update deprecated function ([710a871](https://git.act3-ace.com/ace/data/tool/commit/710a8718dd0bfff106636a8a5229439e286214f3))
* update double init test to pass ([f73990e](https://git.act3-ace.com/ace/data/tool/commit/f73990e8b4baa4841bd747ae21a0e0b0276e880b))
* update exist set for queued files only after successful transfer ([2c7bc3c](https://git.act3-ace.com/ace/data/tool/commit/2c7bc3c6ad8ebcf3beb8f443da479aaf2601ef53))
* update levels to reflect desired output ([aff0f76](https://git.act3-ace.com/ace/data/tool/commit/aff0f76089dd789751ef372bd6cf699965f6f7af))
* upgraded the pipeline to v3.1.0 ([eeba251](https://git.act3-ace.com/ace/data/tool/commit/eeba2512c2e1c532d694fe7f1c2b44806a3e0c29))
* upgraded the pipeline to v3.1.0 ([fecc7a3](https://git.act3-ace.com/ace/data/tool/commit/fecc7a339274ee6217a17ea909c3caec962e04cd))
* util/yamlnode.go not parsing omitempty option, causing entry.yaml to be broken ([f5f1790](https://git.act3-ace.com/ace/data/tool/commit/f5f1790af16f6c87ef18de6a61e2c02afb41ca7c))
* **output:** Change logging to either on (1) or off (0). update printing to use command output ([4da6cb4](https://git.act3-ace.com/ace/data/tool/commit/4da6cb47565f81e66b798553ecdcf0ae453ae48b))
* **work-queue:** correct regression bug in authentication during mirror pull, and include important missing commit integrating work queue into mirror functionality ([11db8f6](https://git.act3-ace.com/ace/data/tool/commit/11db8f6dfb8434e4e1f5b1c93f59a57a8d01ae52))
* add meta command group, started work on set cmd ([4a8a01a](https://git.act3-ace.com/ace/data/tool/commit/4a8a01a7162dd9bf6f4ac02e1343edb9e954246a))
* allowed progress bar disable during archive and extract ([28c915f](https://git.act3-ace.com/ace/data/tool/commit/28c915f9674b57ce40e0fe374be6ed31e5117a3b))
* begin switching cmd style with arglist style ([1ce9a9c](https://git.act3-ace.com/ace/data/tool/commit/1ce9a9c4395f6350831669f98ee1d7b2d5a37333))
* change dataset to bottle ([e4dd68e](https://git.act3-ace.com/ace/data/tool/commit/e4dd68e8765b6ac4f16f0d1e7e2ce7c31f0c1f3e))
* fix tests failing due to old special files name ([27aadf7](https://git.act3-ace.com/ace/data/tool/commit/27aadf710853ac82a51113f4b900d3f3dbbd8990))
* get tests to pass ([d3eb339](https://git.act3-ace.com/ace/data/tool/commit/d3eb339dc7ff076df6b81b1f62ebfe592745a5c1))
* lint errors ([1948f51](https://git.act3-ace.com/ace/data/tool/commit/1948f5173f65ca47a2468e2340d51cd0a4688555))
* log path changes ([21d04d4](https://git.act3-ace.com/ace/data/tool/commit/21d04d47b8851a625023f4cdb426a7e6b0572248))
* log path changes 3 ([69a28fa](https://git.act3-ace.com/ace/data/tool/commit/69a28fa86ad7755459455523a9c72169600a1e9c))
* log path changes 4 ([7ed2745](https://git.act3-ace.com/ace/data/tool/commit/7ed2745a58df780d2f59b7856db6600c14135e41))
* log path changes2 ([2fcf12c](https://git.act3-ace.com/ace/data/tool/commit/2fcf12c0b4f0be658fbfc8765176adb17fc1f27d))
* merge with lastest release ([a53dbb1](https://git.act3-ace.com/ace/data/tool/commit/a53dbb18a997984b6656b096441046b2ccd285d1))
* merge with master ([6152ef7](https://git.act3-ace.com/ace/data/tool/commit/6152ef72ed3ac1d460a412ddc0ac63b3e5b7efcc))
* merge with next ([c230d7e](https://git.act3-ace.com/ace/data/tool/commit/c230d7e8bd55bd863d3285ecc76933d973f0ca21))
* merge with next and fix meta test ([291ba9b](https://git.act3-ace.com/ace/data/tool/commit/291ba9becdcc69c384ab1e02894eecbc9161a598))
* merge with updated testing ([c874b22](https://git.act3-ace.com/ace/data/tool/commit/c874b22f43ab2f0f3f0c5a5753d6b3cc484f97ee))
* new failure test ([b749b39](https://git.act3-ace.com/ace/data/tool/commit/b749b390eb73fe71b9feaeb33f8ecf7a42d6ee8c))
* new failure test 2 ([0cd4368](https://git.act3-ace.com/ace/data/tool/commit/0cd4368ec9326453a7897f614318f573e2190a86))
* new failure test 3 ([43c50b5](https://git.act3-ace.com/ace/data/tool/commit/43c50b5496f28da639e045424cb024dfba83e260))
* omitempty fields that are omitempty-able ([8a8bb7e](https://git.act3-ace.com/ace/data/tool/commit/8a8bb7e9d73a11e2225442e1d4b90efe7217c5dc))
* oops... revert gitlab-ci back to normal ([9ac8190](https://git.act3-ace.com/ace/data/tool/commit/9ac81900b6c7c76255b0b0e315bd7eed987ec0ee))
* output trace logs on non logger errors ([e6f3748](https://git.act3-ace.com/ace/data/tool/commit/e6f37488f9d9e639f943fd8774c8ab62347e1c1c))
* quick logic fix, Looking to solve output inconsistancies ([4f80c0c](https://git.act3-ace.com/ace/data/tool/commit/4f80c0ca9f339149e1259d32ef591f96ce9f0626))
* refactor getmetadata from yaml functionality ([9f46325](https://git.act3-ace.com/ace/data/tool/commit/9f4632512df29da8ecd31e1c025bf6441d1b3f39))
* refactor old switch logic for argsList ([89298fa](https://git.act3-ace.com/ace/data/tool/commit/89298faebe834c6825f807c61f9ec49af6b0350f))
* refactor test_helper to use arg list ([9dc712e](https://git.act3-ace.com/ace/data/tool/commit/9dc712ea400bebc9e4936cc876c7b88ae1cd7607))
* remove old logrus mod import, run go mod tidy ([169a76f](https://git.act3-ace.com/ace/data/tool/commit/169a76f5726efcb4bcae816f51b29758890e47d3))
* satisfy golint errors ([f85dbcd](https://git.act3-ace.com/ace/data/tool/commit/f85dbcd2d431b6cec49284e4de26fd73727e2547))
* satisfy golint warnings ([7b5cdb7](https://git.act3-ace.com/ace/data/tool/commit/7b5cdb72b456637571a10879e9c2f86bd13cace4))
* satisfy golint warnings ([dca2cdc](https://git.act3-ace.com/ace/data/tool/commit/dca2cdc16996fc59967c432b132e4b0722e973f5))
* small changes in schema based of off TODOs ([ff699cf](https://git.act3-ace.com/ace/data/tool/commit/ff699cf391aa8964b704a4b0df1fc8330f7b3503))
* start working on docs. Re-add docs that were removed ([4e9b5e1](https://git.act3-ace.com/ace/data/tool/commit/4e9b5e1f5d9bff8a526455704e385257fc14ad0d))
* test error in CICD ([bb50410](https://git.act3-ace.com/ace/data/tool/commit/bb5041047b73b6a49971af571b60255e958ec236))
* test failure in CICD ([f02a0e3](https://git.act3-ace.com/ace/data/tool/commit/f02a0e365b6e5bfeb2ac9d275336a37ec5d6c2fa))
* test out artifacts for failures ([9ddae56](https://git.act3-ace.com/ace/data/tool/commit/9ddae56a232732dfc6c278401419f8c6bfa850f0))
* test out artifacts for failures 2 ([632a13e](https://git.act3-ace.com/ace/data/tool/commit/632a13eb6bc5deccb5b8664f081a4e6301eb9bee))
* update a few tests and change documentation/comments ([1f93cd7](https://git.act3-ace.com/ace/data/tool/commit/1f93cd722e844900f150abae1687fc8871a3e981))
* update annotations -> labels. map[string]string now instead of a struct ([43e72c3](https://git.act3-ace.com/ace/data/tool/commit/43e72c3723154882b6c1a93710dbc46f41dd7efc))
* update gitlab-ci reference tag ([5f5d6da](https://git.act3-ace.com/ace/data/tool/commit/5f5d6da393ba310a37dc09068d49f25e10d84f3d))
* update go mod vals ([ae61695](https://git.act3-ace.com/ace/data/tool/commit/ae61695a81223f9a37c21c2d91845e4fa495b308))
* update schema and various names based off of suggested changes ([7d88230](https://git.act3-ace.com/ace/data/tool/commit/7d882305294ffe9c26ca73b56cbb9b35359ba803))
* **label-annotations:** change label annotations to contain a prefix in the oci manifest, and change index.json annotations used during mirror to ones that appear in the specification BREAKING CHANGE ([2a99348](https://git.act3-ace.com/ace/data/tool/commit/2a99348181a671f3064e9472319410d4875d1bc4))
* update golang pipeline reference ([d566b38](https://git.act3-ace.com/ace/data/tool/commit/d566b384ea9adea0a646c4dd4c3aa267f4cdc75e))


### Features

* **mirror-src-regex:** Enhance the mirror command source parsing to accept a second column regex for tag matches ([545abe6](https://git.act3-ace.com/ace/data/tool/commit/545abe64ef98719be632852c7bddcd04f7f88197))
* **partial-bottle:** Added tracking for bottle parts not selected when pulling with selectors, allowing partial bottles ([9433254](https://git.act3-ace.com/ace/data/tool/commit/9433254c8c22f61e148bee729c2c9d5dbcb30e14))
* **transfer-worker:** Add a multi-queue, multi-pool concurrent transfer task consumer ([cda9c07](https://git.act3-ace.com/ace/data/tool/commit/cda9c07fcdb8cd8b7e2ac23830dea75c9b4ce9be))
* add tests that expect an error. Refactor old code to accommodate these changes ([284e9ed](https://git.act3-ace.com/ace/data/tool/commit/284e9ed31ccf8ff77b125d18bf62861d01348a75))
* add, set, get commmands implemented and tested. new data schema defined. migration also implemented. new dataset functions for yaml file reading and editing. edits to misc files for usage name change ([d2fc54f](https://git.act3-ace.com/ace/data/tool/commit/d2fc54f6bbad68c3f5291dc15daf08e13f711d8b))
* **unified-dest-exist:** Unify destination existance tracking to not segregate by blob source ([80bc835](https://git.act3-ace.com/ace/data/tool/commit/80bc8353c153c7673be6417d1b4aabba6129c938))
* supressed logs, write trace logs, stop progress bars, add fail tests ([ae82e54](https://git.act3-ace.com/ace/data/tool/commit/ae82e54f2501d75de20e27605867857896d8be50))

# [0.17.0](https://git.act3-ace.com/ace/data/tool/compare/v0.16.1...v0.17.0) (2021-07-23)


### Bug Fixes

* cross server transfer corrections after authentication changes and other updates ([0631080](https://git.act3-ace.com/ace/data/tool/commit/0631080df2f3511095151d1a8eb1804bd4c802f5))
* duplicated code and lint errors ([3314a16](https://git.act3-ace.com/ace/data/tool/commit/3314a1615473a6dc163e394dca06be014135a18a))
* duplicated code and lint errors ([e8c6121](https://git.act3-ace.com/ace/data/tool/commit/e8c6121071c9f5199314556631236a1bbfdaf35a))


### Features

* **cross-server:** integrate cross server transfer into mirror push ([bce9a4b](https://git.act3-ace.com/ace/data/tool/commit/bce9a4ba91005c82d4b9b7ecf5e79a8f802f26f2))
* **cross-server-transfer:** enable unbuffered server to server data streaming from source blob to dest blob ([47a4319](https://git.act3-ace.com/ace/data/tool/commit/47a43195db911214357045aa4205ba26a7dfd291))
* **cross-server-transfer:** enable unbuffered server to server data streaming from source blob to dest blob ([5ae1994](https://git.act3-ace.com/ace/data/tool/commit/5ae1994313efcd7c56c48d85dcbbaeb009a5e4c3))

## [0.16.1](https://git.act3-ace.com/ace/data/tool/compare/v0.16.0...v0.16.1) (2021-07-23)


### Bug Fixes

* add response body logging for manifest put errors ([1f1fce8](https://git.act3-ace.com/ace/data/tool/commit/1f1fce8fee6bf402276f5be8e83b9641eb750959))
* gofmt files, and include mapping test with sample aaco data ([2de3f8e](https://git.act3-ace.com/ace/data/tool/commit/2de3f8e0ecc56022c736cf885cf12bfbe9fdef6c))
* Improve functional testing and include error tests ([3d46bab](https://git.act3-ace.com/ace/data/tool/commit/3d46bab089a69bdc9b3bdc19242dadead83dfd4a))
* preserve authentication token between post and patch requests for layers ([1a95217](https://git.act3-ace.com/ace/data/tool/commit/1a95217b4d1993da742c60200cc6120c4fc397c2))
* properly exit on unaccepted authorization ([2364175](https://git.act3-ace.com/ace/data/tool/commit/23641754b8f82608f1ddf894afe7dafc6eefdfe1))
* rewind file on manifest post failure when httpClient.Do calls close ([411f798](https://git.act3-ace.com/ace/data/tool/commit/411f79859f1108bdf1ca5e4175efa98414a08e7b))
* work around docker auth incorrect parsing of www-authenticate header scope field ([fadb2d1](https://git.act3-ace.com/ace/data/tool/commit/fadb2d1fbc24b299861d072c5b3ced37e4df256c))
* wrap manifest files in no close reader on mirror push ([1357300](https://git.act3-ace.com/ace/data/tool/commit/13573002300dd36e10948191565378a1b7386921))

# [0.16.0](https://git.act3-ace.com/ace/data/tool/compare/v0.15.1...v0.16.0) (2021-07-13)


### Bug Fixes

* update tests based on source declaration differentiation ([8ae8f6a](https://git.act3-ace.com/ace/data/tool/commit/8ae8f6ab6942ad37c69e4e74674a2aef870fb84c))


### Features

* **multidecl:** differentiate repository sources by declaration as well as manifest ID and repository reference ([8aca3ef](https://git.act3-ace.com/ace/data/tool/commit/8aca3efd68de42068505c7602d028a28f2d58ab2))
* **multidest:** allow manifests and blobs to be mapped to multiple desinations on mirror push operations ([225922b](https://git.act3-ace.com/ace/data/tool/commit/225922bcd10f5946fe3c54fc48e892800656add7))

## [0.15.1](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0...v0.15.1) (2021-07-09)


### Bug Fixes

* attempt to close body of nil response after authentication failure ([4a48012](https://git.act3-ace.com/ace/data/tool/commit/4a48012107953b8e3d03d907695ee2bd416bdc35))
* restore capability of retrieving tags when using the repo index -list command ([de471f7](https://git.act3-ace.com/ace/data/tool/commit/de471f72365205a60fed61eba03594c2f8788b00))
* workaround paginated catalog and tags list by specifying a large page count on initial request (1000000) ([fc67fe7](https://git.act3-ace.com/ace/data/tool/commit/fc67fe771fe878dfb4b4fd817ad8b3809f57dd04))

# [0.15.0](https://git.act3-ace.com/ace/data/tool/compare/v0.14.5...v0.15.0) (2021-07-01)


### Bug Fixes

* **ledger:** Track manifest blobs in blob ledger and skip processing matching manifests in index.go.  DO_RELEASE=true ([b8f8bf7](https://git.act3-ace.com/ace/data/tool/commit/b8f8bf74a76ef07641b28dbc7fd67aa5cd0c3957))
* **ocilayout:** avoid recreating oci-layout file if it already exists, to avoid writing in read only scenarios.  DO_RELEASE=true ([6d2a50b](https://git.act3-ace.com/ace/data/tool/commit/6d2a50b021bc61641e00e9785e7c9e810d8f775f))
* **repo index:** change 404 from error to warning, to allow continuation of indexing ([b1b32ce](https://git.act3-ace.com/ace/data/tool/commit/b1b32ceec7b57a669763fc210e23b6f54a131679))
* add a fix msg so the release pipeline will do something ([2745ae6](https://git.act3-ace.com/ace/data/tool/commit/2745ae6dd4f693652bfddf3201aff22809f6ed9b))
* add additional trace logging for showing transfer failure response body data ([a93ed17](https://git.act3-ace.com/ace/data/tool/commit/a93ed1729c26e9710f637c73232e705f4af63f75))
* add digest algorithm label to layer IDs in blob ledger ([902ccfc](https://git.act3-ace.com/ace/data/tool/commit/902ccfc9b3f307eedc569f63849792e4f656ca2f))
* add documentation for new set utility functions) ([efc1f3a](https://git.act3-ace.com/ace/data/tool/commit/efc1f3a0a8ed62392a3631696ed9ad179d7b6e6a))
* add proper releaserc support ([a4f85f9](https://git.act3-ace.com/ace/data/tool/commit/a4f85f9336547999665058d99d49c2fcfb4ddb23))
* disable http keepalive to attempt to avoid too many open files error ([a3e6f20](https://git.act3-ace.com/ace/data/tool/commit/a3e6f201246e0eca227af1721e9e331bb417632b))
* during layer push, make sure response body exists before gathering it while handling errors ([a66b296](https://git.act3-ace.com/ace/data/tool/commit/a66b2963f9254d86227c7280f9aeb00efcc97799))
* exclude test files from golangci-lint and create a Makefile to build with a git based version instead of development ([a34b07d](https://git.act3-ace.com/ace/data/tool/commit/a34b07d027e840d83b8e5fadc2ed19b80a654a90))
* finalize test breaking on sha256: label on digests ([fc79c59](https://git.act3-ace.com/ace/data/tool/commit/fc79c5975971c342cacafe6d8b0802348e040eeb))
* fix .version.yaml output DO_RELEASE=true ([9ceeebb](https://git.act3-ace.com/ace/data/tool/commit/9ceeebbebe8d2d3d6c9520e5d057276eccc4a911))
* fix rc stuff. update docs for testing ([376c880](https://git.act3-ace.com/ace/data/tool/commit/376c8801e66a8185310b1c0fa7cac595ce3ad0ba))
* Forward close calls to storage reader during layer push, and allow linkOnCopy to be optionally disabled for bottle file cache ([b0810a9](https://git.act3-ace.com/ace/data/tool/commit/b0810a9a8ad88da6c1fb8318d7f977088e470a29))
* merge with alpha ([af52eb6](https://git.act3-ace.com/ace/data/tool/commit/af52eb63649f49431a427ee61f1bdcaa3aa1c873))
* merge with master ([2ead8f3](https://git.act3-ace.com/ace/data/tool/commit/2ead8f32ee61c228edc46fa9e9e58487bef07266))
* merge with master ([e6c5662](https://git.act3-ace.com/ace/data/tool/commit/e6c56624c34a33369943293ee68397699485bc57))
* pretty print index.json, correct bug with has instead of tag on sourcelist, add digest alg. in blob ledger entries, deduplicate blob ledger layers ([90f085e](https://git.act3-ace.com/ace/data/tool/commit/90f085e5da5c36f1651c6750127511a95bf8791b))
* put charts back in git lfs ([ff26301](https://git.act3-ace.com/ace/data/tool/commit/ff26301a5360cd046b1d0e7bf06fbfc3161fd53a))
* remove debug line that was limiting max connections to 1 ([910c7b9](https://git.act3-ace.com/ace/data/tool/commit/910c7b90a4daae6aa8dff15ea8a787b7a60d8536))
* supress an expected error during new dir in commit ([3117fac](https://git.act3-ace.com/ace/data/tool/commit/3117facad2822d9b1e5467a449698b2f0989cbf4))
* unhandled errors causing crash in gethttp after authentication ([d846f64](https://git.act3-ace.com/ace/data/tool/commit/d846f64868f2311802b7470af28e4f72e1320074))
* Update user agent string to include build version ([5db9896](https://git.act3-ace.com/ace/data/tool/commit/5db9896d6f93086e59aac8be1a5ff8d160b87801))
* user-agent string replace missing slash ([9779a6e](https://git.act3-ace.com/ace/data/tool/commit/9779a6e39ec8c240227d4508185ab67eb450920e))
* **cache:** move default cache scratch path to a subdir of main cache dir ([b70049d](https://git.act3-ace.com/ace/data/tool/commit/b70049d69296380eb7973ff20add34507507a140))
* **commit:** clear layer digest when resetting metadata in preparation for updating bottle parts. (closes [#123](https://git.act3-ace.com/ace/data/tool/issues/123)) ([f3040c5](https://git.act3-ace.com/ace/data/tool/commit/f3040c589cf1e5832b61ee45b43a2a5b5d94985e))
* **labels:** add label markers for directories containing separate parts that are not otherwise labeled ([80c4206](https://git.act3-ace.com/ace/data/tool/commit/80c4206deafcfc922b04c351891cf9cec058fade))
* **modtime:** properly update part modification time during pull so only change detection has fewer false positives ([838c963](https://git.act3-ace.com/ace/data/tool/commit/838c96335eff9bced263b19ac160aa1b0eba7cfb))
* **prune:** fix util prune to also remove cache scratch items ([ddc5383](https://git.act3-ace.com/ace/data/tool/commit/ddc5383ced3ae4a5e4e0253c63031d32f916e5c4))
* **scopedauth:** Create a scoped authorization manager to work around docker's realm-based authorization caching issue ([2c7d04d](https://git.act3-ace.com/ace/data/tool/commit/2c7d04d04af12c173ca5cc15863230968a1951bb))
* resolve short registry refs with docker style reference parsing ([9ee951d](https://git.act3-ace.com/ace/data/tool/commit/9ee951d0aa894996adfe91c4f0197d815314669f))
* set log output to rootCmd.out properly ([6c06ed0](https://git.act3-ace.com/ace/data/tool/commit/6c06ed057f684c5b5c5a26928e20839b21fa09f9))
* setup and config rootCmd better in cmd tests ([1f3597c](https://git.act3-ace.com/ace/data/tool/commit/1f3597cc0952d8acd6f54cf04b3d0aa45b6c4e67))
* start working on docs. Re-add docs that were removed ([7a77816](https://git.act3-ace.com/ace/data/tool/commit/7a778166e3331120e7785073fca0c5dc50c01a0f))
* update docs ([9f717ff](https://git.act3-ace.com/ace/data/tool/commit/9f717ff6a1fede9851f04dab4afab37f537a2de9))
* update tutorial-labels wording ([7e89004](https://git.act3-ace.com/ace/data/tool/commit/7e890040d1d1ccc923db37c429a8cb9bba7cbdbe))
* use default authorizer for wildcard tag list on mirror pull ([a841c66](https://git.act3-ace.com/ace/data/tool/commit/a841c66869abec4dc2449b270cbcfd11340998d2))
* watch working dest exists for existing blobs rather than final dest exists, and add sha256 specifier in front of blob digests ([c3be019](https://git.act3-ace.com/ace/data/tool/commit/c3be019a4012b9f7de954914fb510cdc2dd8c981))


### Features

* added new testing features to increase test robustness, reset env, check bottle existence ([678814b](https://git.act3-ace.com/ace/data/tool/commit/678814bd1a16cef38af2a17cccc2bf71d92e0b07))

# [0.15.0-alpha.10](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.9...v0.15.0-alpha.10) (2021-06-28)


### Bug Fixes

* **ledger:** Track manifest blobs in blob ledger and skip processing matching manifests in index.go.  DO_RELEASE=true ([b8f8bf7](https://git.act3-ace.com/ace/data/tool/commit/b8f8bf74a76ef07641b28dbc7fd67aa5cd0c3957))
* exclude test files from golangci-lint and create a Makefile to build with a git based version instead of development ([a34b07d](https://git.act3-ace.com/ace/data/tool/commit/a34b07d027e840d83b8e5fadc2ed19b80a654a90))

# [0.15.0-alpha.9](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.8...v0.15.0-alpha.9) (2021-06-28)


### Bug Fixes

* **ocilayout:** avoid recreating oci-layout file if it already exists, to avoid writing in read only scenarios.  DO_RELEASE=true ([6d2a50b](https://git.act3-ace.com/ace/data/tool/commit/6d2a50b021bc61641e00e9785e7c9e810d8f775f))
* user-agent string replace missing slash ([9779a6e](https://git.act3-ace.com/ace/data/tool/commit/9779a6e39ec8c240227d4508185ab67eb450920e))

# [0.15.0-alpha.8](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.7...v0.15.0-alpha.8) (2021-06-24)


### Bug Fixes

* during layer push, make sure response body exists before gathering it while handling errors ([a66b296](https://git.act3-ace.com/ace/data/tool/commit/a66b2963f9254d86227c7280f9aeb00efcc97799))

# [0.15.0-alpha.7](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.6...v0.15.0-alpha.7) (2021-06-23)


### Bug Fixes

* Forward close calls to storage reader during layer push, and allow linkOnCopy to be optionally disabled for bottle file cache ([b0810a9](https://git.act3-ace.com/ace/data/tool/commit/b0810a9a8ad88da6c1fb8318d7f977088e470a29))
* Update user agent string to include build version ([5db9896](https://git.act3-ace.com/ace/data/tool/commit/5db9896d6f93086e59aac8be1a5ff8d160b87801))
* **cache:** move default cache scratch path to a subdir of main cache dir ([b70049d](https://git.act3-ace.com/ace/data/tool/commit/b70049d69296380eb7973ff20add34507507a140))
* **commit:** clear layer digest when resetting metadata in preparation for updating bottle parts. (closes [#123](https://git.act3-ace.com/ace/data/tool/issues/123)) ([f3040c5](https://git.act3-ace.com/ace/data/tool/commit/f3040c589cf1e5832b61ee45b43a2a5b5d94985e))
* **labels:** add label markers for directories containing separate parts that are not otherwise labeled ([80c4206](https://git.act3-ace.com/ace/data/tool/commit/80c4206deafcfc922b04c351891cf9cec058fade))
* **modtime:** properly update part modification time during pull so only change detection has fewer false positives ([838c963](https://git.act3-ace.com/ace/data/tool/commit/838c96335eff9bced263b19ac160aa1b0eba7cfb))
* **prune:** fix util prune to also remove cache scratch items ([ddc5383](https://git.act3-ace.com/ace/data/tool/commit/ddc5383ced3ae4a5e4e0253c63031d32f916e5c4))

# [0.15.0-alpha.6](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.5...v0.15.0-alpha.6) (2021-06-17)


### Bug Fixes

* **scopedauth:** Create a scoped authorization manager to work around docker's realm-based authorization caching issue ([2c7d04d](https://git.act3-ace.com/ace/data/tool/commit/2c7d04d04af12c173ca5cc15863230968a1951bb))
* supress an expected error during new dir in commit ([3117fac](https://git.act3-ace.com/ace/data/tool/commit/3117facad2822d9b1e5467a449698b2f0989cbf4))

# [0.15.0-alpha.5](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.4...v0.15.0-alpha.5) (2021-06-15)


### Bug Fixes

* disable http keepalive to attempt to avoid too many open files error ([a3e6f20](https://git.act3-ace.com/ace/data/tool/commit/a3e6f201246e0eca227af1721e9e331bb417632b))

# [0.15.0-alpha.4](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.3...v0.15.0-alpha.4) (2021-06-15)


### Bug Fixes

* merge with alpha ([af52eb6](https://git.act3-ace.com/ace/data/tool/commit/af52eb63649f49431a427ee61f1bdcaa3aa1c873))
* unhandled errors causing crash in gethttp after authentication ([d846f64](https://git.act3-ace.com/ace/data/tool/commit/d846f64868f2311802b7470af28e4f72e1320074))
* update tutorial-labels wording ([7e89004](https://git.act3-ace.com/ace/data/tool/commit/7e890040d1d1ccc923db37c429a8cb9bba7cbdbe))

# [0.15.0-alpha.3](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.2...v0.15.0-alpha.3) (2021-06-02)


### Bug Fixes

* add additional trace logging for showing transfer failure response body data ([a93ed17](https://git.act3-ace.com/ace/data/tool/commit/a93ed1729c26e9710f637c73232e705f4af63f75))

# [0.15.0-alpha.2](https://git.act3-ace.com/ace/data/tool/compare/v0.15.0-alpha.1...v0.15.0-alpha.2) (2021-06-01)


### Bug Fixes

* remove debug line that was limiting max connections to 1 ([910c7b9](https://git.act3-ace.com/ace/data/tool/commit/910c7b90a4daae6aa8dff15ea8a787b7a60d8536))

# [0.15.0-alpha.1](https://git.act3-ace.com/ace/data/tool/compare/v0.14.6-alpha.4...v0.15.0-alpha.1) (2021-05-28)


### Bug Fixes

* add digest algorithm label to layer IDs in blob ledger ([902ccfc](https://git.act3-ace.com/ace/data/tool/commit/902ccfc9b3f307eedc569f63849792e4f656ca2f))
* add documentation for new set utility functions) ([efc1f3a](https://git.act3-ace.com/ace/data/tool/commit/efc1f3a0a8ed62392a3631696ed9ad179d7b6e6a))
* add proper releaserc support ([a4f85f9](https://git.act3-ace.com/ace/data/tool/commit/a4f85f9336547999665058d99d49c2fcfb4ddb23))
* fix .version.yaml output DO_RELEASE=true ([9ceeebb](https://git.act3-ace.com/ace/data/tool/commit/9ceeebbebe8d2d3d6c9520e5d057276eccc4a911))
* merge with master ([e6c5662](https://git.act3-ace.com/ace/data/tool/commit/e6c56624c34a33369943293ee68397699485bc57))
* pretty print index.json, correct bug with has instead of tag on sourcelist, add digest alg. in blob ledger entries, deduplicate blob ledger layers ([90f085e](https://git.act3-ace.com/ace/data/tool/commit/90f085e5da5c36f1651c6750127511a95bf8791b))
* put charts back in git lfs ([ff26301](https://git.act3-ace.com/ace/data/tool/commit/ff26301a5360cd046b1d0e7bf06fbfc3161fd53a))
* set log output to rootCmd.out properly ([6c06ed0](https://git.act3-ace.com/ace/data/tool/commit/6c06ed057f684c5b5c5a26928e20839b21fa09f9))
* setup and config rootCmd better in cmd tests ([1f3597c](https://git.act3-ace.com/ace/data/tool/commit/1f3597cc0952d8acd6f54cf04b3d0aa45b6c4e67))
* update docs ([9f717ff](https://git.act3-ace.com/ace/data/tool/commit/9f717ff6a1fede9851f04dab4afab37f537a2de9))


### Features

* added new testing features to increase test robustness, reset env, check bottle existence ([678814b](https://git.act3-ace.com/ace/data/tool/commit/678814bd1a16cef38af2a17cccc2bf71d92e0b07))

## [0.14.6-alpha.4](https://git.act3-ace.com/ace/data/tool/compare/v0.14.6-alpha.3...v0.14.6-alpha.4) (2021-05-28)


### Bug Fixes

* fix rc stuff. update docs for testing ([376c880](https://git.act3-ace.com/ace/data/tool/commit/376c8801e66a8185310b1c0fa7cac595ce3ad0ba))
* merge with master ([2ead8f3](https://git.act3-ace.com/ace/data/tool/commit/2ead8f32ee61c228edc46fa9e9e58487bef07266))
* start working on docs. Re-add docs that were removed ([7a77816](https://git.act3-ace.com/ace/data/tool/commit/7a778166e3331120e7785073fca0c5dc50c01a0f))

## [0.14.6-alpha.3](https://git.act3-ace.com/ace/data/tool/compare/v0.14.6-alpha.2...v0.14.6-alpha.3) (2021-05-27)


### Bug Fixes

* add a fix msg so the release pipeline will do something ([2745ae6](https://git.act3-ace.com/ace/data/tool/commit/2745ae6dd4f693652bfddf3201aff22809f6ed9b))

## [0.14.6-alpha.2](https://git.act3-ace.com/ace/data/tool/compare/v0.14.6-alpha.1...v0.14.6-alpha.2) (2021-05-27)


### Bug Fixes

* finalize test breaking on sha256: label on digests ([fc79c59](https://git.act3-ace.com/ace/data/tool/commit/fc79c5975971c342cacafe6d8b0802348e040eeb))

## [0.14.6-alpha.1](https://git.act3-ace.com/ace/data/tool/compare/v0.14.5...v0.14.6-alpha.1) (2021-05-27)


### Bug Fixes

* resolve short registry refs with docker style reference parsing ([9ee951d](https://git.act3-ace.com/ace/data/tool/commit/9ee951d0aa894996adfe91c4f0197d815314669f))
* use default authorizer for wildcard tag list on mirror pull ([a841c66](https://git.act3-ace.com/ace/data/tool/commit/a841c66869abec4dc2449b270cbcfd11340998d2))
* watch working dest exists for existing blobs rather than final dest exists, and add sha256 specifier in front of blob digests ([c3be019](https://git.act3-ace.com/ace/data/tool/commit/c3be019a4012b9f7de954914fb510cdc2dd8c981))

## [0.14.5](https://git.act3-ace.com/ace/data/tool/compare/v0.14.4...v0.14.5) (2021-05-25)


### Bug Fixes

* remove automatic addition of slash when prefix replacing an empty string, change destmap separator to comma ([f0f7205](https://git.act3-ace.com/ace/data/tool/commit/f0f7205beadd97d06528f3c434bad4fc42c71cfe))

## [0.14.4](https://git.act3-ace.com/ace/data/tool/compare/v0.14.3...v0.14.4) (2021-05-25)


### Bug Fixes

* Always add manifests to index, even if they are already present, handling repos with same data, but different tags ([abda645](https://git.act3-ace.com/ace/data/tool/commit/abda64597544d92768a6aa2f2a2eaf6bb47d70e2))
* correct file permissions when creating files on mirror pull ([f617c72](https://git.act3-ace.com/ace/data/tool/commit/f617c72334c45de2319085e534fe707470ec832a))
* sanitize repo map input a bit more to avoid double slashes on bulk-prepend ([d649074](https://git.act3-ace.com/ace/data/tool/commit/d649074b6f307ace92da2fa26b38c18ece465fd8))

## [0.14.3](https://git.act3-ace.com/ace/data/tool/compare/v0.14.2...v0.14.3) (2021-05-24)


### Bug Fixes

* close dest-exist file after writing updates ([94f4708](https://git.act3-ace.com/ace/data/tool/commit/94f4708a37af2b45a3ec913ccdf629eb77562e39))

## [0.14.2](https://git.act3-ace.com/ace/data/tool/compare/v0.14.1...v0.14.2) (2021-05-21)


### Bug Fixes

* functional tests after repo-map change, add quote aware string split and tests, temporarily remove wildcard mirror source test ([706a62a](https://git.act3-ace.com/ace/data/tool/commit/706a62a11a21a892100e57c145f7b3c21a61dc52))
* rework repo map to use a prefix match approach ([aa3ed81](https://git.act3-ace.com/ace/data/tool/commit/aa3ed8130680531a094ee4057d20bb239e551565))

## [0.14.1](https://git.act3-ace.com/ace/data/tool/compare/v0.14.0...v0.14.1) (2021-05-20)


### Bug Fixes

* use source digest media type on mirror push (instead of always oci.manifest) ([e1ccc43](https://git.act3-ace.com/ace/data/tool/commit/e1ccc431cf77e6f815abde2bc72e582e8c1955d2))

# [0.14.0](https://git.act3-ace.com/ace/data/tool/compare/v0.13.0...v0.14.0) (2021-05-20)


### Bug Fixes

* actually supress progress bars during tests ([a9bf0bc](https://git.act3-ace.com/ace/data/tool/commit/a9bf0bcbe79e7f80bf9b3429291549ab239ef15a))
* add missing 401 response authorization for manifest push in mirror push ([a4f3b7f](https://git.act3-ace.com/ace/data/tool/commit/a4f3b7fe5999d732c708e5bb1f4dd41ab452cbaf))
* avoid including manifests in index.json if no related blobs are fetched and manifest already exists ([0e9fe62](https://git.act3-ace.com/ace/data/tool/commit/0e9fe62f494cc3e05e0eda26597ba67a1050263c))
* cached credentials from docker auth causing failure when each repository requires separate permissions ([30e7543](https://git.act3-ace.com/ace/data/tool/commit/30e7543515b04854c59a335eca223aeaacacc0a9))
* changes how commands are fed into cmd tests ([6c0a58d](https://git.act3-ace.com/ace/data/tool/commit/6c0a58dbdd44c543d9cd043ffccf8dffe04fac49))
* defer closing error channel to avoid races ([45f9aa2](https://git.act3-ace.com/ace/data/tool/commit/45f9aa2a6da70e275cf55a43d20fe3107c0d3561))
* disable progress bars during tests ([52b2b42](https://git.act3-ace.com/ace/data/tool/commit/52b2b429e2b6a508f97b89b2ff558ce851dcaa46))
* do not perform download when requested usage is not found, and display valid choices to user ([f9d5e8c](https://git.act3-ace.com/ace/data/tool/commit/f9d5e8c0cc42db5f2ec5bc9b1cc0bc91e40c6ef5))
* fix path for mirror artifacts ([999c2b9](https://git.act3-ace.com/ace/data/tool/commit/999c2b9001e1d09871e0afdb004c7928386b607d))
* gitlab cicd reference to lastest tag ([259349c](https://git.act3-ace.com/ace/data/tool/commit/259349c6b70547780ee9d0160e500b3a72ea5bc1))
* improve thread safety of cache (filesystem storage provider), tweak ref parsing ([7322b4d](https://git.act3-ace.com/ace/data/tool/commit/7322b4dc2f6929c236c025dc0072b7ea377f5b7e))
* load existing index.json during fetch to allow multi step mirror sessions ([fe111cc](https://git.act3-ace.com/ace/data/tool/commit/fe111cccd6ed2ae47293395a48bfec0c84f17f8b))
* merge with ace-dt mirror cmd ([2a2414d](https://git.act3-ace.com/ace/data/tool/commit/2a2414dd067c6c94c3f6e2fff21aa7d7d3fba5a1))
* merge with master ([8c278da](https://git.act3-ace.com/ace/data/tool/commit/8c278da69ff2791ff40992a0f1ffd8d4cfb8a76d))
* merge with master ([8afe658](https://git.act3-ace.com/ace/data/tool/commit/8afe658b8a01ad5135d5dcfc08fa25c0941e4efd))
* missed bottle delete in monotest ([7a95f1e](https://git.act3-ace.com/ace/data/tool/commit/7a95f1ee6bd2930e4bbeeb0ac090fdb965603697))
* move edit to bottle ([44aed7d](https://git.act3-ace.com/ace/data/tool/commit/44aed7db22691ac81040a4766aa6115cd896f662))
* properly mark goroutines as done and block for write opts in routines ([30e449d](https://git.act3-ace.com/ace/data/tool/commit/30e449d0efbe10b1848c9e46796468e55353fda4))
* redo how I manage subcommands in cmd tests ([60a22fe](https://git.act3-ace.com/ace/data/tool/commit/60a22fe11c42f81ce83a8cf32ae0ae52b92eb168))
* remove auth sub command ([6023f52](https://git.act3-ace.com/ace/data/tool/commit/6023f52aeecb5a7e0fbea3827ae95d70bdad154e))
* remove goroutine from commit changes stage ([3d59b2f](https://git.act3-ace.com/ace/data/tool/commit/3d59b2f12a53159c000ef44dc2584677c6d132d8))
* respect show_progress user setting for mirror fetch ([1b823fb](https://git.act3-ace.com/ace/data/tool/commit/1b823fbc9afede8c70a6736b5a7e99a2e2285310))
* sync map ([fa2beaf](https://git.act3-ace.com/ace/data/tool/commit/fa2beaf74380939a73720ec7d4797a071ade6ef0))
* update goroutines to better handle parallel writes ([17d2e87](https://git.act3-ace.com/ace/data/tool/commit/17d2e87d8d9d3671c74f50777ec4db338956ece0))


### Features

* add basic mirror fetch cmd test ([0bf193b](https://git.act3-ace.com/ace/data/tool/commit/0bf193bf2aaf9c455e886b1ec575c90109b3bead))
* add blobledger checking to tests ([982df7d](https://git.act3-ace.com/ace/data/tool/commit/982df7d45fc03a9b7b387f8fcc7fd40a32a36b89))
* add goroutines to archive step ([5047e6a](https://git.act3-ace.com/ace/data/tool/commit/5047e6ab0929687b09b560ac52016e26ea9c33b8))
* add large bottle testing to mono test ([5acae3a](https://git.act3-ace.com/ace/data/tool/commit/5acae3a20113c3fd1ab5720cfdfdb403682de47c))
* added support for mirror push, and source file wildcards ([ab0bb55](https://git.act3-ace.com/ace/data/tool/commit/ab0bb55a062b8efddad3f560169bc31736be4a45))
* decent functional tests for ace-dt mirror fetch ([6edc001](https://git.act3-ace.com/ace/data/tool/commit/6edc001be84c3f68bc8b8c846f45c4e70afd5be1))
* merge with mirror and archive-compress branches ([44c776b](https://git.act3-ace.com/ace/data/tool/commit/44c776bbd73e5a867e632adb8f496d9816887804))
* support mirror finalize ([d6033af](https://git.act3-ace.com/ace/data/tool/commit/d6033af14255867486f55f05ebf560b27e73260e))
* **mirror:** Add mirror command set to fetch layers based on repo name and push them to a different registry ([f93fc2c](https://git.act3-ace.com/ace/data/tool/commit/f93fc2c7e878f64d3872442a31919422809d391c))
* **mirrorfinalize:** add session semantics to mirror by using a working state version of destination exists ([5bfaf53](https://git.act3-ace.com/ace/data/tool/commit/5bfaf53ce80c3d53aff8a7e05b927067c0a6bb4c))
* re-organized commands into sub-command groups ([c8f551d](https://git.act3-ace.com/ace/data/tool/commit/c8f551d32781050e430e6806ef669842e8389ea2))
* support goroutines for each commit / push / archive / digest / etc step ([4f1b3db](https://git.act3-ace.com/ace/data/tool/commit/4f1b3dbf5dde5d4a6ff02cf13e094f50b3aba414))

# [0.13.0](https://git.act3-ace.com/ace/data/tool/compare/v0.12.2...v0.13.0) (2021-04-23)


### Bug Fixes

* **labels:** fix labels not populating from labels.yaml files on commit and push ([404a758](https://git.act3-ace.com/ace/data/tool/commit/404a75822928fc89a134fb4c82ff8eb7c10020d7))


### Features

* **partselect:** Add the ability to select by part title on pull (shortcut for existing functionality), and fix migration error when parts not found ([b0cbde1](https://git.act3-ace.com/ace/data/tool/commit/b0cbde15244fa492e51c55ee577f218e41f69221))
* **usageselector:** Add the ability to select parts based on usage topic names, specified with -u command line option on pull ([3836e8b](https://git.act3-ace.com/ace/data/tool/commit/3836e8bc5ded174982c85690c5798709a220973e))

## [0.12.2](https://git.act3-ace.com/ace/data/tool/compare/v0.12.1...v0.12.2) (2021-04-16)


### Bug Fixes

* **validargs:** use a log trace message vs an error if registry list configuration list isn't found ([1782649](https://git.act3-ace.com/ace/data/tool/commit/17826490c4cb95394fbe03986c200ab506ad4ff2))
* **yaml:** change instances of URL to url to match comments ([b9b2554](https://git.act3-ace.com/ace/data/tool/commit/b9b2554e4f223c130f666bc164ac1105cf7f4f0a))

## [0.12.1](https://git.act3-ace.com/ace/data/tool/compare/v0.12.0...v0.12.1) (2021-04-16)


### Bug Fixes

* **yaml:** Improve yaml generation for bottle metadata to incorporate struct tags properly ([a4d1414](https://git.act3-ace.com/ace/data/tool/commit/a4d141453393574e25126093ea88720f49c7164d))

# [0.12.0](https://git.act3-ace.com/ace/data/tool/compare/v0.11.6...v0.12.0) (2021-04-15)


### Bug Fixes

* **label:** add input validation to label key and values to match k8s restrictions ([c3d09f7](https://git.act3-ace.com/ace/data/tool/commit/c3d09f78f212f551bd936367907e5cfc8282e427))
* **show:** fix content filter on show to prevent transferring full bottle contents when looking for just the config. ([fb07733](https://git.act3-ace.com/ace/data/tool/commit/fb07733a306c92d6badc65d7043d87cc2ced3566))
* initial name change changes ([966c910](https://git.act3-ace.com/ace/data/tool/commit/966c910b3c361535e19a59cf8d5df6a289a9df26))
* repair crashing stream processing pipeline ([5ac94b6](https://git.act3-ace.com/ace/data/tool/commit/5ac94b611cbb6e9d631374bbb5966014101e5c47))
* **migration:** fix migration to correctly calculate content size and digest, and fix a concurrency issue with manifest map ([825104b](https://git.act3-ace.com/ace/data/tool/commit/825104b994ca4ff97690cd9877a976723195d621))


### Features

* **const_content:** remove mutable data from bottle config, format, compressed size, and using content digest ([61cfb23](https://git.act3-ace.com/ace/data/tool/commit/61cfb2315333c4dda0a2f25b363b2f6e83795545))
* **raw-format:** Add raw media type that explicitly indicates a file that is not to be compressed ([2abb59a](https://git.act3-ace.com/ace/data/tool/commit/2abb59a1949251643df8bd1a7227632239bb7127))
* **repomount:** Add the ability to perform a cross-repo mount of a layer if an existing repo source is known ([0d18554](https://git.act3-ace.com/ace/data/tool/commit/0d18554d2a1dad5cfae3f11a7641edb040223803))
* **usage:** Expand usage metadata to allow multiple files to be specified with topics ([8cc3753](https://git.act3-ace.com/ace/data/tool/commit/8cc3753260273c87a5851ffb1fb19f64f5051a5e))
* added test cases for repo ([5866436](https://git.act3-ace.com/ace/data/tool/commit/586643677fd949c10095bae84397a808adcdb5f6))

## [0.11.6](https://git.act3-ace.com/ace/data/tool/compare/v0.11.5...v0.11.6) (2021-04-01)


### Bug Fixes

* **commit:** fix path resolution when committing current directory ([779835c](https://git.act3-ace.com/ace/data/tool/commit/779835cb6465887a7b20a029b264acc1d4135356))
* **index:** continue processing repos after one in the list is reported as not found ([a268dc2](https://git.act3-ace.com/ace/data/tool/commit/a268dc272d694c5f005bfbf9753c6d07052f579d))
* **push:** enforce lowercase for repository and bottle names, and verify valid characters ([aad9444](https://git.act3-ace.com/ace/data/tool/commit/aad944410db61624e49d51a84dac484ed20d1466))

## [0.11.5](https://git.act3-ace.com/ace/data/tool/compare/v0.11.4...v0.11.5) (2021-03-24)


### Bug Fixes

* update datastore writer to accomodate oras change that delivers manifest as content file ([7165537](https://git.act3-ace.com/ace/data/tool/commit/71655371b0835b5df95f8c35a3fa7d08ac9d7ef2))
* **archive:** close files opened during archive operation ([db83d91](https://git.act3-ace.com/ace/data/tool/commit/db83d91e687428fb2bdc97509f07d658941c3286))
* **ci:** updated to go 1.16 ([bfe22fb](https://git.act3-ace.com/ace/data/tool/commit/bfe22fb9b35de71fa84f77de94e2c9829d117240))

## [0.11.4](https://git.act3-ace.com/ace/data/tool/compare/v0.11.3...v0.11.4) (2021-03-18)


### Bug Fixes

* update gitlab-ci to pull in a publishing pipeline fix ([6d67f62](https://git.act3-ace.com/ace/data/tool/commit/6d67f62c21458be8eab29c6c023fd95d136b5b9d))

## [0.11.3](https://git.act3-ace.com/ace/data/tool/compare/v0.11.2...v0.11.3) (2021-03-18)


### Bug Fixes

* update ci to use new pipeline version with corrected publishing destination (gitlab packages) ([952bc9a](https://git.act3-ace.com/ace/data/tool/commit/952bc9a23efe118633a5545ba39be2176d3acc1b))

## [0.11.2](https://git.act3-ace.com/ace/data/tool/compare/v0.11.1...v0.11.2) (2021-03-16)


### Bug Fixes

* resolve cwd instead of using './', corrects problem pushing when inside bottle dir ([b3e4f7c](https://git.act3-ace.com/ace/data/tool/commit/b3e4f7c28e4eabd9bbddf87db486da722cecac44))

## [0.11.1](https://git.act3-ace.com/ace/data/tool/compare/v0.11.0...v0.11.1) (2021-03-15)


### Bug Fixes

* Update pipeline version in .gitlab-ci.yml ([4523426](https://git.act3-ace.com/ace/data/tool/commit/4523426ca6ec454fbaa9ae00f35b03370db5a40a))

# [0.11.0](https://git.act3-ace.com/ace/data/tool/compare/v0.10.1...v0.11.0) (2021-03-12)


### Bug Fixes

* fix url ([46d0887](https://git.act3-ace.com/ace/data/tool/commit/46d08872bcda430701f11730e875a74f59a40928))
* **ci:** updated the pipeline version ([90154cb](https://git.act3-ace.com/ace/data/tool/commit/90154cb291edceb0ef439dce329f0d1570f8c58a))
* remove old code and move version string back to main ([22076ca](https://git.act3-ace.com/ace/data/tool/commit/22076ca8742b097c300d7bf9671cf275f3c09726))
* update delete to pass linter ([16bc223](https://git.act3-ace.com/ace/data/tool/commit/16bc223ae7597e842eb3d40cae89ef94401ad8e5))
* verify ref format for valid tags, and apply URL escapes to ref components ([f014c79](https://git.act3-ace.com/ace/data/tool/commit/f014c7964b692c956198a19c662f05b54b027631))
* **cmderr:** remove usage and extraneous output when a command error occurs (Again) ([8f9e556](https://git.act3-ace.com/ace/data/tool/commit/8f9e556973857f0726bee5f209f95c4cafb32d58))
* **commit:** Remove file processing functions from bottle init, in favor of performing these steps in commit and push ([8817ee6](https://git.act3-ace.com/ace/data/tool/commit/8817ee6750e22a45bed2bf04201b1c0437e53041))
* **pushlayers:** fix crash when pushing zero byte files ([815ca96](https://git.act3-ace.com/ace/data/tool/commit/815ca96f28038e56e87259a90e775f6ea452cd11))
* **timer:** fix how passage of time is calculated for certain timer tests ([2aad721](https://git.act3-ace.com/ace/data/tool/commit/2aad7210ba6bcf0174825e5222eb02418a546f86))


### Features

* **GCP:** Support Google Cloud Platform as registry destination ([17bbeca](https://git.act3-ace.com/ace/data/tool/commit/17bbeca4ded68f9217c77e7918618d7d7668c274))
* **labels:** Add label marker files for better label management, and bottle part partitioning and recursion. ([e5f476a](https://git.act3-ace.com/ace/data/tool/commit/e5f476a8e4b2189c61dac090677833a83d4bdc57))
* add working delete functionality ([45ffd95](https://git.act3-ace.com/ace/data/tool/commit/45ffd95e34b3da181a8289b2f491549408c10656))
* added completion support for GCP and Gitlab registries. ([8d9ca31](https://git.act3-ace.com/ace/data/tool/commit/8d9ca313358fce26232453d133837828fb2cfdaa))
* added proper configuration to bash completion for tag completion ([3a42e30](https://git.act3-ace.com/ace/data/tool/commit/3a42e308dda495290e5250f309fcbcaa684cc1e9))
* added tag completion for show and pull commands ([384dfe3](https://git.act3-ace.com/ace/data/tool/commit/384dfe3576b0a6556eaa6b0d246f79ad459ef204))
* **commit:** Add commit command that applies part and label changes to existing bottle ([208f431](https://git.act3-ace.com/ace/data/tool/commit/208f431679b8e90604630c946fb34726dabfca4e))
* **pushchange:** Detect changed parts on push and update bottle ([4a8573c](https://git.act3-ace.com/ace/data/tool/commit/4a8573c0e4a953531ea0531be124ab56b40da648))

## [0.10.1](https://git.act3-ace.com/ace/data/tool/compare/v0.10.0...v0.10.1) (2021-02-10)


### Bug Fixes

* **auth:** remove dependency on customized oras for authentication ([efebef0](https://git.act3-ace.com/ace/data/tool/commit/efebef08f720d425d8dc7fdacd4d5a8f21283afe))

# [0.10.0](https://git.act3-ace.com/ace/data/tool/compare/v0.9.0...v0.10.0) (2021-02-09)


### Bug Fixes

* **cmderr:** fix a missing error check in archive.go ([265be21](https://git.act3-ace.com/ace/data/tool/commit/265be2177df90ef13c03907a9f590d37760c817c))
* fix additional lint errors in layers.go ([564640a](https://git.act3-ace.com/ace/data/tool/commit/564640a3d8b81e0ac0e4d174dbba01cfbd91e0ca))
* Fix numerous lint complaints from golangci-lint ([64aa1ae](https://git.act3-ace.com/ace/data/tool/commit/64aa1ae8238e7c55301cf8add1582d61dd66dfea))
* **archive:** fix output path generation for non compressed archives ([53a6b67](https://git.act3-ace.com/ace/data/tool/commit/53a6b67bff8c29e74e48270d048864d4158aafde))
* **auth:** cause credential provider to create default store if doesn't exist ([2974ce2](https://git.act3-ace.com/ace/data/tool/commit/2974ce2a1edb85dd4a6acc478036189bebd5a376))
* **backoff:** Fix regression in backoff calculaton ([b8cbdf2](https://git.act3-ace.com/ace/data/tool/commit/b8cbdf2b1913966302106c3d47f6202d88fcafe2))
* **ci:** update gitlab-ci to 2.7.2 ([65d6368](https://git.act3-ace.com/ace/data/tool/commit/65d63685456e9a42645ab71d9648e83dd196e307))
* **docs:** update documentation for the index command for new behavior ([28880c4](https://git.act3-ace.com/ace/data/tool/commit/28880c4167065f1f9aee0b11eec447ffaa41b57a))
* **docs:** updated the docs ([ffb2780](https://git.act3-ace.com/ace/data/tool/commit/ffb27804cdac5c65edfc17111168ff493d0252eb))
* **index:** fix early termination when indexing non bottle repos, and create local index path before save ([9f729cd](https://git.act3-ace.com/ace/data/tool/commit/9f729cdfc6105074251b54762570a3b0451ced47))
* **index:** improve tolerance for including a repo list file when querying an unauthenticated registry ([c123fef](https://git.act3-ace.com/ace/data/tool/commit/c123fefa60fac17d6435ce5a5d9f739874c5b5e5))
* **init:** fix cache disable feature not functional in init ([1478d60](https://git.act3-ace.com/ace/data/tool/commit/1478d6056fdd4e92501609d98279a18ddcebef96))
* **layerpush:** correct missing error indication when push fails for layer ([a25f315](https://git.act3-ace.com/ace/data/tool/commit/a25f3153306bb74b0665573a9dee0b952f79aed3))
* lint complaints for selector inequality feature ([36a99a8](https://git.act3-ace.com/ace/data/tool/commit/36a99a80c26b28b837af8bd8ce07f6696fc87279))
* max chunk size key type mismatch causing missing value ([9ea7471](https://git.act3-ace.com/ace/data/tool/commit/9ea74718149d4754fd798542bfcc8ef1b91b2c10))
* **repo:** fix path determination when creating a new repository.yml file ([5e0bf0e](https://git.act3-ace.com/ace/data/tool/commit/5e0bf0ef9e5124dfbbc9a007b1d127d1784a3b6e))
* **selector:** updated getSelectors to combine multiple selectors with OR logic ([57d6e73](https://git.act3-ace.com/ace/data/tool/commit/57d6e7317b0b311793ad90bc00191c95c2b0c846))
* **semantic release:** include .version.yaml in release ([2b9dd35](https://git.act3-ace.com/ace/data/tool/commit/2b9dd357018bdac991a9a3c55e3bd7a5770710c5))
* **test:** fix tests for ref parsing ([5077c9d](https://git.act3-ace.com/ace/data/tool/commit/5077c9d5d4a0d60c669db6e500222711abc63c23))
* **upgrade:** return error if source config does not contain apiVersion ([a8adfd9](https://git.act3-ace.com/ace/data/tool/commit/a8adfd965f3b93aa2ce8b4ac821eedee2c4ad1d7))
* host name parse return url if parse doesn't identify host ([4f6dea2](https://git.act3-ace.com/ace/data/tool/commit/4f6dea2ce1064a4368313d2e74375392c1318132))
* update error message to not refer to Oras Pull when checking for bottle status during push ([d5dc6bb](https://git.act3-ace.com/ace/data/tool/commit/d5dc6bb93a889e6f124c8f50ef914abe7e7c1257))


### Features

* **chunkupload:** Add chunked upload support for push via modified opencontainers docker resolver code ([a9ffc40](https://git.act3-ace.com/ace/data/tool/commit/a9ffc402106ff4d9a625316b080cb8d62bea6d1a))
* **cmderr:** Add command line return OS error codes ([85b5413](https://git.act3-ace.com/ace/data/tool/commit/85b54135a9aa86527f8ee624e7b69763d7df5059))
* **docs:** added functionality to print tutorials via ace-bottle tutorial cmd ([9dfaa60](https://git.act3-ace.com/ace/data/tool/commit/9dfaa60c454c86ab9a6c481ddacfaa91442705e4))
* **index:** allow explicit repositories to be queried versus searching a catalog api ([3b4618e](https://git.act3-ace.com/ace/data/tool/commit/3b4618e50c39cac55501361eeaa25feb61b0aa5d))
* **index:** Enable index command to work with authenticated repositories through a yaml file repo list and JWT auth ([7e188cb](https://git.act3-ace.com/ace/data/tool/commit/7e188cb159262a0266442c6f7b0e4dfd9fcbc1f8))
* **layerpush:** Add custom chunked blob upload with chunk level retry and backoff ([bc21557](https://git.act3-ace.com/ace/data/tool/commit/bc215576c9e1c97819722cd580e8bf2a0fec0784))
* **layerpush:** add progress and authentication to layer push ([68853c8](https://git.act3-ace.com/ace/data/tool/commit/68853c803eb7e393595b0eb3a9db2443b8d6fd06))
* **migration:** Add config schema version history with version subdirectories ([c8cfc74](https://git.act3-ace.com/ace/data/tool/commit/c8cfc7419f6eafaa2972cdcdcb1fbcab7bcc40d6))
* **migration:** Add version migration framework ([256f5bd](https://git.act3-ace.com/ace/data/tool/commit/256f5bd4c9ef635333135d70212a81ec4aff6b2e))
* **perf:** use compression ratio during archive to skip compression if it doesn't meet a threshold ([670dbef](https://git.act3-ace.com/ace/data/tool/commit/670dbef5cca506d9e8a4988cdcf9e24c17e6100b))
* **repo:** Add repo command set, managing known registries and repositories ([02a7c72](https://git.act3-ace.com/ace/data/tool/commit/02a7c72701fabcdb178e0aceee2d3024da5fd462))
* **scratch:** use configurable scratch location for creating archives, near cache to avoid cross-volume moves ([eaa55d4](https://git.act3-ace.com/ace/data/tool/commit/eaa55d4320fe6aedda83d053c0722367ad331752))
* **search:** Improve search feature to include bottle source information in results ([794d3eb](https://git.act3-ace.com/ace/data/tool/commit/794d3ebb822cd2c9453356fc5722b739ccc137c4))
* **selector:** Implement numerical inequality for selectors ([718792e](https://git.act3-ace.com/ace/data/tool/commit/718792e9be3345c04df40899bc9374449407f8e1))
* **show:** Add bottle schema version to show data output ([6c0a011](https://git.act3-ace.com/ace/data/tool/commit/6c0a0119cd631f16364863b5d275f9847dac2d81))
* **upgrade:** Add upgrade command to upgrade API versions locally, and integrate upgrade step into push ([c8968d9](https://git.act3-ace.com/ace/data/tool/commit/c8968d98a8b997f13886b42b019f9d891dfa0729))

# [0.9.0](https://git.act3-ace.com/ace/data/tool/compare/v0.8.1...v0.9.0) (2020-11-13)


### Bug Fixes

* **archive:** correct calculation of relative path, and inclusion of TLD in tar ([6f4f8b1](https://git.act3-ace.com/ace/data/tool/commit/6f4f8b1da29a24e493a41a4e11c950d11f261da9))
* **archive:** fix tests to correspond to archive and extract changes ([0ad80ea](https://git.act3-ace.com/ace/data/tool/commit/0ad80eab1286e5a44eb15e2858dd7259090b086e))
* **cache:** fix creation of empty cache dir in wrong place, and add nil cache for disabling cache commit ([5080d1d](https://git.act3-ace.com/ace/data/tool/commit/5080d1d8185d3306124953c27c7d81e9d3515a93))
* **init:** add missing digest information output ([5edbf4a](https://git.act3-ace.com/ace/data/tool/commit/5edbf4a2790011ae927c3f44505e2571ad99cd46))
* remove incorrect stream close from std that caused files to be closed before the compressor flushed ([5422fe4](https://git.act3-ace.com/ace/data/tool/commit/5422fe4c8d9ef463cc6c148343f5d38ea0c0bf2c))
* **cache:** translate unix style home path to user home directory path for cross platform support ([382918b](https://git.act3-ace.com/ace/data/tool/commit/382918b007e2864b6d67b070c5441b93691b9ea0))
* **pipestream:** Fix intermittent panic when reading digest from non-closed pipeline ([5d08ffa](https://git.act3-ace.com/ace/data/tool/commit/5d08ffa89d618515789b94a9cb74b61756bafa49))
* remove test code from show command ([0227e7d](https://git.act3-ace.com/ace/data/tool/commit/0227e7d9f66fcae9c6d063cfadc2895e706060dd))
* **cache:** change default cache path to use XDG_CACHE_HOME/ace-bottle ([97b7225](https://git.act3-ace.com/ace/data/tool/commit/97b72250dbf03a745c251050960eb30126557b38))
* **prune:** fix max size calculation to work in MiB ([8f31491](https://git.act3-ace.com/ace/data/tool/commit/8f314913380d103f2336774d9ed6fae8464f25f0))
* **search:** fix error when yaml marshalling time structures ([6ef5cf9](https://git.act3-ace.com/ace/data/tool/commit/6ef5cf9453ce09a71eb712ddd565f87ff91d693a))
* **status:** fix directory date check to avoid  being spotted as a change ([13c0a96](https://git.act3-ace.com/ace/data/tool/commit/13c0a9605dda96b7f0e40f4135ddb93785dceec9))
* **tests:** remove tests for deprecated failure on empty file iterator callback ([d573604](https://git.act3-ace.com/ace/data/tool/commit/d5736047a5c8f3e5395dc4d2a37ae20a8017e2e5))
* **util:** commented out premature local tests ([a3960ed](https://git.act3-ace.com/ace/data/tool/commit/a3960ed0be82a8c734fba1bcb85308bf9f46f034))
* remove uses of path in favor of filepath package ([b286eea](https://git.act3-ace.com/ace/data/tool/commit/b286eea04041369b18ccb1356718358fe6f4b050))


### Features

* **cache:** add a simpler default file cache that does not maintain a metadata index for the cache ([c91b4b2](https://git.act3-ace.com/ace/data/tool/commit/c91b4b252371c5f73218f8dbbb47b0c3e22e6f01))
* **cache:** add caching index and file handler utilities ([409a1eb](https://git.act3-ace.com/ace/data/tool/commit/409a1eb07e13cce5f824017a1f39028cafd86236))
* **cache:** Add CommitMote to allow direct addition of data to cache, and consolidate final writes ([3faa998](https://git.act3-ace.com/ace/data/tool/commit/3faa998d393b0a5e7615a095417e54f2858e6246))
* **cache:** cache local files during init and push ([6af851f](https://git.act3-ace.com/ace/data/tool/commit/6af851fd9aed1d3ba8e095c152e1d72d0c45e116))
* **cache:** hardlink cached file blobs before reading to enable safe deletion ([714178b](https://git.act3-ace.com/ace/data/tool/commit/714178ba8e7c0444f51065261bd4779d7a17bf4a))
* **cache:** integrate cache into init, push, and pull operations ([2523a15](https://git.act3-ace.com/ace/data/tool/commit/2523a150a4c2e19d56aed971163f6b39f9fd8b79))
* **cache:** integrate cache manager into ace bottle operations ([e7c95fc](https://git.act3-ace.com/ace/data/tool/commit/e7c95fcbc2861bd64a01a88ee3663dec5087a8b2))
* **cache:** use double move to avoid nonatomic cross-volume rename collisions ([74a904f](https://git.act3-ace.com/ace/data/tool/commit/74a904ff33b4c968e0a908b8758028097ab7a2a3))
* **filelistgen:** Add a file list generator, python style, allowing a flat loop over a file list iterator with a common interface ([4e2a4d5](https://git.act3-ace.com/ace/data/tool/commit/4e2a4d5354209b076252f42f875d1bdcbd4c4ded))
* **pipestream:** Add gzip compressing pipe segment ([8766fe2](https://git.act3-ace.com/ace/data/tool/commit/8766fe27032afe634dffa0793ebeaacb169192f9))
* **pipestream:** Add modular stream pipeline enabling selective in-stream operations such as counting and digest ([5b9660e](https://git.act3-ace.com/ace/data/tool/commit/5b9660e6fc01eb3b8cba8a158186c3d509b07157))
* **prune:** Add prune command to reduce cache by removing LRU files ([ed0bd55](https://git.act3-ace.com/ace/data/tool/commit/ed0bd55adc6f96030702661e207ced8a0ed381c6))
* **prune:** Update config and cache to support prune command ([fb54e9d](https://git.act3-ace.com/ace/data/tool/commit/fb54e9da004544f7567fe492b5217011d55fa82d))
* **schema:** add error check for schema mismatch on pull ([0c7c177](https://git.act3-ace.com/ace/data/tool/commit/0c7c177a162102ca732b624cd7f3e6f5464008fc))
* **schema:** add uncompressed size to bottle schema ([bdc1855](https://git.act3-ace.com/ace/data/tool/commit/bdc18553a3a632d1037e99fd165917bfeba1f119))
* **status:** add Deleted files to status list, and optionally verify by digest, and show directory details ([61241df](https://git.act3-ace.com/ace/data/tool/commit/61241df8809c4f7b425a311778fd96b92862f66b))
* **status:** Add status command to display cached, updated, and new files in bottle ([2f331bf](https://git.act3-ace.com/ace/data/tool/commit/2f331bf7c734601edde67043295a2d386014e12b))
* **status:** compare uncompressed file and directory sizes ([ccd146f](https://git.act3-ace.com/ace/data/tool/commit/ccd146f5a3fb1f6300c92dd0192ad2a4785323d4))
* **tarstream:** Add a stream based tar archiver, using file list FData instead of os specific file data ([ff89528](https://git.act3-ace.com/ace/data/tool/commit/ff8952840021e8acbf4832421c6941b4e49f7ce1))

# [0.8.0](https://git.act3-ace.com/ace/data/tool/compare/v0.7.5...v0.8.0) (2020-09-22)


### Features

* **index:** Add index command to display bottles and tags based on regex filter ([16336db](https://git.act3-ace.com/ace/data/tool/commit/16336db3b6b3f9e6ac16f4d8b1d8386e5a4fbeff))

## [0.7.4](https://git.act3-ace.com/ace/data/tool/compare/v0.7.3...v0.7.4) (2020-09-17)


### Bug Fixes

* **ci:** minor fix ([6281b74](https://git.act3-ace.com/ace/data/tool/commit/6281b74740e468b2c8bf35e34bb04de092bb7ce1))

## [0.7.3](https://git.act3-ace.com/ace/data/tool/compare/v0.7.2...v0.7.3) (2020-09-17)


### Bug Fixes

* **ci:** Added a version file. ([a66843a](https://git.act3-ace.com/ace/data/tool/commit/a66843a019e1d489237c2e51e206b7e4662fbee9))
* **ci:** Fixed .version.yaml name ([65ea72d](https://git.act3-ace.com/ace/data/tool/commit/65ea72df94fd9cc9a90591a7c0800b2c5f8ca7a4))
* **ci:** Test new pipeline ([5e09461](https://git.act3-ace.com/ace/data/tool/commit/5e094614b987bf72e49ea9e66e3b41adabf1522e))
