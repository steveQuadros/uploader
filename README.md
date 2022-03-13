# Sample Multi Cloud Uploader

## Quick Start
see `example_config.json` for config file format

run `make build`

run `./uploader` to see usage

Additional `make` commands:
- `make run`
- `make test` tests with `-race -cover`
- `make buildvalid` tests, builds, and runs with valid input (assumes valid config file located at `~/.filescom/config.json`) 

## Targets
Targets aws, gcp, and azure, and any specific one can be targeted.

ex:

Just Azure
```
./uploader --provider azure  --file test.txt --config ~/.filescom/config.json -bucket filescomquad -key key.txt
```

All Three

```
./uploader --provider aws --provider azure --provider gcp --file test.txt --config ~/.filescom/config.json -bucket filescomquad -key test.txt
```

## Details
- Buckets are created if they do not exist
- Files are uploaded concurrently to providers
  - There is in issue in Azure which closes the passed in `ReadSeekCloser`, which had to worked around by copying the file. A comment was added to fix this when possible.

## Enhancements
- Additional unit Testing
- Integration Testing using [min.io](https://min.io))
- Supporting multiple credential types
- Supporting env vars in addition to config file