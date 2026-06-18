locals {
  app_name = "groundstation-console"

  labels = {
    "app.kubernetes.io/name"      = local.app_name
    "app.kubernetes.io/part-of"   = "heliosnet"
    "app.kubernetes.io/component" = "ops-console"
  }

  # Error budget derived from the availability SLO, expressed as a percentage of
  # allowed unavailability over the rolling window.
  error_budget_pct = (1 - var.availability_slo) * 100
}
