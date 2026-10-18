#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
start=$(date +%s)
terraform -chdir=terraform destroy -auto-approve
terraform -chdir=terraform apply -auto-approve
t_tf=$(date +%s)
(cd ansible && ansible-playbook site.yml)
end=$(date +%s)
ansible server -b -a "k3s kubectl get nodes"
echo "Terraform: $((t_tf-start))s | Ansible: $((end-t_tf))s | total: $((end-start))s"
