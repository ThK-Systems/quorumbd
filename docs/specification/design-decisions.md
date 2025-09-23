# QuorumBD - Design Decisions

## Foundational Choices

### Target Audience
- Designed for **professional administrators** who understand cluster filesystems
- No "idiot-proofing" - clear conventions and documentation over forced restrictions
- Admin is treated as a partner in operations
- **Admin responsibility**: Regular backups of local state on OS disk

### Storage Abstraction
- **Network Block Device (NBD)** over filesystem-level implementation
- Leverages existing Linux stack (LVM, thin provisioning) rather than reinventing storage management
- **Proxmox integration via ready-to-use LVM Thin Pools** provisioned by the engine

### Cluster Management
- **Arbiter nodes** as first-class concept from inception
- Flexible node roles: any mix of data and arbiter nodes, with odd total count requirement
- **Arbiter functionality**: **GlusterFS-like model** - participate in voting and store namespace/metadata only (no file data)
- Consensus mechanism must work with asymmetric node topologies and storage roles

### Metadata Architecture
- **Local meta-storage** on each node's OS disk (not replicated)
- **Device-mapper based** allocation with node-type specific sizing
- **Separation of concerns**: 
  - RAW WAL device for maximum performance (sequential writes)
  - ext4 filesystem for complex mapping structures (data nodes only)
- **Metadata placement**: Stored at the beginning of the storage device for optimal performance
- **No recursion**: Meta-storage managed separately from data volume I/O path
- **State persistence**: Critical configuration stored locally (admin backup required)

### Configuration Philosophy
- **Clear separation** between static configuration (`/etc/quorumbd/`) and dynamic state (`/var/lib/quorumbd/`)
- Static configuration defines **what** (volumes, cluster layout, meta-storage sizing)
- Dynamic state tracks **how** (runtime mappings, device assignments, final sizes)
- **Local state criticality**: Engine state requires regular backups by admin
- Engine maintains full control over its internal storage layout
- Admin works with ready-to-use Thin Pools provided by the engine

### Security and Networking
- **VLAN segregation** for cluster communication (mandatory)
- **Localhost-only** communication between NBD client and engine (no VLAN overhead)
- Clear separation of management, cluster, and client traffic

### Implementation Approach
- **Pure Go implementation** where possible
- **Minimal external dependencies** for core functionality
- **Practical V1 approach**: Use external tools where Go libraries are immature
- **Long-term vision**: Migrate to pure Go implementation as libraries mature