#!/usr/bin/env bash
set -e

# Step 1: Apply only project-related resources
# Allow this step to fail, so the script continues even if apply-projects.sh fails
./apply-projects.sh "$@" || true

# Step 2: Export KBC_TOKEN from terraform output
export KBC_TOKEN=$(terraform output -raw project_storage_token)
echo "Exported KBC_TOKEN from project_storage_token output."

# Step 3: Run full terraform apply with the token available
terraform apply -var-file=.tfvars "$@" || true