module quorumbd.net/middleware-qemu-nbd

go 1.26.0

require (
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0
	github.com/pelletier/go-toml/v2 v2.2.4
	quorumbd.net/common v0.0.0-00010101000000-000000000000
	quorumbd.net/middleware-common v0.0.0-00010101000000-000000000000
)

replace quorumbd.net/common => ../common

replace quorumbd.net/middleware-common => ../middleware-common
