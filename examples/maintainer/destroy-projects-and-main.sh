#!/usr/bin/env bash
set -e

# This script destroys resources in reverse order of creation to avoid dependency issues.
# 1. Export KBC_TOKEN from terraform output
# 2. Destroy all non-project resources (e.g., configurations)
# 3. Destroy the rest (maintainer, organization, projects)

# Step 1: Export KBC_TOKEN from terraform output
export KBC_TOKEN=$(terraform output -raw project_storage_token)
echo "Exported KBC_TOKEN from project_storage_token output."

# Step 2: Destroy all non-project resources (e.g., configurations)
# Use the helper script to destroy all keboola_component_configuration resources
./destroy-component-configurations.sh "$@" || true

# Step 3: Destroy the rest of the resources (maintainer, organization, projects)
terraform destroy -var-file=.tfvars "$@" 