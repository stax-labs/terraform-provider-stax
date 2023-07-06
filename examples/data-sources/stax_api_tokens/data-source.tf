variable "api_token_id" {
  description = "the user identifier used to filter api tokens list"
}

data "stax_api_tokens" "stax_demo" {
  id = var.api_token_id
  # filters = {
  #   ids = [var.api_token_id]
  # }
}
