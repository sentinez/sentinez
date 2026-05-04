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

#ifndef EDGE_MONITOR_H
#define EDGE_MONITOR_H

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "edge_helper.h"
#include <linux/if_ether.h>
#include <linux/ip.h>

struct {
    __uint(type, BPF_MAP_TYPE_LRU_PERCPU_HASH);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 10240);
} ip_bandwidth SEC(".maps");

static __always_inline int bandwidth_handler(struct xdp_md *ctx) {
    // debug("monitor: bandwidth_handler");

    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;

    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end) {
        return XDP_PASS;
    }

    if (eth->h_proto != __constant_htons(ETH_P_IP)) {
        debug("ARP packet detected");
        return XDP_PASS;
    }

    struct iphdr *iph = (void *)(eth + 1);
    if ((void *)(iph + 1) > data_end)
        return XDP_PASS;

    __u32 src_ip = iph->saddr;
    __u64 pkt_len = data_end - data;

    debug("monitor: bandwidth_handler: src_ip=%x", src_ip);

    __u64 *bytes = bpf_map_lookup_elem(&ip_bandwidth, &src_ip);
    if (bytes) {
        *bytes += pkt_len;
    } else {
        bpf_map_update_elem(&ip_bandwidth, &src_ip, &pkt_len, BPF_ANY);
    }

    return XDP_PASS;
}

#endif