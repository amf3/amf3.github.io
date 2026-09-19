---
title: "Self Hosting Publicly Facing Services"
description: "Understanding Virtual Networking and Segmentation: A Three Part Series"
date: 2026-09-17
draft: false
searchHidden: false
showToc: true
tags: ["networking", "docker", "qemu", "homelab"]
categories: [systems]
---

## Motivation

Online discussions around self hosting public services typically fall into two unhelpful groups.  Alarmists will warn that exposing any service guarantees unwanted security intrusions and offer no path forward.  The other group offers a recipe. It's suggested that some Proxmox or Podman command will solve the issue, but what the command does isn't explained.

Both approaches leave the reader without a clear understanding of the network boundaries being created or why they provide protection.

It's possible to self host a public facing service by limiting its network exposure to the traffic it needs.  Network segmentation and firewall rules provide the foundation for doing this.  My intent with this series is to demonstrate how combining features of the Linux kernel can control exposure of a public service.

## Network Series

Creating trust zones through network segmentation is a deep topic and is why I'm splitting this series into three sections.

* Part 1: [Virtual L2 Networks From Scratch](../bridges/) will create a basic Layer 2 network using `iproute` tooling. This section will describe the virtual network objects created in the host's kernel and how virtual machines and containers connect to the virtual network.

* Part 2: Introduces the Layer 3 components by explaining IP forwarding, routing on the host, and using nftables to implement firewall and mangle rules.

* Part 3: Pulls together the lessons from Parts 1 and 2 to discuss exposing a service to the public Internet.  This is where I'll explain how routing, trust zones, and firewall rules can be used to restrict traffic.

## What's not discussed

Docker/Podman or Proxmox specific details won't be discussed.  I'm not suggesting that hand written `ip` commands be used to replace those platforms.  Instead, understanding the kernel objects created by virtualization stacks is helpful in designing networks and troubleshooting network issues.

This series won't provide a cookbook for publicly hosting a service as there are too many variables to consider.  The intent is to provide enough foundation that one knows what questions to ask during follow up research.

Most changes I describe are ephemeral.  Creating network interfaces during boot up is the job of the init system. Depending on the operating system, managing network interfaces could be the responsibility of systemd, netplan, or rc.d.  Additional research is needed by the reader to make the changes permanent with their init system.

Because this series is focused on networking, I won't discuss application sandboxing with apparmor or selinux.  Both are important topics in preventing a public attacker from escaping a virtual machine or container, but are being ignored due to scope of the series.

## Recommendations

> **Warning:** I do not recommend running examples from this series on a computer that is directly attached to the Internet. Some of the commands can briefly expose the host system to unsolicited traffic, risking attack. 

**What I do recommend:**
   * Testing within a local isolated network environment.
   * Using an ephemeral host like a Virtual Machine or spare PC for testing changes.
   * Using a disposable network from a cloud provider when experimenting with public exposure.
        * Both Google Cloud and Oracle Cloud offer a free tier for cloud services.
        * Save self hosting at home for later, after gaining network and firewall experience.

## Series Updates

For updates on the series, one can either subscribe to the [RSS feed](/articles/index.xml) or follow along on [Bluesky account](https://bsky.app/profile/af9.us).