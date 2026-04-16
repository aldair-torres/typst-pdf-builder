# Generate-typst

An interactive CLI tool written in Go for building Typst documents into production-ready PDFs. It guides you through selecting languages, build options, and cover images, then compiles and assembles the final PDF using `typst` and `qpdf`.

## Requirements

- [Go](https://go.dev/) 1.18+
- [`typst`](https://typst.app/) — must be on your `PATH`
- [`qpdf`](https://qpdf.sourceforge.io/) — must be on your `PATH`

## Usage

```sh
go run typst-pdf-builder.go
```

The tool walks you through each step interactively:

1. **Working directory** — path to your Typst project (defaults to the current directory)
2. **Extraction rules** — loads `.typst-rules.json` if present (see below)
3. **Project scan** — detects root `.typ` files and per-language folders automatically
4. **Build target** — choose between building one or more languages or a multilingual booklet
5. **Build options** — column count, media type, audience, production mode, and an optional cover image

### Expected project structure

```
<project-root>/
├── <lang>/                          # e.g. en/, de/, fr/
│   ├── *.typ                        # main document(s)
│   ├── snippets-vars/
│   │   └── document-info-vars.typ   # document name / short product name
│   └── sharedResources/
│       └── pdf-cover/
│           ├── digital-front-cover.typ
│           ├── digital-back-cover.typ
│           ├── printed-front-cover.typ
│           └── printed-back-cover.typ
├── *.typ                            # root-level file(s) for multilingual booklets
├── .typst-rules.json                # optional — custom extraction rules
└── .resources/
    ├── a4-empty.pdf                 # required for multilingual digital builds
    └── a5-empty.pdf                 # required for multilingual printed builds
```

### Build options

| Option | Values | Description |
|---|---|---|
| Columns | `1`, `2` | Passed to Typst as `--input columns=` (single-language builds only) |
| Media | `digital`, `printed` | Controls which cover templates are used and the empty-page size |
| Audience | any string | Passed to Typst as `--input audience=` (optional) |
| Production | yes/no | Encrypts the output PDF if `BUILD_PASSWORD` env var is set |
| Cover image | file path | Passed to the front-cover template as `--input cover-image=` |

### Extraction rules (`.typst-rules.json`)

The tool reads document metadata (short product name, document name) from a variables file in each language folder. By default it uses built-in generic patterns. If your project has a `.typst-rules.json` in the same directory as the script, the tool will offer to use it instead.

The file defines named rules, each with a regex `pattern` and a `matchMode` of `"first"` or `"all"`:

```json
{
  "shortName": {
    "pattern": "short-product-name\\s*=\\s*\"([^\"]*)\"",
    "matchMode": "all"
  },
  "docName": {
    "pattern": "document-name\\s*=\\s*\\[([^\\]]*)\\]",
    "matchMode": "all"
  },
  "productArg": {
    "pattern": "(?i)(foo|bar)",
    "matchMode": "first"
  }
}
```

`productArg` is optional. When present, it enables a product code lookup where the short name is extracted from a keyed block matching the product argument rather than a flat pattern search.

### Production mode & encryption

When production mode is enabled, the tool reads the `BUILD_PASSWORD` environment variable and passes it to `qpdf` to encrypt the output with AES-256. If the variable is not set, an unprotected PDF is produced with a warning.

```sh
export BUILD_PASSWORD="your-secret"
go run typst-pdf-builder.go
```

### Output

Single-language builds produce:

```
<basename>_<lang>-<DD-MM-YYYY>.pdf
```

Multilingual builds produce:

```
<basename>_all-<DD-MM-YYYY>.pdf
```

Intermediate PDFs (covers, raw compiled output) are removed automatically after assembly.
