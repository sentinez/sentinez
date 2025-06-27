# System architecture

```plaintext
Incoming Request
        ↓
┌──────────────────────┐
│      WAF Proxy       │
│     (Go Service)     │
└────────────┬─────────┘
             ↓
┌──────────────────────┐
│     Rule Engine      │
│  (gRPC / REST API)   │
└────────────┬─────────┘
             ↓
     Forward or Reject
             ↓
┌──────────────────────┐       ┌─────────────────────────┐
│    Logging Stack     │◀────▶ │   Dashboard (React)     │
│    (ES / Prometheus) │       │ Live charts, metrics    │
└──────────────────────┘       └─────────────────────────┘

```

```
┌───────────────────────────────┐
│      User opens browser       │
│   https://tenant-a.example.com|
└────────────┬──────────────────┘
             ↓
┌───────────────────────────────┐
│        DNS Resolver           │
│ *.example.com → 203.0.113.42  │
└────────────┬──────────────────┘
             ↓
┌───────────────────────────────┐
│      Reverse Proxy Layer      │
│ (Go server or NGINX, Caddy)   │
│ Routes all domains to the app │
└────────────┬──────────────────┘
             ↓
┌───────────────────────────────┐
│       Multi-Tenant App        │
│ Reads ctx.Host() = tenant-a   │
│ → handles logic per tenant    │
└───────────────────────────────┘

```