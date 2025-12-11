-- 訪問スケジュールテーブル
CREATE TABLE Visits (
  visit_id STRING(36) NOT NULL,
  patient_id STRING(36) NOT NULL,
  doctor_id STRING(36) NOT NULL,
  scheduled_at TIMESTAMP NOT NULL,
  duration_minutes INT64 NOT NULL DEFAULT (30),
  visit_type STRING(20) NOT NULL, -- 'regular', 'emergency', 'initial'
  status STRING(20) NOT NULL DEFAULT ('scheduled'), -- 'scheduled', 'in_progress', 'completed', 'canceled'
  notes STRING(1000),
  canceled_reason STRING(500),
  deleted BOOL NOT NULL DEFAULT (false),
  created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  updated_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  CONSTRAINT fk_visits_patient FOREIGN KEY (patient_id) REFERENCES Patients(patient_id),
  CONSTRAINT fk_visits_doctor FOREIGN KEY (doctor_id) REFERENCES Doctors(doctor_id),
) PRIMARY KEY (visit_id);

-- インデックス: 患者ごとの訪問履歴取得用
CREATE INDEX idx_visits_patient ON Visits(patient_id, scheduled_at DESC);

-- インデックス: 医師ごとのスケジュール取得用
CREATE INDEX idx_visits_doctor_schedule ON Visits(doctor_id, scheduled_at);

-- インデックス: 日付範囲検索用
CREATE INDEX idx_visits_date_range ON Visits(scheduled_at);
