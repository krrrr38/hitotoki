locals {
  secret_access = {
    LINE_NOTIFY_TOKEN = data.google_secret_manager_secret.line_notify_token.secret_id
    STORAGE           = data.google_secret_manager_secret.storage.secret_id
  }
  secret_write = {
    STORAGE = data.google_secret_manager_secret.storage.secret_id
  }
}

resource "google_service_account" "this" {
  project    = var.project_id
  account_id = "hitotoki"
}

resource "google_secret_manager_secret_iam_member" "read" {
  for_each = local.secret_access

  project   = var.project_id
  secret_id = each.value
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.this.email}"
}

resource "google_secret_manager_secret_iam_member" "write" {
  for_each = local.secret_write

  project   = var.project_id
  secret_id = each.value
  role      = "roles/secretmanager.secretVersionManager"
  member    = "serviceAccount:${google_service_account.this.email}"
}
