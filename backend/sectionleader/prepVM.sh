#!/bin/bash
set -euo pipefail

if [ $# -ne 1 ]; then
  echo "Usage: $0 <base_path>"
  exit 1
fi

base=$1

# Generate ssh key without passphrase
ssh-keygen -f "${base}/id_rsa" -N "" -q

# Copy public key to authorized_keys inside squashfs-root directory
cp -v "${base}/id_rsa.pub" "${base}/squashfs-root/root/.ssh/authorized_keys"

# Set ownership of squashfs-root recursively to root:root
sudo chown -R root:root "${base}/squashfs-root"

# Create a 400M ext4 image file named after base argument
truncate -s 400M "${base}/fs.ext4"

# Format ext4 filesystem with squashfs-root as the directory content
sudo mkfs.ext4 -d "${base}/squashfs-root" -F "${base}/fs.ext4"

