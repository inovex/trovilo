# trovilo

trovilo collects and prepares files from Kubernetes ConfigMaps for Prometheus & friends.

## Philosophy

This simple helper tool aims to collect ConfigMaps (files) via the Kubernetes API and writes them down into the filesystem, that may be internally processed by apps like Prometheus or Grafana. It focuses to serve this purpose only on a very generic way. This means it's not meant to work with a specific app only and won't contain such specific code. Instead we try to provide an extensive UI and try keep the code maintainable.

Contributions are highly appreciated. :)

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

Setup your GO environment if not already done:

```
$ export GOPATH=${HOME}/GOPATH GOBIN=${HOME}/GOPATH/bin
```

```
$ go get -u github.com/inovex/trovilo/cmd/trovilo
$ $GOBIN/trovilo --help
usage: trovilo --config=CONFIG [<flags>]

Trovilo collects and prepares files from Kubernetes ConfigMaps for Prometheus & friends

Flags:
  -h, --help                   Show context-sensitive help (also try --help-long and --help-man).
      --config=CONFIG          YAML configuration file.
      --kubeconfig=KUBECONFIG  Optional kubectl configuration file. If undefined we expect trovilo is running in a pod.
      --log-level="info"       Specify log level (debug, info, warn, error).
      --log-json               Enable JSON-formatted logging on STDOUT.
  -v, --version                Show application version.
```

## Deployment

Deploy the binary to your target systems or use the [official Docker image](https://hub.docker.com/r/inovex/trovilo/). Notice: The *tools*-tagged Docker image additionally contains useful tools for verify or post-deploy commands.

Simple trovilo example configuration file:

```
$ cat trovilo-config.yaml
# Which namespace to check (empty string means all namespaces)
#namespace: ""
jobs:
  # Arbitrary name for identification (and troubleshooting in logs)
  - name: alert-rules
    # Kubernetes-styled label selector to define how to find ConfigMaps
    selector:
      type: prometheus-alerts
    verify:
      # Example verification step to check whether the contents of the ConfigMap are valid Prometheus alert files. %s will be replaced by the ConfigMap's file path(s).
      - name: verify alert rule validity
        cmd: ["promtool", "check", "rules", "%s"]
    target-dir: /etc/prometheus-alerts/
    # Enable directory flattening so all ConfigMap files will be placed into a single directory
    flatten: true
    # After successfully verifying the ConfigMap and deploying it into the target-dir, run the following commands to trigger (e.g. Prometheus) manual config reloads
    post-deploy:
      - name: reload prometheus
        cmd: ["curl", "-s", "-X", "POST", "http://localhost:9090/-/reload"]
  # Another job example, but for Grafana dashboards (JSON model)
  - name: grafana-dashboards
    selector:
      type: grafana-dashboards
    target-dir: tmp/target-grafana-dashboards/
```

Full example Kubernetes deployment with Prometheus:

```
$ kubectl apply \
  -f https://raw.githubusercontent.com/inovex/trovilo/master/examples/k8s/alert-rules-team1.yaml \
  -f https://raw.githubusercontent.com/inovex/trovilo/master/examples/k8s/prometheus-config.yaml \
  -f https://raw.githubusercontent.com/inovex/trovilo/master/examples/k8s/trovilo-config.yaml \
  -f https://raw.githubusercontent.com/inovex/trovilo/master/examples/k8s/deployment.yaml
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details
