import createNextIntlPlugin from 'next-intl/plugin';

const withNextIntl = createNextIntlPlugin('./i18n/request.ts');

/** @type {import('next').NextConfig} */
const nextConfig = {
  transpilePackages: ["@sentinez/ui", "@sentinez/proto"],
  assetPrefix: "",
  sassOptions: {},
  redirects: () => {
    return [
      {
        source: "/",
        destination: "/auth",
        permanent: true
      },
      {
        source: "/console",
        destination: "/console/domain",
        permanent: true
      }
    ]
  }
}

export default withNextIntl(nextConfig)
