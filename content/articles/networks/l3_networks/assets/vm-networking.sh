#!/bin/ash

hostname guestVM
echo "nameserver 1.1.1.1" > /etc/resolv.conf
ip link set dev lo up
ip addr add 192.168.100.2/24 dev eth0
ip link set dev eth0 up
ip route add default via 192.168.100.1 dev eth0
