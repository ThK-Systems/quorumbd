# QuorumBD - Design Decisions

## Foundational Choices

### Target Audience
- Designed for **professional administrators** who understand cluster filesystems
- No "idiot-proofing" - clear conventions and documentation over forced restrictions
- Admin is treated as a partner in operations

### Storage Abstraction
- **Network Block Device (NBD)** over filesystem-level implementation
- Leverages existing Linux stack (LVM, thin provisioning) rather than reinventing storage management
- **Proxmox integration** via LVM Thin Pool provisioned by the engine

### Cluster Management
- **Arbiter nodes** as first-class concept from inception
- Flexible node roles: any mix of data and arbiter nodes, with odd total count requirement
- **Arbiter functionality**: **GlusterFS-like model** - participate in voting and store namespace/metadata only (no file data)
- Consensus mechanism must work with asymmetric node topologies and storage roles

### Metadata Architecture
- **Dedicated Meta-Volume** for cluster-wide metadata (WAL, global LBA mapping)
- **Separation of concerns**: Metadata operations isolated from data volume I/O path
- **Prevents circular dependencies**: Metadata writes don't go through the WAL they manage
- **Volume-based data**: Each data volume contains only user data blocks

### Configuration Philosophy
- **Clear separation** between static configuration (`/etc/quorumbd/`) and dynamic state (`/var/lib/quorumbd/`)
- Static configuration defines **what** (volumes, cluster layout)
- Dynamic state tracks **how** (runtime mappings, device assignments)
- Engine maintains full control over its internal storage layout
- Admin works with ready-to-use Thin Pools provided by the engine

### Security and Networking
- **VLAN segregation** for cluster communication (mandatory)
- **Localhost-only** communication between NBD client and engine (no VLAN overhead)
- Clear separation of management, cluster, and client traffic