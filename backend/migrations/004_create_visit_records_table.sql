-- 訪問記録テーブル
CREATE TABLE VisitRecords (
  record_id STRING(36) NOT NULL,
  visit_id STRING(36) NOT NULL,
  blood_pressure_systolic INT64,
  blood_pressure_diastolic INT64,
  pulse INT64,
  temperature FLOAT64,
  spo2 INT64,
  chief_complaint STRING(1000),
  findings STRING(2000),
  prescription STRING(1000),
  next_visit_date DATE,
  photos ARRAY<STRING(500)>, -- Cloud Storage URIs
  created_by STRING(36) NOT NULL, -- doctor_id
  created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  updated_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  CONSTRAINT fk_visit_records_visit FOREIGN KEY (visit_id) REFERENCES Visits(visit_id),
  CONSTRAINT fk_visit_records_doctor FOREIGN KEY (created_by) REFERENCES Doctors(doctor_id),
) PRIMARY KEY (record_id);

-- インデックス: 訪問IDごとの記録取得用
CREATE INDEX idx_visit_records_visit ON VisitRecords(visit_id);

-- インデックス: 作成日時ソート用
CREATE INDEX idx_visit_records_created_at ON VisitRecords(created_at DESC);
