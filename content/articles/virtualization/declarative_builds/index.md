---
title: "Declarative Builds"
description: "Demonstrating an Opinionated Build Process for Container Images"
date: 2026-04-17T23:20:00
draft: false
searchHidden: false
showtoc: true
categories: [containers, cloud]
images: ["./assets/container_loading.jpeg"]
---

I previously described how I think [container base images](../base_image/) should be built. This post turns that 
idea into a working example.

Usually a container image starts with something like `FROM: python:3.x` or `FROM debian:bookworm-slim`.
What gets pulled in is more than just a runtime. It's an entire filesystem assembled by an upstream
distribution.

The filesystem often includes tools and binaries that aren't required for the application.  They're not
necessarily harmful, but they are unexplained.  When a container image includes a Perl interpreter,
USB utilities, or terminal tooling that the application never uses, it raises a simple question.  *Why is
that there?*

Minimal and hardened images try to address the question by removing as much as possible after the
fact.  This helps reduce size and surface area but it doesn't fully solve the problem.  Artifacts in
the image can still remain without clear intent.

This post takes a different approach.  Instead of starting with a full image and removing what doesn't
belong, we start with an empty filesystem and add only what is explicitly required.

## Build Overview

Instead of asking "What should I remove from this image?", we begin by asking "What do I need to add for my application to work?". That question inverts current container image builds by starting with an
empty root filesystem.

Files or build artifacts, are added only when declared in a manifest or as a discovered ELF dependency.  Existing container tools like Docker buildx or Buildah are still used.

The root filesystem starts as an empty staging directory, a process reads a manifest file and copies artifacts into the staging directory and exits.  Library dependencies are auto discovered for each
listed binary at runtime and copied into the staging directory. Container tooling like Docker buildx is then used to convert the staging directory into a container image.  

## The Manifest

Building the root filesystem starts with the manifest YAML file.  The manifest lists locations of artifacts to include, data to copy, directories and symlinks, to create in the empty root filesystem.

```yaml
input:
  backend: buildroot
  mode: staging
  path: /path/to/buildroot/output/staging

binaries:
  - BUILDROOT/usr/bin/netcat:/usr/bin/netcat
  - BUILDROOT/usr/sbin/nginx:/usr/sbin/nginx

data:
  - BUILDROOT/etc/passwd:/etc/passwd
  - BUILDROOT/etc/nginx/nginx.conf:/etc/nginx/nginx.conf
  - ./local/config/motd:/etc/motd

directories:
  - /var/log/nginx
  - /var/run
  - /tmp

symlinks:
  - /usr/bin/netcat:/usr/bin/nc
  - /usr/lib:/lib
```

Here's an overview of the manifest file.  The format will likely change as this is a work in progress.

* **input** - defines the artifact backend. I'm using [Buildroot](https://buildroot.org) to create build artifacts, but this could be expanded to other backends like fetching artifactory URLs.
* **binaries** - These are binaries to include with the empty root filesystem.  ELF dependencies are dynamically found and copied over.  Currently this is done by capturing lddtree output from paix-utils.
* **data** - Data is simply copied over from the source.
* **directories** - Directories to create in the root filesystem
* **symlinks** - Symlinks will be created in the root filesystem

## Build Output

The first step is to assemble the rootfs by reading artifacts from the buildroot staging directory.  assemble.py is a Python script that reads the manifest file. When reading the manifest, assemble.py
creates specified directories, symbolic links, then the specified data files.  Binaries are handled a bit differently.  In addition to copying the specified binary, assemble.py will run lddtree against each binary and copy each linked library mentioned in lddtree's output.

```
$ ./assemble.py unbound_container.yml my_root_fs
Rootfs built at: /home/ubuntu/just_enough/project_manifest/tools/my_root_fs
```

The assembled filesystem which is only 12MB in size.

```
$ tree my_root_fs/
my_root_fs/
├── bin -> /usr/bin
├── etc
│   ├── group
│   ├── passwd
│   └── unbound
│       └── unbound.conf
├── lib
│   ├── ld-linux-aarch64.so.1
│   ├── libc.so.6
│   ├── libcrypto.so.3
│   ├── libevent-2.1.so.7
│   ├── libsodium.so.23
│   └── libssl.so.3
├── lib64 -> /lib
├── sbin -> /usr/sbin
├── usr
│   ├── lib -> /lib
│   └── sbin
│       └── unbound
└── var
    ├── cache
    │   └── unbound
    └── run

14 directories, 10 files

$ du -sh my_root_fs/
12M     my_root_fs/
```

## Demo

I used this Dockerfile with `docker build` to convert the staging directory into a container image.

```Dockerfile
FROM scratch
COPY --chown=0:0 my_root_fs /
EXPOSE 53/tcp
EXPOSE 53/udp
ENTRYPOINT ["/usr/sbin/unbound", "-d", "-c", "/etc/unbound/unbound.conf"] 
```

This is how the image appears.

```shell
$ docker history unbound-minimal:latest 
IMAGE          CREATED       CREATED BY                                      SIZE      COMMENT
43ac3c4ba387   3 hours ago   ENTRYPOINT ["/usr/sbin/unbound" "-d" "-c" "/…   0B        buildkit.dockerfile.v0
<missing>      3 hours ago   EXPOSE [53/udp]                                 0B        buildkit.dockerfile.v0
<missing>      3 hours ago   EXPOSE [53/tcp]                                 0B        buildkit.dockerfile.v0
<missing>      3 hours ago   COPY --chown=0:0 my_root_fs/ / # buildkit       11.7MB    buildkit.dockerfile.v0
```

The Docker run command to launch the container.

```shell
$ docker run -it --rm -p 53:53/udp -p 53:53/tcp  unbound-minimal:latest 
Apr 17 06:43:35 unbound[1:0] notice: init module 0: validator
Apr 17 06:43:35 unbound[1:0] notice: init module 1: iterator
Apr 17 06:43:35 unbound[1:0] info: start of service (unbound 1.21.1).
```

The dig command is run in a second terminal to validate that the resolver works.

```shell
$ dig @127.0.0.1 +answer bsky.app 

; <<>> DiG 9.18.39-0ubuntu0.24.04.3-Ubuntu <<>> @127.0.0.1 +answer bsky.app
; (1 server found)
;; global options: +cmd
;; Got answer:
;; ->>HEADER<<- opcode: QUERY, status: NOERROR, id: 60188
;; flags: qr rd ra; QUERY: 1, ANSWER: 8, AUTHORITY: 0, ADDITIONAL: 1

;; OPT PSEUDOSECTION:
; EDNS: version: 0, flags:; udp: 1232
;; QUESTION SECTION:
;bsky.app.                      IN      A

;; ANSWER SECTION:
bsky.app.               60      IN      A       3.15.241.112
bsky.app.               60      IN      A       16.59.153.152
bsky.app.               60      IN      A       3.134.210.175
bsky.app.               60      IN      A       18.220.107.168
bsky.app.               60      IN      A       3.12.186.67
bsky.app.               60      IN      A       18.225.51.158
bsky.app.               60      IN      A       3.150.190.10
bsky.app.               60      IN      A       3.135.42.242

;; Query time: 32 msec
;; SERVER: 127.0.0.1#53(127.0.0.1) (UDP)
;; WHEN: Thu Apr 16 23:44:28 PDT 2026
;; MSG SIZE  rcvd: 165
```

With the unbound log output showing the query was received.

```shell
Apr 17 06:44:28 unbound[1:0] info: 172.17.0.1 bsky.app. A IN
Apr 17 06:44:28 unbound[1:0] info: resolving bsky.app. A IN
Apr 17 06:44:28 unbound[1:0] info: response for bsky.app. A IN
Apr 17 06:44:28 unbound[1:0] info: reply from <.> 1.1.1.1#53
Apr 17 06:44:28 unbound[1:0] info: query response was ANSWER
Apr 17 06:44:28 unbound[1:0] info: 172.17.0.1 bsky.app. A IN NOERROR 0.032742 0 165
```

## Wrapping Up

This demo shows a different way to build container images.  Instead of inheriting an unknown
filesystem and removing what doesn't belong, the image is assembled from known artifacts with
explicit intent. Every file is accounted for and nothing exists by accident.  

This process does introduce more upfront work.  A manifest needs to be defined and the build process
is more deliberate than a typical Dockerfile.  The benefit is gaining complete control over the
final image and the ability to reproduce it exactly from known inputs.

The tooling is still evolving, and there are gaps around permissions, ownership and input sources.  But the model works as
container images can be built declaratively without relying on base image inheritance.

The question isn’t whether this approach is possible, but whether this level of control is worth it for your use case.
