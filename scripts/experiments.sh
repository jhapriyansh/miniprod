#!/usr/bin/env bash
# Run from the Mac: bash scripts/experiments.sh {gray|pod|vm}
# Probes and load run on the Dell over ssh (the balancer is reachable only there).
# kubectl runs from the Mac through the 6443 tunnel; incus targets the Dell remote.
set -u

MODE=${1:-}
DELL=trexible@trexible.local
TARGET_POD=web-78d68bd9f7-bv5wc   # live v2 pod
TARGET_VM=w2
LOG=/tmp/exp-${MODE:-none}-$(date +%Y%m%d-%H%M%S).log
LOADPID=

ms() { python3 -c 'import time; print(int(time.time()*1000))'; }
ts() { printf '%s  %s\n' "$(date +%H:%M:%S)" "$*" | tee -a "$LOG"; }

# probe N: N requests from the Dell. Prints "ok X, failed Y, slowest Zs".
probe() {
  ssh "$DELL" "bash -s $1" <<'REMOTE'
ok=0; bad=0; slow=0
for _ in $(seq 1 $1); do
  out=$(curl -s -o /dev/null -m 3 -w '%{http_code} %{time_total}' http://10.95.18.210:8080/)
  code=${out% *}; t=${out#* }
  if [ "$code" = 200 ]; then ok=$((ok+1)); else bad=$((bad+1)); fi
  slow=$(awk -v a="$slow" -v b="$t" 'BEGIN{print (b>a)?b:a}')
done
echo "ok $ok, failed $bad, slowest ${slow}s"
REMOTE
}

start_load() {
  ssh "$DELL" "while :; do curl -s -o /dev/null -m 3 http://10.95.18.210:8080/ & sleep 0.1; done" &
  LOADPID=$!
}

stop_load() {
  [ -n "$LOADPID" ] && kill "$LOADPID" 2>/dev/null
  ssh "$DELL" "pkill -f '10.95.18.210:8080' || true" 2>/dev/null
}
trap stop_load EXIT

case "$MODE" in
  gray|pod|vm) ;;
  *) echo "usage: $0 {gray|pod|vm}"; exit 1 ;;
esac

ts "start mode=$MODE"
start_load
sleep 10
ts "baseline: $(probe 20)"

case "$MODE" in
  gray)
    # Freeze only the worker that serves requests. The master (pid 1) does not.
    WORKER=$(kubectl exec "$TARGET_POD" -c nginx -- sh -c 'ps -o pid,args | grep "worker process" | grep -v grep | head -1 | awk "{print \$1}"')
    ts "freeze worker $WORKER in $TARGET_POD"
    kubectl exec "$TARGET_POD" -c nginx -- kill -STOP "$WORKER"
    STATE=$(kubectl exec "$TARGET_POD" -c nginx -- sh -c "grep State /proc/$WORKER/status | awk '{print \$2}'")
    ts "worker $WORKER state=$STATE"
    [ "$STATE" = T ] || { ts "freeze failed, aborting"; kill -9 "$LOADPID" 2>/dev/null; exit 1; }
    kubectl exec "$TARGET_POD" -c nginx -- wget -T 3 -qO- http://127.0.0.1:8080/ >/dev/null 2>&1
    ts "direct check exit=$? (124 or 1 = hung)"
    for _ in $(seq 1 20); do ts "during freeze: $(probe 5)"; sleep 1; done
    kubectl exec "$TARGET_POD" -c nginx -- kill -CONT "$WORKER"
    ts "thaw worker $WORKER"
    ;;
  pod)
    ts "delete $TARGET_POD"
    T0=$(ms)
    kubectl delete pod "$TARGET_POD" --wait=false >/dev/null
    for _ in $(seq 1 60); do
      READY=$(kubectl get pods -l app=web --no-headers | grep -c ' 2/2 .*Running')
      ts "pods ready=$READY: $(probe 5)"
      [ "$READY" -eq 3 ] && break
      sleep 1
    done
    ts "recovered in $(( $(ms) - T0 )) ms"
    ;;
  vm)
    ts "force-stop worker $TARGET_VM"
    T0=$(ms)
    incus stop "dell:$TARGET_VM" --force --project iac
    for _ in $(seq 1 60); do ts "after kill ($(( $(ms) - T0 )) ms): $(probe 5)"; sleep 1; done
    ts "restart $TARGET_VM"
    incus start "dell:$TARGET_VM" --project iac
    ;;
esac

ts "end mode=$MODE"
echo "log: $LOG"
