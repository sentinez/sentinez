CLIENT_NS := client
GATEWAY_NS := gateway

CLIENT_IF := veth-client
DMZ_IF := veth0

DMZ_UPLINK := dmz-uplink
HOST_UPLINK := host-uplink

CLIENT_IP := 203.0.113.10/24
DMZ_IP := 203.0.113.1/24

HOST_UPLINK_IP := 192.168.100.1/24
DMZ_UPLINK_IP := 192.168.100.2/24

INTERNET_IF := $(shell ip route get 8.8.8.8 | grep -oP 'dev \K\S+' | head -1)

.PHONY: setup
setup: clean create

.PHONY: create
create:
	@echo "Create namespaces..."
	ip netns add $(CLIENT_NS)
	ip netns add $(GATEWAY_NS)

	@echo "Create public network..."
	ip link add $(CLIENT_IF) type veth peer name $(DMZ_IF)

	ip link set $(CLIENT_IF) netns $(CLIENT_NS)
	ip link set $(DMZ_IF) netns $(GATEWAY_NS)

	ip netns exec $(CLIENT_NS) \
		ip addr add $(CLIENT_IP) dev $(CLIENT_IF)

	ip netns exec $(GATEWAY_NS) \
		ip addr add $(DMZ_IP) dev $(DMZ_IF)

	ip netns exec $(CLIENT_NS) ip link set lo up
	ip netns exec $(CLIENT_NS) ip link set $(CLIENT_IF) up

	ip netns exec $(GATEWAY_NS) ip link set lo up
	ip netns exec $(GATEWAY_NS) ip link set $(DMZ_IF) up

	@echo "Create DMZ uplink..."
	ip link add $(DMZ_UPLINK) type veth peer name $(HOST_UPLINK)

	ip link set $(DMZ_UPLINK) netns $(GATEWAY_NS)

	ip addr add $(HOST_UPLINK_IP) dev $(HOST_UPLINK)
	ip link set $(HOST_UPLINK) up

	ip netns exec $(GATEWAY_NS) \
		ip addr add $(DMZ_UPLINK_IP) dev $(DMZ_UPLINK)

	ip netns exec $(GATEWAY_NS) \
		ip link set $(DMZ_UPLINK) up

	@echo "Routes..."
	ip netns exec $(CLIENT_NS) \
		ip route add default via 203.0.113.1

	ip netns exec $(GATEWAY_NS) \
		ip route add default via 192.168.100.1

	@echo "Enable forwarding..."
	sysctl -w net.ipv4.ip_forward=1

	@echo "Enable NAT..."
	iptables -t nat -C POSTROUTING \
		-s 192.168.100.0/24 \
		-o $(INTERNET_IF) \
		-j MASQUERADE 2>/dev/null || \
	iptables -t nat -A POSTROUTING \
		-s 192.168.100.0/24 \
		-o $(INTERNET_IF) \
		-j MASQUERADE

	@echo "Allow forwarding..."
	iptables -C FORWARD -i $(HOST_UPLINK) -j ACCEPT 2>/dev/null || \
	iptables -A FORWARD -i $(HOST_UPLINK) -j ACCEPT

	iptables -C FORWARD -o $(HOST_UPLINK) -j ACCEPT 2>/dev/null || \
	iptables -A FORWARD -o $(HOST_UPLINK) -j ACCEPT

.PHONY: attach-xdp
attach-xdp:
	ip netns exec $(GATEWAY_NS) \
		ip link set dev $(DMZ_IF) \
		xdp obj xdp.o sec xdp

.PHONY: detach-xdp
detach-xdp:
	ip netns exec $(GATEWAY_NS) \
		ip link set dev $(DMZ_IF) xdp off

.PHONY: shell-client
shell-client:
	ip netns exec $(CLIENT_NS) bash

.PHONY: shell-gateway
shell-gateway:
	ip netns exec $(GATEWAY_NS) bash

.PHONY: test-public
test-public:
	ip netns exec $(CLIENT_NS) \
		curl http://203.0.113.1:7443

.PHONY: test-internet
test-internet:
	ip netns exec $(GATEWAY_NS) \
		curl -I https://google.com

.PHONY: status
status:
	@echo "=== CLIENT ==="
	ip netns exec $(CLIENT_NS) ip addr
	ip netns exec $(CLIENT_NS) ip route

	@echo
	@echo "=== GATEWAY ==="
	ip netns exec $(GATEWAY_NS) ip addr
	ip netns exec $(GATEWAY_NS) ip route

	@echo
	@echo "=== XDP ==="
	ip netns exec $(GATEWAY_NS) bpftool net || true

.PHONY: clean
clean:
	-iptables -t nat -D POSTROUTING \
		-s 192.168.100.0/24 \
		-o $(INTERNET_IF) \
		-j MASQUERADE

	-ip netns del $(CLIENT_NS)
	-ip netns del $(GATEWAY_NS)

	-ip link del $(HOST_UPLINK)


curl:
	sudo ip netns exec client curl -vk https://203.0.113.1:7443

tracing:
	cat /sys/kernel/tracing/trace_pipe

tracing.debug:
	cat /sys/kernel/debug/tracing/trace_pipe