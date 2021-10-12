terraform {
  required_version = ">= 0.13"

  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "3.0.0"
    }
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.11"
    }
    ct = {
      source  = "poseidon/ct"
      version = "0.7.0"
    }
  }
}
