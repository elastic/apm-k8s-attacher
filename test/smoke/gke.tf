provider "google" {
  project = var.gcp_project
  region  = var.gcp_region
  zone    = var.gcp_zone
}

# google_client_config and kubernetes provider must be explicitly specified like the following.
data "google_client_config" "default" {}

resource "google_compute_network" "kube_network" {
  name = "apm-attacher-smoke"
}

resource "google_compute_subnetwork" "kube_subnet" {
  name          = "apm-attacher-smoke-subnet"
  ip_cidr_range = "10.10.0.0/16"
  region        = var.gcp_region
  network       = google_compute_network.kube_network.id
  secondary_ip_range {
    ip_cidr_range = "10.11.0.0/16"
    range_name    = "pods"
  }
  secondary_ip_range {
    ip_cidr_range = "10.12.0.0/16"
    range_name    = "svc"
  }
}

module "gke" {
  source                     = "terraform-google-modules/kubernetes-engine/google"
  project_id                 = var.gcp_project
  name                       = "gke-smoketest-1"
  region                     = var.gcp_region
  zones                      = ["${var.gcp_region}-a", "${var.gcp_region}-b", "${var.gcp_region}-c"]
  network                    = google_compute_network.kube_network.name
  subnetwork                 = google_compute_subnetwork.kube_subnet.name
  ip_range_pods              = google_compute_subnetwork.kube_subnet.secondary_ip_range[0].range_name
  ip_range_services          = google_compute_subnetwork.kube_subnet.secondary_ip_range[1].range_name
  http_load_balancing        = false
  network_policy             = false
  horizontal_pod_autoscaling = true
  filestore_csi_driver       = false

  node_pools = [
    {
      name               = "default-node-pool"
      machine_type       = "e2-medium"
      node_locations     = "${var.gcp_region}-b,${var.gcp_region}-c"
      min_count          = 1
      max_count          = 2
      local_ssd_count    = 0
      spot               = false
      disk_size_gb       = 100
      disk_type          = "pd-standard"
      image_type         = "COS_CONTAINERD"
      enable_gcfs        = false
      enable_gvnic       = false
      auto_repair        = true
      auto_upgrade       = true
      preemptible        = false
      initial_node_count = 1
    },
  ]

  node_pools_oauth_scopes = {
    all = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }

  node_pools_labels = {
    all = {}
    default-node-pool = {
      default-node-pool = true
    }
  }

  node_pools_metadata = {
    all = {}
    default-node-pool = {
      node-pool-metadata-custom-value = "my-node-pool"
    }
  }

  node_pools_taints = {
    all = []
    default-node-pool = [
      {
        key    = "default-node-pool"
        value  = true
        effect = "PREFER_NO_SCHEDULE"
      },
    ]
  }

  node_pools_tags = {
    all = []
    default-node-pool = [
      "default-node-pool",
    ]
  }
}
