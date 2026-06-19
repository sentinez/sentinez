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

//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#include "senz_security.h"
#include "senz_monitor.h"

SEC("xdp")
int senz_main(struct xdp_md *ctx)
{
    enum xdp_action action = bandwidth_handler(ctx);
    if (action != XDP_PASS) {
        return action;
    }

    return security_rule_handler(ctx);
}

char LICENSE[] SEC("license") = "GPL";
