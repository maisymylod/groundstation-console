output "service_endpoint" {
  description = "In-cluster DNS name of the console service."
  value       = "${kubernetes_service.console.metadata[0].name}.${var.namespace}.svc.cluster.local"
}

output "dashboard_uid" {
  description = "UID of the managed SLO dashboard."
  value       = grafana_dashboard.slo.uid
}

output "error_budget_pct" {
  description = "Allowed unavailability over the rolling window, derived from the SLO."
  value       = local.error_budget_pct
}
