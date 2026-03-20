module github.com/ez-connect/webhook/plugins/echo

replace github.com/ez-connect/webhook/pkg/core => ../../pkg/core

replace github.com/ez-connect/webhook => ../..

go 1.26.1

require github.com/ez-connect/webhook/pkg/core v0.0.0-00010101000000-000000000000

require github.com/pelletier/go-toml/v2 v2.2.4 // indirect
