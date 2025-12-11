-- 医師テーブル
CREATE TABLE Doctors (
  doctor_id STRING(36) NOT NULL,
  name STRING(100) NOT NULL,
  email STRING(100) NOT NULL,
  phone STRING(15),
  specialization STRING(50),
  license_number STRING(20),
  deleted BOOL NOT NULL DEFAULT (false),
  created_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
  updated_at TIMESTAMP NOT NULL OPTIONS (allow_commit_timestamp=true),
) PRIMARY KEY (doctor_id);

-- インデックス: メールアドレス検索用（ユニーク制約として機能）
CREATE UNIQUE INDEX idx_doctors_email ON Doctors(email);
