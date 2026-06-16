# MCP server

oasdiff is available as a hosted [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server, so AI assistants like Claude and Cursor can run it for you. Ask "does this change break the API?" and the assistant calls oasdiff on your specs and reads back the result.

It is a remote server over the Streamable HTTP transport, hosted at `https://api.oasdiff.com/mcp`. No install, no API key.

## Tools

- detect breaking changes between two OpenAPI specs
- generate a changelog
- diff two specs
- validate a spec

Specs are passed as text by the assistant; external `$ref` URLs are not resolved, so a spec must be self-contained.

## Connect

Claude Code:

```
claude mcp add --transport http oasdiff https://api.oasdiff.com/mcp
```

Cursor, in `~/.cursor/mcp.json` (global) or `.cursor/mcp.json` (per project):

```json
{
  "mcpServers": {
    "oasdiff": {
      "url": "https://api.oasdiff.com/mcp"
    }
  }
}
```

Any MCP client that speaks Streamable HTTP works: point it at `https://api.oasdiff.com/mcp`.

Full guide: [oasdiff.com/docs/mcp](https://www.oasdiff.com/docs/mcp)
