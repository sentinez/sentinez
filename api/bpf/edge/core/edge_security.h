/*
   Copyright 2025 Duc-Hung Ho

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

#ifndef EDGE_SECURITY_H
#define EDGE_SECURITY_H

#include <linux/bpf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <bpf/bpf_helpers.h>

#define MAX_BLOCK 128   // number of IP / CIDR can block

struct ip_rule {
    __u32 ip;    // network prefix
    __u32 mask;  // CIDR mask
};

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, MAX_BLOCK);
    __type(key, __u32);
    __type(value, struct ip_rule);
} blocklist SEC(".maps");

static __always_inline int security_rule_handler(struct xdp_md *ctx) {
    bpf_printk("[quadrum] security rule");

    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_ABORTED;

    // IPv4
    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return XDP_PASS;

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_ABORTED;

    __u32 src_ip = iph->saddr;   // network byte order

    // --- CHECK BLOCKLIST ---
    for (__u32 i = 0; i < MAX_BLOCK; i++) {
        __u32 idx = i; // <-- use a local, initialized stack slot for map key

        struct ip_rule *r = bpf_map_lookup_elem(&blocklist, &idx);
        if (!r)
            continue;

        if (r->mask == 0)
            continue;   // rule empty

        // match: (ip & mask) == prefix
        if ((src_ip & r->mask) == r->ip) {
            bpf_printk("[quadrum] BLOCKED: ip=%x\n", src_ip);
            return XDP_DROP;
        }
    }

    return XDP_PASS;
}

#endif
