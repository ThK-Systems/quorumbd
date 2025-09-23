# QuorumBD - Distributed Network Block Device

## Project Overview
QuorumBD is a network block device solution designed as a modern replacement for GlusterFS, which is facing declining development. It targets homelab and small-to-medium business (SMB) environments with a focus on being a lightweight, reliable distributed storage system optimized for **consumer hardware**.

## Core Philosophy
The system prioritizes **data consistency** above all else, while maintaining performance at least equivalent to GlusterFS. It is built with the philosophy of "working with the admin, not against them," assuming operators understand core concepts of clustering and filesystems.

## Key Specifications
- **Implementation Language**: Go
- **Node Scale**: 3 to 9 nodes (odd number required for consensus)
- **Minimum Requirements**: At least 2 data nodes + 1 arbiter node
- **Network Environment**: 1-2.5 Gigabit Ethernet
- **Hardware Target**: Consumer-grade SSDs and hardware
- **Node Types**: 
  - **Data Nodes**: Store actual block data (minimum 2 required)
  - **Arbiter Nodes**: Participate in voting and store metadata only (GlusterFS-like model)
- **Resource Constraints**: Maximum 300-400 MB RAM usage
- **Version 1 Scope**: No erasure coding, primary integration with Proxmox

## Architecture Highlights
- **Multi-volume support** with independent LVM Thin Pools
- **Ready-to-use LVM Thin Pools** exposed to Proxmox for immediate VM deployment
- **Local meta-storage** management with device-mapper
- **NBD-based block storage** with direct Proxmox integration
- **Flexible node roles** supporting arbiters for resource-efficient operation
- **Optimized for consumer SSDs** without enterprise hardware requirements
- **Arbiter nodes store metadata** reducing storage requirements for consensus nodes

## Proxmox Integration
QuorumBD volumes are presented to Proxmox as **fully configured LVM Thin Pools**, allowing administrators to immediately create virtual machines without additional storage configuration.

## Status
*Early design phase - concept development in progress*

## Quick Start (Planned)
1. Install QuorumBD package on all nodes
2. Configure cluster in `/etc/quorumbd/quorumbd.toml`
3. Start `quorumbd-engine.service`
4. Create volumes via REST API or CLI tool
5. Add discovered Thin Pools in Proxmox storage configuration