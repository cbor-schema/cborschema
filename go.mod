module github.com/cbor-schema/cborschema

go 1.20

require (
	github.com/fxamacker/cbor/v2 v2.5.0-beta2
	github.com/stretchr/testify v1.8.3
	github.com/x448/float16 v0.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/fxamacker/cbor/v2 v2.5.0-beta2 => github.com/ldclabs/cbor/v2 v2.5.0-stg5
