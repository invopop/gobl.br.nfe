# GOBL ➡️ Brazil NF-e / NFC-e

Brazil NF-e / NFC-e addon for [GOBL](https://github.com/invopop/gobl).

Released under the Apache 2.0 [LICENSE](https://github.com/invopop/gobl.br.nfe/blob/main/LICENSE), Copyright 2026 [Invopop S.L.](https://invopop.com).

[![Lint](https://github.com/invopop/gobl.br.nfe/actions/workflows/lint.yaml/badge.svg)](https://github.com/invopop/gobl.br.nfe/actions/workflows/lint.yaml)
[![Test Go](https://github.com/invopop/gobl.br.nfe/actions/workflows/test.yaml/badge.svg)](https://github.com/invopop/gobl.br.nfe/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/invopop/gobl.br.nfe)](https://goreportcard.com/report/github.com/invopop/gobl.br.nfe)
[![codecov](https://codecov.io/gh/invopop/gobl.br.nfe/graph/badge.svg)](https://codecov.io/gh/invopop/gobl.br.nfe)
[![GoDoc](https://godoc.org/github.com/invopop/gobl.br.nfe?status.svg)](https://godoc.org/github.com/invopop/gobl.br.nfe)
![Latest Tag](https://img.shields.io/github/v/tag/invopop/gobl.br.nfe)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/invopop/gobl.br.nfe)

This module implements Brazil's **NF-e / NFC-e** electronic invoicing documents
as a GOBL tax addon:

- **NF-e / NFC-e** (`br-nfe-v4`) — Brazil's NF-e 4.00 layout, covering both NF-e
  (goods) and NFC-e (consumer) electronic invoices.

Brazil's service invoices (NFS-e) live in a separate module,
[`gobl.br.nfse`](https://github.com/invopop/gobl.br.nfse).

Unlike the format converters in the GOBL ecosystem, this is a true **addon**: it
registers extensions, normalizers, and validation rules into GOBL's global
registry. It lives in its own module so that only projects handling Brazil NF-e
documents take on its weight.

## Layout

- `addon/` — the GOBL addon: extensions, normalizers, scenarios, and validation
  rules that register into GOBL on import. This package is kept dependency-light
  so importing it never pulls in conversion tooling.
- the module root (and future subpackages) is reserved for converters and other
  Brazil NF-e tooling that build on the addon.

## Usage

Add a blank import of the **addon** so it registers itself, then use GOBL as
normal:

```go
import (
	"github.com/invopop/gobl"
	_ "github.com/invopop/gobl.br.nfe/addon"
)
```

Declare the addon on a document (or let the regime/scenario add it) and
`Calculate` + `Validate` will run the full Brazil NF-e normalization and rules.

> **Note**: the `br-nfe-v4` key is listed in GOBL core's approved external-addon
> registry, so it is recognised as a valid `$addons` value in the JSON Schema.
> The runtime check stays strict, however: a document declaring this addon will
> fail validation with `add-on must be registered` unless this module is
> imported. Any service that processes Brazil NF-e documents must import it.

## Development

The addon builds on core GOBL features (the `regimes/br` regime and the approved
external-addon registry) that are not yet in a tagged release. The `go.mod`
therefore pins `github.com/invopop/gobl` to a commit on the core
`extract-br-addons` branch (a pseudo-version); bump it to the release tag once
core is published.

```sh
go test ./...
```

### Examples

`examples/` holds sample documents with their expected JSON envelopes under
`examples/out/`. They are verified via GOBL's shared `pkg/examples` helpers.
Regenerate the golden output after intentional changes with:

```sh
go test . -run TestExamples -update
```

## License

Apache 2.0 — see [LICENSE](./LICENSE).
