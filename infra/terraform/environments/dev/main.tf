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
    "artifactregistry.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "iam.googleapis.com",
    "iap.googleapis.com",
    "cloudarmor.googleapis.com",
  ])

  service            = each.value
  disable_on_destroy = false
}

# Cloud Spanner Instance
resource "google_spanner_instance" "main" {
  name             = var.spanner_instance_name
  config           = "regional-${var.region}"
  display_name     = "Visitas ${var.environment} Instance"
  processing_units = var.spanner_processing_units
  labels = {
    environment = var.environment
    project     = "visitas"
  }

  depends_on = [google_project_service.required_apis]
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

  encryption {
    default_kms_key_name = google_kms_crypto_key.storage_key.id
  }

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
resource "google_project_organization_policy" "allowed_locations" {
  project    = var.project_id
  constraint = "constraints/gcp.resourceLocations"

  list_policy {
    allow {
      values = [
        "in:asia-locations", # Asia全般
        "asia-northeast1",   # 東京
        "asia-northeast2",   # 大阪
      ]
    }
  }

  depends_on = [google_project_service.required_apis]
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
