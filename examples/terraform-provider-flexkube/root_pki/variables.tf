# Optional variables
variable "rsa_bits" {
  description = "Default number of RSA bits for certificates."
  type        = string
  default     = "4096"
}

variable "organization" {
  description = "Organization field for certificates."
  type        = string
  # TODO pick better default here
  default     = "TODO"
}

variable "validity_period_hours" {
  description = "Validity of Root CA in hours"
  type        = string
  default     = 8760
}
