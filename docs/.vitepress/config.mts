import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "Sentinéz Engineering",
  description: "engineering documents",
  themeConfig: {
    logo: '/logo.png'
  },

  locales: {
    root: {
      label: 'English',
      lang: 'en',
      themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        nav: [
          { text: 'Home', link: '/' },
          { text: 'Style Guide', link: '/uber-guide' }
        ],

        sidebar: [
          {
            text: 'Introduction',
            items: [
              { text: 'Engineering', link: '/markdown-examples' },
              { text: 'Runtime API Examples', link: '/api-examples' },
              { text: 'Performance', link: '/edge-stats' }
            ]
          }
        ],

        socialLinks: [
          { icon: 'github', link: 'https://github.com/sentinez' }
        ]
      },
    },

    vi: {
      label: 'Tiếng Việt',
      lang: 'vi',
      link: '/vi/',
      themeConfig: {
        // https://vitepress.dev/reference/default-theme-config
        nav: [
          { text: 'Home', link: '/vi/' },
          { text: 'Examples', link: '/markdown-examples' }
        ],

        sidebar: [
          {
            text: 'Examples',
            items: [
              { text: 'Kỹ thuật', link: '/markdown-examples' },
              { text: 'Runtime API Examples', link: '/api-examples' }
            ]
          }
        ],

        socialLinks: [
          { icon: 'github', link: 'https://github.com/sentinez' }
        ]
      },
    }
  }
})
