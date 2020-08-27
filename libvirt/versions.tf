terraform {
  required_version = ">= 0.13"

  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "2.1.2"
    }
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.2"
    }
    ct = {
      source  = "poseidon/ct"
      version = "0.6.1"
    }
  }
}
