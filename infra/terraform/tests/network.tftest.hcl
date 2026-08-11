# Terraform contract test notes (run with terraform test when providers available).

run "network_subnets" {
  command = plan

  module {
    source = "./modules/network"
  }

  variables {
    name = "nexora-test"
    cidr = "10.40.0.0/16"
  }

  assert {
    condition     = length(output.public_subnets) == 3
    error_message = "expected 3 public subnets"
  }
}
