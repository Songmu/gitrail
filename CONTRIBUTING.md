# Contributing

## Repository Settings

Repository settings are managed with
[gh-infra](https://github.com/babarot/gh-infra) using
[`.github/infra.yaml`](.github/infra.yaml). Install the extension to review
changes locally:

```console
% gh extension install babarot/gh-infra --pin v0.13.0
% make infra-plan
```

Changes merged to `main` are applied by
[`.github/workflows/gh-infra.yaml`](.github/workflows/gh-infra.yaml) using the
repository's `GITHUB_TOKEN`. No additional repository administration
credentials are stored.
