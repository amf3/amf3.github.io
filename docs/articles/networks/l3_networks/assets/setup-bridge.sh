#!/bin/bash

# Create the bridge
sudo ip link add dev br0 type bridge
sudo ip link set dev br0 up
sudo ip addr add 192.168.100.1/24 dev br0

# Create the tap device
sudo ip tuntap add dev tap0 mode tap user $USER
sudo ip link set dev tap0 master br0
sudo ip link set dev tap0 up

# Enable packet fowarding between network interfaces
sudo sysctl -w net.ipv4.ip_forward=1