# =================== ETCD CA ======================
resource "tls_private_key" "etcd_ca" {
  algorithm = "RSA"
  rsa_bits  = var.rsa_bits
}

resource "tls_cert_request" "etcd_ca" {
  key_algorithm   = tls_private_key.etcd_ca.algorithm
  private_key_pem = tls_private_key.etcd_ca.private_key_pem

  subject {
    # This is CN recommended by K8s: https://kubernetes.io/docs/setup/best-practices/certificates/
    common_name  = "etcd-ca"
    organization = var.organization
  }
}

resource "tls_locally_signed_cert" "etcd_ca" {
  cert_request_pem = tls_cert_request.etcd_ca.cert_request_pem

  ca_key_algorithm   = var.root_ca_algorithm
  ca_private_key_pem = var.root_ca_key
  ca_cert_pem        = var.root_ca_cert

  is_ca_certificate     = true

  # TODO make it configurable
  validity_period_hours = 8760

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
  ]
}

# =================== PEER CERTS ======================
resource "tls_private_key" "peer" {
  count = length(var.peer_ips)

  algorithm = "RSA"
  rsa_bits  = var.rsa_bits
}

resource "tls_cert_request" "peer" {
  count = length(var.peer_ips)

  key_algorithm   = tls_private_key.peer[count.index].algorithm
  private_key_pem = tls_private_key.peer[count.index].private_key_pem

  subject {
    common_name  = "etcd-${var.peer_names[count.index]}"
    organization = var.organization
  }

  ip_addresses = [
    "127.0.0.1",
    var.peer_ips[count.index],
  ]

  dns_names = [
    "localhost",
    "etcd-${var.peer_names[count.index]}",
    # TODO add FQDN for peer IP?
  ]
}

resource "tls_locally_signed_cert" "peer" {
  count = length(var.peer_ips)

  cert_request_pem = tls_cert_request.peer[count.index].cert_request_pem

  ca_key_algorithm   = tls_cert_request.etcd_ca.key_algorithm
  ca_private_key_pem = tls_private_key.etcd_ca.private_key_pem
  ca_cert_pem        = tls_locally_signed_cert.etcd_ca.cert_pem

  # TODO Again, configurable
  validity_period_hours = 8760

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}
