# MCP LSP Bridge

## Logging Configuration

The MCP LSP Bridge provides a flexible logging system with the following features:

### Configuration Options

The logger can be configured through the `lsp_config.json` file under the `global` section:

```json
{
    "global": {
        "log_file_path": "/var/log/mcp-lsp-bridge/bridge.log",
        "log_level": "info",
        "max_log_files": 5
    }
}
```

#### Configuration Parameters

- `log_file_path`: Path to the log file. Defaults to a temporary file if not specified.
- `log_level`: Logging verbosity. Options:
  - `info`: Default, logs informational messages
  - `debug`: Logs debug messages in addition to info and error
  - `error`: Logs only error messages
- `max_log_files`: Maximum number of log files to keep before rotation. Default is 5.

### Programmatic Configuration

You can also configure logging programmatically:

```go
logConfig := logger.LoggerConfig{
    LogPath:     "/path/to/logfile.log",
    LogLevel:    "debug",
    MaxLogFiles: 3,
}
logger.InitLogger(logConfig)
```

### Logging Methods

- `logger.Info()`: Log informational messages
- `logger.Debug()`: Log debug messages (only when log level is "debug")
- `logger.Error()`: Log error messages

## Running Tests

```bash
go test ./logger
```