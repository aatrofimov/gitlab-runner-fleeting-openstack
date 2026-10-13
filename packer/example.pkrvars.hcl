# Example variables for the image bake. Copy to vars.pkrvars.hcl and fill in
# real values (vars.pkrvars.hcl is gitignored).
# Usage: packer build -var-file=vars.pkrvars.hcl .

source_image_id      = "00000000-0000-0000-0000-000000000000" # Debian genericcloud image UUID
flavor               = "<flavor>"                             # HFL2 flavor name or ID (local disk)
network_id           = "00000000-0000-0000-0000-000000000000" # private network UUID
security_group       = "default"                              # SG that allows SSH
external_net         = "<external-net>"                       # external network for the FIP
availability_zone    = ""                                     # AZ of the private network (optional)
ssh_private_key_file = ""                                     # path to private key of 'packer-builder'
