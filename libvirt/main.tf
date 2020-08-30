provider "libvirt" {
  uri = "qemu:///system"
}

variable "core_public_keys" {
  type = list(string)
}

resource "libvirt_pool" "pool" {
  name = "libflexkube"
  type = "dir"
  path = "${abspath(path.module)}/pool"
}

resource "libvirt_volume" "base" {
  name   = "flexkube-base"
  source = "${abspath(path.module)}/flatcar_production_qemu_image.img"
  pool   = libvirt_pool.pool.name
  format = "qcow2"
}

resource "libvirt_network" "network" {
  name      = "flexkube"
  mode      = "nat"
  domain    = "k8s.local"
  addresses = [var.nodes_cidr]

  dns {
    local_only = true
    # can specify local names here
  }
}

data "ct_config" "controller" {
  count = var.controllers_count

  content = templatefile("./templates/ct_config.yaml.tmpl", {
    core_public_keys = var.core_public_keys
    hostname         = local.controller_names[count.index]
  })
}

resource "libvirt_ignition" "controller" {
  count = var.controllers_count

  name = "flexkube-ignition-controller-${count.index}"
  pool = libvirt_pool.pool.name

  content = data.ct_config.controller[count.index].rendered
}

resource "libvirt_volume" "controller-disk" {
  count = var.controllers_count

  name           = local.controller_names[count.index]
  base_volume_id = libvirt_volume.base.id
  pool           = libvirt_pool.pool.name
  format         = "qcow2"
}

resource "libvirt_domain" "controller_machine" {
  count  = var.controllers_count
  name   = local.controller_names[count.index]
  vcpu   = 2
  memory = 4096

  disk {
    volume_id = libvirt_volume.controller-disk[count.index].id
  }

  fw_cfg_name     = "opt/org.flatcar-linux/config"
  coreos_ignition = libvirt_ignition.controller[count.index].id

  graphics {
    listen_type = "address"
  }

  network_interface {
    network_id = libvirt_network.network.id
    hostname   = local.controller_names[count.index]
    addresses  = [local.controller_ips[count.index]]
  }
}

resource "libvirt_volume" "worker-disk" {
  count = var.workers_count

  name           = local.worker_names[count.index]
  base_volume_id = libvirt_volume.base.id
  pool           = libvirt_pool.pool.name
  format         = "qcow2"
}

data "ct_config" "worker" {
  count = var.workers_count

  content = templatefile("./templates/ct_config.yaml.tmpl", {
    core_public_keys = var.core_public_keys
    hostname         = local.worker_names[count.index]
  })
}

resource "libvirt_ignition" "worker" {
  count = var.workers_count

  name = "flexkube-ignition-worker-${count.index}"
  pool = libvirt_pool.pool.name

  content = data.ct_config.worker[count.index].rendered
}

resource "libvirt_domain" "worker_machine" {
  count  = var.workers_count
  name   = local.worker_names[count.index]
  vcpu   = 2
  memory = 2048

  disk {
    volume_id = libvirt_volume.worker-disk[count.index].id
  }

  fw_cfg_name     = "opt/org.flatcar-linux/config"
  coreos_ignition = libvirt_ignition.worker[count.index].id

  graphics {
    listen_type = "address"
  }

  network_interface {
    network_id = libvirt_network.network.id
    hostname   = local.worker_names[count.index]
    addresses  = [local.worker_ips[count.index]]
  }
}
