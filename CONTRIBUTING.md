# Contributing

## Repository Settings

Repository settings are managed with
[gh-infra](https://github.com/babarot/gh-infra) using
[`.github/infra.yaml`](.github/infra.yaml). Install the extension, review the
plan, and apply changes using the maintainer's local GitHub credentials:

```console
% gh extension install babarot/gh-infra --pin v0.13.0
% gh infra plan .github/infra.yaml
% gh infra apply .github/infra.yaml
```

Repository administration credentials are not stored in GitHub Actions.
