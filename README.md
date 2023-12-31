# atlas-deploy-action

## Deprecation Notice

This action has been deprecated in favor of the `ariga/atlas-action/migrate/apply` action.

Please see [the docs](https://github.com/ariga/atlas-action#arigaatlas-actionmigrateapply) for more
details.

## About

A GitHub Action to deploy versioned migrations with [Atlas](https://atlasgo.io).

## Supported Workflows

- Local - the migration directory is checked in to the repository.
- Cloud - the migration directory is [connected to Atlas Cloud](https://atlasgo.io/cloud/directories).
  Runs are reported to your Atlas Cloud account.

## Examples 

### Local Workflow

```yaml
name: Deploy Database Migrations
on:
  push:
    branches:
      - master
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy Atlas Migrations
        uses: ariga/atlas-deploy-action@v0
        with:
          url: ${{ secrets.DATABASE_URL }}
          dir: path/to/migrations
```

### Cloud Workflow

```yaml
name: Deploy Database Migrations
on:
  push:
    branches:
      - master
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy Atlas Migrations
        uses: ariga/atlas-deploy-action@v0
        with:
          url: ${{ secrets.DATABASE_URL }}
          cloud-token: ${{ secrets.ATLAS_CLOUD_TOKEN }}
          cloud-dir: hello # replace with your directory name
```

## Reference

### Inputs

- `url`: URL to target database (should be passed as a secret). (Required)
- `dir`: Local path of the migration directory in the repository. (Optional)
- `cloud-token`: Token for using Atlas Cloud (should be passed as a secret). (Optional)
- `cloud-dir`: Name of the migration directory in the cloud. (Must be set if `cloud-token` is set)
- `cloud-tag`: Tag of the migration version in the cloud. (Optional)

Note: Either `dir` or `cloud-dir` must be set. If both are provided, an error will be thrown.

### Outputs

- `error`: Error message if any.
- `current`: Current migration version.
- `target`: Target migration version.
- `pending_count`: Number of pending migrations.
- `applied_count`: Number of applied migrations.

## License

This project is licensed under the [Apache License, Version 2.0](LICENSE).