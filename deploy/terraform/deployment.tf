# Kubernetes deployment and service for the console backend. The container reads
# its broker and fixture settings from the environment (see internal/config).

resource "kubernetes_namespace" "console" {
  metadata {
    name   = var.namespace
    labels = local.labels
  }
}

resource "kubernetes_deployment" "console" {
  metadata {
    name      = local.app_name
    namespace = kubernetes_namespace.console.metadata[0].name
    labels    = local.labels
  }

  spec {
    replicas = var.replicas

    selector {
      match_labels = {
        "app.kubernetes.io/name" = local.app_name
      }
    }

    template {
      metadata {
        labels = local.labels
        annotations = {
          "prometheus.io/scrape" = "true"
          "prometheus.io/port"   = "8080"
          "prometheus.io/path"   = "/api/health"
        }
      }

      spec {
        container {
          name  = "console"
          image = var.image

          port {
            name           = "http"
            container_port = 8080
          }

          env {
            name  = "CONSOLE_ADDR"
            value = ":8080"
          }

          env {
            name  = "TELEMETRY_TOPIC"
            value = "heliosnet.telemetry"
          }

          readiness_probe {
            http_get {
              path = "/api/health"
              port = "http"
            }
            initial_delay_seconds = 3
            period_seconds        = 10
          }

          liveness_probe {
            http_get {
              path = "/api/health"
              port = "http"
            }
            initial_delay_seconds = 10
            period_seconds        = 20
          }

          resources {
            requests = {
              cpu    = "100m"
              memory = "128Mi"
            }
            limits = {
              cpu    = "500m"
              memory = "256Mi"
            }
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "console" {
  metadata {
    name      = local.app_name
    namespace = kubernetes_namespace.console.metadata[0].name
    labels    = local.labels
  }

  spec {
    selector = {
      "app.kubernetes.io/name" = local.app_name
    }

    port {
      name        = "http"
      port        = 80
      target_port = "http"
    }
  }
}
