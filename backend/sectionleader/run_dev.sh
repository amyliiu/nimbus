#!/bin/bash
rm -f server.log
sudo rm -rf /home/tswu/frpc/nimbus
sudo rm -rf _data/
make
#sudo ./server-sectionleader