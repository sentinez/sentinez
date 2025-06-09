# System architecture

```plaintext
Incoming Request
        ↓
+-----------------+
|    WAF Proxy    |
|   (Go Service)  |
+-----------------+
        ↓
+-----------------+
|   Rule Engine   |
| (gRPC/REST API) |
+-----------------+
        ↓
 Forward or Reject
        ↓
+-----------------+      +--------------------+
|  Logging Stack  | <--> |  Dashboard (React) |
|   (ES / Prom)   |      | Live charts, table |
+-----------------+      +--------------------+
```           