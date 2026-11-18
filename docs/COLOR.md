# Color, Pipes and Redirects
When outputting changes to a Unix terminal, oasdiff automatically adds colors with ANSI color escape sequences.  
If output is piped into another process or redirected to a file, oasdiff disables color.  
To control color manually, use the `--color` flag with `always` or `never`.
