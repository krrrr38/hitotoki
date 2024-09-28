resource "google_artifact_registry_repository" "this" {
  project       = var.project_id
  location      = var.region
  repository_id = "hitotoki"
  format        = "DOCKER"
}
