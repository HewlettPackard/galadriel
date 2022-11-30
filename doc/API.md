# Open API Specifications

## Common Definition between APIs 

- [Common Definitions](../pkg/common/api/schemas.yaml)


## Galadriel Server APIs

- [Harvester API](../pkg/server/api/harvester/harvester.yaml)
- [Management API](../pkg/server/api/management/management.yaml)


## Generating code boilerplate from specs

Tool: [oapi-codegen](https://github.com/deepmap/oapi-codegen)

Current Version: 1.12.4

1. Download openapi code generation tool:

    > `go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest`

2. Go to where the specs are (listed above), and generate the boilerplate

    > `oapi-codegen -config <config-file> <openapi_spec.yaml>`    

    Example: 
    
    2.1 While in the [management](../pkg/server/api/management/) directory, run the following command:
    - `oapi-codegen -config management.cfg.yaml management.yaml`

## Generating code boilerplate using script

This make target automates the process above.

1. While in the [root directory](/), run the following command:
    > `make generate-spec`