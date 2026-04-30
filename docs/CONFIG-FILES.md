# Configuration Files
The `oasdiff` command can read its configuration from a file.  
This is useful for complex configurations or repeated usage patterns.  
The config file should be named oasdiff.{json,yaml,yml,toml,hcl} and placed in the directory where the command is run.  
For example, see [oasdiff.yaml](../examples/oasdiff.yaml).

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
