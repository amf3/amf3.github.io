#!/bin/bash

qemu-system-aarch64 \
    -machine type=virt,gic-version=3 \
    -cpu cortex-a72 \
    -m 1024 \
    -smp 2 \
    -bios /usr/share/qemu-efi-aarch64/QEMU_EFI.fd \
    -cdrom alpine-standard-3.24.1-aarch64.iso \
    -netdev tap,id=net0,ifname=tap0,script=no,downscript=no \
    -device virtio-net-pci,netdev=net0,mac=52:54:00:12:34:56 \
    -nographic \
    -serial mon:stdio