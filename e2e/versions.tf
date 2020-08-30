terraform {
  required_version = ">= 0.13"

  required_providers {
    null = {
      source  = "hashicorp/null"
      version = "2.1.2"
    }
    local = {
      source  = "hashicorp/local"
      version = "1.4.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "2.2.1"
    }
    flexkube = {
      source  = "flexkube-testing/flexkube"
      version = "0.1.0"
    }
  }
}
