# miniprod

A miniature production environment on one laptop. Terraform and Ansible build a four-node k3s cluster on Incus VMs. A Go load balancer routes traffic to nginx pods using heartbeats sent over UDP, and failure experiments measure how the system detects and recovers.

## What's in here

- `terraform/`: four Incus VMs with fixed IPs and cloud-init
- `ansible/`: hardening, node_exporter, k3s prerequisites, the k3s server and agents, and the balancer
- `cmd/balancer/`: the Go load balancer (TCP proxy and UDP telemetry receiver)
- `cmd/sidecar/`: the telemetry sidecar that runs beside each nginx pod
- `internal/`: telemetry parser, health state machine, routing, and proxy, all unit-tested
- `k8s/`: the web Deployment and its nginx config
- `scripts/`: rebuild timing and the experiment and batch runners
- `docs/`: design decisions, experiment results, and the gray-failure postmortem

## Quick start

```bash
cd terraform && terraform apply
cd ../ansible && ansible-playbook site.yml
```

Build the balancer binary first with `scripts/build-balancer.sh`, since the balancer role copies it onto `cp`.

## Read next

Start with [docs/README.md](docs/README.md) for the architecture and design decisions, then [docs/experiments.md](docs/experiments.md) for the results, and [docs/postmortem-gray-failure.md](docs/postmortem-gray-failure.md) for the main finding.

## Status and limits

The balancer and its experiments are complete. Known limits, including the single point of failure, unauthenticated telemetry, and the 2-second detection threshold, are listed in the docs.
