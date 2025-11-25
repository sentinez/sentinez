#include "xdp_quadrum_kern.h"

#define SAMPLE_LEN 128
#define MAX_ENTRIES 131072

struct flow_key
{
    __u32 src_ip;
    __u32 des_ip;
    __u16 src_port;
    __u16 des_port;
    __u8 proto;
    __u8 pad[3];
};

struct flow_event {
    struct flow_key key;
    __u32 payload_len;
    char payload[SAMPLE_LEN];
};

struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __type(key, struct flow_key);
};

SEC("xdp")
int xdp_quadrum_prog(struct xdp_md *ctx)
{

    // pointer of the start of the packet data
    void *data = (void *)(long)ctx->data;

    // pointer of the end of the packet data
    void *data_end = (void *)(long)ctx->data_end;

    // ethernet header size: 14 bytes
    // ip header size: 20 bytes
    // tcp/udp header size: 20 bytes

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_ABORTED;

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_ABORTED;

    int pkg_size = data_end - data;
    bpf_printk("[quadrum] packet size: %d bytes\n", pkg_size);

    return XDP_PASS;
}
