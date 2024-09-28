locals {
  cloudrun = {
    envs = {
      GCP_PROJECT_ID          = var.project_id
      GCP_SECRET_MANAGER_NAME = data.google_secret_manager_secret.storage.secret_id
    }
    secrets = {
      LINE_NOTIFY_TOKEN = data.google_secret_manager_secret.line_notify_token.secret_id
    }
  }
}

resource "google_cloud_run_v2_job" "this" {
  name     = "hitotoki"
  location = var.region

  template {
    task_count = 1
    template {
      timeout = "120s"
      containers {
        image = "${var.region}-docker.pkg.dev/${var.project_id}/hitotoki/hitotoki"

        dynamic "env" {
          for_each = local.cloudrun.envs
          content {
            name  = env.key
            value = env.value
          }
        }

        dynamic "env" {
          for_each = local.cloudrun.secrets
          content {
            name = env.key
            value_source {
              secret_key_ref {
                version = "latest"
                secret  = env.value
              }
            }
          }
        }
      }

      service_account = google_service_account.this.email
    }
  }

  lifecycle {
    ignore_changes = [
      client,
      client_version,
      # template[0].template[0].containers[0].image,
      template[0].labels["client.knative.dev/nonce"]
    ]
  }

  depends_on = [
    google_secret_manager_secret_iam_member.read,
  ]
}
