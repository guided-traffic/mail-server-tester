# Mail Server Tester

A continuous mail server testing tool that monitors mail delivery between multiple mail servers and exposes metrics for Prometheus monitoring.

## Features

- Continuous bidirectional mail testing between servers
- Prometheus metrics endpoint for monitoring
- Support for TLS/SSL connections
- Configurable test intervals
- Support for multiple external mail servers
- Docker/Podman containerization support

## How it works

The tool performs bidirectional mail tests:
1. **Test Server → External Servers**: Sends test emails from your main server to external servers
2. **External Servers → Test Server**: Sends test emails from external servers back to your main server

Each test measures:
- Mail delivery success/failure
- Delivery duration
- Error details (if any)

## Installation

### Building from source

```bash
go mod download
go build -o mail-server-tester .
```

### Using Docker

```bash
docker build -t mail-server-tester -f Containerfile .
docker run -v $(pwd)/config.yaml:/root/config.yaml mail-server-tester
```

### Using Podman

```bash
podman build -t mail-server-tester .
podman run -v $(pwd)/config.yaml:/root/config.yaml mail-server-tester
```

## Configuration

Create a `config.yaml` file with your mail server configurations:

```yaml
interval_minutes: 60  # Test interval in minutes

testserver:
  name: testserver
  smtp_server: smtp.main.example.com
  smtp_port: 587
  smtp_user: user@main.example.com
  smtp_password: password
  imap_server: imap.main.example.com
  imap_port: 993
  imap_user: user@main.example.com
  imap_password: password
  tls: true
  skip_cert_verify: false

external_servers:
  - name: external1
    smtp_server: smtp.ext1.example.com
    smtp_port: 587
    smtp_user: user@ext1.example.com
    smtp_password: password1
    imap_server: imap.ext1.example.com
    imap_port: 993
    imap_user: user@ext1.example.com
    imap_password: password1
    tls: true
    skip_cert_verify: true

  - name: external2
    smtp_server: smtp.ext2.example.com
    smtp_port: 587
    smtp_user: user@ext2.example.com
    smtp_password: password2
    imap_server: imap.ext2.example.com
    imap_port: 993
    imap_user: user@ext2.example.com
    imap_password: password2
    tls: false
    skip_cert_verify: false
```

### Configuration Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `interval_minutes` | Minutes between test cycles | 60 |
| `name` | Server identifier for metrics | - |
| `smtp_server` | SMTP server hostname | - |
| `smtp_port` | SMTP server port | - |
| `smtp_user` | SMTP username | - |
| `smtp_password` | SMTP password | - |
| `imap_server` | IMAP server hostname | - |
| `imap_port` | IMAP server port | - |
| `imap_user` | IMAP username | - |
| `imap_password` | IMAP password | - |
| `tls` | Enable TLS/SSL | false |
| `skip_cert_verify` | Skip certificate verification | false |

## Usage

### Command line options

```bash
# Use default config.yaml
./mail-server-tester

# Use custom config file
./mail-server-tester --configpath /path/to/config.yaml
```

### Running with custom config

```bash
./mail-server-tester --configpath /path/to/your/config.yaml
```

## Monitoring

### Prometheus Metrics

The application exposes Prometheus metrics on port 8080 at the `/metrics` endpoint.

Available metrics:
- `mail_test_success{from="server1",to="server2"}` - Test success (1) or failure (0)
- `mail_test_duration_seconds{from="server1",to="server2"}` - Test duration in seconds
- `mail_test_error_total{from="server1",to="server2"}` - Error count per test

### Example Prometheus scrape config

```yaml
scrape_configs:
  - job_name: 'mail-server-tester'
    static_configs:
      - targets: ['localhost:8080']
```

## Dependencies

- Go 1.24+
- [github.com/emersion/go-imap](https://github.com/emersion/go-imap) - IMAP client
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML configuration parsing

## License

This project is open source. Please check the repository for license details.
