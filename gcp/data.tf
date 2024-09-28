data "google_secret_manager_secret" "line_notify_token" {
  project   = var.project_id
  secret_id = "HITOTOKI_LINE_NOTIFY_TOKEN"
}

data "google_secret_manager_secret" "storage" {
  project   = var.project_id
  secret_id = "HITOTOKI_STORAGE"
}
