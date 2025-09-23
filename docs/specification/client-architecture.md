# QuorumBD - Client Architecture

## Core Components

### NBD Integration
- Direct communication between kernel NBD driver and QuorumBD engine
- No intermediate client bridge - engine acts as NBD server backend
- Multiple NBD devices supported (one per volume)
- **Isolation**: Each volume uses separate socket connection on localhost

### Volume Management
- **Multiple independent volumes** (GlusterFS-like model)
- Each volume maps to a dedicated NBD device
- **Strict separation**: Data volumes only (no meta-volumes)
- **Final output**: Ready-to-use LVM Thin Pools for Proxmox
- Engine manages volume-to-NBD mapping in state file
- **Single engine process** handles all volumes

### NBD Device Discovery
The engine automatically finds and uses available NBD devices:

1. **Scan `/sys/block/` directory** for entries starting with "nbd"
2. **Check device status** by reading `/sys/block/nbdX/size`
3. **Select first available device** where size equals "0" (unused)
4. **Reserve device** by storing mapping in state file

### Local Meta-Storage Management

#### Device-Mapper Configuration
- **Engine-managed**: Creates/removes device-mapper devices on startup/shutdown
- **Node-type specific sizing** (from configuration):
  - Data Nodes: 20GB WAL + 40GB Mapping
  - Arbiter Nodes: 1GB WAL (no mapping)
- **Persistent sizing**: Final sizes stored in local state
- **Optimal placement**: Meta-storage allocated at beginning of storage device

#### Meta-Storage Layout

**Data Nodes:**
- /dev/qb_meta_wal (20GB) : RAW device, direct block-I/O for WAL
- /dev/qb_meta_map (40GB) : ext4 filesystem for mapping/B+Tree

**Arbiter Nodes:**
- /dev/qb_meta_wal (1GB) : RAW device, cluster metadata only
- No mapping storage (no data storage)

### LVM Integration Strategy
- Engine maintains **full control** over LVM configuration on NBD devices
- **UUID management**: Engine generates unique PV-UUID per volume to avoid conflicts
- **LVM tagging**: All engine-managed objects tagged with `quorumbd-volume=<name>` and `quorumbd-managed=true`

### Data Volumes (Volume 1, 2, ...) 
- **Purpose**: User data storage for Proxmox VMs
- **VG**: `qbd_vg_<volumename>`
- **LV Types**: Thin Pool LV only
- **Contents**: Thin-Pool-LV (exclusively for Proxmox VM data)
- **No metadata storage**: All metadata handled by local meta-storage

### Administration Interface
- **REST API** for volume management (create, resize, status)
- **Python CLI tool** for administrative operations
- Headless engine daemon controlled exclusively via API

### System Integration
- Single **systemd service**: `quorumbd-engine.service`
- Engine handles own LVM activation during startup
- **Local state management**: Critical state stored on OS disk (requires admin backups)
- No external dependencies or activation sequences required

### Proxmox Workflow
1. Admin creates QuorumBD data-volume "vm-data" via REST API
2. Engine provisions complete LVM stack on NBD device
3. Engine runs `vgchange -ay` for all volume groups
4. **Proxmox auto-discovers** Thin Pool via LVM scan
5. Admin adds Thin Pool as LVM-Thin storage in Proxmox GUI