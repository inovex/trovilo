# trovilo

Trovilo collects and prepares files from Kubernetes ConfigMaps for Prometheus & friends

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
jobs:
  - name: alert-rules
    selector:
      type: prometheus-alerts
    verify:
      - name: verify alert rule validity
        cmd: ["promtool", "check", "rules", "%s"]
    target-dir: /etc/prometheus-alerts/
    flatten: true
    post-deploy:
      - name: reload prometheus
        cmd: ["curl", "-s", "-X", "POST", "http://localhost:9090/-/reload"]
```

Full example Kubernetes deployment with Prometheus:

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus
  namespace: monitoring
data:
  trovilo-config.yaml: |
    jobs:
      - name: alert-rules
        selector:
          type: prometheus-alerts
        verify:
          - name: verify alert rule validity
            cmd: ["promtool", "check", "rules", "%s"]
        target-dir: /etc/prometheus-alerts/
        flatten: true
        post-deploy:
          - name: reload prometheus
            cmd: ["curl", "-s", "-X", "POST", "http://localhost:9090/-/reload"]
  prometheus.yml: |-
    global:
      #scrape_interval: 1m
      #evaluation_interval: 1m
      external_labels:
    ...
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: prometheus
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      name: prometheus
      labels:
        app: prometheus
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9090"
    spec:
      # Allow non-root user to access PersistentVolume
      #securityContext:
      #  runAsUser: 65534
      #  fsGroup: 65534
      # Optional SA for prometheus itself
      #serviceAccountName: prometheus
      volumes:
        - name: config-volume
          configMap:
            name: prometheus
        # Optional AWS EBS volume
        #- name: data
        #  persistentVolumeClaim:
        #    claimName: prometheus-ebs
        - name: prometheus-alerts
          emptyDir:
            medium: Memory
      containers:
      - name: prometheus
        image: prom/prometheus:v2.3.1
        args:
          - '--config.file=/etc/prometheus/prometheus.yml'
          - '--web.enable-lifecycle'
          - '--storage.tsdb.retention=90d'
          - '--storage.tsdb.path=/prometheus/data/'
        ports:
          - containerPort: 9090
        volumeMounts:
          - name: config-volume
            mountPath: /etc/prometheus
            readOnly: true
          # Optional AWS EBS volume
          #- name: data
          #  mountPath: /prometheus/data/
          - name: prometheus-alerts
            mountPath: /etc/prometheus-alerts
        resources:
          limits:
            cpu: 1
            memory: 10Gi
          requests:
            cpu: 500m
            memory: 1Gi
      - name: trovilo
        image: inovex/trovilo:tools-2045130-dev
        args:
          - '--config=/etc/prometheus/trovilo-config.yaml'
          - '--log-json'
          #- '--log-level=debug'
        volumeMounts:
          - name: config-volume
            mountPath: /etc/prometheus
            readOnly: true
          - name: prometheus-alerts
            mountPath: /etc/prometheus-alerts
        resources:
          limits:
            cpu: 100m
            memory: 200mi
          requests:
            cpu: 100m
            memory: 200mi
```

See ``examples/configmap/`` for demo Prometheus alert configmaps.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details
