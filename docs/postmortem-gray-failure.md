# Postmortem: a frozen nginx worker kept receiving traffic

## Summary

A frozen nginx worker stayed in the balancer's pool with its heartbeat still reporting healthy, so the balancer kept routing requests to it. With the original sidecar (v1), roughly 1 request in 20 failed during a freeze, and this was stable across five batches. A v2 sidecar that beats only after a local probe succeeds shortens the failure window, but does not remove it.

## Impact

- v1: 5 of 100 probes failed (5%) across five freeze batches.
- v2: 4 of 100 probes failed (4%) across five freeze batches. The difference is within the noise of five batches of 20 and should not be read as a real improvement on its own.
- Timed v2 freeze: detected 2.1 s after the freeze, recovered 1 s after the thaw, one of 20 requests failed in the batch.

## Timeline (timed v2 freeze, UTC)

- 17:08:54 worker frozen
- 17:08:56 balancer marks the pod Suspect (silence crossed 2 s)
- 17:08:59 worker thawed
- 17:09:00 balancer marks the pod Healthy again

## Root cause

1. **v1:** the sidecar sent a beat every second regardless of whether nginx answered. A frozen worker therefore looked healthy to the balancer. The sidecar even reported zeros on a failed stub_status poll, which still counted as a beat.
2. **v2:** the sidecar gates its beat on a local probe. The gate fired within a second in the logs. The remaining failures come from the detection window: the balancer needs about 2 s of silence before it stops routing, and requests in that window go to the frozen pod.

## Contributing factors

- The SUSPECT threshold (S=2) with 1-second beats is tight. A healthy pod whose probe runs slow loses a beat and flaps, so the threshold can't simply be lowered to shrink the window without causing more flapping.
- The balancer routes only to Healthy pods, so once a pod is Suspect it gets no new requests. Failures are therefore limited to the window before detection.

## What we would change

- Shorten the detection window without flapping: either make the beat cheaper so the probe never misses a beat (it currently can take most of a second), or raise the beat rate and keep S in beats.
- Add per-backend metrics on the balancer so state changes and routed counts are visible without the journal.
- Authenticate telemetry with an HMAC, so only the pods can send beats.

## Open questions

- The exact share of failures attributable to the detection window versus in-flight connections, which the current logs don't separate.
- Whether a lower S with a faster probe removes the flapping without adding false DEAD states.
- Whether the one-in-twenty rate holds over a longer run.

## Status

Documented as measured. The shallow-heartbeat failure is reproduced and explained. The v2 gate is verified at the sidecar and in the balancer's transition log. Remaining work is listed above.
