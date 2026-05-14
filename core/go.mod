module quorumbd.net/core

go 1.26.0

require (
	github.com/google/uuid v1.6.0
	github.com/pelletier/go-toml/v2 v2.3.0
	quorumbd.net/common v0.0.0-00010101000000-000000000000
)

require github.com/go-ozzo/ozzo-validation/v4 v4.3.0

require (
	go.etcd.io/bbolt v1.4.3 // indirect
	golang.org/x/sys v0.29.0 // indirect
)

replace quorumbd.net/common => ../common
