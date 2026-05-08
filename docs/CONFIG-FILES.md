# Configuration Files
The `oasdiff` command can read its configuration from a file.  
This is useful for complex configurations or repeated usage patterns.

## Default lookup

By default, `oasdiff` looks for a config file in the directory where the command is run, in this order:

1. `.oasdiff.{json,yaml,yml,toml,hcl}` — preferred (dotfile convention)
2. `oasdiff.{json,yaml,yml,toml,hcl}` — legacy fallback, kept for back-compat with existing setups

`.oasdiff.*` is recommended for new setups. Existing `oasdiff.*` users don't need to migrate — both filenames continue to work.

For example, see [.oasdiff.yaml](../examples/.oasdiff.yaml).

## Explicit override

To use a different filename or path, pass `--config <path>` or set the `OASDIFF_CONFIG` environment variable. When either is set, the default lookup is skipped and the file at the given path must exist (missing or malformed file is an error).

Precedence: `--config <path>` > `OASDIFF_CONFIG` > default `.oasdiff.*` lookup.

```sh
# Explicit flag (per-invocation)
oasdiff diff --config ./my-config.yaml base.yaml revision.yaml

# Environment variable (set once for a shell or CI workflow)
export OASDIFF_CONFIG=./my-config.yaml
oasdiff diff base.yaml revision.yaml
```

The configuration file supports the exact same flags that are supported by the command-line.
Notes:
1. Command-line flags take precedence over configuration file settings.
2. **Boolean flags**: to set a boolean flag to `false` on the command line, use `=` syntax: `--flag=false`.
   Without `=`, cobra treats the value as a positional argument: `--flag false` ≠ `--flag=false`.
   In a configuration file, use standard YAML/JSON boolean syntax: `flag: false`.
   Example: `--allow-external-refs=false` (CLI) vs `allow-external-refs: false` (config file).
3. Some of the flags define paths to additional configuration files:
    - `err-ignore`:              configuration file for ignoring errors
    - `severity-levels`:         configuration file for custom severity levels
    - `warn-ignore`:             configuration file for ignoring warnings
    - `template`:                custom Go template file for changelog generation

   **Relative paths in these flags are resolved against the config file's directory**, not the process's current working directory. So when you write `err-ignore: rules.txt` in `path/to/.oasdiff.yaml`, oasdiff reads `path/to/rules.txt`. Absolute paths and paths set via CLI flag are not rewritten.
