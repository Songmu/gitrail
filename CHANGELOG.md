# Changelog

## [v0.0.21](https://github.com/Songmu/gitrail/compare/v0.0.20...v0.0.21) - 2026-09-23

- Fix gitsign OIDC authentication in tagpr by @Songmu in https://github.com/Songmu/gitrail/pull/80

## [v0.0.20](https://github.com/Songmu/gitrail/compare/v0.0.19...v0.0.20) - 2026-09-23

- use latest tagpr and setup-gitsign by @Songmu in https://github.com/Songmu/gitrail/pull/77
- restore setup-go by @Songmu in https://github.com/Songmu/gitrail/pull/79

## [v0.0.19](https://github.com/Songmu/gitrail/compare/v0.0.18...v0.0.19) - 2026-09-22

- update tagpr to v1.20.4 by @Songmu in https://github.com/Songmu/gitrail/pull/75

## [v0.0.18](https://github.com/Songmu/gitrail/compare/v0.0.17...v0.0.18) - 2026-09-22

- update ghr to v0.18.5 by @Songmu in https://github.com/Songmu/gitrail/pull/73

## [v0.0.17](https://github.com/Songmu/gitrail/compare/v0.0.16...v0.0.17) - 2026-09-22

- update install.sh with insmith by @Songmu in https://github.com/Songmu/gitrail/pull/69
- Separate tagpr release dispatch permissions by @Songmu in https://github.com/Songmu/gitrail/pull/71
- update goxz to newest by @Songmu in https://github.com/Songmu/gitrail/pull/72

## [v0.0.16](https://github.com/Songmu/gitrail/compare/v0.0.15...v0.0.16) - 2026-09-21

- install.sh: comment on signer-digest/source-digest usage by @Songmu with @Copilot in https://github.com/Songmu/gitrail/pull/65
- Refine release automation and policies by @Songmu in https://github.com/Songmu/gitrail/pull/67
- Harden curl downloads in install.sh by @Songmu with @Copilot in https://github.com/Songmu/gitrail/pull/68

## [v0.0.15](https://github.com/Songmu/gitrail/compare/v0.0.14...v0.0.15) - 2026-09-19

- Enforce trusted release provenance by @Songmu in https://github.com/Songmu/gitrail/pull/63

## [v0.0.14](https://github.com/Songmu/gitrail/compare/v0.0.13...v0.0.14) - 2026-09-19

- Specify repository for release publishing by @Songmu in https://github.com/Songmu/gitrail/pull/61

## [v0.0.13](https://github.com/Songmu/gitrail/compare/v0.0.12...v0.0.13) - 2026-09-19

- Use gocredits directly by @Songmu in https://github.com/Songmu/gitrail/pull/56
- Harden installer verification by @Songmu in https://github.com/Songmu/gitrail/pull/58
- Update vendored shlib to v2026.08.30 by @Songmu in https://github.com/Songmu/gitrail/pull/59
- Use reusable workflow for SLSA release provenance by @Songmu in https://github.com/Songmu/gitrail/pull/60

## [v0.0.12](https://github.com/Songmu/gitrail/compare/v0.0.11...v0.0.12) - 2026-09-19

- Update release action for goxz v0.11.1 by @Songmu in https://github.com/Songmu/gitrail/pull/54

## [v0.0.11](https://github.com/Songmu/gitrail/compare/v0.0.10...v0.0.11) - 2026-09-19

- remove release and upload target from Makfile by @Songmu in https://github.com/Songmu/gitrail/pull/50
- Harden action attestation verification by @Songmu in https://github.com/Songmu/gitrail/pull/52
- Use goxz action for release builds by @Songmu in https://github.com/Songmu/gitrail/pull/53

## [v0.0.10](https://github.com/Songmu/gitrail/compare/v0.0.9...v0.0.10) - 2026-09-19

- Add prepare-release make target for tagpr command by @Songmu with @Copilot in https://github.com/Songmu/gitrail/pull/47
- Add provenance attestations for release artifacts by @Songmu with @Copilot in https://github.com/Songmu/gitrail/pull/49

## [v0.0.9](https://github.com/Songmu/gitrail/compare/v0.0.8...v0.0.9) - 2026-09-18

- Bump reviewdog/action-misspell from 1.27.0 to 1.29.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/43
- Bump github.com/itchyny/gojq from 0.12.18 to 0.12.19 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/44
- Bump codecov/codecov-action from 7.0.0 to 7.1.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/42
- Bump reviewdog/action-staticcheck from 1.29.0 to 1.32.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/41
- Bump reviewdog/action-actionlint from 1.72.0 to 1.74.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/40
- Bump actions/setup-go from 6.5.0 to 7.0.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/39
- Add composite action for installing gitrail by @Songmu in https://github.com/Songmu/gitrail/pull/46

## [v0.0.8](https://github.com/Songmu/gitrail/compare/v0.0.7...v0.0.8) - 2026-09-17

- instruction installation in SKILL.md by @Songmu in https://github.com/Songmu/gitrail/pull/37

## [v0.0.7](https://github.com/Songmu/gitrail/compare/v0.0.6...v0.0.7) - 2026-09-16

- Bump reviewdog/action-staticcheck from 1.28.0 to 1.29.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/33
- Bump actions/setup-go from 6.4.0 to 6.5.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/32
- Bump Songmu/tagpr from 1.18.2 to 1.20.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/28
- Bump actions/checkout from 6.0.2 to 7.0.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/31
- Bump codecov/codecov-action from 6.0.0 to 7.0.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/30
- Improve gitrail skill metadata and schema packaging by @Songmu in https://github.com/Songmu/gitrail/pull/35
- Add gojq-backed jq output filtering by @Songmu with @Copilot in https://github.com/Songmu/gitrail/pull/36

## [v0.0.6](https://github.com/Songmu/gitrail/compare/v0.0.5...v0.0.6) - 2026-04-15
- udpate ghr by @Songmu in https://github.com/Songmu/gitrail/pull/23

## [v0.0.5](https://github.com/Songmu/gitrail/compare/v0.0.4...v0.0.5) - 2026-04-15
- fix ghr by @Songmu in https://github.com/Songmu/gitrail/pull/21

## [v0.0.4](https://github.com/Songmu/gitrail/compare/v0.0.3...v0.0.4) - 2026-04-15
- Bump Songmu/tagpr from 1.17.1 to 1.18.1 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/18
- Bump actions/setup-go from 6.3.0 to 6.4.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/17
- introduce ghr actions by @Songmu in https://github.com/Songmu/gitrail/pull/20
- Bump github.com/Songmu/skillsmith from 0.0.1 to 0.1.0 by @dependabot[bot] in https://github.com/Songmu/gitrail/pull/15

## [v0.0.3](https://github.com/Songmu/gitrail/compare/v0.0.2...v0.0.3) - 2026-03-18
- Add Renamed status for pure renames by @Songmu in https://github.com/Songmu/gitrail/pull/10

## [v0.0.2](https://github.com/Songmu/gitrail/compare/v0.0.1...v0.0.2) - 2026-03-15
- Add Agent Skills and skillsmith integration by @Copilot in https://github.com/Songmu/gitrail/pull/8

## [v0.0.1](https://github.com/Songmu/gitrail/commits/v0.0.1) - 2026-03-15
- Implement gitrail: Git history file change tracker CLI and library by @Copilot in https://github.com/Songmu/gitrail/pull/3
- Rename StartCommit/EndCommit to From/To by @Songmu in https://github.com/Songmu/gitrail/pull/4
- Refactor trail_test.go: consolidate into table-driven tests by @Songmu in https://github.com/Songmu/gitrail/pull/5
- Update README.md and add JSON Schema for NDJSON output by @Songmu in https://github.com/Songmu/gitrail/pull/6
