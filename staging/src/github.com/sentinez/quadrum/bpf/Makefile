# TARGET = xdp_lb
# TARGET = packetdrop
TARGET = xdp_quadrum

# For xdp_liz, make and also make user. The others don't have userspace programs

USER_TARGET = ${TARGET:=_user}
BPF_TARGET = ${TARGET:=_kern}
BPF_C = ${BPF_TARGET:=.c}
BPF_OBJ = ${BPF_C:.c=.o}

xdp: $(BPF_OBJ)
	bpftool net detach xdpgeneric dev veth0
	rm -f /sys/fs/bpf/$(TARGET)
	bpftool prog load $(BPF_OBJ) /sys/fs/bpf/$(TARGET)
	bpftool net attach xdpgeneric pinned /sys/fs/bpf/$(TARGET) dev veth0 

user: $(USER_TARGET)

$(USER_TARGET): %: %.c  
	gcc -Wall $(CFLAGS) -I../bpf/include -I../bpf/include/bpf/include/uapi -L../bpf/include -o $@  \
	 $< -l:libbpf.a -lelf -lz

$(BPF_OBJ): %.o: %.c
	clang -S \
	    -target bpf \
	    -D __BPF_TRACING__ \
	    -I../bpf/include\
	    -Wall \
	    -Wno-unused-value \
	    -Wno-pointer-sign \
	    -Wno-compare-distinct-pointer-types \
	    -Werror \
	    -O2 -emit-llvm -c -o ${@:.o=.ll} $<
	llc -march=bpf -filetype=obj -o $@ ${@:.o=.ll}

clean:
	bpftool net detach xdpgeneric dev veth0
	rm -f /sys/fs/bpf/$(TARGET)
	rm $(BPF_OBJ)
	rm ${BPF_OBJ:.o=.ll}




veth.init:
	ip link add veth0 type veth peer name veth1

veth.add:
	ip addr add 10.0.0.1/24 dev veth0
	ip addr add 10.0.0.2/24 dev veth1

veth.up:
	ip link set veth0 up
	ip link set veth1 up

tracing:
	cat /sys/kernel/tracing/trace_pipe

tracing.debug:
	cat /sys/kernel/debug/tracing/trace_pipe