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
- Engine manages volume-to-NBD mapping in state file
- **Single engine process** handles all volumes

### NBD Device Discovery
The engine automatically finds and uses available NBD devices:

1. **Scan `/sys/block/` directory** for entries starting with "nbd"
2. **Check device status** by reading `/sys/block/nbdX/size`
3. **Select first available device** where size equals "0" (unused)
4. **Reserve device** by storing mapping in state file

### LVM Integration Strategy
- Engine maintains **full control** over LVM configuration on NBD devices
- **UUID management**: Engine generates unique PV-UUID per volume to avoid conflicts
- **LVM tagging**: All engine-managed objects tagged with `quorumbd-volume=<name>` and `quorumbd-managed=true`
- Per-volume structure created automatically:
  - Volume Group (e.g., `qbd_vg_volumename`)
  - Fixed-size Meta LVs for engine internals (WAL, LBA mapping)
  - Thin Pool LV for user/Proxmox consumption

### Administration Interface
- **REST API** for volume management (create, resize, status)
- **Python CLI tool** for administrative operations
- Headless engine daemon controlled exclusively via API

### System Integration
- Single **systemd service**: `quorumbd-engine.service`
- Engine handles own LVM activation during startup
- No external dependencies or activation sequences required