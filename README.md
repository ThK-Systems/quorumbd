# QuorumBD - Distributed Network Block Device

## Project Overview
QuorumBD is a network block device solution designed as a modern replacement for GlusterFS, which is facing declining development. It targets homelab and small-to-medium business (SMB) environments with a focus on being a lightweight, reliable distributed storage system.

## Core Philosophy
The system prioritizes **data consistency** above all else, while maintaining performance at least equivalent to GlusterFS. It is built with the philosophy of "working with the admin, not against them," assuming operators understand core concepts of clustering and filesystems.

## Key Specifications
- **Implementation Language**: Go
- **Node Scale**: 3 to 9 nodes (odd number required for consensus)
- **Network Environment**: 1-2.5 Gigabit Ethernet
- **Node Types**: Flexible mix of Data Nodes (store data) and Arbiter Nodes (participate in voting only)
- **Resource Constraints**: Maximum 300-400 MB RAM usage
- **Version 1 Scope**: No erasure coding, primary integration with Proxmox

## Architecture Highlights
- **Multi-volume support** with independent LVM Thin Pools
- **Dedicated metadata volume** for cluster-wide management
- **NBD-based block storage** with direct Proxmox integration
- **Flexible node roles** supporting arbiters for resource-efficient operation

## Status
*Early design phase - concept development in progress*
