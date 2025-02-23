---
title: "Getting started with cloud-init for unattended Linux deployments"
date: 2025-02-21T23:06:35-08:00
draft: false
searchHidden: false
showtoc: true
categories: [cloud]
robotsNoIndex: false
---

## Intro

Cloud compute companies like GCP, AWS, or Azure offer a management API for allocating resources. In the on-premise space,
services such as Docker or Incus provide APIs for managing containers or virtual machines (VMs). But what about installing 
the operating system (OS) on bare-metal hosts? What API exists for this task? This is where
[cloud-init](https://github.com/canonical/cloud-init) enters the picture, providing the ability to provision VMs or
bare-metal hardware.

cloud-init is a useful tool that doesn't rely on network services like PXE as a dependency.  Its simplicity saves time by
removing the need to navigate OS installation menus, while ensuring user accounts and installed software packages are consistent
across hosts. So why should one bother using cloud-init if they are managing a single host at home? In the event
the OS needs to be reinstalled due to failure, cloud-init allows one to quickly restore the system to a known state.

This example will use cloud-init to configure a Personal Package Archive (PPA), install Docker, and create a user account inside a Ubuntu VM.

## Prerequisite

I find that using cloud-init with Multipass is a easy way to get started.  Multipass is a virtual machine manager that
works with Linux, MacOS (arm & intel), and Windows.  When launching a new VM, Multipass is capable of initializing the VM with cloud-init.
If Multipass isn't already installed, this link will provide instructions for installing
[Multipass](https://canonical.com/multipass/install).  For this cloud-init introduction, I'm using Multipass on a M2 Macbook running MacOS Sequoia.

## cloud-init

Like many infrastructure tools, the input data for cloud-init is a YAML file.  For specifics of this schema, consult the official cloud-init
[documentation](https://cloudinit.readthedocs.io/en/latest/index.html).  There one will find that cloud-init input file
will need to be [prefixed](https://cloudinit.readthedocs.io/en/latest/tutorial/qemu.html#define-the-configuration-data-files) with `#cloud-config`.

### Package Management

Lets get started with package management for our Multipass instance.  This section will show how to add an external PPA (software repository) to
the VM with cloud-init to provide additional software packages and define a list of software packages to be installed on the VM.

#### Add External PPA

Add the 3rd-party [PPA](https://cloudinit.readthedocs.io/en/latest/reference/modules.html#apt-configure) provided by Docker, Inc.

```yaml
# Add Docker's PPA for Ubuntu
apt:
  sources:
    docker.list:
      # This snippet comes from https://stackoverflow.com/a/62540068
      source: deb [arch=arm64] https://download.docker.com/linux/ubuntu $RELEASE stable
      # Key ID can be found with “gpg --show-keys <(curl -s https://download.docker.com/linux/ubuntu/gpg)”
      keyid: 9DC858229FC7DD38854AE2D88D81803C0EBFCD88 
```

Should the GPG key ID for the Docker PPA change, I have left a comment above on how to find that value.  
This is how the GPG output appears in 2025.

```bash
$ gpg --show-keys <(curl -s https://download.docker.com/linux/ubuntu/gpg)
pub   rsa4096 2017-02-22 [SCEA]
      9DC858229FC7DD38854AE2D88D81803C0EBFCD88
uid                      Docker Release (CE deb) <docker@docker.com>
sub   rsa4096 2017-02-22 [S]
```

#### Define Package List

Specify a list of [packages](https://cloudinit.readthedocs.io/en/latest/reference/modules.html#package-update-upgrade-install) to install.

```yaml
# Update the list of packages available online
package_update: true
# Upgrade all installed packages
package_upgrade: true

# Install docker & other utilities
packages:
  - apt-transport-https
  - ca-certificates
  - curl
  - gnupg-agent
  - software-properties-common
  - docker-ce
  - docker-ce-cli
  - containerd.io
  - docker-buildx-plugin
  - docker-compose-plugin
```

### User Management

Here a new new [user](https://cloudinit.readthedocs.io/en/latest/reference/yaml_examples/user_groups.html) account is created and added
to the docker group with cloud-init.  Its likely our user will require both a password & ssh key for remote access.  A public ssh key and a
password hash is needed for cloud-init input.

#### Secrets: Generating a Password Hash

To create a password hash, use the `mkpasswd` command from Ubuntu's whois package.  This example will
hash the weak password of "abc123" with the sha512 algorithm.  A password better than "abc123" should be used if following these examples.

```bash
$ mkpasswd -m sha-512 "abc123"
$6$EkwQ38oDCPnJDuui$QKw3IISzY3emHXgJ/QHeEH8xyzGOKB3N6.bU/wAkwf4KDRsreB2iApa/EHULbunx6v9o9Q8foq4K.d8WtHukU/
```

As mkpasswd is specific to Linux and doesn't work with MacOS, one can alternatively use `openssl` to create a password hash.

```bash
$ echo abc123 |  openssl passwd -6 -stdin  
$6$tdPON3RwkVViXg41$4O9euMZeGFJQXgJ3bvP3YtVcCw9BwIMHLLkix1s/R7woSuAAFvWWtrqqQ.33ESzgcUi9/HdEwelqB9jJUIrpU0
```

#### Secrets: Generating a SSH public private key pair

To create a SSH key pair, use ssh-keygen: `ssh-keygen -t ed25519 -f ./docker_vm_key -C "app@docker_vm" -P abc123`.  This will create a public & private
ssh key in the current directory, with the easily guessable passphrase of `abc123`.  Once again, use a better passphrase if following these examples.

#### Defining the User Account

This defines an application account named "app".  The `ssh_authorized_keys` value comes from the contents of docker_vm_key.pub.  
As a convenience, the [public](./assets/docker_vm_key.pub) and [private](./assets/docker_vm_key) ssh keys from this example are provided.

```yaml
# create the docker group
groups:
  - docker

users:
  - name: app
    groups: [docker, admin, users]
    gecos: Application User
    shell: /bin/bash
    lock_passwd: true
    passwd: $6$tdPON3RwkVViXg41$4O9euMZeGFJQXgJ3bvP3YtVcCw9BwIMHLLkix1s/R7woSuAAFvWWtrqqQ.33ESzgcUi9/HdEwelqB9jJUIrpU0
    ssh_authorized_keys:
      - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIHsPNGa1NJLd4edDLRI033Sw33Nkl6qO+52qNAhY556C app@docker_vm
```

### Putting it all together

I've combined the YAML snippets into a single file named docker-install.yaml which can be downloaded [here](./assets/docker-install.yaml).  
Run the following to see cloud-init in action. This will create a virtual machine with 2 virtual CPU cores, 2 GB of ram,
with a 2GB virtual disk using the LTS release of Ubuntu.  Depending on your Internet speed, this may take a few minutes as
you'll be downloading packages from the Internet.

```bash
$ multipass launch -n docker-demo --cloud-init docker-install.yaml -c 2 -m 2G -d 4G lts
Starting docker-demo \
Waiting for initialization to complete |
Launched: docker-demo
```

To find the new VM and access it over SSH with the private key so a docker command can be ran from a remote shell.

```shell
% mp list                                                                                                                             
Name                    State             IPv4             Image
docker-demo             Running           192.168.64.32    Ubuntu 24.04 LTS        
                                          172.17.0.1     

% ssh -l app -i ./docker_vm_key 192.168.64.32
 The authenticity of host '192.168.64.32 (192.168.64.32)' can't be established.
 ED25519 key fingerprint is SHA256:EUqLjr9n9CyjKY6Y8EzNQGomeEtpePMFo5BXjO8YfHY.
 This key is not known by any other names.                                 
 Are you sure you want to continue connecting (yes/no/[fingerprint])? yes
 ...

app@docker-demo:~$ docker run hello-world
Unable to find image 'hello-world:latest' locally
latest: Pulling from library/hello-world
c9c5fd25a1bd: Pull complete 
Digest: sha256:e0b569a5163a5e6be84e210a2587e7d447e08f87a0e90798363fa44a0464a1e8
Status: Downloaded newer image for hello-world:latest

Hello from Docker!
...

```

## Conclusion

Several cloud-init basics have been covered in this introduction. Like adding a PPA, installing software packages, and creating a user account.  
While I understand that installing Docker in my example might not represent the typical workflow.  Combining cloud-init concepts with Multipass
creates a local mini-cloud on my Macbook.  I can quickly iterate through cloud-init data file changes for other platforms like AWS or on-premise hardware.

cloud-init is capable of much more, like formatting hard drives or managing network interfaces.  These & other topics will be covered in followups
which I will announce on [Bluesky](https://bsky.app/profile/amf3.bsky.social).  Follow me for notifications of when its made available.  Otherwise,
try out these examples and let me know what works.

