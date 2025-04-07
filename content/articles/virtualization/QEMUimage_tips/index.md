---
title: "Tips for working with qemu images"
date: 2025-04-06T07:41:29-04:00
draft: false
searchHidden: false
showtoc: true
categories: [cloud, virtualization, tips]
---

# About QEMU images

QEMU uses files to emulate storage devices, and the features available
depend on how those files are created.  While QEMU can emulate disks from Parallels and VirtualBox, I’m going to 
focus on the formats most commonly used in automation and scripting, **raw** and **qcow2**.

The default format is raw and raw offers the fewest features.  It's just plain storage.  The other format qcow2
supports compression, snapshots, and copy-on-write in addition to storage.

## Raw Format

Creating a raw disk with qemu-img is similar to using dd to create a block-based file. One can see this with 
the output of **qemu-img info**.

Here I create two storage devices, one with qemu-img which defaults to the raw format and another with the
dd command.

```sh
% qemu-img create my_disk.img +1m         
Formatting 'my_disk.img', fmt=raw size=1048576

% dd if=/dev/zero of=my_block.file count=1 bs=1m
1+0 records in
1+0 records out
```

Now let's use **qemu-img info** to confirm there's no difference between the two files.

```sh
% qemu-img info my_disk.img  
image: my_disk.img
file format: raw
virtual size: 1 MiB (1048576 bytes)
disk size: 1 MiB
Child node '/file':
    filename: my_disk.img
    protocol type: file
    file length: 1 MiB (1048576 bytes)
    disk size: 1 MiB

% qemu-img info my_block.file 
image: my_block.file
file format: raw
virtual size: 1 MiB (1048576 bytes)
disk size: 1 MiB
Child node '/file':
    filename: my_block.file
    protocol type: file
    file length: 1 MiB (1048576 bytes)
    disk size: 1 MiB
```

## Qcow2 Format

Creating a disk in **qcow2** format enables zlib compression by default.

```sh
% qemu-img create -f qcow2 my_disk.img 1M 
Formatting 'my_disk.img', fmt=qcow2 cluster_size=65536 extended_l2=off compression_type=zlib size=1048576 lazy_refcounts=off refcount_bits=16

% qemu-img info my_disk.img 
image: my_disk.img
file format: qcow2
virtual size: 1 MiB (1048576 bytes)
disk size: 196 KiB
cluster_size: 65536
Format specific information:
    compat: 1.1
    compression type: zlib
    lazy refcounts: false
    refcount bits: 16
    corrupt: false
    extended l2: false
Child node '/file':
    filename: my_disk.img
    protocol type: file
    file length: 192 KiB (197120 bytes)
    disk size: 196 KiB
```

# Tip One - Resize an image file

It's possible to grow or shrink a QEMU storage device.  Think of this as expanding the physical SSD itself, not the filesystem 
that sits on it.  **Important,** when shrinking a image with negative values, 
**always shrink the filesystem first** using resize2fs before running qemu-img resize **or risk data corruption.**

```sh
% qemu-img resize my_disk.img +1m  
Image resized.
```

When inspecting the new disk image, we see the new capacity is 2MB but the file size on disk is under 200KB.  This is because qcow2 supports
copy-on-write and compression.

```sh
% qemu-img info my_disk.img
image: my_disk.img
file format: qcow2
virtual size: 2 MiB (2097152 bytes)
disk size: 196 KiB
cluster_size: 65536
Format specific information:
    compat: 1.1
    compression type: zlib
    lazy refcounts: false
    refcount bits: 16
    corrupt: false
    extended l2: false
Child node '/file':
    filename: my_disk.img
    protocol type: file
    file length: 192 KiB (197120 bytes)
    disk size: 196 KiB

% ls -lh my_disk.img 
-rw-r--r--  1 adam  staff   192K Apr  6 10:19 my_disk.img
```

If I were to resize a QEMU storage file formatted as raw, the file size on disk of 2MB matches the image capacity of 2MB as raw 
doesn't support compression or copy-on-write.

```sh
% qemu-img create raw_disk.img +2m
Formatting 'raw_disk.img', fmt=raw size=2097152

% ls -lh raw_disk.img 
-rw-r--r--  1 adam  staff   2.0M Apr  6 10:22 raw_disk.img
```

# Tip Two - Snapshots

Snapshots are supported with qcow2 devices.  These are handy for creating a base disk image that's shareable and later modified
for other purposes.  Snapshots also make a great backup point before making image changes.

To create a snapshot named "my_first_snapshot" inside an existing qcow2 image.

```sh
% qemu-img snapshot -c my_first_snapshot my_disk.img 
```

To list existing snapshots

```sh
% qemu-img snapshot -l my_disk.img 
Snapshot list:
ID      TAG               VM_SIZE                DATE        VM_CLOCK     ICOUNT
1       my_first_snapshot      0 B 2025-04-06 10:37:07  0000:00:00.000          0
```

To revert or "apply" a snapshot

```sh
% qemu-img snapshot -a my_first_snapshot my_disk.img 
```

To delete a snapshot from a file

```sh
% qemu-img snapshot -d my_first_snapshot my_disk.img 
```

# Tip Three - Modify a QEMU image

While many online guides suggest using the Network Block Device (NBD) kernel driver in Linux to mount and modify QEMU images, I 
use a different process that also works on MacOS.  My preferred method is to boot a VM using QEMU and attaching the image as a data drive.

This example uses the [extended x86_64 Alpine Linux ISO](https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-extended-3.21.3-x86_64.iso)
and a QEMU command that mounts the image as a data drive.  The Alpine extended ISO lets you log in as root with an empty password, 
which makes quick edits easy.

```sh
#/bin/sh
qemu-system-x86_64 \
  -m 2G -smp cpus=4 -serial stdio \
  -boot once=d  \
  -drive file=./my_disk.img,format=qcow2,media=disk,cache=unsafe \
  -drive file=./alpine-extended-3.21.2-x86_64.iso,format=raw,media=cdrom \
  -nic user,model=virtio-net-pci,hostfwd=tcp::2222-:22
```

Once logged in, you'll see the QEMU file we want to modify listed as /dev/sda.  The device hasn't been formatted with a filesystem, but if 
one were present it could be mounted within the VM, files edited within the image, and then unmounted.

# Tip Four - Transfer a QEMU image to bare-metal

It's possible to use a QEMU image with bare-metal by converting it to **raw** format.  Use the following to convert the image from qcow2 to raw.

```sh
% qemu-img convert -f qcow2 -O raw my_disk.img raw_disk.img

% qemu-img info raw_disk.img 
image: raw_disk.img
file format: raw
virtual size: 10 MiB (10485760 bytes)
disk size: 10 MiB
Child node '/file':
    filename: raw_disk.img
    protocol type: file
    file length: 10 MiB (10485760 bytes)
    disk size: 10 MiB
```

Once we have the raw image, the **dd** command can be used to write the data to either a USB stick or physical SSD.  To avoid 
any destructive commands let's pretend raw_disk2.img represents /dev/sdc, your verified USB thumb drive.

```sh
% dd if=raw_disk.img of=raw_disk2.img bs=1m 
10+0 records in
10+0 records out
10485760 bytes transferred in 0.006266 secs (1673437600 bytes/sec)
```

Because our raw file is only 10MB in size, only 10MB will be used on the thumb drive.  This is where follow up tools like LVM or 
resize2fs will extend the filesystem to fill the entire thumb drive.  Tools used for expansion depends on how the filesystem was created.

# Putting it all together

Enough with the documentation, let's put it into practice with a real use case.  Presume that after reading my [cloud-init
tutorials](../../../articles/cloudinit/intro/) you wish to modify the [Alpine Linux cloud-init image](https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/cloud/nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2) before installation.

We can see the downloaded file is a qcow2 image with a capacity of 200Mb from **qemu-img info**. 

```sh
 % qemu-img info nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2 
image: nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2
file format: qcow2
virtual size: 200 MiB (209715200 bytes)
disk size: 181 MiB
...
```

As we want to install our java app into the installer, we need to add space to the image with **qemu-img resize**.  But first, 
let’s create a snapshot. That way, if we make a mistake, we won’t need to re-download the cloud-init image.

```sh
% qemu-img snapshot -c no_modifications nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2 

% qemu-img resize nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2 +800M
Image resized.

 % qemu-img info nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2        
image: nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2
file format: qcow2
virtual size: 0.977 GiB (1048576000 bytes)
disk size: 197 MiB
cluster_size: 65536
Snapshot list:
ID      TAG               VM_SIZE                DATE        VM_CLOCK     ICOUNT
1       no_modifications      0 B 2025-04-06 15:23:50  0000:00:00.000          0
Format specific information:
...
```

I'm still using the Alpine extended ISO to boot the VM. Alpine cloud images require setup for ssh key authentication to login and an
empty root password is much easier to use.

```sh
% qemu-system-x86_64 \
    -m 2G -smp cpus=4 -serial stdio \
    -boot once=d  \
    -drive file=./nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2,format=qcow2,media=disk,cache=unsafe \
    -drive file=./alpine-extended-3.21.2-x86_64.iso,format=raw,media=cdrom \
    -nic user,model=virtio-net-pci,hostfwd=tcp::2222-:22
```

Login as root and mount the disk device under /mnt

```sh
localhost:~# mount /dev/sda /mnt 
localhost:~# ls /mnt
bin         home        mnt         run         tmp
boot        lib         opt         sbin        usr
dev         lost+found  proc        srv         var
etc         media       root        sys
```

Then make changes to the cloud image, unmount the filesystem and you're done.

```sh
localhost:~# echo "Adam Faris was here" > /mnt/etc/motd 

localhost:~# cat /mnt/etc/motd 
Adam Faris was here

localhost:~# umount /mnt

localhost:~# poweroff
```

Finally, convert our modified cloud image from qcow2 format to raw format, then use dd to write the raw image to a USB device.

```sh
% qemu-img convert -f qcow2 -O nocloud_alpine-3.21.2-x86_64-bios-cloudinit-r0.qcow2 alpine_cloudinit.raw
% dd if=alpine_cloudinit.raw bs=1m of=/dev/...
```

With the modified image written to the USB device, you can now boot a physical machine from it. Thanks for sticking with me until 
the end. If you find this content useful, follow me on [BlueSky social](https://@amf3.bsky.social) for future announcements.