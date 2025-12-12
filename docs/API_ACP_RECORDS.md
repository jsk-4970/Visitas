# ACP Records API Reference

## Base URL
```
/api/v1/patients/{patient_id}/acp-records
```

## Authentication
All endpoints require Firebase Authentication token in the `Authorization` header.

---

## Endpoints

### 1. Create ACP Record

**POST** `/api/v1/patients/{patient_id}/acp-records`

Creates a new Advance Care Planning record for a patient.

#### Request Body

```json
{
  "recorded_date": "2025-12-12",
  "status": "active",
  "decision_maker": "patient",
  "proxy_person_id": "key_person_uuid",
  "directives": {
    "dnar": true,
    "ventilator": false,
    "feeding_tube": false,
    "palliative_care_only": true,
    "hospitalization": "avoid_if_possible",
    "cpr": false
  },
  "values_narrative": "患者は自然な形での終末期を希望。延命措置は望まない。緩和ケアを優先。",
  "legal_documents": {
    "living_will": "gs://bucket/documents/living_will.pdf",
    "power_of_attorney": "gs://bucket/documents/poa.pdf"
  },
  "discussion_log": [
    {
      "date": "2025-12-10",
      "participants": ["主治医", "患者", "長女"],
      "location": "患者宅",
      "summary": "終末期医療についての意向確認。患者の価値観と希望を詳細に聞き取り。",
      "decisions": ["DNAR同意", "緩和ケア優先"]
    }
  ],
  "data_sensitivity": "highly_confidential",
  "access_restricted_to": ["doctor_uid_123", "nurse_uid_456"],
  "created_by": "doctor_uid_123"
}
```

#### Response (201 Created)

```json
{
  "acp_id": "acp_uuid_789",
  "patient_id": "patient_uuid_123",
  "recorded_date": "2025-12-12T00:00:00Z",
  "version": 1,
  "status": "active",
  "decision_maker": "patient",
  "proxy_person_id": "key_person_uuid",
  "directives": { ... },
  "values_narrative": "患者は自然な形での終末期を希望...",
  "legal_documents": { ... },
  "discussion_log": [ ... ],
  "data_sensitivity": "highly_confidential",
  "access_restricted_to": ["doctor_uid_123", "nurse_uid_456"],
  "created_by": "doctor_uid_123",
  "created_at": "2025-12-12T10:30:00Z"
}
```

#### Validation Rules

| Field | Required | Values | Notes |
|-------|----------|--------|-------|
| `recorded_date` | ✅ | Date | Date when ACP was recorded |
| `status` | ✅ | `draft`, `active`, `superseded` | Current status |
| `decision_maker` | ✅ | `patient`, `proxy`, `guardian` | Who makes decisions |
| `proxy_person_id` | ⚠️ | UUID | Required if decision_maker is `proxy` or `guardian` |
| `directives` | ✅ | JSON object | DNAR, ventilator preferences, etc. |
| `created_by` | ✅ | String | User ID who created the record |

---

### 2. Get ACP Record by ID

**GET** `/api/v1/patients/{patient_id}/acp-records/{id}`

Retrieves a specific ACP record.

#### Response (200 OK)

```json
{
  "acp_id": "acp_uuid_789",
  "patient_id": "patient_uuid_123",
  "recorded_date": "2025-12-12T00:00:00Z",
  "version": 1,
  "status": "active",
  "decision_maker": "patient",
  "directives": { ... },
  "created_by": "doctor_uid_123",
  "created_at": "2025-12-12T10:30:00Z"
}
```

#### Error Responses

- `404 Not Found`: ACP record not found

---

### 3. List ACP Records

**GET** `/api/v1/patients/{patient_id}/acp-records`

Lists all ACP records for a patient with optional filtering.

#### Query Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `status` | string | Filter by status | `?status=active` |
| `decision_maker` | string | Filter by decision maker | `?decision_maker=patient` |
| `recorded_from` | date | Filter from date (YYYY-MM-DD) | `?recorded_from=2025-01-01` |
| `recorded_to` | date | Filter to date (YYYY-MM-DD) | `?recorded_to=2025-12-31` |
| `limit` | int | Max records to return (default: 100) | `?limit=10` |
| `offset` | int | Skip records (default: 0) | `?offset=20` |

#### Example Request

```
GET /api/v1/patients/patient_123/acp-records?status=active&limit=10
```

#### Response (200 OK)

```json
[
  {
    "acp_id": "acp_uuid_789",
    "patient_id": "patient_uuid_123",
    "recorded_date": "2025-12-12T00:00:00Z",
    "version": 2,
    "status": "active",
    "decision_maker": "patient",
    "directives": { ... },
    "created_by": "doctor_uid_123",
    "created_at": "2025-12-12T10:30:00Z"
  },
  {
    "acp_id": "acp_uuid_456",
    "patient_id": "patient_uuid_123",
    "recorded_date": "2025-11-01T00:00:00Z",
    "version": 1,
    "status": "superseded",
    "decision_maker": "patient",
    "directives": { ... },
    "created_by": "doctor_uid_123",
    "created_at": "2025-11-01T14:20:00Z"
  }
]
```

---

### 4. Get Latest Active ACP

**GET** `/api/v1/patients/{patient_id}/acp-records/latest`

Retrieves the latest active ACP record for a patient.

#### Response (200 OK)

```json
{
  "acp_id": "acp_uuid_789",
  "patient_id": "patient_uuid_123",
  "recorded_date": "2025-12-12T00:00:00Z",
  "version": 2,
  "status": "active",
  "decision_maker": "patient",
  "directives": {
    "dnar": true,
    "ventilator": false,
    "feeding_tube": false,
    "palliative_care_only": true
  },
  "created_by": "doctor_uid_123",
  "created_at": "2025-12-12T10:30:00Z"
}
```

#### Error Responses

- `404 Not Found`: Patient not found or no active ACP record exists

---

### 5. Get ACP History

**GET** `/api/v1/patients/{patient_id}/acp-records/history`

Retrieves the complete version history of all ACP records for a patient.

#### Response (200 OK)

```json
[
  {
    "acp_id": "acp_uuid_789",
    "version": 2,
    "status": "active",
    "recorded_date": "2025-12-12T00:00:00Z",
    "decision_maker": "patient",
    "directives": { ... },
    "created_by": "doctor_uid_123",
    "created_at": "2025-12-12T10:30:00Z"
  },
  {
    "acp_id": "acp_uuid_456",
    "version": 1,
    "status": "superseded",
    "recorded_date": "2025-11-01T00:00:00Z",
    "decision_maker": "patient",
    "directives": { ... },
    "created_by": "doctor_uid_123",
    "created_at": "2025-11-01T14:20:00Z"
  }
]
```

---

### 6. Update ACP Record

**PUT** `/api/v1/patients/{patient_id}/acp-records/{id}`

Updates an existing ACP record.

#### Request Body

All fields are optional. Only provided fields will be updated.

```json
{
  "status": "superseded",
  "directives": {
    "dnar": true,
    "ventilator": true,
    "feeding_tube": false,
    "palliative_care_only": false
  },
  "values_narrative": "患者の意向が変更。短期的な人工呼吸は許可。",
  "discussion_log": [
    {
      "date": "2025-12-15",
      "participants": ["主治医", "患者", "配偶者"],
      "summary": "意向の再確認と変更"
    }
  ]
}
```

#### Response (200 OK)

```json
{
  "acp_id": "acp_uuid_789",
  "patient_id": "patient_uuid_123",
  "recorded_date": "2025-12-12T00:00:00Z",
  "version": 1,
  "status": "superseded",
  "decision_maker": "patient",
  "directives": {
    "dnar": true,
    "ventilator": true,
    "feeding_tube": false,
    "palliative_care_only": false
  },
  "values_narrative": "患者の意向が変更。短期的な人工呼吸は許可。",
  "created_by": "doctor_uid_123",
  "created_at": "2025-12-12T10:30:00Z"
}
```

#### Error Responses

- `404 Not Found`: ACP record not found
- `400 Bad Request`: Validation error

---

### 7. Delete ACP Record

**DELETE** `/api/v1/patients/{patient_id}/acp-records/{id}`

Deletes an ACP record.

#### Response (204 No Content)

No response body.

#### Error Responses

- `404 Not Found`: ACP record not found

---

## Data Models

### ACPRecord

| Field | Type | Description |
|-------|------|-------------|
| `acp_id` | string | Unique identifier for the ACP record |
| `patient_id` | string | Patient UUID |
| `recorded_date` | date | Date when ACP was recorded |
| `version` | int | Version number for change tracking |
| `status` | string | `draft`, `active`, or `superseded` |
| `decision_maker` | string | `patient`, `proxy`, or `guardian` |
| `proxy_person_id` | string | UUID of proxy (if applicable) |
| `directives` | object | Medical directives (DNAR, ventilator, etc.) |
| `values_narrative` | string | Patient's values and preferences narrative |
| `legal_documents` | object | Links to legal documents |
| `discussion_log` | array | History of ACP discussions |
| `data_sensitivity` | string | Sensitivity classification |
| `access_restricted_to` | array | List of user IDs with access |
| `created_by` | string | User ID who created the record |
| `created_at` | timestamp | Creation timestamp |

### Directives Object Structure

```json
{
  "dnar": true,                      // Do Not Attempt Resuscitation
  "ventilator": false,               // Mechanical ventilation
  "feeding_tube": false,             // Artificial nutrition/hydration
  "palliative_care_only": true,      // Palliative care preference
  "hospitalization": "avoid_if_possible", // Hospital admission preference
  "cpr": false,                      // Cardiopulmonary resuscitation
  "dialysis": false,                 // Dialysis preference
  "blood_transfusion": true,         // Blood transfusion preference
  "antibiotics": "comfort_only"      // Antibiotic use preference
}
```

### Discussion Log Entry

```json
{
  "date": "2025-12-10",
  "participants": ["主治医", "患者", "家族"],
  "location": "患者宅",
  "duration_minutes": 45,
  "summary": "終末期医療についての意向確認",
  "decisions": ["DNAR同意", "緩和ケア優先"],
  "patient_understanding": "良好",
  "family_consensus": true,
  "follow_up_required": false
}
```

---

## Error Responses

### 400 Bad Request

```json
{
  "error": "invalid status: unknown_status"
}
```

### 404 Not Found

```json
{
  "error": "ACP record not found"
}
```

### 500 Internal Server Error

```json
{
  "error": "Failed to retrieve ACP records"
}
```

---

## Security Notes

1. **Authentication**: All endpoints require valid Firebase JWT token
2. **Authorization**: Access logged in audit trail
3. **Data Sensitivity**: Default `highly_confidential` classification
4. **Access Control**: Use `access_restricted_to` for granular permissions
5. **Audit Trail**: All access logged with user ID, timestamp, and action

---

## Best Practices

### Creating ACP Records

1. Always involve the patient (or proxy) in discussions
2. Document all discussion participants in `discussion_log`
3. Use clear, specific language in `values_narrative`
4. Link legal documents when available
5. Set status to `draft` until finalized

### Updating ACP Records

1. Mark previous version as `superseded` when creating new version
2. Document reason for changes in `discussion_log`
3. Notify care team of updates
4. Ensure patient/proxy consent for changes

### Version Management

1. Never delete historical ACP records
2. Use `superseded` status for outdated records
3. Maintain complete audit trail
4. Use `GetACPHistory` to review changes over time

---

## Integration Examples

### Flutter/Dart (Mobile App)

```dart
// Create ACP record
final response = await http.post(
  Uri.parse('$baseUrl/patients/$patientId/acp-records'),
  headers: {
    'Authorization': 'Bearer $firebaseToken',
    'Content-Type': 'application/json',
  },
  body: jsonEncode({
    'recorded_date': DateTime.now().toIso8601String(),
    'status': 'active',
    'decision_maker': 'patient',
    'directives': {
      'dnar': true,
      'ventilator': false,
    },
    'created_by': currentUserId,
  }),
);
```

### JavaScript/TypeScript (Web)

```typescript
// Get latest ACP
const getLatestACP = async (patientId: string) => {
  const response = await fetch(
    `/api/v1/patients/${patientId}/acp-records/latest`,
    {
      headers: {
        'Authorization': `Bearer ${firebaseToken}`,
      },
    }
  );
  return await response.json();
};
```

---

## Testing

Use the following test cases:

1. ✅ Create ACP with all required fields
2. ✅ Create ACP with proxy decision maker
3. ✅ Validate proxy_person_id requirement
4. ✅ Update ACP status to superseded
5. ✅ Get latest active ACP
6. ✅ Get complete ACP history
7. ✅ Filter by status
8. ✅ Filter by date range
9. ✅ Validate invalid status
10. ✅ Handle non-existent patient

---

**Last Updated**: 2025-12-12
**API Version**: v1
