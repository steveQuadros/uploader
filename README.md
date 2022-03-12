# Sample Multi Cloud Uploader

## Quick Start
run `make build`

run `./uploader` to see usage

## Targets
Currently targets aws, gcp, and azure, and any specific one can be target.

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