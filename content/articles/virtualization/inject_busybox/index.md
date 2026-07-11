---
title: "No Shell in Your Container? No Problem"
description: "Injecting Busybox into Minimal/Distroless Containers"
date: 2026-06-10T11:01:27-07:00
lastmod: 2026-07-11T11:58:20-07:00
draft: false
searchHidden: false
showtoc: true
categories: [Docker, Cloud]
images: ["./assets/rowboat.jpg"]
---

![A minimal rowboat is like a minimal container. Photo taken on the Isle of Malta, a historical shipping center.](./assets/rowboat.jpg)

Minimal or distroless container images provide only the components required to run an application. By design, they exclude shells and 
common troubleshooting utilities. While this keeps images small and secure, reality kicks in when you need to debug something. You 
could find yourself needing a shell for runtime initialization scripts, health checks, or troubleshooting.

This article shows how to add Busybox to a minimal container image and use it as the container's shell.

## Download Busybox

Busybox combines many common Unix utilities into a single executable.  The executable also includes the Almquist Shell (ash). 
There are a few ways to obtain the executable.  

* Download a [precompiled](https://busybox.net) binary file from the Busybox project.
* Copy the executable from a [existing](https://github.com/amf3/just_enough/pkgs/container/just_enough%2Fbusybox/909494266?tag=latest) container image.

For me, copying the executable from an existing container image is simpler because it avoids concerns about CPU architecture.  Here's
a partial Dockerfile example.

```dockerfile

FROM ghcr.io/amf3/just_enough/busybox:latest AS my-busybox   # This image contains a statically compiled Busybox binary.
FROM ghcr.io/amf3/just_enough/unbound_dns:latest             # This is a minimalist unbound_dns image that does not have a shell

COPY --from=my-busybox /bin/busybox /bin/busybox             # Add the Busybox binary to the minimal container

```

## Configure the Shell

Docker and Podman use /bin/sh by default when processing shell form RUN instructions.  Since Busybox provides ash, we need to change
the default shell used for RUN instructions.

```dockerfile
SHELL ["/bin/busybox", "sh", "-c"]       # Set the default shell to Busybox
```

Busybox can create symbolic links for utilities such as `ls`, `cat`, or `wget`, but I prefer not to create the symlinks in
minimal container images.  Calling commands through `/bin/busybox` keeps the image contents explicit and avoids adding additional
filesystem entries.

If you still would like the convenience of using named shell commands, create the symbolic links with:

```dockerfile
RUN /bin/busybox --install -s /bin       # Install the symlinked program names for shell commands to the /bin directory
```

## Using the Shell

Now that we have a working shell environment, let's see some examples.

### Startup Scripts

This example initializes the environment by creating a temporary directory before starting the application.

```dockerfile
FROM ghcr.io/amf3/just_enough/busybox:latest AS my-busybox
FROM ghcr.io/amf3/just_enough/unbound_dns:latest             

COPY --from=my-busybox /bin/busybox /bin/busybox             
SHELL ["/bin/busybox", "sh", "-c"]

ENTRYPOINT ["/bin/busybox", "sh", "-c"]
CMD ["echo 'Initializing...'; mkdir -p /tmp/data; exec /usr/sbin/unbound -d -c /etc/unbound/unbound.conf"]
```

The `exec` is needed to replace the shell and ensure signals are delivered correctly.

### Health Checks

This example adds a health check to a minimal container image using Busybox.

```dockerfile
HEALTHCHECK CMD ["/bin/busybox","wget","-q","-O","-","http://127.0.0.1:8080/health"]
```

### Interactive Troubleshooting

Busybox also provides an easy way to inspect a running container interactively.

```shell
docker exec -it my_running_container /bin/busybox sh
```

## Using Busybox as a build tool

Sometimes Busybox is only needed when building the container image.  
[Docker buildx](https://docs.docker.com/reference/cli/docker/buildx/build/) supports bind mounting files from one 
build stage into another.  Because the file is mounted for the duration of the RUN instruction, it's never
committed into the resulting image layer.  This allows using the Busybox binary during the build without copying
it into the final image.

This example bind mounts the Busybox binary from the busybox:latest container image.  The mounted binary is
used to run wget and chown while constructing a scratch image.  Because Busybox is mounted, it's not copied
into the final image.

```Dockerfile
FROM ghcr.io/amf3/just_enough/busybox:latest AS bb_shell
FROM scratch

RUN --mount=type=bind,from=bb_shell,source=/usr/bin/busybox,target=/usr/bin/busybox \
    ["/usr/bin/busybox", "wget", "-O", "/etc/unbound/root.hints", "https://www.internic.net/domain/named.cache"]
RUN --mount=type=bind,from=bb_shell,source=/usr/bin/busybox,target=/usr/bin/busybox \
    ["/usr/bin/busybox", "chown", "65534:65534", "/etc/unbound/root.hints"]
```

Other possibilities include changing file permissions, updating ownership, or extracting archives.  The approach
provides options for making modifications without adding unnecessary binaries to the final image.

## Conclusion

By adding a single file to the image, we gain support for health checks, initialization scripts, and interactive troubleshooting.
Busybox is a single binary that doesn't need a full distribution to work.  While this blurs the definition of a minimal
container, adding a single Busybox binary is a reasonable compromise between functionality and the philosophy of keeping container
images small and focused.
