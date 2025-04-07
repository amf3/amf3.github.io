---
title: Unattended Ubuntu Installs - Virtual Machines to Bare-Metal
date: 2025-03-03T23:07:23-08:00
draft: false
searchHidden: false
showtoc: true
categories: [cloud, automation, operations]
---

## Intro

In a [previous post](../../../articles/cloudinit/intro/), I discussed using cloud-init and Multipass as a method of provisioning
virtual machines on a local computer with a cloud-like API.  Today I am going
to dive deeper with Ubuntu and how their autoinstall API can simplify on-premise host provisioning.

[autoinstall](https://canonical-subiquity.readthedocs-hosted.com/en/latest/intro-to-autoinstall.html)
is a tool that allows for unattended installations of Ubuntu, ensuring consistency, reporducibility, and providing automation
across a fleet of hosts.  In this post I'll walk through an example of using autoinstall
to configure networking, local storage, and demonstrate shell command execution during provisioning.  


## Prerequisite

Because the final target is a bare-metal instance, I find it quicker to iterate & test autoinstall changes with
QEMU on my macOS M2 laptop.  QEMU is a hardware emulator
which runs on Linux, macOS, & Windows, allowing the emulation of different CPUs, network cards, or storage devices.
Instructions to [install QEMU](https://www.qemu.org/download/#macos) can be found online.  For macOS, this can be as simple as
running `brew install qemu`.

Next we need the Ubuntu install media which can be downloaded
[here](https://ubuntu.com/download/server).

### QEMU overview

Let's get started by creating a virtual disk drive for installing Ubuntu.  This can be done with
`qemu-img create -f qcow2 virtual_disk.img 6G` which creates a 6GB virtual disk named
virtual_disk.img in the current directory.

In the example below, the `-boot once=d` option instructes QEMU to boot from the virtual CD-ROM on first startup. After which QEMU will
boot from the virtual disk. The other options initialize a 4 core CPU with 4GB of memory.  The `-net user,hostfwd` string will
port forward from localhost on the host system to port 22 on the virtual machine.  If additional port forwarding is needed, like testing 
https traffic on port 443 of the VM, multiple hostfwd options seperated by commas can used.  Be sure to adjust the filename and path to the Ubuntu ISO
as needed.

```code
qemu-system-x86_64 -hda virtual_disk.img -boot once=d  -cdrom ./ubuntu-24.10-live-server-amd64.iso -m 4096 -smp cpus=4 -net nic,model=virtio -net user,hostfwd=tcp:127.0.0.1:2222-:22
```

## Autoinstall

Autoinstall is included as part of the Ubuntu boot ISO and works with other provisioning tools from Canonical like
[Subiquity](https://canonical-subiquity.readthedocs-hosted.com/en/latest/),
[Curtin](https://curtin.readthedocs.io/en/latest/topics/overview.html), or [cloud-init](https://cloudinit.readthedocs.io/en/latest/).
When reading Autoinstall documentation, it's useful to know which tool is being used during each install stage as often those
options are passed to the underlying provisioning tool.

Like Kickstart for RHEL, Autoinstall is Ubuntu's answer to unattended
installations, and uses YAML files for data input.  Autoinstall uses default locations for finding the YAML files and locations
can also be specified in the GRUB menu when the instance boots.  Locations are specified as either a filepath or a URL.  I'll be
using a URL for the file locations.

Lets create the empty YAML files and use python to create a simple webserver to serve the files.  In another terminal type the following as the
webserver runs in the foreground.  Use cntl-c to terminate the python webserver when it's no longer needed.

```code
touch user-data meta-data network-config vendor-data
python3 -m http.server -b 127.0.0.1 -d $PWD 8080
```

Next start the virtual machine.

```code
qemu-system-x86_64 -hda virtual_disk.img -boot once=d  -cdrom ./ubuntu-24.10-live-server-amd64.iso -m 4096 -smp cpus=4 -net nic,model=virtio -net user,hostfwd=tcp:127.0.0.1:2222-:22
```

This will open a QEMU console window where we'll interact with the GRUB menu to specify the YAML file locations. Change focus to
the console window, highlight "Try or Install Ubuntu Server" and hit the `"e"` key to edit the grub menu.

On the line starting with “linux /casper/vmlinuz” add:  `autoinstall ds=nocloud\;s=http://10.0.2.2:8080/`  before
the three dashes.  The grub menu should look something like this when the edits are complete.

```code
linux   /casper/vmlinuz autoinstall ds=nocloud\;s=http://10.0.2.2:8080/ ---
initrd  /casper/initrd  
```

Exit grub and boot by following the on-screen instructions to hit F10 or cntl-x.  Watch the terminal running the python webserver and
requests for the autoinstall YAML files should be seen.  As they are empty config files, an interactive menu-driven session will present itself
in the QEMU console window.  To cancel the install, close the QEMU console window.

The GRUB modification tells autoinstall to use the
[nocloud](https://cloudinit.readthedocs.io/en/latest/reference/datasources/nocloud.html) plugin from cloud-init to download
its configuration at the specified URL.  QEMU assigns the special IP address of `10.0.2.2` to the host system when using `-net user`.  This allows the 
VM to reach services running on the host such as our local Python HTTP server and why autoinstall is able to download its configurations over HTTP.

The YAML block should be added to the user-data file that was created earlier.  The other files will remain empty.
The [minimal config example](https://github.com/canonical/subiquity/blob/main/doc/howto/autoinstall-quickstart.rst) assigns a
hostname of my-qemu-vm, creates an admin account named ubuntu, and assigns the ubuntu user the password of abc123.  
It's possible to generate a different secure password hash with openssl, as shown in 
this example: `echo abc123 |  openssl passwd -6 -stdin`.  Restart the QEMU VM so it boots from the 
virtual CD-ROM and modify the GRUB menu so it loads the new config when the VM boots.  

```yaml
#cloud-config
autoinstall:
  version: 1
  identity:
     hostname: my-qemu-vm
     username: ubuntu
     password: $6$xK2amorOU9tK4jt4$zLA1RZUpo4CzyDBzPDHCT61FLOngjWpV7Q/BH9KieLsJ/VG8r/Y88YIMLIOL.vc4ZHees40IAqORxjqa7GKti/
     # password is "abc123"

```

Autoinstall will take several minutes to complete and will reboot when done.  In some stages autoinstall can look stuck in some stages.
Remember that Linux virtual consoles are available to inspect running processes.  Virtual consoles are accessible by typing alt + left/right
arrow key or using alt + F2 or alt + F3.  (Use the option key for alt when using a Mac keyboard.)  Eventually the VM will reboot and the login 
prompt should be visible if everything went as expected.

Autoinstall has a list of defaults it uses when the option is present in the user-data file.  After logging into the QEMU instance, it's
possible to view the specified values from the user-data YAML file that have been merged into the defaults.

```code
sudo less /var/log/installer/autoinstall-user-data
```

Before continuing lets enable the ssh server, allow passwords for ssh login, and minimize the number of packages used during
the install.  Other options like locale or the keyboard setup can be found in the autoinstall-user-data file and added ot the example below.  Restarting
the QEMU VM and modifying the GRUB menu to reinstall the host OS is needed to apply the new changes to the YAML file.  Reinstalling
the OS also demonstrates the ease of initializing a system to a known state with autoinstall & cloud-init configs.

```yaml
#cloud-config
autoinstall:
   version: 1
   identity:
      hostname: my-qemu-vm
      username: ubuntu
      password: $6$xK2amorOU9tK4jt4$zLA1RZUpo4CzyDBzPDHCT61FLOngjWpV7Q/BH9KieLsJ/VG8r/Y88YIMLIOL.vc4ZHees40IAqORxjqa7GKti/
      # password is "abc123"
   ssh:
      # Install SSH server and allow password logins
      allow-pw: true
      install-server: true
   source:
      # id can also be ubuntu-server 
      id: ubuntu-server-minimal

```

### Networking

Both autoinstall and cloud-init support a netplan-formatted network configuration, meaning the YAML network example will work with
either installer.

Network device names are different between distributions that use Systemd (Ubuntu, Fedora) vs OpenRC (Alpine).  Where OpenRC
will use easily found device names like "eth0", or "eth1",
[Systemd](https://www.freedesktop.org/software/systemd/man/latest/systemd.net-naming-scheme.html) will use the PCI slot number.
A Systemd example might look like "enp2s0", where "en" means ethernet, and "p2s0" is the
[physical PCI slot](https://www.freedesktop.org/wiki/Software/systemd/PredictableNetworkInterfaceNames/).
This value will change based on which slot a PCI card is plugged into.  Luckily
[autoinstall](https://canonical-subiquity.readthedocs-hosted.com/en/latest/reference/autoinstall-reference.html#network)
lets us wildcard the device names.

This network example will work with either OpenRC or Systemd device names.  It's similar to what's used by
[Ubuntu's LiveCD](https://git.launchpad.net/livecd-rootfs/tree/live-build/ubuntu-server/includes.chroot.ubuntu-server-minimal.ubuntu-server.installer/etc/cloud/cloud.cfg#n23).

```yaml
network:
  version: 2
  ethernets:
    my-en-devices:
        match:
            # This will match Systemd naming conventions for ethernet devices which start with "en" and set them to use DHCPv4
            name: "en*"
        dhcp4: true
    my-eth-devices:
        match:
            # This will match OpenRC naming conventions like "eth0"
            name: "eth*"
        addresses:
          # This will specify a static network address
          - 10.10.10.2/24
        nameservers:
          # We can modify the DNS search path & specify DNS name servers.
          search:
            - "mycompany.local"
          addresses:
            - 10.10.10.253
            - 8.8.8.8
```

### Storage

Configuring
[storage](https://canonical-subiquity.readthedocs-hosted.com/en/latest/reference/autoinstall-reference.html#storage)
can be complex when configuring per partition byte offsets. Luckily we can provide a storage device name and let defaults
handle the details.  I'll show a basic lvm example but the other supported layouts are direct, and zfs.

Here we specify a LVM configuration with a sizing policy to use the entire disk for the logical volume.  If sizing-policy
were set to `scaled`, free space would be left on the storage device for things like snapshots or further expansion.

```yaml
storage:
    layout:
      name: lvm
      sizing-policy: all
```

Its possible to target a specific drive to wipe and install a new OS with a
[match](https://canonical-subiquity.readthedocs-hosted.com/en/latest/reference/autoinstall-reference.html#disk-selection-extensions)
statement.  There are multiple ways to select a storage device, model name, serial number, path, whether
its rotational or not, or even big or little in size.  These values can be found in smartctl output, which
comes from the smartmontools package.

```code
ubuntu@my-qemu-vm:~$ sudo apt-get install -y smartmontools
... install stuff ...

ubuntu@my-qemu-vm:~$ sudo smartctl -i /dev/sda
smartctl 7.4 2023-08-01 r5530 [x86_64-linux-6.11.0-18-generic] (local build)
Copyright (C) 2002-23, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF INFORMATION SECTION ===
Device Model:     QEMU HARDDISK
Serial Number:    QM00001
Firmware Version: 2.5+
User Capacity:    8,589,934,592 bytes [8.58 GB]
Sector Size:      512 bytes logical/physical
TRIM Command:     Available, deterministic
Device is:        Not in smartctl database 7.3/5528
ATA Version is:   ATA/ATAPI-7, ATA/ATAPI-5 published, ANSI NCITS 340-2000
Local Time is:    Tue Mar  4 05:45:56 2025 UTC
SMART support is: Available - device has SMART capability.
SMART support is: Enabled
```

If we wanted to match this disk by wild-carding the model name, we would use the following.

```yaml
#cloud-config
autoinstall:
   storage:
      layout:
         name: lvm
         sizing-policy: all
         match:
            model: QEMU*
```

Alternatively if our on-premise hardware instance had a 1GB SSD for the OS and a second 12GB spinning disk for data storage, we could
use a match with size `size: smallest` to install the OS on the 1GB disk.

```yaml
#cloud-config
autoinstall:
   storage:
      layout:
         name: lvm
         sizing-policy: all
         match:
            size: smallest
```

### Commands

Running arbitrary commands is possible when autoinstall runs.  Commands are specified as a list and run under "sh -c".
Its  possible to specify if commands should run early in the autoinstall process, late, or when an error occurs.

For example we want to hit a web endpoint when the installer has completed.

```yaml
#cloud-config
autoinstall:
   late-commands:
      - curl -H 'Content-Type: application/json' --data '{"host": "'$HOSTNAME'"}' http://myapi.example.com/success
      - echo "Install Success"  > /var/log/my.log
```

To run a command before the autoinstall process runs, like downloading internal x509 certificates:

```yaml
#cloud-config
autoinstall:
   early-commands:
      - mkdir /etc/ssl/mycerts
      - wget -O /etc/ssl/mycerts/internal.pem "http://x509api.example.com/certs/$HOSTNAME"
```

Or reporting an error when autoinstall fails

```yaml
#cloud-config
autoinstall:
   error-commands:
      - echo "Install failed" > /var/log/my.log
      - curl -H 'Content-Type: application/json' --data '{"host": "'$HOSTNAME'"}' http://myapi.example.com/failures
```

### cloud-init

It's possible to invoke cloud-init from autoinstall, allowing for additional functionality. This is done by placing
the cloud-init entries under a user-data key.  Here's a cloud-init example that installs a few packages. 

cloud-init and 
autoinstall sometimes perform similar tasks. When configuring a storage device with cloud-init, I found it was
better to use autoinstall as the cloud-init changes were overwritten.

```yaml
#cloud-config
autoinstall:
   user-data:
      package_update: true    # update the list of available packages
      package_upgrade: true   # upgrade currently installed packages.
      packages:
      - curl
      - ca-certificates

```

### Other

It's possible to configure a local proxy for downloading software packages.  Running apt-cacher-ng as a package
proxy inside a docker container on my laptop helps when I'm on a high latency Internet connection.

```yaml
#cloud-config
autoinstall:
   proxy: http://10.0.2.2:3142
```

## Provision a physical host

A complete autoinstall user-data file can be downloaded from [here](./assets/user-data).  It contains all the examples listed in this post.

Provisioning a physical host is very similar to using QEMU. The only change is when starting the python
webserver.  Instead of python binding to 127.0.0.1, have it bind to all interfaces so configs can be downloaded
by remote hosts.

```code
python3 -m http.server -d $PWD 8080
```

A USB thumb drive is needed to make the Ubuntu ISO available to the physical host; and a monitor & keyboard are needed
to modify the GRUB menu when the on-premise hosts boots. When modifying the GRUB menu,
instead of using http://10.0.2.2 in the nocloud URL, specify the hostname of the host running the python webserver.  In my scenario, the hostname would
resolve to my personal laptop.

## Wrapping Up

By leveraging autoinstall, it's possible to reliably reproduce system setups, whether for virtual machines or bare-metal hosts.
In this post, autoinstall was explored as a method to streamline unattended provisioning for Ubuntu instances.  Using a QEMU-based test environment, 
it was possible to quickly iterate on autoinstall configurations by modifying the GRUB menu to pull configuration files over HTTP.  The process demonstrated
how to format storage devices, set up networking, and run shell commands during installation.

Next steps?  If looking to extend this setup, consider integrating additional automation, such as PXE boot for network-based installs or using cloud-init to 
interact with configuration management systems like Puppet or Chef.  If you have insights from your own experiences, feel free to share them 
on [Bluesky](https://bsky.app/profile/amf3.bsky.social).
