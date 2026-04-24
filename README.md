# Generate-typst

An interactive CLI tool written in Go for building Typst documents into production-ready PDFs. It guides you through selecting languages, build options, and cover images, then compiles and assembles the final PDF using `typst` and `qpdf`.

## Requirements

- [Go](https://go.dev/) 1.18+
- [`typst`](https://typst.app/) — must be on your `PATH`
- [`qpdf`](https://qpdf.sourceforge.io/) — must be on your `PATH`

## Usage

```sh
go run .
```

The tool walks you through each step interactively:

1. **Working directory** — path to your Typst project (defaults to the current directory)
2. **Project scan** — detects root `.typ` files and per-language folders automatically
3. **Build target** — choose between building one or more languages, a multilingual booklet, or changing the working directory
4. **Build options** — column count, media type, audience, production mode, cover image, and optional extra Typst arguments

### Expected project structure

```
<project-root>/
├── <lang>/                          # e.g. en/, de/, fr/
│   ├── *.typ                        # main document(s)
│   ├── snippets-vars/
│   │   └── document-info-vars.typ   # document name / full product name
│   └── sharedResources/
│       └── pdf-cover/
│           ├── digital-front-cover.typ
│           ├── digital-back-cover.typ
│           ├── printed-front-cover.typ
│           └── printed-back-cover.typ
├── *.typ                            # root-level file(s) for multilingual booklets
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
| Production | yes/no | Encrypts the output PDF if `TYPST_PDF_BUILD_PASSWORD` env var is set |
| Cover image | file path | Passed to the front-cover template as `--input cover-image=` |
| Extra Typst args | any flags | Appended to every `typst compile` call (covers + main document) |

### Main menu

The main menu is shown after every build and offers:

- **Build one or more languages** — compile selected language folders individually
- **Build multilingual booklet** — compile a root-level `.typ` file with covers from the `en/` folder
- **Change working directory** — switch to a different Typst project without restarting
- **Exit**

The screen is cleared after a menu choice and after confirming "Build another document?" to keep the output readable.

### Extra Typst arguments

At the end of the build-options steps, you can supply arbitrary flags that will be forwarded to all `typst compile` invocations in that build. Arguments are shell-split, so quoted values with spaces work as expected:

```
Extra args passed to all typst compile calls (Enter to skip) []: --font-path /my/fonts --input debug=true
```

### Production mode & encryption

When production mode is enabled, the tool reads the `TYPST_PDF_BUILD_PASSWORD` environment variable and passes it to `qpdf` to encrypt the output with AES-256. If the variable is not set, an unprotected PDF is produced with a warning.

```sh
export TYPST_PDF_BUILD_PASSWORD="your-secret"
go run .
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

## Source layout

| File | Responsibility |
|---|---|
| `main.go` | Entry point, dependency check, directory prompt, main loop |
| `ui.go` | Terminal colors, log helpers, `prompt`, `yesno`, `pickOne`, `clearScreen` |
| `state.go` | Global state shared across all steps |
| `scanner.go` | Project scan and scan-results display |
| `menu.go` | Main menu, language selection, `.typ` file picker |
| `params.go` | Build-option prompts and shell-split helper |
| `build.go` | `typst` and `qpdf` invocations, PDF assembly and cleanup |
