terraform {
  required_version = ">= 1.1.8, < 2.0.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">=4.29.0"
    }
    ec = {
      source  = "elastic/ec"
      version = ">=0.4.1"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">=2.6.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">=2.13.1"
    }
  }
}

provider "ec" {}

module "ec_deployment" {
  source                 = "github.com/elastic/apm-server//testing/infra/terraform/modules/ec_deployment?depth=1"
  deployment_name_prefix = "apm-attacher"
  stack_version          = "\\d.\\d.\\d?(\\d)$" // use the latest released version.
  region                 = "gcp-us-west2"
  deployment_template    = "gcp-compute-optimized"
  integrations_server    = true
}

resource "local_file" "chart_values" {
  content  = <<EOT
apm:
  secret_token: ${module.ec_deployment.apm_secret_token}
  namespaces:
    - default
webhookConfig:
  agents:
    nodejs:
      environment:
        ELASTIC_APM_SERVER_URL: "${module.ec_deployment.apm_url}"
        ELASTIC_APM_LOG_LEVEL: "info"
EOT
  filename = "custom.yaml"
}

provider "helm" {
  kubernetes {
    host                   = "https://${module.gke.endpoint}"
    token                  = data.google_client_config.default.access_token
    cluster_ca_certificate = base64decode(module.gke.ca_certificate)
  }
}

resource "helm_release" "apm_attacher" {
  name             = "apm-attacher"
  chart            = "../../charts/apm-attacher"
  namespace        = "elastic-apm"
  create_namespace = true
  values           = ["${local_file.chart_values.content}"]
}

provider "kubernetes" {
  host                   = "https://${module.gke.endpoint}"
  token                  = data.google_client_config.default.access_token
  cluster_ca_certificate = base64decode(module.gke.ca_certificate)
}

resource "kubernetes_deployment_v1" "nodejs-demo" {
  depends_on = [
    helm_release.apm_attacher // Install the apm-attacher chart first.
  ]
  metadata {
    name      = "nodejs-demo-app"
    namespace = "default"
  }

  spec {
    replicas = 1
    selector {
      match_labels = {
        "app" = "nodejs-demo"
      }
    }
    template {
      metadata {
        annotations = {
          "co.elastic.apm/attach" = "nodejs"
        }
        labels = {
          "app" = "nodejs-demo"
        }
      }
      spec {
        container {
          name  = "nodejs"
          image = "docker.elastic.co/observability/nodejs-hello-world:latest"
          port {
            container_port = 8080
          }
          env {
            name  = "ELASTIC_APM_SERVICE_NAME"
            value = "nodejs-demo-app"
          }
          liveness_probe {
            http_get {
              path = "/"
              port = 8080
            }
            initial_delay_seconds = 10
            period_seconds        = 5
          }
        }
      }
    }
  }
}
