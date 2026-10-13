# Source image: the UUID of a current Debian genericcloud image in Glance.
variable "source_image_id" {
  type        = string
  description = "UUID of the Debian genericcloud source image"
}

variable "flavor" {
  type        = string
  description = "Nova flavor used while baking the image"
}

variable "network_id" {
  type        = string
  description = "Neutron network UUID to attach the build server to"
}

variable "security_group" {
  type        = string
  description = "Security group name applied to the build server"
}

variable "external_net" {
  type        = string
  description = "External network used to allocate the provisioning floating IP"
}

variable "availability_zone" {
  type        = string
  default     = ""
  description = "Availability zone of the private network (empty = cloud default)"
}

variable "image_name" {
  type        = string
  default     = "gitlab-builder"
  description = "Name of the baked image"
}

variable "ssh_username" {
  type        = string
  default     = "debian"
  description = "SSH user of the source image"
}
