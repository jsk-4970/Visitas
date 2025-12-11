package models

import "time"

type Patient struct {
	PatientID         string    `json:"patient_id" spanner:"patient_id"`
	NameLast          string    `json:"name_last" spanner:"name_last"`
	NameFirst         string    `json:"name_first" spanner:"name_first"`
	BirthDate         string    `json:"birth_date" spanner:"birth_date"` // YYYY-MM-DD format
	Gender            string    `json:"gender" spanner:"gender"`
	PostalCode        string    `json:"postal_code" spanner:"postal_code"`
	AddressPrefecture string    `json:"address_prefecture" spanner:"address_prefecture"`
	AddressCity       string    `json:"address_city" spanner:"address_city"`
	AddressStreet     string    `json:"address_street" spanner:"address_street"`
	AddressBuilding   string    `json:"address_building,omitempty" spanner:"address_building"`
	Phone             string    `json:"phone" spanner:"phone"`
	EmergencyContact  string    `json:"emergency_contact,omitempty" spanner:"emergency_contact"`
	InsuranceNumber   string    `json:"insurance_number,omitempty" spanner:"insurance_number"`
	InsuranceSymbol   string    `json:"insurance_symbol,omitempty" spanner:"insurance_symbol"`
	CopayRate         int       `json:"copay_rate,omitempty" spanner:"copay_rate"`
	PrimaryDiagnosis  string    `json:"primary_diagnosis,omitempty" spanner:"primary_diagnosis"`
	Allergies         string    `json:"allergies,omitempty" spanner:"allergies"`
	Notes             string    `json:"notes,omitempty" spanner:"notes"`
	Deleted           bool      `json:"deleted" spanner:"deleted"`
	CreatedAt         time.Time `json:"created_at" spanner:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" spanner:"updated_at"`
}

type PatientCreateRequest struct {
	NameLast          string `json:"name_last" validate:"required"`
	NameFirst         string `json:"name_first" validate:"required"`
	BirthDate         string `json:"birth_date" validate:"required"`
	Gender            string `json:"gender"`
	PostalCode        string `json:"postal_code"`
	AddressPrefecture string `json:"address_prefecture"`
	AddressCity       string `json:"address_city"`
	AddressStreet     string `json:"address_street"`
	AddressBuilding   string `json:"address_building"`
	Phone             string `json:"phone" validate:"required"`
	EmergencyContact  string `json:"emergency_contact"`
	InsuranceNumber   string `json:"insurance_number"`
	InsuranceSymbol   string `json:"insurance_symbol"`
	CopayRate         int    `json:"copay_rate"`
	PrimaryDiagnosis  string `json:"primary_diagnosis"`
	Allergies         string `json:"allergies"`
	Notes             string `json:"notes"`
}

type PatientUpdateRequest struct {
	NameLast          string `json:"name_last"`
	NameFirst         string `json:"name_first"`
	BirthDate         string `json:"birth_date"`
	Gender            string `json:"gender"`
	PostalCode        string `json:"postal_code"`
	AddressPrefecture string `json:"address_prefecture"`
	AddressCity       string `json:"address_city"`
	AddressStreet     string `json:"address_street"`
	AddressBuilding   string `json:"address_building"`
	Phone             string `json:"phone"`
	EmergencyContact  string `json:"emergency_contact"`
	InsuranceNumber   string `json:"insurance_number"`
	InsuranceSymbol   string `json:"insurance_symbol"`
	CopayRate         int    `json:"copay_rate"`
	PrimaryDiagnosis  string `json:"primary_diagnosis"`
	Allergies         string `json:"allergies"`
	Notes             string `json:"notes"`
}

type PatientListResponse struct {
	Patients   []Patient `json:"patients"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}
