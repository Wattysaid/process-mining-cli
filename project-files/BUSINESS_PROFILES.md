# Business Profiles

**Document version:** R1.00 (2026-01-11)

## 1. Purpose
Business profiles capture recurring context (systems, data sources, security constraints) so users do not repeat setup steps for each project.

## 2. Storage
- Location: `.business/` in the project directory
- Format: YAML, one file per business

## 3. Required fields
- `name`
- `industry`
- `region`

## 4. Optional fields
- `systems` (ERP, CRM, procurement platforms)
- `data_sources` (DBs, files, APIs, warehouses)
- `security` (PII handling, network constraints, offline requirements)
- `default_connectors` (preferred connector types and read-only modes)

## 5. Example profile
```yaml
name: Acme Manufacturing
industry: Manufacturing
region: APAC
systems:
  - SAP
  - Coupa
data_sources:
  - Postgres
  - S3
security:
  pii: restricted
  offline_only: false
default_connectors:
  - postgres
  - s3
```
