# Access Control Implementation Status

## 🎯 Overview
This document tracks the implementation of access control and structured logging for the 5 new Phase 1 domains as identified in the consistency verification report.

**Date**: 2025-12-12
**Status**: Service Layer Complete ✅ | Handler Layer In Progress ⏳

---

## ✅ Completed: Service Layer (5/5 Files)

All 5 service files have been successfully updated with:
- ✅ Access control using `patientRepo.CheckStaffAccess()`
- ✅ Structured logging with `logger.InfoContext`, `logger.WarnContext`, `logger.ErrorContext`
- ✅ Additional `createdBy`, `updatedBy`, `deletedBy`, `requestorID` parameters
- ✅ 403 Forbidden errors for unauthorized access
- ✅ Consistent error messages

### Updated Service Files:
1. **visit_schedule_service.go** (8 methods updated)
   - `CreateVisitSchedule` - added `createdBy` param
   - `GetVisitSchedule` - added `requestorID` param
   - `ListVisitSchedules` - added `requestorID` param
   - `UpdateVisitSchedule` - added `updatedBy` param
   - `DeleteVisitSchedule` - added `deletedBy` param
   - `GetUpcomingSchedules` - added `requestorID` param
   - `AssignStaff` - added `assignedBy` param
   - `AssignVehicle` - added `assignedBy` param
   - `UpdateStatus` - added `updatedBy` param

2. **clinical_observation_service.go** (7 methods updated)
   - `CreateClinicalObservation` - added `createdBy` param
   - `GetClinicalObservation` - added `requestorID` param
   - `ListClinicalObservations` - added `requestorID` param
   - `UpdateClinicalObservation` - added `updatedBy` param
   - `DeleteClinicalObservation` - added `deletedBy` param
   - `GetLatestObservationByCategory` - added `requestorID` param
   - `GetTimeSeriesData` - added `requestorID` param

3. **care_plan_service.go** (6 methods updated)
   - `CreateCarePlan` - added `createdBy` param
   - `GetCarePlan` - added `requestorID` param
   - `ListCarePlans` - added `requestorID` param
   - `UpdateCarePlan` - added `updatedBy` param
   - `DeleteCarePlan` - added `deletedBy` param
   - `GetActiveCarePlans` - added `requestorID` param

4. **medication_order_service.go** (6 methods updated)
   - `CreateMedicationOrder` - added `createdBy` param
   - `GetMedicationOrder` - added `requestorID` param
   - `ListMedicationOrders` - added `requestorID` param
   - `UpdateMedicationOrder` - added `updatedBy` param
   - `DeleteMedicationOrder` - added `deletedBy` param
   - `GetActiveOrders` - added `requestorID` param

5. **acp_record_service.go** (7 methods updated)
   - `CreateACPRecord` - added `createdBy` param
   - `GetACPRecord` - added `requestorID` param
   - `ListACPRecords` - added `requestorID` param
   - `UpdateACPRecord` - added `updatedBy` param
   - `DeleteACPRecord` - added `deletedBy` param
   - `GetLatestACP` - added `requestorID` param
   - `GetACPHistory` - added `requestorID` param

---

## ⏳ In Progress: Handler Layer (0/5 Files)

The following 5 handler files need to be updated to:
1. Get user ID from context using `middleware.GetUserIDFromContext(r.Context())`
2. Return 401 Unauthorized if user ID is not found
3. Pass user ID to all service method calls
4. Handle 403 Forbidden errors for access denied

### Handler Files to Update:

#### 1. visit_schedules.go
**Methods to update** (8 methods):
```go
// Example pattern for CreateVisitSchedule
func (h *VisitScheduleHandler) CreateVisitSchedule(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    patientID := chi.URLParam(r, "patient_id")

    // ADD: Get user ID from context
    userID, ok := middleware.GetUserIDFromContext(ctx)
    if !ok {
        respondError(w, http.StatusUnauthorized, "Unauthorized")
        return
    }

    var req models.VisitScheduleCreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.WarnContext(ctx, "Invalid request body", map[string]interface{}{
            "error": err.Error(),
        })
        respondError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // MODIFY: Pass userID as createdBy
    schedule, err := h.visitScheduleService.CreateVisitSchedule(ctx, patientID, &req, userID)
    if err != nil {
        // ADD: Handle access denied errors
        if strings.Contains(err.Error(), "access denied") {
            respondError(w, http.StatusForbidden, err.Error())
            return
        }
        logger.ErrorContext(ctx, "Failed to create visit schedule", err)
        respondError(w, http.StatusInternalServerError, "Failed to create visit schedule")
        return
    }

    respondJSON(w, http.StatusCreated, schedule)
}
```

**Methods**:
- `CreateVisitSchedule` - pass userID as `createdBy`
- `GetVisitSchedule` - pass userID as `requestorID`
- `GetVisitSchedules` (List) - pass userID as `requestorID`
- `UpdateVisitSchedule` - pass userID as `updatedBy`
- `DeleteVisitSchedule` - pass userID as `deletedBy`
- `GetUpcomingSchedules` - pass userID as `requestorID`
- `AssignStaff` - pass userID as `assignedBy`
- `AssignVehicle` - pass userID as `assignedBy`
- `UpdateScheduleStatus` - pass userID as `updatedBy`

#### 2. clinical_observations.go
**Methods to update** (7 methods):
- `CreateClinicalObservation` - pass userID as `createdBy`
- `GetClinicalObservation` - pass userID as `requestorID`
- `GetClinicalObservations` (List) - pass userID as `requestorID`
- `UpdateClinicalObservation` - pass userID as `updatedBy`
- `DeleteClinicalObservation` - pass userID as `deletedBy`
- `GetLatestObservation` - pass userID as `requestorID`
- `GetTimeSeriesObservations` - pass userID as `requestorID`

#### 3. care_plans.go
**Methods to update** (6 methods):
- `CreateCarePlan` - pass userID as `createdBy`
- `GetCarePlan` - pass userID as `requestorID`
- `GetCarePlans` (List) - pass userID as `requestorID`
- `UpdateCarePlan` - pass userID as `updatedBy`
- `DeleteCarePlan` - pass userID as `deletedBy`
- `GetActiveCarePlans` - pass userID as `requestorID`

#### 4. medication_orders.go
**Methods to update** (6 methods):
- `CreateMedicationOrder` - pass userID as `createdBy`
- `GetMedicationOrder` - pass userID as `requestorID`
- `GetMedicationOrders` (List) - pass userID as `requestorID`
- `UpdateMedicationOrder` - pass userID as `updatedBy`
- `DeleteMedicationOrder` - pass userID as `deletedBy`
- `GetActiveMedicationOrders` - pass userID as `requestorID`

#### 5. acp_records.go
**Methods to update** (7 methods):
- `CreateACPRecord` - pass userID as `createdBy`
- `GetACPRecord` - pass userID as `requestorID`
- `GetACPRecords` (List) - pass userID as `requestorID`
- `UpdateACPRecord` - pass userID as `updatedBy`
- `DeleteACPRecord` - pass userID as `deletedBy`
- `GetLatestACPRecord` - pass userID as `requestorID`
- `GetACPRecordHistory` - pass userID as `requestorID`

---

## 🔧 Implementation Pattern

### 1. Get User ID from Context
```go
userID, ok := middleware.GetUserIDFromContext(r.Context())
if !ok {
    respondError(w, http.StatusUnauthorized, "Unauthorized")
    return
}
```

### 2. Pass User ID to Service
```go
// Create operations
result, err := h.service.CreateXXX(ctx, patientID, &req, userID)

// Read operations
result, err := h.service.GetXXX(ctx, patientID, xxxID, userID)

// Update operations
result, err := h.service.UpdateXXX(ctx, patientID, xxxID, &req, userID)

// Delete operations
err := h.service.DeleteXXX(ctx, patientID, xxxID, userID)
```

### 3. Handle Access Denied Errors
```go
if err != nil {
    if strings.Contains(err.Error(), "access denied") {
        respondError(w, http.StatusForbidden, err.Error())
        return
    }
    // Other error handling...
}
```

---

## 📊 Estimated Impact

### Total Methods to Update: 34 methods across 5 handler files
- visit_schedules.go: 8 methods
- clinical_observations.go: 7 methods
- care_plans.go: 6 methods
- medication_orders.go: 6 methods
- acp_records.go: 7 methods

### Lines of Code Added: ~350-400 lines
- User ID extraction: ~5 lines per method
- Access control error handling: ~5 lines per method

---

## 🚀 Next Steps

1. ⏳ **Update Handler Layer** (Priority: HIGH)
   - Update all 5 handler files following the pattern above
   - Ensure middleware.GetUserIDFromContext is used consistently
   - Add proper error handling for 401/403 responses

2. 🧪 **Verify Build** (Priority: HIGH)
   ```bash
   cd backend && go build ./cmd/api
   ```

3. ✅ **Run Tests** (Priority: MEDIUM)
   ```bash
   cd backend && go test ./...
   ```

4. 📝 **Update OpenAPI Spec** (Priority: LOW)
   - Document security requirements in backend/openapi.yaml
   - Add 401/403 response codes to all endpoints

---

## 🎯 Success Criteria

- ✅ All service methods have access control checks
- ✅ All service methods use structured logging
- ⏳ All handler methods extract and pass user ID
- ⏳ All handler methods return 401 for missing auth
- ⏳ All handler methods return 403 for access denied
- ⏳ Build completes with zero errors
- ⏳ Consistency score improves from 73.8% to 95%+

---

## 📈 Progress Tracking

| Component | Status | Files | Methods | Complete |
|-----------|--------|-------|---------|----------|
| Service Layer | ✅ Done | 5/5 | 34/34 | 100% |
| Handler Layer | ⏳ In Progress | 0/5 | 0/34 | 0% |
| **Total** | **⏳ 50%** | **5/10** | **34/68** | **50%** |

---

## 🔍 References

- **Consistency Report**: `/CONSISTENCY_VERIFICATION_REPORT.md`
- **Completion Report**: `/PHASE1_COMPLETION_REPORT.md`
- **Example Implementation**: `internal/services/medical_condition_service.go` (existing correct pattern)
- **Example Handler**: `internal/handlers/medical_conditions.go` (existing correct pattern)
