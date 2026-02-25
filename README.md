# lingualeo

Lingualeo API console helper for translating words, optionally pronouncing them, visualizing pictures, and adding words to your dictionary.

[![CI](https://github.com/trezorg/lingualeo/actions/workflows/ci.yml/badge.svg)](https://github.com/trezorg/lingualeo/actions/workflows/ci.yml)
[![Release](https://github.com/trezorg/lingualeo/actions/workflows/release.yml/badge.svg)](https://github.com/trezorg/lingualeo/actions/workflows/release.yml)

## Install

### Option 1: install prebuilt binary

```bash
curl -sfL https://raw.githubusercontent.com/trezorg/lingualeo/main/install.sh | bash
```

Install into a custom directory:

```bash
curl -sfL https://raw.githubusercontent.com/trezorg/lingualeo/main/install.sh | bash -s -- -d /your/bin/dir
```

Install a specific version:

```bash
curl -sfL https://raw.githubusercontent.com/trezorg/lingualeo/main/install.sh | bash -s -- -v v1.2.3
```

### Option 2: build/install with Go

```bash
go install github.com/trezorg/lingualeo/cmd/lingualeo@latest
```

## Authentication and configuration

`lingualeo` requires your Lingualeo email and password.

You can provide credentials either:
- directly via CLI flags (`--email`, `--password`), or
- via a config file (`--config`) in TOML, YAML, or JSON.

Default config file names searched automatically:
- `~/lingualeo.toml`
- `~/lingualeo.yml`
- `~/lingualeo.yaml`
- `~/lingualeo.json`
- same filenames in the current working directory

### Example config (TOML)

```toml
email = "email@gmail.com"
password = "password"
add = false
sound = true
player = "mplayer"
download = false
visualize = false
reverse_translate = false
request_timeout = "30s"
log_level = "INFO"
```

### Example config (YAML)

```yaml
email: email@gmail.com
password: password
add: false
sound: true
player: mplayer
download: false
visualize: false
reverse_translate: false
request_timeout: 30s
log_level: INFO
```

## Common usage scenarios

Show help:

```bash
lingualeo --help
```

Translate one or more words:

```bash
lingualeo --email you@example.com --password 'secret' hello world
```

Use a config file:

```bash
lingualeo --config ./lingualeo.toml hello
```

Add a word with custom translation:

```bash
lingualeo add -t "custom translation" hello
```

Pronounce words using a player:

```bash
lingualeo --sound --player "mpv --no-video" hello
```

Download sound before playing (for players that cannot stream URLs):

```bash
lingualeo --sound --download --player "mplayer" hello
```

Enable reverse translation mode:

```bash
lingualeo --reverse-translate привет
```

## Development

Build:

```bash
make build
```

Generate mocks:

```bash
make generate
```

Lint:

```bash
make lint
```

Run tests:

```bash
make test
```

Full local check before PR:

```bash
make clean cache generate lint test
```
