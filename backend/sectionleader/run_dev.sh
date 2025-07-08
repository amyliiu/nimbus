#!/bin/bash
rm -f server.log
sudo rm -rf /home/tswu/frpc/nimbus
sudo mkdir /home/tswu/frpc/nimbus
sudo rm -rf _data/
sudo rm -rf /var/lib/cni/networks
sudo rm -rf /etc/cni/conf.d
make
#sudo ./server-sectionleader