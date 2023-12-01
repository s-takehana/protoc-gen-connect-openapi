# protoc-gen-connect-openapi

`protoc-gen-connect-openapi` generates OpenAPI YAML files for [Connect](https://connectrpc.com/docs/introduction) from Protocol Buffers definition.

:warning: `protoc-gen-connect-openapi` supports OpenAPI 3.0.

## Usage

0. Install and configure [`buf`](https://buf.build/docs/installation)

1. Install `protoc-gen-connect-openapi`

```shell
go install github.com/s-takehana/protoc-gen-connect-openapi@latest
```

2. Create a OpenAPI template

```yaml
info:
  title: Example API
  version: 0.1.0
```

3. Configure a `buf.gen.yaml` file

```yaml
version: v1
plugins:
  - name: connect-openapi
    out: .
    opt:
      - template=path/to/protoc-gen-connect-openapi_template.yaml
```

4. Execute `buf`

```shell
buf generate
```
