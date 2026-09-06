import{_ as a,c as n,o as p,ah as e}from"./chunks/framework.MgE-7MWY.js";const m=JSON.parse('{"title":"System architecture","description":"","frontmatter":{},"headers":[],"relativePath":"system-architecture.md","filePath":"system-architecture.md"}'),l={name:"system-architecture.md"};function i(t,s,c,r,o,h){return p(),n("div",null,[...s[0]||(s[0]=[e(`<h1 id="system-architecture" tabindex="-1">System architecture <a class="header-anchor" href="#system-architecture" aria-label="Permalink to “System architecture”">​</a></h1><div class="language-plaintext"><button title="Copy Code" class="copy"></button><span class="lang">plaintext</span><pre class="shiki shiki-themes github-light github-dark" style="--shiki-light:#24292e;--shiki-dark:#e1e4e8;--shiki-light-bg:#fff;--shiki-dark-bg:#24292e;" tabindex="0" dir="ltr"><code><span class="line"><span>Incoming Request</span></span>
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
<span class="line"><span>└──────────────────────┘       └─────────────────────────┘</span></span></code></pre></div><div class="language-"><button title="Copy Code" class="copy"></button><span class="lang"></span><pre class="shiki shiki-themes github-light github-dark" style="--shiki-light:#24292e;--shiki-dark:#e1e4e8;--shiki-light-bg:#fff;--shiki-dark-bg:#24292e;" tabindex="0" dir="ltr"><code><span class="line"><span>┌───────────────────────────────┐</span></span>
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
<span class="line"><span>└───────────────────────────────┘</span></span></code></pre></div>`,3)])])}const u=a(l,[["render",i]]);export{m as __pageData,u as default};
