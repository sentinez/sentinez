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

export default nextConfig
