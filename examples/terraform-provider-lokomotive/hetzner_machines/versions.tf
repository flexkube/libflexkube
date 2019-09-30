terraform {
  required_version = "= 0.12.9"
}

provider "hcloud" {
  version = "~> 1.12.0"
  token   = var.hcloud_token
}

provider "tls" {
  version = "= 2.1.0"
}

provider "null" {
  version = "~> 2.1.2"
}

provider "sshcommand" {
  version = "~> 0.1.2"
}

provider "ct" {
  version = "~> 0.3.2"
}

provider "wireguard" {
  version = "~> 0.1.0"
}

