# miniprod

A miniature production environment on one laptop. Terraform and Ansible build four Incus VMs into a multi-node k3s cluster. A Go load balancer routes to nginx pods using telemetry heartbeats, and failure experiments measure how the system detects and recovers.

## Architecture

- **Provisioning:** Terraform (lxc/incus provider) creates four VMs with fixed IPs on a private bridge. cloud-init creates an `ansible` user.
- **Configuration:** Ansible roles for hardening (ufw, fail2ban), node_exporter, k3s prerequisites, the k3s server and agents, and the balancer.
- **Cluster:** k3s v1.37.1. The control plane is tainted `dedicated=control-plane:NoSchedule`, so workloads run only on the three workers.
- **Workload:** a `web` Deployment of six nginx pods, each with a telemetry sidecar.
- **Balancer (Go):** a TCP proxy on :8080 and a UDP telemetry receiver on :9000. Backends are Healthy, Suspect, or Dead, based on heartbeat silence.

## Design decisions

- **Telemetry is the heartbeat.** No separate health check. Silence is the failure signal. Packets carry their own `addr`, because pod traffic leaving the cluster is masqueraded.
- **Staged downward transitions.** HEALTHY → SUSPECT → DEAD, one level per tick, so the balancer reacts twice.
- **Recovery needs R=3 consecutive beats per level.** A beat more than 2×T after the last restarts the streak, which tolerates jitter. DEAD recovers through SUSPECT, never directly to HEALTHY.
- **New backends start SUSPECT.** They take traffic only after earning HEALTHY.
- **Least-connections routing** uses the balancer's own active-connection count, with ties rotated so sequential traffic spreads. The rotation was a real bug, found by the sequential batch.
- **Connect retries once** on a refused dial, with no state change on a refused dial. A 503 is returned only before any byte reaches the client.

## Results

See [experiments.md](experiments.md) for the tables and [postmortem-gray-failure.md](postmortem-gray-failure.md) for the gray-failure analysis.

## Known limits

- The balancer is a single point of failure, by design.
- Telemetry is unauthenticated. Anything on the pod network can send beats.
- The shallow sidecar (v1) keeps beating while its worker is frozen. The v2 gate fixes this only within the detection window.
- Detection is limited by the 2-second SUSPECT threshold with 1-second beats, which flaps under normal jitter.
- The Mac cannot reach the balancer directly; testing runs from the Dell.
- No per-backend metrics yet.
