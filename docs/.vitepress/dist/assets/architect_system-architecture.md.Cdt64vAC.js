import{G as e,W as t,n,rt as r}from"./chunks/framework.DBm3e_pg.js";var i=JSON.parse(`{"title":"System architecture","description":"","frontmatter":{},"headers":[],"relativePath":"architect/system-architecture.md","filePath":"architect/system-architecture.md"}`),a={name:`architect/system-architecture.md`};function o(n,i,a,o,s,c){return r(),t(`div`,null,[...i[0]||=[e(`<h1 id="system-architecture" tabindex="-1">System architecture <a class="header-anchor" href="#system-architecture" aria-label="Permalink to “System architecture”">​</a></h1><div class="language-plaintext"><button title="Copy code" data-copied="Copied" class="copy"></button><span class="lang">plaintext</span><pre class="shiki shiki-themes github-light github-dark" style="--shiki-light:#24292e;--shiki-dark:#e1e4e8;--shiki-light-bg:#fff;--shiki-dark-bg:#24292e;" tabindex="0" dir="ltr"><code><span class="line"><span>Incoming Request</span></span>
<span class="line"><span>        ↓</span></span>
<span class="line"><span>┌──────────────────────┐</span></span>
<span class="line"><span>│      WAF Proxy       │</span></span>
<span class="line"><span>│     (Go Service)     │</span></span>
<span class="line"><span>└────────────┬─────────┘</span></span>
<span class="line"><span>             ↓</span></span>
<span class="line"><span>┌──────────────────────┐</span></span>
<span class="line"><span>│     Rule Engine      │</span></span>
<span class="line"><span>│  (gRPC / REST API)   │</span></span>
<span class="line"><span>└────────────┬─────────┘</span></span>
<span class="line"><span>             ↓</span></span>
<span class="line"><span>     Forward or Reject</span></span>
<span class="line"><span>             ↓</span></span>
<span class="line"><span>┌──────────────────────┐       ┌─────────────────────────┐</span></span>
<span class="line"><span>│    Logging Stack     │◀────▶ │   Dashboard (React)     │</span></span>
<span class="line"><span>│    (ES / Prometheus) │       │ Live charts, metrics    │</span></span>
<span class="line"><span>└──────────────────────┘       └─────────────────────────┘</span></span></code></pre></div><div class="language-"><button title="Copy code" data-copied="Copied" class="copy"></button><span class="lang"></span><pre class="shiki shiki-themes github-light github-dark" style="--shiki-light:#24292e;--shiki-dark:#e1e4e8;--shiki-light-bg:#fff;--shiki-dark-bg:#24292e;" tabindex="0" dir="ltr"><code><span class="line"><span>┌───────────────────────────────┐</span></span>
<span class="line"><span>│      User opens browser       │</span></span>
<span class="line"><span>│   https://tenant-a.example.com|</span></span>
<span class="line"><span>└────────────┬──────────────────┘</span></span>
<span class="line"><span>             ↓</span></span>
<span class="line"><span>┌───────────────────────────────┐</span></span>
<span class="line"><span>│        DNS Resolver           │</span></span>
<span class="line"><span>│ *.example.com → 203.0.113.42  │</span></span>
<span class="line"><span>└────────────┬──────────────────┘</span></span>
<span class="line"><span>             ↓</span></span>
<span class="line"><span>┌───────────────────────────────┐</span></span>
<span class="line"><span>│      Reverse Proxy Layer      │</span></span>
<span class="line"><span>│ (Go server or NGINX, Caddy)   │</span></span>
<span class="line"><span>│ Routes all domains to the app │</span></span>
<span class="line"><span>└────────────┬──────────────────┘</span></span>
<span class="line"><span>             ↓</span></span>
<span class="line"><span>┌───────────────────────────────┐</span></span>
<span class="line"><span>│       Multi-Tenant App        │</span></span>
<span class="line"><span>│ Reads ctx.Host() = tenant-a   │</span></span>
<span class="line"><span>│ → handles logic per tenant    │</span></span>
<span class="line"><span>└───────────────────────────────┘</span></span></code></pre></div>`,3)]])}var s=n(a,[[`render`,o]]);export{i as __pageData,s as default};