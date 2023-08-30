---
title: "Elastic Block Store Design for Home Lab Use"
date: 2023-08-25T07:38:23-07:00
draft: true
categories: [homelab, architecture]
---

## Intro

I have a small home lab that I use for testing different software stacks. Lately I have been choosing a distributed 
storage solution for use at home.  Usually I look at industry and scale down the solution.  This has lead to 
conversations with colleagues and the question of "How did Amazon Web Services implement their Elastic Block Store 
offering?".  

## Background

There are solutions that provide network storage like iSCSI, CIFS, or NFS.  These protocols can be grouped into two
data access methods.  One method will access data as a *file* & the method is to access data as a *block device*.  Block 
storage is a device that stores data in evenly spaced chunks called blocks.  An example block storage device is 
the hard drive inside a computer.

CIFS & NFS are protocols where data is accessed as a file.  iSCSI will access data at the block level. Sometimes 
backup solutions like Apple's TimeMachine will obfuscate this by treating the file as a block device. From the 
perspective of the storage server, it's providing NFS or CIFS access to the file so its file based access.

For me, iSCSI is the more interesting use case as it allows a physical host, VM, or container [^1], to access the block 
device over the network. Having access to the underlying block device provides flexibility with data management.  An 
example is being able to snapshot the block device as a backup or a point in time image for use in future 
provisioning.

In my home lab I have tinkered with iSCSI and found that data access latency is bad without a dedicated storage-only 
network connection.  Even with a dedicated connection, its arguable that a 1Gb/s network isn't 
fast enough for a SAN.  Typically when SANs are mentioned companies pull out the checkbook because it involves custom 
hardware. Elastic Block Storage is no exception to this statement.

## EBS Findings

Elastic Block Storage (EBS) is a storage offering from Amazon AWS.  EBS is a service that emulates a local block device
where data is persistently stored over the network.

AWS uses a combined approach to providing a performant storage solution to clients. [^2]  This is done by using a 
custom network stack that uses datagrams, specialized hardware that provides a bridge between storage & networking, 
and a minimized Linux kernel running the EC2 servers.

### Network

AWS replaced TCP with a datagram based protocol named Scalable Reliable Datagram (SRD).  SRD supports equal cost 
multipathing, congestion control, and reliable out of order delivery, leaving ordering to networking (OSI) layers above
SRD.  

A summary of why SRD matters is it removes latency caused by the three-way TCP handshake & congestion control delays 
in TCP.

### Data Processing Units

Data Processing Units (DPU) are an interesting bit of hardware.  The DPU is a PCI card that has a onboard CPU, Memory, & 
SFP+ networking ports.  The DPU plugs into the server and gains access to the server's PCI lanes.  This allows it 
emulate both a storage device, network interface cards (NIC), and provide out of band management.

I quickly realized the DPU, which AWS calls a Nitro card in marketing, is the magic sauce for EBS.  The Nitro card will 
represent itself as a NVME storage device to the server's operating system (OS) and also handles networking to the EBS 
volumes without the server's OS knowing it.  AWS even moved the SRD network stack into the Nitro card which means 
networking is handled by the hardware, not software, making networking more efficient.

In addition to emulating the storage device, the Nitro card will also provide general networking for the server OS.  One 
other function the Nitro card performs, is during server resets (reboots), the Nitro card will checksum the bios
to validate it hasn't been tampered with. Remember that its a computer on a PCI card with access to the PCI lanes of 
the server its plugged into.

### Kernel

AWS announced in 2019 that they were moving away from Xen to using KVM with Qemu for hosting their EC2 offering.  This 
is relevant to the NVME storage interface provided by the Nitro card.

Qemu has the ability to directly access a NVME device by using a user-space driver.  Because the Nitro card provides 
both storage and system network interfaces, there's no need for the linux kernel on the server to need a
storage or network driver.  This leads to a very lean kernel on the EC2 host, essentially turning it into a appliance.

## Follow up

Now we know that AWS is doing interesting things with both hardware & software, how can it be applied to the home lab?
Well it can't because purchasing a pair of DPUs from Nvidia or Dell is beyond my budget. In addition to the $3000-$5000
cost per card, I also need to factor in the 200 watts of power used by each DPU.

I can get creative with networking by bonding multiple interfaces on the host into a single virtual interface.  
Bonding won't go beyond or above line speed, but it helps when there are multiple clients accessing
data on the same physical host.  Bonding is something I have experience with and I remember the entire connection can be 
fragile when one of the links from the bonded pair fails.  This was back in 2010, so hopefully bonding works 
better in 2023.

Similar to AWS, I am also using KVM with Qemu in my home lab.  For storage, Gluster FS has caught my eye.  With 
distributed filesystems, managing the filesystem metadata can be complicated and Gluster FS has a simple metadata model.
Like NFS or CIFS, Gluster FS is another file based system. While my goal was to implement a local block store, because 
Qemu has a Gluster driver built into it, the result is a file based storage system that appears as a block store to Qemu.

## References

Bouffler B. In the search for performance, there’s more than one way to build a network. Amazon Web Services. 
June 22, 2021.  Accessed August 20, 2023.
https://aws.amazon.com/blogs/hpc/in-the-search-for-performance-theres-more-than-one-way-to-build-a-network/

The Security Design of the AWS Nitro System. Amazon Web Services. November 18, 2022. Accessed August 20, 2023. 
https://docs.aws.amazon.com/whitepapers/latest/security-design-of-aws-nitro-system/security-design-of-aws-nitro-system.html

Baligh H. Serag E. Talaat S. Gaballah Y. DPUS: Acceleration Through Disaggregation. Dell Technologies. 2021. 
Accessed August 20, 2023
https://education.dell.com/content/dam/dell-emc/documents/en-us/2021KS_Baligh-DPUs_Acceleration_Through_Disaggregation.pdf

Disk Images. Qemu Project August 2023. Accessed August 20, 2023.
https://qemu-project.gitlab.io/qemu/system/images.html#nvme-disk-images


[^1]: Kubernetes can access block storage with Container Storage Interface (CSI) plugins. Docker volumes also have 
plugins that support block storage.
[^2]: I am not affiliated with AWS nor do I have internal contacts with AWS.  The content on this page is my 
interpretation from reading publicly available marketing documentation.  I feel I'm close enough in my interpretation 
for a layman's understanding of how EBS is implemented. 

