terraform {
  required_version = ">= 0.13"

  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "3.0.0"
    }
    libvirt = {
      source  = "invidian/libvirt"
      version = "0.6.10-rc1"
    }
    ct = {
      source  = "poseidon/ct"
      version = "0.7.0"
    }
  }
}
