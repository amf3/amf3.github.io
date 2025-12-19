---
title: "Decoupling Compute and Storage"
description: "Why a Two-Node Docker Swarm with ZFS Snapshots Is Enough"
date: 2025-12-16T21:51:16-08:00
draft: false
searchHidden: false
showtoc: true
categories: [homelab, architecture]
---

## Who Should Not Read This Post

This post does not discuss:

- data support for multiple writers
- having zero downtime
- cloud-scale best practices

Instead I discuss embracing single points of failure (SPOF) in exchange for simplicity within
environments that can tolerate downtime.

## Permission to Not Over-Engineer

In environments that tolerate downtime, having a SPOF is great because they are simple to deploy and it's obvious what failed.

Lets examine the typical home router.  In addition to routing packets between the LAN and the Internet, the home router also
provides DHCP and DNS services.  Then someone learns of PiHole and runs it on a separate device like a Raspberry PI. Now the router
queries PiHole for adblock entries during DNS resolution.  Ad-free browsing works works great; until HTTP requests
stop working.  With this scenario, it's not obvious if it's an issue on the router or an issue with PiHole.

I recently ran into a situation on deciding how to improve my container deployments.

Currently I use Docker Engine with Compose files to manage running containers. Container storage is either named volumes or bind mounts,
both of which write to the local disk of the host running Docker Engine.  This works but it has a few issues:

- Services go offline during planned maintenance on the host
- Container placement is manual. Having a second hardware instance doesn't help.
- Docker Compose deployments cause container restarts
- Failed Docker Compose deployments have no rollback

I started to look at K3S, Nomad, and Docker Swarm to address these short comings. While I still plan on using a single hardware
instance, orchestration would help with downtime during planned maintenance.  This led to thinking how the design would change by adding
a second host and because of distributed storage, would I be trapped by choices made today?  

## Compute Does Not Need to Scale with Storage

I know many distributed storage solutions use a consensus model requiring three hosts. Does this mean I'm required to run three computers
and pay for the additional power usage?

It feels strange to admit, but this led to two evenings of not knowing how to proceed. I assumed that adding a second
host meant I also needed distributed storage and now I need three hosts. This assumption stalled everything.  Eventually I realized
its okay to have orchestration for compute but not storage.

Thinking about storage and compute as separate problems was the unlock I needed to move forward. I won't need the complications of
running Kubernetes or Nomad. The simpler solution of using Docker Swarm can solve my deployment requirements with either one or two
hardware instances.

## Compute: What Swarm Provides

Swarm is responsible for compute orchestration and overlay networking.  It does not provide storage orchestration or multi-writer safety.

Swarm mode is enabled on each Docker instance, and the instance registers with the cluster as either a manager role or worker role. The
Swarm worker provides CPU and memory resources to the running containers and must have sufficient capacity for its workloads.

As a Swarm manager, the instance is responsible for orchestration, (scheduling, container placement, maintaining cluster state) by using RAFT
consensus. RAFT needs a quorum to make decisions and a quorum is calculated as "(N-1)/2 + 1", where N is the number of Swarm
managers.  In a two node deployment I can only have a single manager, otherwise quorum will never be reached.

In a two node Swarm deployment, one node would run both the worker and manager. The second node only joins as a worker.  Write heavy apps are assigned
to the manager instance and stateless apps assigned to the worker node. Swarm labels are applied to services to ensure container placement.  

## Storage

Swarm itself does not manage storage. Swarm presumes storage is managed externally to its processes. For example, CephFS
would be mounted by external systems on the host OS and treated as a local filesystem by Swarm.  A take away is Swarm does not provide
multi-writer safety. If two containers attempt to write to the same dataset, Swarm does nothing to provide locking or prevent data
corruption.  It assumes the underlying storage layer handles this.

Databases have built in multi-writer support by providing locking at either the table or row level.  Usually they are capable of replicating
data to a secondary instance.  Because of these features, they solve my multi-instance storage problem by scaling independently of the container
platform.

Local filesystems are more challenging. Detected filesystem changes must be transferred to secondary storage. I use ZFS snapshots with
send and receive over the network.  Snapshots are taken on the manager instance and replicated to the worker node every 15 minutes using
[zrepl](https://github.com/zrepl/zrepl).

This approach is a tradeoff for simplicity and assumes one is okay with a bounded data-loss window. If the primary ZFS dataset gets corrupted, up
to 15 minutes of data could be lost.  Write heavy applications are pinned to the manager node so their datasets can be replicated.

Another challenge is application-level locking.  Databases enforce locking in their clients, but local filesystems do not. While mechanisms
like flock or fcntl exist, they only work if the application uses them.  To avoid data corruption caused by lack of locking, I make sure
to deploy only a single container for each dataset.  This greatly simplifies the storage design by avoiding the need of distributed locking.

For container storage, I use a mix of named volumes and bind mounts. Persistent data that requires replication uses bind mounts. Ephemeral data
that needs to persist between container restarts uses named volumes. Because bind mounts reference host paths, the same directory structure and
ownership (UID) must exist on both hardware instances. Configuration management is used to ensure consistency between the two hardware instances.

## Failure Modes

Single points of failure are not accidental in this design. They provide predictable and observable failure modes, that act as constraints for
modeling the system.

- **Worker node failure:** This is a non-event. Stateless services are rescheduled onto the manager node by Swarm. No data recovery is required.

- **Manager node failure:** This is an expected control-plane failure where manual intervention is required. The worker node is promoted to manager and services restarted on the
  manager. This is a deliberate tradeoff in exchange for simplicity.

- **Data corruption on the manager node:** This is treated the same as a manager failure. The worker node is promoted to manager and services are restored using the most recent ZFS snapshot.

- **Replication lag:** Replication is monitored to ensure snapshot transfer time remains below the snapshot interval. If replication falls behind, the response is to reduce the snapshot interval or increase network capacity. This keeps data loss bounded and predictable.

- **Multi-writer data corruption:** If because of misconfiguration, more than one container (writer) is launched for the same dataset and corruption occurs, stop all containers for the application and restore the data from the last known-good ZFS snapshot.

Depending on snapshot age and response time, expected downtime for these scenarios is 15-30 minutes.  The goal is not uninterrupted
service, but fast understandable recovery.

## When the solution is smaller than the problem

I'm not defending Swarm in this post. I'm defending clear failure modes, explicit ownership, and operational practicality. The goal is to start
with requirements and choose the smallest system that satisfies them.

Because this design keeps the data independent of orchestration, it provides an escape hatch to K3s or Nomad without having to decouple data from Swarm.

If my availability requirements change, then I would add a third low-power node like a Raspberry Pi to establish a three node quorum.  I would also
migrate from Swarm to K3s to leverage the Longhorn distributed filesystem which is only available under Kubernetes. For multi-writer workloads, shared
filesystems like NFS could work if the container application supports file locking. Neither Swarm or Longhorn provides multi-writer locking on behalf of
the application.

If you have any thoughts on this post, feel free to share them on [Bluesky](https://bsky.app/profile/af9.us).
