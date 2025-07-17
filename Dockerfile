# You'll want to use this as a base but add your language servers
# Also a volume with your project in it.
FROM alpine:latest
COPY mcp-lsp-bridge /usr/bin/mcp-lsp-bridge

CMD ["mcp-lsp-bridge"]
