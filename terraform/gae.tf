resource "google_app_engine_application" "poster" {
  project     = var.gcloud_project_id
  location_id = "us-central"
}
