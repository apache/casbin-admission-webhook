# Example Casbin Policy Configurations

This directory contains example policy files for different use cases.

## Basic RBAC Example

```csv
# Admin role can do everything
p, admin, *, *

# Developer role can manage pods in development namespace
p, developer, pods/development, CREATE
p, developer, pods/development, UPDATE  
p, developer, pods/development, DELETE

# Viewer role can only read
p, viewer, */*, GET
```

## Namespace-based Access Control

```csv
# Team A can manage everything in team-a namespace
p, team-a-admin, */team-a, *

# Team B can manage everything in team-b namespace  
p, team-b-admin, */team-b, *

# SRE team can manage everything in all namespaces
p, sre, */*, *
```

## Resource-type Restrictions

```csv
# Allow creating configmaps and secrets in dev namespace
p, developer, configmaps/dev, CREATE
p, developer, configmaps/dev, UPDATE
p, developer, configmaps/dev, DELETE
p, developer, secrets/dev, CREATE
p, developer, secrets/dev, UPDATE
p, developer, secrets/dev, DELETE

# Deny production access
# (No rule = deny by default)

# Allow read access to all resources
p, developer, */*, GET
```

## Service Account Example

```csv
# Allow specific service account full access to a namespace
p, system:serviceaccount:default:my-app, */default, *

# Allow cluster-admin service account full access
p, system:serviceaccount:kube-system:cluster-admin, */*, *
```

## Mixed User and Service Account

```csv
# Human users
p, alice@example.com, pods/development, *
p, bob@example.com, */*, GET

# Service accounts
p, system:serviceaccount:production:app-deployer, pods/production, CREATE
p, system:serviceaccount:production:app-deployer, pods/production, UPDATE
p, system:serviceaccount:production:app-deployer, pods/production, DELETE

# Admin group
p, admin-group, */*, *
```

## Notes

- **Subject**: Can be a username, email, or service account name
- **Object**: Format is `{resource}/{namespace}` or use `*` for wildcards
- **Action**: Kubernetes operation (CREATE, UPDATE, DELETE, GET, etc.) or `*` for all
- Empty lines and lines starting with `#` are ignored
- The wildcard `*` matches any value
- If no rule matches, the request is denied by default
