#ifndef EDGE_LOG_H
#define EDGE_LOG_H

#define debug(fmt, ...) \
    bpf_printk("[senz:edge][%s:%d] " fmt, __FILE__, __LINE__, ##__VA_ARGS__)

#endif
