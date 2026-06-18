# SLO dashboards and alert rules for the console, managed as code so the
# observability surface ships with the service. The dashboard JSON lives in
# deploy/grafana and is also usable by importing it directly into Grafana.

resource "grafana_folder" "console" {
  title = "Heliosnet / groundstation-console"
}

resource "grafana_dashboard" "slo" {
  folder      = grafana_folder.console.id
  config_json = file("${path.module}/../grafana/console-slo-dashboard.json")
}

# Alert rule group: availability and latency SLO burn for the console API.
resource "grafana_rule_group" "console_slo" {
  name             = "console-slo"
  folder_uid       = grafana_folder.console.uid
  interval_seconds = 60

  rule {
    name      = "console-availability-burn"
    condition = "C"

    data {
      ref_id         = "A"
      datasource_uid = var.prometheus_datasource_uid
      relative_time_range {
        from = 3600
        to   = 0
      }
      model = jsonencode({
        refId = "A"
        expr  = "sum(rate(console_requests_total{status=~\"5..\"}[5m])) / sum(rate(console_requests_total[5m]))"
      })
    }

    data {
      ref_id         = "C"
      datasource_uid = "__expr__"
      relative_time_range {
        from = 3600
        to   = 0
      }
      model = jsonencode({
        refId      = "C"
        type       = "threshold"
        expression = "A"
        conditions = [{
          evaluator = {
            type   = "gt"
            params = [local.error_budget_pct / 100]
          }
        }]
      })
    }

    for            = "5m"
    no_data_state  = "OK"
    exec_err_state = "Alerting"

    annotations = {
      summary = "Console API error ratio is burning the availability SLO budget."
    }
    labels = {
      severity = "critical"
      service  = local.app_name
    }
  }

  rule {
    name      = "console-latency-slo"
    condition = "C"

    data {
      ref_id         = "A"
      datasource_uid = var.prometheus_datasource_uid
      relative_time_range {
        from = 3600
        to   = 0
      }
      model = jsonencode({
        refId = "A"
        expr  = "histogram_quantile(0.99, sum(rate(console_request_duration_seconds_bucket{route=\"/api/fleet\"}[5m])) by (le)) * 1000"
      })
    }

    data {
      ref_id         = "C"
      datasource_uid = "__expr__"
      relative_time_range {
        from = 3600
        to   = 0
      }
      model = jsonencode({
        refId      = "C"
        type       = "threshold"
        expression = "A"
        conditions = [{
          evaluator = {
            type   = "gt"
            params = [var.latency_slo_ms]
          }
        }]
      })
    }

    for            = "10m"
    no_data_state  = "OK"
    exec_err_state = "Alerting"

    annotations = {
      summary = "Fleet-summary p99 latency is above the SLO target."
    }
    labels = {
      severity = "warning"
      service  = local.app_name
    }
  }
}
