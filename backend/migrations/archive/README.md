# Migration Archive

This directory contains archived migration files for historical reference and documentation purposes.

## Directory Structure

```
archive/
├── README.md (this file)
├── google-sql-dialect/          # Original migrations using Google SQL dialect
├── duplicate_files/             # Duplicate migration files (not conforming to DATABASE_REQUIREMENTS.md)
└── 001_create_patients_enhanced.sql  # Early enhanced version
```

## Archive Categories

### 1. `google-sql-dialect/` - Original Google SQL Dialect Migrations

**Date Archived**: 2025-12-12
**Reason**: Migration from Google SQL dialect to PostgreSQL dialect

These files were the initial migrations created using Cloud Spanner's Google SQL dialect. The project has since migrated to using the PostgreSQL interface for better compatibility and JSONB support.

**Files**:
- `001_create_patients_table.sql`
- `002_create_doctors_table.sql`
- `003_create_visits_table.sql`
- `004_create_visit_records_table.sql`

**Do not use**: These are obsolete and replaced by PostgreSQL dialect equivalents.

---

### 2. `duplicate_files/` - Duplicate Migration Numbers

**Date Archived**: 2025-12-12
**Reason**: Resolved migration number conflicts (003-010 duplicates)

These files had duplicate migration numbers with files defined in DATABASE_REQUIREMENTS.md. They represent alternative implementations that were superseded by the canonical definitions.

**Files**:
- `003_create_staff_patient_assignments.sql` (conflicted with `003_create_patient_social_profiles.sql`)
- `004_create_audit_patient_access_logs.sql` (conflicted with `004_create_patient_coverages.sql`)
- `005_create_patient_social_profiles.sql` (duplicate, wrong number)
- `006_create_patient_coverages.sql` (duplicate, wrong number)
- `007_create_medical_conditions.sql` (duplicate, wrong number)
- `008_create_allergy_intolerances.sql` (duplicate, wrong number)
- `009_create_view_my_patients.sql` (conflicted with `009_create_acp_records.sql`)
- `010_create_indexes.sql` (conflicted with `010_create_medication_orders.sql`)

**Note**: Some of these tables may still be valuable and could be re-introduced with correct migration numbers in the future.

---

### 3. `001_create_patients_enhanced.sql` - Early Enhanced Version

**Date Archived**: Prior to 2025-12-12
**Reason**: Superseded by canonical `001_create_patients.sql` aligned with DATABASE_REQUIREMENTS.md

An early iteration of the patients table with different field definitions.

---

## Usage Guidelines

### When to Reference These Files

1. **Understanding Project Evolution**: To see how the schema design evolved
2. **Recovering Lost Definitions**: If a table from `duplicate_files/` needs to be re-implemented
3. **Debugging Historical Issues**: If production data was created with old schemas
4. **Documentation**: For onboarding new team members

### When NOT to Use These Files

- **Never apply these migrations to production or development databases**
- **Do not copy DDL from here without reviewing against current DATABASE_REQUIREMENTS.md**
- **These are for reference only, not active migrations**

---

## Canonical Migration Source

The current, canonical migration definitions are in:
- **Location**: `/backend/migrations/` (parent directory)
- **Specification**: [docs/DATABASE_REQUIREMENTS.md](../../../docs/DATABASE_REQUIREMENTS.md)
- **Files**: `001_create_patients.sql` through `015_create_staff_tables.sql`

---

## Maintenance

- **Retention Policy**: Keep indefinitely for historical reference
- **Git Status**: Committed to version control
- **Review Date**: Review annually to assess if still needed

---

**Last Updated**: 2025-12-12
**Maintainer**: Technical Lead
