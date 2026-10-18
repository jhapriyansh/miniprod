terraform{
	required_providers {
		incus = {
			source = "lxc/incus"
			version = "~> 1.0"
		}
	}
}

provider "incus" {
	config_dir = "/Users/priyanshukumarjha/Library/Application Support/incus"
	generate_client_certificates = false
	accept_remote_certificate = false
	default_remote = "dell"
}

locals {
	vms = {
		cp = { ip = "10.95.18.210", memory = "1536MiB", cpu = "2" }
		w1 = { ip = "10.95.18.211", memory = "1024MiB", cpu = "1" }
		w2 = { ip = "10.95.18.212", memory = "1024MiB", cpu = "1" }
		w3 = { ip = "10.95.18.213", memory = "1024MiB", cpu = "1" }
	}
}

resource "incus_instance" "vm" {
	for_each = local.vms

	remote = "dell"
	name = each.key
	project = "iac"
	image = "images:debian/12/cloud"
	type = "virtual-machine"
	profiles = ["default"]

	device {
		name = "eth0"
		type = "nic"
		properties = {
			network = "incusbr0"
			"ipv4.address" = each.value.ip
		}
	}	

	config = {
		"limits.cpu" = each.value.cpu
		"limits.memory" = each.value.memory
		"cloud-init.user-data" = file("${path.module}/cloud-init.yaml")
	}
}

output "vm_ips" {
	value = local.vms
}
