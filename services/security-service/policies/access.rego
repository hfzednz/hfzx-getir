package nexora.access

default allow = false

allow {
  input.action == "read"
}

allow {
  input.action == "write"
  input.identityTrust >= 0.7
}
