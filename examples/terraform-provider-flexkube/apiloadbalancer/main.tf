provider "flexkube" {}

resource "flexkube_apiloadbalancer_pool" "apiloadbalancer" {
  config = templatefile("./config.yaml.tmpl", {
    metrics_bind_addresses = ["127.0.0.1"]
    servers = ["127.0.0.1"]
  })
}
