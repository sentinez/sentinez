import { readFile } from 'node:fs/promises';
import path from 'node:path';

export async function GET(_request: Request, { params }: { params: Promise<{ path: string[] }> }) {
  const { path: segments } = await params;

  const filePath = path.join(process.cwd(), 'public', 'assets', ...segments);

  try {
    const file = await readFile(filePath);

    return new Response(file, {
      headers: {
        'Content-Type': getContentType(filePath),
        'Cache-Control': 'public, max-age=31536000, immutable',
      },
    });
  } catch {
    return new Response('Not Found', {
      status: 404,
    });
  }
}

function getContentType(filePath: string) {
  const ext = path.extname(filePath).toLowerCase();

  const types: Record<string, string> = {
    '.jpg': 'image/jpeg',
    '.jpeg': 'image/jpeg',
    '.png': 'image/png',
    '.gif': 'image/gif',
    '.webp': 'image/webp',
    '.svg': 'image/svg+xml',
    '.ico': 'image/x-icon',
    '.avif': 'image/avif',
  };

  return types[ext] ?? 'application/octet-stream';
}
