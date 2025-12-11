variable "project_id" {
  description = "GCP Project ID"
  type        = string
  default     = "visitas-dev"
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "asia-northeast1" # 東京
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "spanner_instance_name" {
  description = "Cloud Spanner instance name"
  type        = string
  default     = "visitas-dev-instance"
}

variable "spanner_database_name" {
  description = "Cloud Spanner database name"
  type        = string
  default     = "visitas-dev-db"
}

variable "spanner_processing_units" {
  description = "Cloud Spanner processing units (100 PU = 1 node)"
  type        = number
  default     = 100 # 開発環境は最小構成
}
