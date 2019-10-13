# Required variables
variable "hcloud_token" {
  type        = string
  description = "Token required for accessing hcloud API."
}

variable "domain" {
  type        = string
  description = "Name of domain for cluster nodes."
}

variable "management_cidrs" {
  type        = list(string)
  description = "CIDRs, which are allowed to SSH to nodes."
}

variable "management_ssh_keys" {
  type        = list(string)
  description = "SSH keys which will be allowed to connect to the nodes. Should be specified in following format: '<key type> <key> <key name>'."
}

# Optional variables
variable "environment" {
  type        = string
  default     = "testing"
  description = "Name of the environment to run."
}

variable "controller_nodes_count" {
  default     = 1
  description = "Number of controller nodes to deploy."
}

variable "hcloud_server_location" {
  type        = string
  default     = "fsn1"
  description = "Valid hcloud location."
}

variable "hcloud_server_image" {
  type        = string
  default     = "ubuntu-18.04"
  description = "Initial hcloud server image. Will be replaced by Flatcar Edge anyway."
}

variable "controller_nodes_type" {
  type        = string
  default     = "cx11"
  description = "hcloud machine type for controller nodes."
}

variable "ssh_port" {
  default     = "22"
  description = "SSH port."
}

variable "core_user_password" {
  type        = string
  default     = ""
  description = "Core user password for console login."
}

variable "wireguard_port" {
  default     = "51820"
  description = "Port for running wireguard."
}

variable "wireguard_cidr" {
  default     = "10.44.0.0/16"
  description = "CIDR for wireguard network. It will run etcd."
}

variable "pods_cidr" {
  default     = "10.42.0.0/16"
  description = "CIDR for pods."
}

variable "hetzner_cidr" {
  default     = "10.40.0.0/16"
  description = "CIDR for Hetzner private network."
}
