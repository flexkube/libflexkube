# Taken from https://raw.githubusercontent.com/poseidon/terraform-render-bootstrap/master/tls-etcd.tf
# =================== PROVIDERS ======================
provider "local" {
  version = "= 1.3.0"
}
provider "tls" {
  version = "= 2.1.0"
}

# =================== VARIABLES ======================
variable "asset_dir" {
  description = "Path to directory where generated certificates should be placed."
  type        = string
  default     = "pki"
}

variable "rsa_bits" {
  description = "Default number of RSA bits for certificates."
  type        = string
  default     = "1024"
}

variable "organization" {
  description = "Organization field for certificates."
  type        = string
  # TODO pick better default here
  default     = "TODO"
}

variable "peers" {
  description = "Map of etcd peers"
  default     = [
    {
      name = "bar"
      ip = "10.200.200.26"
    },
    {
      name = "baz"
      ip = "10.200.200.22"
    },
  ]
}

# =================== ROOT CA ======================
resource "tls_private_key" "root_ca" {
  algorithm = "RSA"
  rsa_bits  = var.rsa_bits
}

resource "tls_self_signed_cert" "root_ca" {
  key_algorithm   = tls_private_key.root_ca.algorithm
  private_key_pem = tls_private_key.root_ca.private_key_pem

  subject {
    common_name  = "root-ca"
    organization = var.organization
  }

  is_ca_certificate     = true
  # TODO make it configurable, root cert should be valid for a long time
  validity_period_hours = 8760

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "cert_signing",
  ]
}

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

  ca_key_algorithm   = tls_self_signed_cert.root_ca.key_algorithm
  ca_private_key_pem = tls_private_key.root_ca.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.root_ca.cert_pem

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
  count = length(var.peers)

  algorithm = "RSA"
  rsa_bits  = var.rsa_bits
}

resource "tls_cert_request" "peer" {
  count = length(var.peers)

  key_algorithm   = tls_private_key.peer[count.index].algorithm
  private_key_pem = tls_private_key.peer[count.index].private_key_pem

  subject {
    common_name  = "etcd-${var.peers[count.index].name}"
    organization = var.organization
  }

  ip_addresses = [
    "127.0.0.1",
    var.peers[count.index].ip,
  ]

  dns_names = [
    "localhost"
    # TODO add FQDN for peer IP?
  ]
}

resource "tls_locally_signed_cert" "peer" {
  count = length(var.peers)

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

# Output to local files
# TODO make it outputs so this code can be used as module
resource "local_file" "etcd_ca_crt" {
  content  = tls_locally_signed_cert.etcd_ca.cert_pem
  filename = "${var.asset_dir}/etcd-ca.crt"
}

resource "local_file" "etcd_peer_crt" {
  count = length(var.peers)

  content   = tls_locally_signed_cert.peer[count.index].cert_pem
  filename  = "${var.asset_dir}/${var.peers[count.index].name}.crt"
}

resource "local_file" "etcd_peer_key" {
  count = length(var.peers)

  sensitive_content = tls_private_key.peer[count.index].private_key_pem
  filename  = "${var.asset_dir}/${var.peers[count.index].name}.key"
}
