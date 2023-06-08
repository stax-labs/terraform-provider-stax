resource "stax_account_type" "production" {
  name = "production"

  # create_before_destroy enables clean up account types in one pass when also removing refernces in stax_account resources
  lifecycle {
    create_before_destroy = true
  }
}