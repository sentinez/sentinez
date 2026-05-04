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

#include "edge_helper.h"

#define MAX_BLOCK 128   // number of IP / CIDR can block

struct ip_lpm_key {
    __u32 prefixlen; // number of bits in mask (e.g. 24 for /24)
    __u32 ip;        // network byte order
};

struct {
    __uint(type, BPF_MAP_TYPE_LPM_TRIE);
    __uint(max_entries, 1024);
    __uint(map_flags, BPF_F_NO_PREALLOC);
    __type(key, struct ip_lpm_key);
    __type(value, __u8); // dummy value (just mark blocked)
} blocklist SEC(".maps");

static __always_inline int security_rule_handler(struct xdp_md *ctx) {
    // debug("security rule");

    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    // Ethernet header
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_ABORTED;

    // Only IPv4
    if (eth->h_proto != __constant_htons(ETH_P_IP))
        return XDP_PASS;

    // IP header
    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_ABORTED;

    __u32 src_ip = iph->saddr; // network byte order
    // debug("security_rule_handler: src_ip=%x", src_ip);

    // --- CHECK BLOCKLIST ---
    struct ip_lpm_key key = {
        .prefixlen = 32, // full IP for lookup
        .ip = src_ip,
    };

    __u8 *blocked = bpf_map_lookup_elem(&blocklist, &key);
    if (blocked) {
        debug("BLOCKED: ip=%x\n", src_ip);
        return XDP_DROP;
    }

    return XDP_PASS;
}

#endif
