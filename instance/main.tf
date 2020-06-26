// Copyright (c) 2018, Oracle and/or its affiliates. All rights reserved.

####################
# Subnet Datasource
####################
data "oci_core_subnet" "this" {
  count     = length(var.subnet_ocids)
  subnet_id = element(var.subnet_ocids, count.index)
}

############
# Instance
############
resource "oci_core_instance" "this" {
  count                = var.instance_count
  availability_domain  = var.compute_availability_domain_list[count.index % length(var.compute_availability_domain_list)]
  compartment_id       = var.compartment_ocid
  display_name         = var.instance_display_name == "" ? "" : var.instance_count != "1" ? "${var.instance_display_name}_${count.index + 1}" : var.instance_display_name
  extended_metadata    = var.extended_metadata
  ipxe_script          = var.ipxe_script
  preserve_boot_volume = var.preserve_boot_volume
  shape                = var.shape

  create_vnic_details {
    assign_public_ip = var.assign_public_ip
    display_name     = var.vnic_name == "" ? "" : var.instance_count != "1" ? "${var.vnic_name}_${count.index + 1}" : var.vnic_name
    hostname_label   = var.hostname_label == "" ? "" : var.instance_count != "1" ? "${var.hostname_label}-${count.index + 1}" : var.hostname_label
    private_ip = element(
      concat(var.private_ips, [""]),
      length(var.private_ips) == 0 ? 0 : count.index,
    )
    skip_source_dest_check = var.skip_source_dest_check
    subnet_id              = data.oci_core_subnet.this[count.index % length(data.oci_core_subnet.this.*.id)].id
  }

  metadata = {
    ssh_authorized_keys = file(var.ssh_authorized_keys)
    user_data           = var.user_data
  }

  source_details {
    boot_volume_size_in_gbs = var.boot_volume_size_in_gbs
    source_id               = var.source_ocid
    source_type             = var.source_type
  }

  timeouts {
    create = var.instance_timeout
  }
}

##################################
# Instance Credentials Datasource
##################################
data "oci_core_instance_credentials" "this" {
  count       = var.resource_platform != "linux" ? var.instance_count : 0
  instance_id = oci_core_instance.this[count.index].id
}

