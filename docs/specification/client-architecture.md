# QuorumBD - Client Architecture

## Core Components

### NBD Integration
- Direct communication between kernel NBD driver and QuorumBD engine
- No intermediate client bridge - engine acts as NBD server backend
- Multiple NBD devices supported (one per volume)
- **Isolation**: Each volume uses separate socket connection on localhost
- **Parallel request handling**: Asynchronous protocol with handle-based response matching

### Volume Management
- **Multiple independent volumes** (GlusterFS-like model)
- Each volume maps to a dedicated NBD device
- **Strict separation**: Data volumes only (no meta-volumes)
- **Final output**: Ready-to-use LVM Thin Pools for Proxmox
- Engine manages volume-to-NBD mapping in state file
- **Single engine process** handles all volumes
- **Global LBA namespace**: Each volume gets unique LBA range to avoid offset collisions

### NBD Device Discovery
The engine automatically finds and uses available NBD devices:

1. **Scan `/sys/block/` directory** for entries starting with "nbd"
2. **Check device status** by reading `/sys/block/nbdX/size`
3. **Select first available device** where size equals "0" (unused)
4. **Reserve device** by storing mapping in state file

### Local Meta-Storage Management

#### Implementation Strategy
- **Go-focused implementation** with minimal external dependencies
- **Device-mapper integration** via Go libraries or syscalls
- **Filesystem operations** using Go libraries where available
- **External tool wrapping** for mature utilities (mkfs.ext4, etc.)

#### Device-Mapper Configuration
- **Engine-managed**: Creates/removes device-mapper devices programmatically
- **Node-type specific sizing** (from configuration):
  - Data Nodes: 20GB WAL + 40GB Mapping
  - Arbiter Nodes: 1GB WAL (no mapping)
- **Persistent sizing**: Final sizes stored in local state
- **Optimal placement**: Meta-storage allocated at beginning of storage device

#### Automated Setup Process
1. **Storage allocation**: Engine manages raw storage areas programmatically
2. **Device-mapper setup**: Creates `/dev/qb_meta_wal` and `/dev/qb_meta_map` devices
3. **Filesystem creation**: Formats mapping area with ext4 (data nodes only)
4. **Persistent configuration**: Stores layout in local state for reboot persistence

#### Meta-Storage Layout

**Data Nodes:**
- `/dev/qb_meta_wal` (20GB) : RAW device, direct block-I/O for WAL (sequential writes)
- `/dev/qb_meta_map` (40GB) : ext4 filesystem for mapping/B+Tree (complex structures)

**Arbiter Nodes:**
- `/dev/qb_meta_wal` (1GB) : RAW device, cluster metadata only
- No mapping storage (arbiters store no user data)

### LVM Integration Strategy
- Engine maintains **full control** over LVM configuration on NBD devices
- **UUID management**: Engine generates unique PV-UUID per volume to avoid conflicts
- **LVM tagging**: All engine-managed objects tagged with `quorumbd-volume=<name>` and `quorumbd-managed=true`
- **Clear naming**: `qbd_` prefix for all engine-managed objects

### Data Volumes (Volume 1, 2, ...) 
- **Purpose**: User data storage for Proxmox VMs
- **VG**: `qbd_vg_<volumename>`
- **LV Types**: Thin Pool LV only
- **Contents**: Thin-Pool-LV (exclusively for Proxmox VM data)
- **No metadata storage**: All metadata handled by local meta-storage
- **Admin access**: Thin Pools available for immediate VM deployment

### Administration Interface
- **REST API** for volume management (create, resize, status, cluster info)
- **Python CLI tool** for administrative operations
- Headless engine daemon controlled exclusively via API
- **Volume lifecycle**: Create, resize, delete, status monitoring

### System Integration
- Single **systemd service**: `quorumbd-engine.service`
- Engine handles own LVM activation during startup (`vgchange -ay` for all volume groups)
- **Local state management**: Critical state stored on OS disk (requires admin backups)
- **No external dependencies**: Pure Go implementation prioritized
- **Automatic recovery**: Engine restores meta-storage devices on startup

### Proxmox Workflow
1. Admin creates QuorumBD data-volume "vm-data" via REST API
2. Engine provisions complete LVM stack on NBD device
3. Engine runs `vgchange -ay` for all volume groups
4. **Proxmox auto-discovers** Thin Pool via LVM scan
5. Admin adds Thin Pool as LVM-Thin storage in Proxmox GUI
6. VMs created directly on Thin Pool with no additional configuration

### Resize Operations
- **Volume resize**: Admin updates volume size via REST API
- **LBA remapping**: Engine adjusts global LBA ranges logically
- **No data migration**: Physical blocks remain in place, only mapping changes
- **Online operation**: Resize possible while volume is in use

### Error Handling
- **NBD errors**: Standard NBD error codes returned to kernel
- **Storage errors**: Automatic retry and failover mechanisms
- **Cluster errors**: Quorum-based decision making for consistency
- **Admin notifications**: API-based error reporting and logging