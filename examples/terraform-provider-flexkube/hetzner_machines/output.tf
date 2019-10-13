# Output IP addresses, which should be allowed on firewall for wireguard traffic
output "node_private_ips" {
  value = hcloud_server_network.network.*.ip
}

output "node_public_ips" {
  value = hcloud_server.controller_nodes.*.ipv4_address
}

output "node_wireguard_ips" {
  value = null_resource.controller_wireguard_configuration.*.triggers.wireguard_ip
}

output "node_names" {
  value = hcloud_server.controller_nodes.*.name
}

output "provisioning_private_key" {
  description = "SSH Private key used for provisioning"
  value       = tls_private_key.terraform.private_key_pem
  sensitive   = true
}
