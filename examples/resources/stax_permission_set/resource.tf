resource "stax_permission_set" "data-scientist" {
  name                 = "data-scientist"
  max_session_duration = 28800
  description          = "Data Scientist Role. "
  aws_managed_policy_arns = [
    "arn:aws:iam::aws:policy/job-function/DataScientist",
    "arn:aws:iam::aws:policy/AWSBillingReadOnlyAccess"
  ]
  tags = {
    "owner" : "stax-demo@stax.io"
  }
}
