#!/usr/bin/env bash
# Usage: bash scripts/gray-batch.sh <pod> <runs>
# One gray-failure batch per run: counter before, freeze, 20 probes from the Dell, thaw, counter after.
set -u
POD=$1
RUNS=${2:-1}
DELL=trexible@trexible.local

counter() {
  kubectl exec "$POD" -c nginx -- wget -qO- http://127.0.0.1:8081/stub_status | sed -n 3p | awk '{print $1}'
}

worker() {
  kubectl exec "$POD" -c nginx -- sh -c 'ps -o pid,args | grep "worker process" | grep -v grep | head -1 | awk "{print \$1}"'
}

for run in $(seq 1 "$RUNS"); do
  before=$(counter)
  W=$(worker)
  echo "freeze $(date +%H:%M:%S) $W"
  kubectl exec "$POD" -c nginx -- kill -STOP "$W"
  probes=$(ssh "$DELL" "bash -s" <<'PROBE' 2>/dev/null
for i in $(seq 1 20); do
  curl -s -o /dev/null -m 3 -w '%{http_code}\n' http://10.95.18.210:8080/
done | sort | uniq -c
PROBE
)
  kubectl exec "$POD" -c nginx -- kill -CONT "$W"
  echo "thaw   $(date +%H:%M:%S)"
  after=$(counter)
  echo "run $run: counter $before -> $after"
  echo "$probes" | sed 's/^/    /'
done
