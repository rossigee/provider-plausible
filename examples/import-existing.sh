#!/bin/bash
set -e

echo "Importing existing Plausible resources..."

# List of known domains that likely have Plausible tracking
DOMAINS=(
    "debs.example.com"
    "dynamicip.example.org" 
    "vault.example.com"
    "sso.example.com"
    "vaultwarden.example.org"
)

echo "Checking provider status..."
kubectl get providers.pkg.crossplane.io provider-plausible -o wide

echo "Checking provider configuration..."
kubectl get providerconfigs.plausible.m.crossplane.io default -o yaml

echo "Creating sites for import..."
for domain in "${DOMAINS[@]}"; do
    echo "Creating site for $domain..."
    
    # Create the site resource with the external-name annotation for import
    cat <<EOF | kubectl apply -f -
apiVersion: site.plausible.m.crossplane.io/v1beta1
kind: Site
metadata:
  name: $(echo "$domain" | tr '.' '-')
  annotations:
    crossplane.io/external-name: "$domain"
spec:
  forProvider:
    domain: "$domain"
    timezone: Asia/Bangkok
  providerConfigRef:
    name: default
EOF
done

echo "Waiting for sites to sync..."
sleep 10

echo "Checking site status..."
kubectl get sites.site.plausible.m.crossplane.io -o wide

echo "Checking site details..."
for domain in "${DOMAINS[@]}"; do
    site_name=$(echo "$domain" | tr '.' '-')
    echo "=== Site: $domain ==="
    kubectl get sites.site.plausible.m.crossplane.io "$site_name" -o yaml | grep -A 10 -B 5 conditions || echo "No conditions found"
    echo
done

echo "Import process completed. Check the status above for any errors."
echo "If sites show as 'Synced: True', they have been successfully imported from Plausible."