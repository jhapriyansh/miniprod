# Experiments

All runs used six web pods (two per worker where scheduled, one per worker after the rollout), the balancer on `cp`, and probes sent from the Dell to the balancer's :8080. Times are from the experiment logs and the balancer journal (UTC in the journal, IST in the scripts).

## Setup and baseline

- Six pods, sidecar v1 for the first runs and v2 for the gray-failure comparison.
- Baseline of 20 probes before each experiment: 20 of 20 succeeded in every run.

## E1: Pod kill

| Event | Result |
|---|---|
| Pod deleted | t = 0 |
| Replacement Running, all six 2/2 | about 107 s |
| Failed probes during the window | 0 of about 100 |
| Slowest probe | 36 ms |

Replacement time is dominated by container start on the worker. The balancer routed around the missing pod and admitted the replacement only after its beats were valid.

## E2: VM kill (worker w2, forced stop)

| Event | Result |
|---|---|
| First probe after the kill | 1 request at about 2.0 s, then 200 |
| Failed probes in the following 95 s | 0 of about 470 |
| Slowest probe after the first | under 10 ms |

The first slow request is the connect timeout on a dead backend, followed by a retry to a healthy one. Measured one event only, so treat the 2.0 s as a single observation.

## E3: Gray failure (frozen nginx worker)

Five batches per sidecar version. Each batch: read the counter, freeze the worker, send 20 probes from the Dell, thaw, read the counter.

| Batch | v1 failed of 20 | v2 failed of 20 |
|---|---|---|
| 1 | 1 | 1 |
| 2 | 1 | 1 |
| 3 | 1 | 0 |
| 4 | 1 | 1 |
| 5 | 1 | 1 |
| Total | 5 of 100 (5%) | 4 of 100 (4%) |

v2 sends a beat only after a local probe of nginx succeeds, so a frozen worker stops beating within about 1 second. The balancer's detection, measured in the timed batch, took 2.1 s (Healthy → Suspect), and recovery after thaw took 1 s.

## Timed freeze (v2, pod 10.42.1.44)

| Event | Time (UTC) |
|---|---|
| Freeze | 17:08:54 |
| Balancer: Healthy → Suspect | 17:08:56 (2.1 s after freeze) |
| Thaw | 17:08:59 |
| Balancer: Suspect → Healthy | 17:09:00 (1 s after thaw) |
| Failed probes in batch | 1 of 20 |

## Known observation: flapping at S=2

With 1-second beats and a 2-second SUSPECT threshold, a healthy pod whose probe runs slow flaps HEALTHY → SUSPECT → HEALTHY on about a 4-second cycle, with silence of about 2.15 s at each drop. The state machine behaves as specified; the threshold is tight for this beat interval.
