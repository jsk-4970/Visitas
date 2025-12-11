terraform {
  required_version = ">= 1.5"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    google-beta = {
      source  = "hashicorp/google-beta"
      version = "~> 5.0"
    }
  }

  # Terraform State をCloud Storageに保存（初回はローカルで実行してから設定）
  # backend "gcs" {
  #   bucket = "stunning-grin-480914-n1-terraform-state"
  #   prefix = "terraform/state"
  # }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

provider "google-beta" {
  project = var.project_id
  region  = var.region
}

# GCP Project
data "google_project" "project" {
  project_id = var.project_id
}

# Enable Required APIs
resource "google_project_service" "required_apis" {
  for_each = toset([
    "spanner.googleapis.com",
    "firestore.googleapis.com",
    "storage.googleapis.com",
    "run.googleapis.com",
    "cloudbuild.googleapis.com",
    "firebase.googleapis.com",
    "identitytoolkit.googleapis.com", # Firebase Authentication / Identity Platform
    "secretmanager.googleapis.com",   # Secret Manager for Firebase credentials
    "artifactregistry.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
    "iap.googleapis.com",
    "cloudkms.googleapis.com",
  ])

  service            = each.value
  disable_on_destroy = false
}

# Cloud Spanner Instance with CMEK encryption
resource "google_spanner_instance" "main" {
  name             = var.spanner_instance_name
  config           = "regional-${var.region}"
  display_name     = "Visitas ${var.environment} Instance"
  processing_units = var.spanner_processing_units

  # CMEK encryption for enhanced security (3省2ガイドライン準拠)
  encryption_config {
    kms_key_name = google_kms_crypto_key.spanner_key.id
  }

  labels = {
    environment = var.environment
    project     = "visitas"
  }

  depends_on = [
    google_project_service.required_apis,
    google_kms_crypto_key.spanner_key
  ]
}

# Cloud Spanner Database
resource "google_spanner_database" "main" {
  instance                 = google_spanner_instance.main.name
  name                     = var.spanner_database_name
  version_retention_period = "7d"
  deletion_protection      = var.environment == "prod" ? true : false

  # DDLはマイグレーションスクリプトで管理するため、ここでは空のまま
  ddl = []
}

# Cloud Storage Bucket for backups
resource "google_storage_bucket" "backups" {
  name          = "${var.project_id}-backups"
  location      = var.region
  force_destroy = var.environment != "prod"

  uniform_bucket_level_access = true

  versioning {
    enabled = true
  }

  lifecycle_rule {
    condition {
      age = 90
    }
    action {
      type = "Delete"
    }
  }

  labels = {
    environment = var.environment
    project     = "visitas"
  }
}

# Cloud Storage Bucket for medical images
resource "google_storage_bucket" "medical_images" {
  name          = "${var.project_id}-medical-images"
  location      = var.region
  force_destroy = var.environment != "prod"

  uniform_bucket_level_access = true

  # 開発環境ではデフォルト暗号化を使用（本番環境では CMEK を検討）
  # encryption {
  #   default_kms_key_name = google_kms_crypto_key.storage_key.id
  # }

  lifecycle_rule {
    condition {
      age = 2555 # ~7 years (医療記録保存期間)
    }
    action {
      type = "Delete"
    }
  }

  labels = {
    environment = var.environment
    project     = "visitas"
    data_type   = "phi" # Protected Health Information
  }

  depends_on = [google_project_service.required_apis]
}

# KMS Key Ring
resource "google_kms_key_ring" "main" {
  name     = "visitas-${var.environment}-keyring"
  location = var.region

  depends_on = [google_project_service.required_apis]
}

# KMS Crypto Key for Storage
resource "google_kms_crypto_key" "storage_key" {
  name            = "storage-encryption-key"
  key_ring        = google_kms_key_ring.main.id
  rotation_period = "7776000s" # 90 days

  lifecycle {
    prevent_destroy = true
  }
}

# KMS Crypto Key for Spanner (CMEK)
resource "google_kms_crypto_key" "spanner_key" {
  name            = "spanner-encryption-key"
  key_ring        = google_kms_key_ring.main.id
  rotation_period = "7776000s" # 90 days

  lifecycle {
    prevent_destroy = true
  }
}

# Grant Spanner service account permission to use the KMS key
resource "google_kms_crypto_key_iam_member" "spanner_key_user" {
  crypto_key_id = google_kms_crypto_key.spanner_key.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:service-${data.google_project.project.number}@gcp-sa-spanner.iam.gserviceaccount.com"
}

# Artifact Registry for Docker images
resource "google_artifact_registry_repository" "docker" {
  location      = var.region
  repository_id = "visitas-${var.environment}"
  description   = "Visitas ${var.environment} Docker repository"
  format        = "DOCKER"

  labels = {
    environment = var.environment
    project     = "visitas"
  }

  depends_on = [google_project_service.required_apis]
}

# Service Account for Cloud Run
resource "google_service_account" "cloud_run" {
  account_id   = "visitas-${var.environment}-run"
  display_name = "Visitas ${var.environment} Cloud Run Service Account"
  description  = "Service account for Visitas Cloud Run services"
}

# Grant Spanner Database User role
resource "google_spanner_database_iam_member" "cloud_run_spanner" {
  instance = google_spanner_instance.main.name
  database = google_spanner_database.main.name
  role     = "roles/spanner.databaseUser"
  member   = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Grant Storage Object Admin role (for medical images bucket)
resource "google_storage_bucket_iam_member" "cloud_run_storage" {
  bucket = google_storage_bucket.medical_images.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Firestore Database (Native Mode)
resource "google_firestore_database" "main" {
  project     = var.project_id
  name        = "(default)"
  location_id = var.region
  type        = "FIRESTORE_NATIVE"

  depends_on = [google_project_service.required_apis]
}

# Organization Policy: リージョン制限（日本国内のみ）
# Note: 組織レベルの権限が必要なため、個人プロジェクトではコメントアウト
# resource "google_project_organization_policy" "allowed_locations" {
#   project    = var.project_id
#   constraint = "constraints/gcp.resourceLocations"
#
#   list_policy {
#     allow {
#       values = [
#         "in:asia-locations", # Asia全般
#         "asia-northeast1",   # 東京
#         "asia-northeast2",   # 大阪
#       ]
#     }
#   }
#
#   depends_on = [google_project_service.required_apis]
# }

# Identity Platform Configuration (Firebase Authentication)
# Note: Firebase project must be manually set up in Firebase Console first
# This configuration enables Identity Platform API for the project

# Service Account for Firebase Admin SDK
resource "google_service_account" "firebase_admin" {
  account_id   = "firebase-admin-${var.environment}"
  display_name = "Firebase Admin SDK Service Account (${var.environment})"
  description  = "Service account for Firebase Admin SDK to manage authentication"
}

# Grant Firebase Admin role to service account
resource "google_project_iam_member" "firebase_admin" {
  project = var.project_id
  role    = "roles/firebase.admin"
  member  = "serviceAccount:${google_service_account.firebase_admin.email}"

  depends_on = [google_project_service.required_apis]
}

# Create service account key (for local development and Cloud Run)
resource "google_service_account_key" "firebase_admin_key" {
  service_account_id = google_service_account.firebase_admin.name
}

# Store the service account key in Secret Manager for secure access
resource "google_secret_manager_secret" "firebase_service_account" {
  secret_id = "firebase-service-account-${var.environment}"

  replication {
    user_managed {
      replicas {
        location = var.region
      }
    }
  }

  labels = {
    environment = var.environment
    project     = "visitas"
  }

  depends_on = [google_project_service.required_apis]
}

resource "google_secret_manager_secret_version" "firebase_service_account" {
  secret      = google_secret_manager_secret.firebase_service_account.id
  secret_data = base64decode(google_service_account_key.firebase_admin_key.private_key)
}

# Grant Cloud Run service account access to the secret
resource "google_secret_manager_secret_iam_member" "cloud_run_firebase_secret" {
  secret_id = google_secret_manager_secret.firebase_service_account.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Grant Cloud Run service account permission to use KMS for application-level encryption
resource "google_kms_crypto_key_iam_member" "cloud_run_kms_user" {
  crypto_key_id = google_kms_crypto_key.spanner_key.id
  role          = "roles/cloudkms.cryptoKeyEncrypterDecrypter"
  member        = "serviceAccount:${google_service_account.cloud_run.email}"
}

# Outputs
output "spanner_instance_name" {
  value       = google_spanner_instance.main.name
  description = "Cloud Spanner instance name"
}

output "spanner_database_name" {
  value       = google_spanner_database.main.name
  description = "Cloud Spanner database name"
}

output "backup_bucket_name" {
  value       = google_storage_bucket.backups.name
  description = "Backup bucket name"
}

output "medical_images_bucket_name" {
  value       = google_storage_bucket.medical_images.name
  description = "Medical images bucket name"
}

output "artifact_registry_url" {
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.docker.repository_id}"
  description = "Artifact Registry URL for Docker images"
}

output "cloud_run_service_account_email" {
  value       = google_service_account.cloud_run.email
  description = "Cloud Run service account email"
}

output "firebase_service_account_email" {
  value       = google_service_account.firebase_admin.email
  description = "Firebase Admin SDK service account email"
}

output "firebase_secret_name" {
  value       = google_secret_manager_secret.firebase_service_account.secret_id
  description = "Secret Manager secret name for Firebase service account key"
  sensitive   = true
}
