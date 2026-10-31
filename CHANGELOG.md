# Changelog

## [1.1.0](https://github.com/muhlba91/hochschule-burgenland-bswe-director/compare/v1.0.0...v1.1.0) (2026-10-31)


### Features

* add api versioning ([fd5f31a](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/fd5f31acda51cf81d742c0a6248b38185eba2c5f))
* add auth for endpoints ([826b6b7](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/826b6b74505a18d5440863ada440659f556adf82))
* add basic authentication and rate limiting ([b5d0a83](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/b5d0a8310999e182f002ceb040bc62a0543ecd32))
* add basic wait and retry logic for callbacks ([bf09d48](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/bf09d489139df405584c7a71457a79d11fd7ff2a))
* add helm chart ([d5fe534](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/d5fe534e37e60b5ecf695f7f38680b5a2723aefa))
* add initial request handling ([b0f58ef](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/b0f58ef7bd8b344a4e3df2fb9b403973e5c10d69))
* basic session handling and flow communication ([85b183f](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/85b183fe942b7a084b1de370390fb42884fca65b))
* implement basic turn logic ([08ec0d2](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/08ec0d2cc60a12f0cf8eac384f284583ba22d044))
* implement basic winner detection ([d4bec1c](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/d4bec1cea3fa3771020ac5f0b32c1e110a2d5c18))
* initial commit ([7715ec5](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/7715ec586af3cf1d51c76905613122841bc5e1a1))
* initialize basic session creation ([f830187](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/f830187e37fdac54e52ca32df16a570191c929b9))
* initialize generic base structure ([fcb5b7e](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/fcb5b7efa731a97a864309e225d3f857913a1074))
* upgrade go to v1.27.x ([651bddc](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/651bddc948bcb57eb6f87b1fdc4b750277086460))


### Bug Fixes

* always add callback secret but only validate signature when env var is set ([0d7d3d5](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/0d7d3d559f6a4cc5cbda995e6c8188ab6943bf08))
* **ci:** fix golangci error ([11e69c9](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/11e69c91ef29e43f2cbf5bfb165cbdfbe5fb3aff))
* declare card types for pokemon ([185c24c](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/185c24cda125614e1ab791d0da02bfddd7092565))
* fix deck value of opponent ([c88feef](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/c88feef78041218b60156d615d8944c3a0f6d1b0))


### Miscellaneous Chores

* **ci:** fix github workflow concurrency ([a40dd4b](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/a40dd4b0efafc448d1b80ee6b571e441c8ed59eb))
* cleanup fixme annotations ([2fc4c97](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/2fc4c97733c37919fb5fff10b07e5e23fdc195c0))
* **deps:** update actions/attest-build-provenance action to v4.2.2 ([d52694f](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/d52694f286dd432b28cfdee624f4fba4c564811a))
* **deps:** update actions/checkout action to v7.0.1 ([8b21924](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/8b219247c7e07b1513f3ba23b4ee1a4a264372b2))
* **deps:** update actions/checkout digest to d23441a ([c69d87a](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/c69d87a95ad689a9cc5a63596dd8155598a79cca))
* **deps:** update actions/upload-artifact action to v7.0.2 ([9e3b0e7](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/9e3b0e78cad34d93ee11e491f9d9fbf5ce39cfb8))
* **deps:** update anchore/sbom-action action to v0.24.3 ([c25472e](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/c25472e7db869f26827e3a568558dd67882c24a2))
* **deps:** update dependencies ([d48f3cf](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/d48f3cfce7d0643b38177d739c38b3f1cb08a153))
* **deps:** update docker/login-action action to v4.6.0 ([b0a2864](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/b0a286442b127e2026e3233336b0afe37d91e91c))
* **deps:** update docker/login-action action to v4.7.0 ([b5eed9b](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/b5eed9b3e2319ebfd96cb3167bfda919eba3a0e6))
* **deps:** update docker/setup-buildx-action action to v4.4.1 ([b98045d](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/b98045da10d35aed10175d9e5f020bdc166a45d6))
* **deps:** update docker/setup-qemu-action action to v4.4.0 ([a4df431](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/a4df4311ff50d87d4395504e82fb48de6975783e))
* **deps:** update gcr.io/distroless/static-debian12:nonroot docker digest to afa5c87 ([b3695b1](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/b3695b10d36c745376c35fd9ad187c9b3b118c2b))
* **deps:** update github/codeql-action action to v4.38.3 ([288d6fe](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/288d6fe32f8f78149fee79a9146ccfdfe1e95202))
* **deps:** update github/codeql-action action to v4.39.0 ([2b8cdf3](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/2b8cdf3cd75e2b26d2becb184b9b77f60f942775))
* **deps:** update github/codeql-action action to v4.39.1 ([b803cd1](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/b803cd105e3cfa1bb66879087c9f586fefd61d60))
* **deps:** update go module directive to v1.27.3 ([fb16f06](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/fb16f0630aaf59360a4c53143d467ff6598c7d87))
* **deps:** update goreleaser/goreleaser-action action to v7.3.0 ([ca7c156](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/ca7c1562ee0ca3bb64e7de756bf61975cf63e1e9))
* **deps:** update hadolint/hadolint-action action to v3.5.0 ([f435369](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/f435369f1afcba9b3619c5dcfb8e9307402e0011))
* **deps:** update jdx/mise-action action to v5.2.4 ([60e14cb](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/60e14cb8b1fde72743a6c264fa233486e4d799f2))
* **deps:** update jdx/mise-action action to v5.2.5 ([27eca21](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/27eca2175f74373f555396e8763ad7617a37ff10))
* **deps:** update labstack/echo to v5 ([78a9c4a](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/78a9c4a7b3d5f8126cc98b0e9ac2d4de8fba9e91))
* **deps:** update module github.com/go-redsync/redsync/v4 to v4.18.0 ([4cb3f14](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/4cb3f1427f4b0fb2d89bc3b21126cb9d92bda89d))
* **deps:** update module github.com/go-redsync/redsync/v4 to v4.18.1 ([cc6e706](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/cc6e706002eb192d96d2ec0cafce0190f4a8b7c3))
* **deps:** update module github.com/labstack/echo/v4 to v5 ([ee15ba7](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/ee15ba7787b7509f8182ed5499da9c557e756b82))
* **deps:** update module github.com/labstack/echo/v5 to v5.4.1 ([76af9d6](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/76af9d6c71151f9247d216ba9594bbf23e9f2d92))
* **deps:** update module github.com/labstack/echo/v5 to v5.4.2 ([3f5b748](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/3f5b7489a2987d8504efb084b99f9557e399196a))
* **deps:** update module github.com/labstack/echo/v5 to v5.4.4 ([10e7de2](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/10e7de2f7139dacbfe76dad506de10ee3389b7a3))
* **deps:** update module github.com/redis/go-redis/v9 to v9.23.0 ([b2468be](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/b2468be4c2256bb8ad67aef7b26f64f0fded90d2))
* **deps:** update ossf/scorecard-action action to v2.5.0 ([551b417](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/551b41770e2070100caefcee6e9dda61631dd2a6))
* **deps:** update siderolabs/conform action to v0.1.0-alpha.32 ([d751a03](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/d751a03fd4af2207bae3aa0ced151b018e3e85d1))
* **deps:** update step-security/harden-runner action to v2.22.1 ([a400736](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/a400736e0c74ef851178af93761ed960275dcaa1))
* **deps:** update step-security/harden-runner action to v2.22.2 ([ae72e23](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/ae72e23eb4cd0376ffe7c367bddf0aa1541c8cb8))
* **deps:** update step-security/harden-runner action to v2.22.3 ([0ba4161](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/0ba4161959b35c70ed9da3432d1db5a2760abdc4))
* migrate to mise ([5ccad2f](https://github.com/muhlba91/hochschule-burgenland-bswe-director/commit/5ccad2f3f148495a707ba525bb996e966f526087))
