# Changes made outside the code

| Date | Change | Where it lives | In the repo? |
|---|---|---|---|
| 2026-10-18 | Pod tolerations for unreachable / not-ready nodes lowered from 300 s to <your value> on the test deployment, edited live with `kubectl edit` | cluster only | only if saved in k8s/web.yaml |
| 2026-10-18 | Worker RAM raised from 768 MiB to 1 GiB | terraform/main.tf | yes |
| 2026-10-18 | Control plane tainted so pods run only on workers | ansible k3s_server template | yes |
| 2026-10-18 | k3s traefik, servicelb and metrics-server disabled | ansible k3s_server template | yes |
