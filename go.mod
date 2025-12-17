module github.com/sentinez/sentinez

go 1.25.3

replace (
	github.com/sentinez/core => ./staging/src/github.com/sentinez/core
	github.com/sentinez/sentinez/api => ./api
	github.com/sentinez/shared => ./staging/src/github.com/sentinez/shared
	github.com/sentinez/x/httphz => ./staging/src/github.com/sentinez/x/httphz
)

require (
	buf.build/go/protovalidate v1.1.0
	github.com/Masterminds/squirrel v1.5.4
	github.com/a-h/templ v0.3.960
	github.com/bytedance/sonic v1.14.2
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/corazawaf/coraza/v3 v3.3.3
	github.com/go-webauthn/webauthn v0.15.0
	github.com/goccy/go-yaml v1.19.1
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/go-cmp v0.7.0
	github.com/gorilla/websocket v1.5.3
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3
	github.com/jackc/pgx/v5 v5.7.6
	github.com/jmoiron/sqlx v1.4.0
	github.com/joho/godotenv v1.5.1
	github.com/pashagolub/pgxmock/v2 v2.12.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/sentinez/core v0.0.0-00010101000000-000000000000
	github.com/sentinez/sentinez/api v0.0.0
	github.com/sentinez/shared v0.0.0-00010101000000-000000000000
	github.com/sony/gobreaker v1.0.0
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.11.1
	golang.org/x/crypto v0.46.0
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20251209175733-2a1774d88802.1 // indirect
	cel.dev/expr v0.25.1 // indirect
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/bytedance/gopkg v0.1.3 // indirect
	github.com/bytedance/sonic/loader v0.4.0 // indirect
	github.com/cloudresty/ulid v1.2.1 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/corazawaf/libinjection-go v0.2.3 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/go-webauthn/x v0.1.26 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/google/go-tpm v0.9.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/consul/api v1.33.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-metrics v0.5.4 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/serf v0.10.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/magefile/mage v1.15.1-0.20241126214340-bdc92f694516 // indirect
	github.com/matoous/go-nanoid/v2 v2.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/petar-dambovaliev/aho-corasick v0.0.0-20250424160509-463d218d4745 // indirect
	github.com/planetscale/vtprotobuf v0.6.1-0.20240319094008-0393e58bdf10 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/stoewer/go-strcase v1.3.1 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/valllabh/ocsf-schema-golang v1.0.3 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/fx v1.24.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.1 // indirect
	golang.org/x/arch v0.23.0 // indirect
	golang.org/x/exp v0.0.0-20251209150349-8475f28825e9 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	rsc.io/binaryregexp v0.2.0 // indirect
)
