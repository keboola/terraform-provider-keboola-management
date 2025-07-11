#!/usr/bin/env bash
set -e

# Find all resource blocks containing "project" in their type
targets=$(grep -hoP 'resource "\K[^"]*project[^"]*"\s*"[^"]*"' *.tf | \
  awk '{print $1"."$2}' | tr -d '"' | sort | uniq)

# Build the -target flags
target_flags=""
for t in $targets; do
  target_flags="$target_flags -target=$t"
done

# Run terraform apply with all project-related targets
echo "Running: terraform apply -var-file=.tfvars $target_flags $@"
terraform apply -var-file=.tfvars $target_flags "$@"