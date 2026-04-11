# Artifact Build Notes

A short collection of notes on how I built these artifacts.  I didn't want to distract from the article with build details,
so they are posted here in case anyone wants to try building this themselves.

## Ubuntu VM Notes (The Build Environment)

My artifacts were created within a multipass VM (ubuntu noble) on my M2 Mac.  The VM was created with a cloud-config file

```
#cloud-config
users:
  - default
  - name: ubuntu
    gecos: Ubuntu
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: users, admin
    shell: /bin/bash
    ssh_import_id: None
    lock_passwd: true
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQ... Rest of public ssh key here ... LpNUtGFTWUnCj
apt:
  sources:
    docker.list:
      # Part of this example comes from https://stackoverflow.com/a/62540068
      source: deb [arch=arm64] https://download.docker.com/linux/ubuntu $RELEASE stable
      # Key ID can be found with “gpg --show-keys <(curl -s https://download.docker.com/linux/ubuntu/gpg)”
      keyid: 9DC858229FC7DD38854AE2D88D81803C0EBFCD88 
package_update: true
package_upgrade: true
packages:
  - nala
runcmd:
 - [ nala, update]
 - [ nala, install, git, openjdk-8-jdk-headless, gradle, jq, yq, -y ]
 - [ nala, install, openjdk-8-jdk-headless, -y ]
 - [ nala, install, gradle, -y ]
 - [ nala, install, jq, -y ]
 - [ nala, install, yq, -y ]
 - [ nala, install, sysstat, -y ]
 - [ nala, install, build-essential, -y ]
 - [ nala, install, ncurses-base, -y ]
 - [ nala, install, ncurses-bin, -y ]
 - [ nala, install, libncurses-dev, -y ]
 - [ nala, install, dialog, -y ]
 - [ nala, install, python3, -y ]
 - [ nala, install, tree, -y ]
 - [ nala, install, python3-ijson, -y ]
 - [ nala, install, python3-aiohttp, -y ]
 - [ nala, install, python3-matplotlib, -y ] 
 - [ nala, install, python3-numpy, -y ]
 - [ nala, install, qemu-system, -y ]
 - [ nala, install, ca-certificates curl, -y ]
 - [ nala, install, pkg-config, -y ]
 - [ nala, install, libssl-dev, -y ]
 - [ nala, install, libgdbm-dev, -y ]
 - [ nala, install, libnss3-dev, -y ]
 - [ nala, install, zlib1g-dev, -y ]
 - [ nala, install, libreadline-dev, -y ]
 - [ nala, install, libffi-dev, -y ]
 - [ nala, install, libsqlite3-dev, -y ]
 - [ nala, install, libbz2-dev, -y ]
 - [ nala, install, docker-ce, docker-ce-cli, containerd.io, docker-buildx-plugin, docker-compose-plugin, -y]
```

The VM was initalized with these multipass create options.  The `lts` tag current maps to the Ubuntu Noble release.

```
multipass launch -vv -n image-builder -c 6 -m 8G -d 64G --cloud-init ./cloud-init/justenough.yaml lts
```

## Busybox Notes

Download from running `git clone https://github.com/mirror/busybox`

Build steps:

```shell
make allnoconfig                                                                               # Creates near empty config with all applets disabled
sed -i 's/# CONFIG_FEATURE_PREFER_APPLETS is not set/CONFIG_FEATURE_PREFER_APPLETS=y/' .config # prefer busybox applets over external binaries in PATH
sed -i 's/# CONFIG_FEATURE_SH_STANDALONE is not set/CONFIG_FEATURE_SH_STANDALONE=y/' .config   # shell can run applets without needing PATH symlinks
make menuconfig                                                                                # add commands and features to busybox
make                                                                               # builds the busybox binary
make install CONFIG_PREFIX=/path/to/rootfs                                         # copies binary to CONFIG_PREFIX path and symlinks applets to busybox
```

The shell can be started with `/path/to/rootfs/bin/busybox ash`.

## Python Notes

Download a python [release](https://www.python.org/downloads/source/) and untar it.

```shell
wget https://www.python.org/ftp/python/3.14.3/Python-3.14.3.tgz
tar xzf Python-3.14.3.tgz
cd Python-3.14.3
mkdir ./my_staging_dir
./configure --prefix=/usr --enable-optimizations
make -j$(nproc)
./python -m test -j4
make install DESTDIR=$PWD/my_staging_dir
```

The final assembled artifact is in the my_staging_dir.
