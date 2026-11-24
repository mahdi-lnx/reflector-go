
# reflector-go

A concurrent, configurable command-line utility written in Go for discovering, ranking, and generating Arch Linux mirrorlists.

Inspired by the original [reflector](https://wiki.archlinux.org/title/Reflector), `reflector-go` retrieves mirror metadata from the 
official Arch Linux mirror status API, filters mirrors according to user-defined criteria, optionally measures their download performance 
directly from the user's machine, and ranks them based on the selected sorting criteria.

It produces pacman compatible mirrorlists or JSON output for scripting and automation.

> [!NOTE]
> This project is under active development.
> Core functionality is implemented and functional, however some automated test coverage is currently not finished and is being expanded.

## Features

- **Concurrent benchmarking** — measures mirror download rates using a configurable worker pool and HTTP timeouts.
- **Flexible filtering** — selects mirrors by country name or country code, protocol, synchronization age, completion percentage, and active status.
- **Multiple ranking strategies** — ranks mirrors by measured download rate, official mirror score, or most recent synchronization time.
- **Atomic file output** — saves the generated mirrorlist to a user-specified file using `--save`, safely replacing existing files without exposing partially written output.
- **Script-friendly output** — generates Pacman configuration, compact JSON, or human-readable JSON.
- **Configuration validation** — rejects invalid arguments and incompatible options.
- **Structured logging** — optional diagnostic logs are sent to stderr, keeping stdout suitable for pipelines and redirection.

## Installation

Requires Go 1.27.2 or later

```sh
go install github.com/mahdi-lnx/reflector-go@latest
```

Ensure that your Go binary directory (`$HOME/go/bin` by default) is included in your `PATH`.

## Usage

Generate a mirrorlist using the default configuration:

```sh
reflector-go
```

Select the 10 fastest active HTTPS mirrors in Canada and the United States:

```sh
reflector-go --country CA,US --number 10
```

Select recently synchronized mirrors without running download benchmarks:

```sh
reflector-go --age 12h --no-benchmark --sort age --number 10
```

Export mirror metadata as JSON, ranked by the official mirror score:

```sh
reflector-go --no-benchmark --sort score --format pretty-json
```

Benchmark mirrors using fewer concurrent workers and a longer request timeout:

```sh
reflector-go --threads 8 --request-timeout 15s --verbose
```

Save the generated mirrorlist to a file:

```sh
reflector-go --country CA,US --number 10 --save mirrorlist
```

Save the generated mirrorlist directly to the Pacman configuration directory:

```sh
sudo "$(command -v reflector-go)" --country CA,US --number 10 --save /etc/pacman.d/mirrorlist
```

> [!IMPORTANT]
> The last command replaces the existing Pacman mirrorlist.
> Review the generated output and back up your existing mirrorlist
> before replacing it.

Root privileges are not required for ordinary benchmarking or writing
to stdout.

## Command-line options

| Option | Default | Description |
| --- | --- | --- |
| `--country` | `all` | Filter by country names or country codes. Use `all` to disable country filtering |
| `--protocol` | `https` | Filter by protocol (`http` or `https`). Use `all` to include both |
| `--age` | `0s` | Maximum time since the mirror last synchronized. Zero disables age filtering |
| `--min-completion` | `100` | Minimum completion percentage, between 0 and 100 (inclusive) |
| `--active-only` | `true` | Include only active mirrors. Use `--active-only=false` to disable |
| `--sort` | `download-rate` | Sort by `download-rate`, `score`, or `age` |
| `--number` | `0` | Maximum number of output mirrors. Zero means unlimited |
| `--threads` | `47` | Maximum number of concurrent benchmark workers |
| `--request-timeout` | `7s` | HTTP client timeout per request |
| `--no-benchmark` | `false` | Skip download benchmarking. Requires sorting by `score` or `age` |
| `--format` | `pacman` | Output format: `pacman`, `json`, or `pretty-json` |
| `--save` | stdout | Atomically save the generated output to the specified file |
| `--verbose` | `false` | Enable informational and warning logs |
| `--help` | — | Display available command-line options |

Values supplied to `--country`, `--protocol`, `--sort`, and `--format`
are case-insensitive. For example, `HTTPS` and `https`, as well as
`ALL` and `all`, are treated identically.

### A note on `--threads`

Despite the name of the `--threads` option, the benchmarking workers
are not operating system threads but goroutines, which are lightweight, higher-level concurrency primitives managed
by the Go runtime.

Explaining what goroutines are and how they work is beyond the scope
of this README. However, it is perfectly practical to have hundreds
of goroutines running concurrently, even on a single-core machine,
especially for I/O-bound workloads such as network requests.

### Multi-value flags

The `--country` and `--protocol` options accept comma-separated values
or repeated occurrences. Explicit values replace the defaults, and
subsequent occurrences accumulate values.

The keyword `all` clears the accumulated restriction. Any values
specified after `all` establish a new restriction.

For example, the following two commands are equivalent:

```sh
reflector-go --country ca,us
reflector-go --country ca --country us
```

The following command includes mirrors using either HTTP or HTTPS:

```sh
reflector-go --protocol all
```

## Testing

The existing tests pass and cover:

- CLI configuration parsing and validation.
- Multi-value flag parsing.
- Atomic file creation, writing, replacement, and cleanup.

Automated test coverage for the `mirrorlist` package and the main package is planned.

## Acknowledgments

This project is inspired by the original [reflector](https://wiki.archlinux.org/title/Reflector) utility.

Mirror metadata is provided by the official [Arch Linux mirror status API](https://archlinux.org/mirrors/status/json/).

This is an independent project and is not an official Arch Linux utility.
