# Helm Repository für Mail Server Tester

Dieses Repository enthält den Helm Chart für den Mail Server Tester.

## Installation

Um den Chart zu installieren, füge zuerst das Repository hinzu:

```bash
helm repo add mail-server-tester https://guided-traffic.github.io/mail-server-tester/deploy/helm/repo
helm repo update
```

Dann kannst du den Chart installieren:

```bash
helm install mail-server-tester mail-server-tester/mail-server-tester
```

## Konfiguration

Siehe die [values.yaml](deploy/helm/mail-server-tester/values.yaml) für die verfügbaren Konfigurationsoptionen.
