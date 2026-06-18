provider "grafana" {
  url  = var.grafana_url
  auth = var.grafana_auth
}

# Uses the ambient kubeconfig (in-cluster config or KUBE_CONFIG_PATH). No
# credentials are committed; this module is validated with `terraform validate`,
# not applied against a live cluster in CI.
provider "kubernetes" {}
