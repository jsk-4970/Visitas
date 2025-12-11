-- 患者テーブル
CREATE TABLE Patients (
  patient_id STRING(36) NOT NULL,
  name_last STRING(50) NOT NULL,
  name_first STRING(50) NOT NULL,
  birth_date DATE NOT NULL,
  gender STRING(10),
  postal_code STRING(8),
  address_prefecture STRING(10),
  address_city STRING(50),
  address_street STRING(100),
  address_building STRING(100),
  phone STRING(15),
  emergency_contact STRING(15),
  insurance_number STRING(20),
  insurance_symbol STRING(20),
  copay_rate INT64,
  primary_diagnosis STRING(200),
  allergies STRING(500),
  notes STRING(2000),
  deleted BOOL NOT NULL DEFAULT (false),
  created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  updated_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (patient_id);

-- インデックス: 名前検索用
CREATE INDEX idx_patients_name ON Patients(name_last, name_first);

-- インデックス: 住所検索用
CREATE INDEX idx_patients_address ON Patients(address_prefecture, address_city);

-- インデックス: 作成日時ソート用
CREATE INDEX idx_patients_created_at ON Patients(created_at DESC);
