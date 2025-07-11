#!/usr/bin/env bash
set -e

# This script destroys all keboola_component_configuration resources in the current directory's .tf files.
# It dynamically finds all such resources and builds the appropriate -target flags for terraform destroy.
# Usage: ./destroy-component-configurations.sh [additional terraform destroy args]

# Find all resource blocks of type keboola_component_configuration
configs=$(grep -hoP 'resource "keboola_component_configuration"\s*"[^"]*"' *.tf | \
  awk '{print $2"."$3}' | tr -d '"' | sort | uniq)

# Build the -target flags
config_targets=""
for c in $configs; do
  config_targets="$config_targets -target=$c"
done

# Run terraform destroy with all keboola_component_configuration targets
if [ -z "$config_targets" ]; then
  echo "No keboola_component_configuration resources found. Nothing to destroy."
  exit 0
fi

echo "Running: terraform destroy -var-file=.tfvars $config_targets $@"
terraform destroy -var-file=.tfvars $config_targets "$@" 