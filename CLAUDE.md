# mail-server-tester

Go CLI that continuously sends/receives test mails between a "testserver" and N external servers, exposes results as Prometheus metrics on `:8080/metrics`. Single binary, deployable via Containerfile + Helm chart.

## Layout

- `main.go` — entry: arg parsing (`--configpath`, `--exit-on-connection-error`, `--help`), startup connection verify, metrics HTTP server, infinite test loop (`cfg.IntervalMinutes`, default 60).
- `config.go` — YAML config (`Config`, `ServerConfig`). Fields: smtp/imap host+port+user+pw, `mail_address`, `tls`, `skip_cert_verify`.
- `test.go` — `RunMailTests`: bidirectional send→sleep 10s→IMAP fetch latest. Results in global `MailTestResults` (reset each loop).
- `smtp.go` — `SendTestMail`. TLS path uses `tls.Dial` + `smtp.NewClient`; non-TLS uses `smtp.SendMail`. `mail_address` overrides `smtp_user` as From.
- `imap.go` — `FetchLatestMail`. Selects INBOX, fetches last message by seq.
- `connection_verify.go` — `VerifyAllConnections` at startup. `VerifySMTPConnection` / `VerifyIMAPConnection`. German user-facing output.
- `main.go:metricsHandler` — emits `mail_test_success`, `mail_test_duration_seconds`, `mail_test_error_total` with `from`/`to` labels.
- Tests: `config_test.go`, `connection_test.go`, `connection_verify_test.go`, `main_test.go`.
- `config.yaml` — example/default. `local_config.yml` — local dev (gitignored? check).
- `Containerfile` — multi-stage Go 1.24 alpine build, binary in `/app`, exposes 8080.
- `deploy/helm/mail-server-tester/` — Helm chart (deployment, configmap, service, servicemonitor, prometheusrules, serviceaccount).
- `.github/workflows/{build,release}.yml` — CI + semantic-release.
- `.releaserc.json` + `package.json` — semantic-release with conventional commits. Commit prefixes: `feat:`, `fix:`, `chore:`, etc.

## Build / Run

```bash
go mod download
go build -o mail-server-tester .
./mail-server-tester --configpath config.yaml
```

Container: `docker build -t mail-server-tester -f Containerfile .`

## Wichtig

- Sprache: User-facing logs/errors auf Deutsch (siehe `connection_verify.go`, `main.go`). Code-Kommentare gemischt. Beibehalten.
- Recipient-resolution: `MailAddress` first, fallback `IMAPUser` (test.go) bzw. `SMTPUser` (smtp.go für From).
- TLS-Pfad und Plain-Pfad sind getrennt in `smtp.go` und `connection_verify.go` — Änderungen in beiden Zweigen nötig.
- `MailTestResults` ist globaler Slice, jeden Lauf in `main.go` zurückgesetzt. Tests laufen parallel pro Cycle (`sync.WaitGroup` in `test.go`); Writes werden über `resultsMu` (`sync.Mutex`) serialisiert. Metrics-Handler liest ohne Lock — akzeptiert, weil Reset+Cycle-Run sequenziell zwischen Cycles ablaufen.
- `RunMailTests` macht zu Beginn jedes Cycles `CleanupOldTestMails` pro Postfach (Bulk-Cleanup per Subject-Prefix für Altlasten). Während der parallelen Tests löscht `WaitAndCleanTestMail` nur die exakt gematchte Mail — sonst würden parallele Tests gegen dieselbe Inbox (mehrere Externals → Testserver) sich gegenseitig Mails wegputzen.
- Zustellzeit-Konfiguration: `delivery_timeout_minutes` (Default 30) + `delivery_poll_seconds` (Default 5). Poll-Loop in `imap.go:WaitAndCleanTestMail` sucht periodisch per exaktem Subject.
- Releases: semantic-release auf `main` aus Conventional Commits. Chart-Version in `deploy/helm/.../Chart.yaml` (assets-Pfad im `.releaserc.json` zeigt allerdings auf `charts/*/Chart.yaml` — möglicher Mismatch, prüfen falls Release-Asset-Update fehlschlägt).
- Go 1.24, IMAP via `github.com/emersion/go-imap` v1.2.1, SMTP via stdlib `net/smtp`.

## Konventionen für künftige Änderungen

- Commits: Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`).
- Pre-built Binary `mail-server-tester` liegt im Repo-Root — nicht versehentlich überschreiben/committen ohne Grund.
- `node_modules/` ist nur für semantic-release toolchain.
