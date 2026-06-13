---
title: "No Shell in Your Container? No Problem"
description: "Injecting BusyBox into Minimal/Distroless Containers"
date: 2026-06-10T11:01:27-07:00
draft: false
searchHidden: false
showtoc: true
categories: [Docker, Cloud]
images: ["./assets/rowboat.jpeg"]

---

![A minimal rowboat is like a minimal container. Photo taken on the Isle of Malta, a historical shipping center.](./assets/rowboat.jpeg)

Minimal or distroless container images provide only the components required to run an application.  By design they exclude shells and 
common troubleshooting utilities.  While this keeps images small and focused, there are situations where the shell is useful.  Shell
scripts may be required during runtime initialization, health checks, or general troubleshooting.

This article shows how to add BusyBox to a minimal container image and use it as the container's shell.

## Download BusyBox

BusyBox combines many common Unix utilities into a single executable.  The executable also includes the Almquist Shell (ash). 
There are a few ways to obtain the executable.  

* Download a [precompiled](https://busybox.net) binary file from the BusyBox project.
* Copy the executable from a [pre-existing](https://github.com/amf3/just_enough/pkgs/container/just_enough%2Fbusybox/909494266?tag=latest) container image.

For me, copying the executable from an existing container image is simpler because it avoids concerns about CPU architecture.  Here's
a partial Dockerfile example.

```dockerfile

FROM ghcr.io/amf3/just_enough/busybox:latest AS my-busybox   # This image contains a statically compiled BusyBox binary.
FROM ghcr.io/amf3/just_enough/unbound_dns:latest             # This is a minimalist unbound_dns image that does not have a shell

COPY --from=my-busybox /bin/busybox /bin/busybox.            # Add the BusyBox binary to the minimal container

```

## Configure the Shell

Docker and Podman use /bin/sh by default when processing shell form RUN instructions.  Since BusyBox provides ash, we need to change
the default shell used for RUN instructions.

```dockerfile
SHELL ["/bin/busybox", "sh", "-c"]       # Set the default shell to busybox
```

BusyBox can create symbolic links for utilities such as `ls`, `cat`, or `wget`, but I prefer not to create the symlinks in
minimal container images.  Calling commands through `/bin/busybox` keeps the image contents explicit and avoids adding additional
filesystem entries.

If you still would like the convenience of using named shell commands, create the symbolic links with:

```dockerfile
RUN /bin/busybox --install -s /bin       # Install the symlinked program names for shell commands to the /bin directory
```

## Using the Shell

Now that we have a working shell environment, let's see some examples.

### Startup Scripts

This example initializes the environment by creating a temporary directory before starting the application

```dockerfile
FROM ghcr.io/amf3/just_enough/busybox:latest AS my-busybox
FROM ghcr.io/amf3/just_enough/unbound_dns:latest             

COPY --from=my-busybox /bin/busybox /bin/busybox             
SHELL ["/bin/busybox", "sh", "-c"]

ENTRYPOINT ["/bin/busybox", "sh", "-c"]
CMD ["echo 'Initializing...'; mkdir -p /tmp/data; exec /usr/sbin/unbound -d -c /etc/unbound/unbound.conf"]
```

The exec is needed to replace the shell and ensure signals are delivered correctly.

### Health Checks

This example adds a health check to a minimal container image using BusyBox.

```dockerfile
HEALTHCHECK CMD ["/bin/busybox","wget","-q","-O","-","http://127.0.0.1:8080/health"]
```

### Interactive Troubleshooting

BusyBox also provides an easy way to inspect a running container interactively.

```shell
docker exec -it my_running_container /bin/busybox sh
```

## Conclusion

By adding a single file to the image, we gain support for health checks, initialization scripts, and interactive troubleshooting.
BusyBox is a single binary that doesn't need a full distribution to work.  While this blurs the definition of a minimal
container, adding a single BusyBox binary is a reasonable compromise between functionality and the philosophy of keeping container
images small and focused.
