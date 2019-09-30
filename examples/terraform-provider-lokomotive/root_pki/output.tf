output "root_ca_cert" {
  value       = tls_self_signed_cert.root_ca.cert_pem
  description = "Root CA certificate."
}

output "root_ca_key" {
  value       = tls_private_key.root_ca.private_key_pem
  description = "Root CA private key."
  sensitive   = true
}

output "root_ca_algorithm" {
  value       = tls_private_key.root_ca.algorithm
  description = "Root CA key algorithm."
}
