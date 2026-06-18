variable "grafana_url" {
  type        = string
  description = "Base URL of the Grafana instance that hosts the console SLO dashboards and alerts."
  default     = "http://localhost:3000"
}

variable "grafana_auth" {
  type        = string
  description = "Grafana service-account token used to manage dashboards and alert rules."
  sensitive   = true
  default     = ""
}

variable "prometheus_datasource_uid" {
  type        = string
  description = "UID of the Prometheus datasource backing the dashboards and alert rules."
  default     = "prometheus"
}

variable "namespace" {
  type        = string
  description = "Kubernetes namespace the console service is deployed into."
  default     = "heliosnet"
}

variable "image" {
  type        = string
  description = "Container image reference for the console backend."
  default     = "registry.internal/heliosnet/groundstation-console:0.1.0"
}

variable "replicas" {
  type        = number
  description = "Number of console backend replicas."
  default     = 2
}

variable "availability_slo" {
  type        = number
  description = "Target availability for the console API as a fraction (for example 0.995)."
  default     = 0.995

  validation {
    condition     = var.availability_slo > 0 && var.availability_slo < 1
    error_message = "availability_slo must be a fraction between 0 and 1."
  }
}

variable "latency_slo_ms" {
  type        = number
  description = "Target p99 latency for the fleet-summary read path, in milliseconds."
  default     = 50
}

variable "notification_contact_point" {
  type        = string
  description = "Name of the Grafana contact point alert rules notify."
  default     = "heliosnet-ops"
}
