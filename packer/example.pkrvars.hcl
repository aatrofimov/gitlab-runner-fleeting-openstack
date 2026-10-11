# Example variables for the image bake. Copy to vars.pkrvars.hcl and fill in
# real values (vars.pkrvars.hcl is gitignored).
# Usage: packer build -var-file=vars.pkrvars.hcl .

source_image_id = "00000000-0000-0000-0000-000000000000" # Debian genericcloud image UUID
flavor          = "<flavor>"                             # flavor name or UUID
network_id      = "00000000-0000-0000-0000-000000000000" # private network UUID
security_group  = "default"                              # SG that allows SSH
external_net    = "<external-net>"                       # external network for the FIP
