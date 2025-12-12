# ACP Records - Architecture Summary

## System Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Client Applications                           │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐  │
│  │  Flutter Mobile  │  │   Web Admin UI   │  │   API Clients    │  │
│  └────────┬─────────┘  └────────┬─────────┘  └────────┬─────────┘  │
└───────────┼────────────────────┼────────────────────┼──────────────┘
            │                    │                    │
            └────────────────────┴────────────────────┘
                                 │
                        [HTTPS/TLS 1.3]
                                 │
┌────────────────────────────────┼──────────────────────────────────┐
│                    Firebase Authentication                          │
│                    (JWT Token Validation)                           │
└────────────────────────────────┼──────────────────────────────────┘
                                 │
┌────────────────────────────────▼──────────────────────────────────┐
│                      API Layer (Go Chi Router)                     │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  Middleware Stack                                             │ │
│  │  • AuthMiddleware (Firebase)                                  │ │
│  │  • AuditLoggerMiddleware                                      │ │
│  │  • RequestID, Logger, Recoverer, Timeout                      │ │
│  │  • CORS                                                        │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                     │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  Routes: /api/v1/patients/{patient_id}/acp-records           │ │
│  │  • POST   /                  - Create ACP Record              │ │
│  │  • GET    /                  - List ACP Records               │ │
│  │  • GET    /latest            - Get Latest Active ACP          │ │
│  │  • GET    /history           - Get Complete History           │ │
│  │  • GET    /{id}              - Get ACP by ID                  │ │
│  │  • PUT    /{id}              - Update ACP Record              │ │
│  │  • DELETE /{id}              - Delete ACP Record              │ │
│  └──────────────────┬───────────────────────────────────────────┘ │
└─────────────────────┼─────────────────────────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────────────────────────┐
│                    Handler Layer                                   │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  ACPRecordHandler                                             │ │
│  │  • CreateACPRecord(w, r)      - HTTP request handling         │ │
│  │  • GetACPRecord(w, r)         - URL param extraction          │ │
│  │  • GetACPRecords(w, r)        - Query param parsing           │ │
│  │  • UpdateACPRecord(w, r)      - JSON encoding/decoding        │ │
│  │  • DeleteACPRecord(w, r)      - Error mapping                 │ │
│  │  • GetLatestACP(w, r)         - HTTP status codes             │ │
│  │  • GetACPHistory(w, r)        - Response formatting           │ │
│  └──────────────────┬───────────────────────────────────────────┘ │
└─────────────────────┼─────────────────────────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────────────────────────┐
│                    Service Layer                                   │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  ACPRecordService                                             │ │
│  │  • CreateACPRecord()          - Business logic                │ │
│  │  • GetACPRecord()             - Validation rules              │ │
│  │  • ListACPRecords()           - Status validation             │ │
│  │  • UpdateACPRecord()          - Decision maker validation     │ │
│  │  • DeleteACPRecord()          - Proxy person ID validation    │ │
│  │  • GetLatestACP()             - Data sensitivity validation   │ │
│  │  • GetACPHistory()            - Patient existence check       │ │
│  └──────────────────┬───────────────────────────────────────────┘ │
│                     │                                              │
│  Dependencies:      │                                              │
│  • ACPRecordRepository                                            │
│  • PatientRepository (for existence validation)                   │
└─────────────────────┼─────────────────────────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────────────────────────┐
│                  Repository Layer                                  │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  ACPRecordRepository                                          │ │
│  │  • Create()                   - Data access operations        │ │
│  │  • GetByID()                  - SQL query building            │ │
│  │  │  • List()                     - Filter construction          │ │
│  │  • Update()                   - JSONB field handling          │ │
│  │  • Delete()                   - Spanner mutations             │ │
│  │  • GetLatestACP()             - Composite key operations      │ │
│  │  • GetACPHistory()            - Index utilization             │ │
│  │  • scanACPRecord()            - Row scanning                  │ │
│  └──────────────────┬───────────────────────────────────────────┘ │
│                     │                                              │
│  Dependencies:      │                                              │
│  • SpannerRepository (Cloud Spanner client wrapper)               │
└─────────────────────┼─────────────────────────────────────────────┘
                      │
┌─────────────────────▼─────────────────────────────────────────────┐
│                   Data Layer                                       │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  Cloud Spanner (PostgreSQL Interface)                        │ │
│  │  ┌────────────────────────────────────────────────────────┐  │ │
│  │  │  Table: acp_records                                     │  │ │
│  │  │  Primary Key: (patient_id, acp_id)                      │  │ │
│  │  │  Columns:                                                │  │ │
│  │  │    • acp_id (varchar(36))                               │  │ │
│  │  │    • patient_id (varchar(36)) [FK → patients]           │  │ │
│  │  │    • recorded_date (date)                               │  │ │
│  │  │    • version (int)                                      │  │ │
│  │  │    • status (varchar(20))                               │  │ │
│  │  │    • decision_maker (varchar(20))                       │  │ │
│  │  │    • proxy_person_id (varchar(36))                      │  │ │
│  │  │    • directives (jsonb)                                 │  │ │
│  │  │    • values_narrative (text)                            │  │ │
│  │  │    • legal_documents (jsonb)                            │  │ │
│  │  │    • discussion_log (jsonb)                             │  │ │
│  │  │    • data_sensitivity (varchar(20))                     │  │ │
│  │  │    • access_restricted_to (jsonb)                       │  │ │
│  │  │    • created_by (varchar(36))                           │  │ │
│  │  │    • created_at (timestamptz)                           │  │ │
│  │  │                                                           │  │ │
│  │  │  Indexes:                                                │  │ │
│  │  │    • idx_acp_active ON (patient_id, version DESC)       │  │ │
│  │  │      WHERE status = 'active'                            │  │ │
│  │  └────────────────────────────────────────────────────────┘  │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                    │
│  Region: asia-northeast1 (Tokyo)                                  │
│  Encryption: CMEK (Customer-Managed Encryption Keys)              │
│  Replication: Multi-region for 99.99% availability                │
└────────────────────────────────────────────────────────────────────┘

┌────────────────────────────────────────────────────────────────────┐
│                    Audit & Logging                                 │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  Cloud Audit Logs                                             │ │
│  │  • All ACP record access logged                               │ │
│  │  • User ID, timestamp, action tracked                         │ │
│  │  • Retention: 5 years (compliance requirement)                │ │
│  └──────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────────────┘
```

## Data Flow

### Create ACP Record Flow

```
1. Client → POST /api/v1/patients/{patient_id}/acp-records
   ├─ Headers: Authorization: Bearer {firebase_token}
   └─ Body: { recorded_date, status, decision_maker, directives, ... }

2. AuthMiddleware → Validates Firebase JWT token
   └─ Extracts user_id from token

3. AuditLoggerMiddleware → Logs access attempt
   └─ Records: user_id, patient_id, action, timestamp

4. ACPRecordHandler.CreateACPRecord()
   ├─ Parses request body
   ├─ Extracts patient_id from URL
   └─ Calls service layer

5. ACPRecordService.CreateACPRecord()
   ├─ Validates patient exists (PatientRepository)
   ├─ Validates status (draft|active|superseded)
   ├─ Validates decision_maker (patient|proxy|guardian)
   ├─ Validates proxy_person_id (required if proxy/guardian)
   ├─ Validates data_sensitivity
   └─ Validates directives (required)

6. ACPRecordRepository.Create()
   ├─ Generates UUID for acp_id
   ├─ Sets version = 1
   ├─ Converts json.RawMessage → sql.NullString for JSONB fields
   ├─ Creates Spanner mutation
   └─ Applies mutation to database

7. Cloud Spanner
   ├─ Validates foreign key constraint (patient_id)
   ├─ Inserts row into acp_records table
   └─ Returns success

8. Response
   ├─ Status: 201 Created
   └─ Body: ACPRecord object with acp_id, created_at, etc.
```

### Get Latest Active ACP Flow

```
1. Client → GET /api/v1/patients/{patient_id}/acp-records/latest

2. AuthMiddleware → Validates token

3. AuditLoggerMiddleware → Logs access

4. ACPRecordHandler.GetLatestACP()
   └─ Extracts patient_id from URL

5. ACPRecordService.GetLatestACP()
   └─ Validates patient exists

6. ACPRecordRepository.GetLatestACP()
   ├─ SQL Query:
   │  SELECT * FROM acp_records
   │  WHERE patient_id = @patient_id AND status = 'active'
   │  ORDER BY version DESC, recorded_date DESC
   │  LIMIT 1
   │
   ├─ Uses idx_acp_active index for fast retrieval
   └─ Converts sql.NullString → json.RawMessage

7. Response
   ├─ Status: 200 OK
   └─ Body: Latest active ACPRecord
```

## Component Interaction Matrix

```
┌────────────┬─────────┬────────┬──────────┬────────────┬─────────┐
│ Component  │ Handler │ Service│Repository│  Spanner   │ Patient │
├────────────┼─────────┼────────┼──────────┼────────────┼─────────┤
│ Handler    │    -    │   ✓    │    -     │     -      │    -    │
│ Service    │    -    │   -    │    ✓     │     -      │    ✓    │
│ Repository │    -    │   -    │    -     │     ✓      │    -    │
│ Spanner    │    -    │   -    │    -     │     -      │    ✓    │
│ Patient    │    -    │   -    │    -     │     -      │    -    │
└────────────┴─────────┴────────┴──────────┴────────────┴─────────┘

Legend: ✓ = Direct dependency
```

## Security Layers

```
┌────────────────────────────────────────────────────────────┐
│ Layer 1: Network Security                                  │
│ • TLS 1.3 encryption for all HTTPS traffic                 │
│ • Cloud Armor DDoS protection                              │
│ • Identity-Aware Proxy (IAP) for admin access              │
└────────────────────────────────────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────────┐
│ Layer 2: Authentication                                     │
│ • Firebase Authentication (JWT tokens)                     │
│ • Token validation on every request                        │
│ • User identity verification                               │
└────────────────────────────────────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────────┐
│ Layer 3: Authorization                                      │
│ • Patient assignment validation                            │
│ • Access control lists (access_restricted_to field)        │
│ • Data sensitivity checks                                  │
└────────────────────────────────────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────────┐
│ Layer 4: Data Protection                                    │
│ • CMEK encryption at rest (Cloud Spanner)                  │
│ • Field-level sensitivity classification                   │
│ • Audit logging of all access                              │
└────────────────────────────────────────────────────────────┘
                           ▼
┌────────────────────────────────────────────────────────────┐
│ Layer 5: Compliance                                         │
│ • 3省2ガイドライン adherence                                │
│ • 5-year audit log retention                               │
│ • Data residency in Japan (asia-northeast1)               │
└────────────────────────────────────────────────────────────┘
```

## Version Management Workflow

```
Patient's ACP Journey:

Initial ACP (Version 1)
┌──────────────────────────────────────┐
│ acp_id: uuid-001                     │
│ version: 1                           │
│ status: active                       │
│ directives: { dnar: true }           │
│ recorded_date: 2025-01-01            │
└──────────────────────────────────────┘
              │
              │ Patient's wishes change
              ▼
Update: Mark as superseded
┌──────────────────────────────────────┐
│ acp_id: uuid-001                     │
│ version: 1                           │
│ status: superseded ← Updated         │
│ directives: { dnar: true }           │
│ recorded_date: 2025-01-01            │
└──────────────────────────────────────┘
              │
              │ Create new version
              ▼
New ACP (Version 2)
┌──────────────────────────────────────┐
│ acp_id: uuid-002                     │
│ version: 2                           │
│ status: active                       │
│ directives: { dnar: true,            │
│               ventilator: true }     │
│ recorded_date: 2025-06-01            │
└──────────────────────────────────────┘

GetLatestACP() → Returns uuid-002 (version 2, active)
GetACPHistory() → Returns [uuid-002, uuid-001] (complete history)
```

## Performance Optimization

```
┌────────────────────────────────────────────────────────────┐
│ Query Optimization Strategy                                │
├────────────────────────────────────────────────────────────┤
│                                                             │
│ 1. Index Usage                                             │
│    • idx_acp_active covers most common query               │
│    • (patient_id, version DESC) with status filter         │
│    • O(log n) lookup instead of full table scan            │
│                                                             │
│ 2. Composite Primary Key                                   │
│    • (patient_id, acp_id) enables efficient lookups        │
│    • Co-located data for same patient                      │
│    • Parent-child relationship (if using INTERLEAVE)       │
│                                                             │
│ 3. JSONB Optimization                                      │
│    • Store complex objects without multiple joins          │
│    • Flexible schema for directives                        │
│    • No need for separate tables                           │
│                                                             │
│ 4. Pagination                                              │
│    • limit/offset support prevents memory exhaustion       │
│    • Default limit: 100 records                            │
│    • Client can adjust based on needs                      │
│                                                             │
│ 5. Single-Query Operations                                 │
│    • GetLatestACP: Single query with LIMIT 1               │
│    • No multiple round-trips to database                   │
│    • Efficient version sorting with DESC                   │
└────────────────────────────────────────────────────────────┘
```

## Error Handling Flow

```
Error occurs at any layer
         │
         ▼
┌─────────────────────────────┐
│ Catch error                 │
└─────────────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ Log error with context      │
│ • User ID                   │
│ • Patient ID                │
│ • Action attempted          │
│ • Error details             │
└─────────────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ Map to appropriate          │
│ HTTP status code            │
│ • 400: Validation error     │
│ • 404: Not found            │
│ • 500: Server error         │
└─────────────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ Return user-friendly        │
│ error message               │
│ (sanitized, no internals)   │
└─────────────────────────────┘
         │
         ▼
┌─────────────────────────────┐
│ Client receives error       │
│ and displays to user        │
└─────────────────────────────┘
```

## File Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go ✅ (Modified - Added ACP routing)
│
├── internal/
│   ├── models/
│   │   └── acp_record.go ✅ (New)
│   │
│   ├── repository/
│   │   ├── spanner.go (Base repository)
│   │   └── acp_record_repository.go ✅ (New)
│   │
│   ├── services/
│   │   └── acp_record_service.go ✅ (New)
│   │
│   └── handlers/
│       └── acp_records.go ✅ (New)
│
└── migrations/
    └── 009_create_acp_records.sql ✅ (Existing)

docs/
├── ACP_RECORDS_IMPLEMENTATION.md ✅ (New)
└── API_ACP_RECORDS.md ✅ (New)
```

## Technology Stack

```
┌─────────────────────────────────────────────────────────────┐
│ Programming Language                                        │
│ • Go 1.22+ (Goroutines, Channels, Context)                  │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Web Framework                                               │
│ • Chi v5 (Lightweight router, middleware support)           │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Database                                                     │
│ • Cloud Spanner (PostgreSQL Interface)                      │
│ • JSONB support for flexible data                           │
│ • 99.99% availability SLA                                   │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Authentication                                               │
│ • Firebase Authentication (JWT)                             │
│ • Identity Platform integration                             │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│ Logging & Monitoring                                         │
│ • Custom logger package (structured logging)                │
│ • Cloud Audit Logs                                          │
└─────────────────────────────────────────────────────────────┘
```

## Deployment Architecture

```
┌────────────────────────────────────────────────────────────┐
│ Client Layer                                               │
│ • Flutter Mobile App (iOS/Android)                         │
│ • Web Admin Interface (Flutter Web)                        │
└────────────────────────────────────────────────────────────┘
                          ▼
┌────────────────────────────────────────────────────────────┐
│ Load Balancer Layer                                        │
│ • Cloud Load Balancing (HTTPS)                             │
│ • SSL/TLS Termination                                      │
└────────────────────────────────────────────────────────────┘
                          ▼
┌────────────────────────────────────────────────────────────┐
│ Application Layer (Cloud Run)                              │
│ • Go API Server (containerized)                            │
│ • Auto-scaling (0 to N instances)                          │
│ • Region: asia-northeast1 (Tokyo)                          │
└────────────────────────────────────────────────────────────┘
                          ▼
┌────────────────────────────────────────────────────────────┐
│ Database Layer                                              │
│ • Cloud Spanner                                            │
│ • Multi-region replication                                 │
│ • CMEK encryption                                          │
└────────────────────────────────────────────────────────────┘
```

---

**Document Status**: Complete
**Last Updated**: 2025-12-12
**Architecture Version**: 1.0
