resource "google_service_account" "scheduler" {
  project    = var.project_id
  account_id = "hitotoki-scheduler"
}

resource "google_project_iam_member" "scheduler" {
  for_each = toset([
    "roles/run.invoker",
  ])

  role    = each.key
  member  = "serviceAccount:${google_service_account.scheduler.email}"
  project = var.project_id
}


resource "google_cloud_scheduler_job" "this" {
  name             = "hitotoki"
  description      = "hitotoki"
  schedule         = "*/3 8-17 * * *"
  time_zone        = "Asia/Tokyo"
  attempt_deadline = "240s"

  http_target {
    http_method = "POST"
    uri         = "https://${var.region}-run.googleapis.com/apis/run.googleapis.com/v1/namespaces/${var.project_id}/jobs/${google_cloud_run_v2_job.this.name}:run"

    oauth_token {
      service_account_email = google_service_account.scheduler.email
    }
  }

  retry_config {
    retry_count = 3
  }
}
