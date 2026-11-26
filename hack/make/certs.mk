# The certificates of the internal mTLS (edge -> centraldata); included
# by the root Makefile, before the cmd/*/Makefile whose sz.<service>.cert
# issue them with hack/mkcert-internal.sh. The CA and every certificate
# share one folder (gitignored), which deploy/docker mounts file by file.
SZ_INTERNAL_CERT_DIR ?= $(SZ_ROOT)/deploy/certs/internal

# A CN is the name of the service, identity.* of core/szgrpc/client: the
# name its peers verify its certificate by, and the name of its service
# token (szservicetoken -name). The edges of a cluster share theirs.
SZ_CENTRALDATA_CERT_NAME ?= sentinez.apps.centraldata
SZ_EDGE_CERT_NAME ?= sentinez.dmz.edge
# names centraldata is reached by besides its CN
SZ_CENTRALDATA_SAN ?= localhost 127.0.0.1
