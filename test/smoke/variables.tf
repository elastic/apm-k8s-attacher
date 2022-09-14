variable "gcp_project" {
  type        = string
  description = "GCP Project name"
  default     = "elastic-apm"
}

variable "gcp_region" {
  type        = string
  description = "GCP region"
  default     = "us-west2"
}

variable "gcp_zone" {
  type        = string
  description = "GCP zone"
  default     = "us-west2-b"
}
