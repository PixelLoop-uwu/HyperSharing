import { sha256Hex } from '@/shared/lib/hash'
import type { FileUploadOut, PartEtagIn } from './types'

/**
 * Upload chunk size in bytes. Must match the backend's
 * ``settings.s3_multipart_part_size``: the client hashes each part *before*
 * ``POST /posts/draft``, so the size is pinned here rather than read from the
 * API. The server echoes the same value back as ``FileUploadOut.part_size``.
 */
export const PART_SIZE = 8 * 1024 * 1024

export interface FileDigest {
  hash: string
  size: number
  content_type: string | null
  part_checksums: string[]
}

/**
 * Compute the whole-file SHA-256 plus one SHA-256 per part (in part order).
 * Reads the file once into memory — fine for the per-file size ceiling.
 */
export async function computeFileDigest(file: File): Promise<FileDigest> {
  const buffer = await file.arrayBuffer()
  const hash = await sha256Hex(buffer)
  const partCount = Math.ceil(file.size / PART_SIZE)
  const part_checksums: string[] = []
  for (let i = 0; i < partCount; i++) {
    const start = i * PART_SIZE
    const end = Math.min(file.size, start + PART_SIZE)
    part_checksums.push(await sha256Hex(buffer.slice(start, end)))
  }
  return {
    hash,
    size: file.size,
    content_type: file.type || null,
    part_checksums,
  }
}

/** PUT one part straight to S3; returns the quoted ETag to hand back on publish. */
async function uploadPart(url: string, checksum_sha256: string, body: Blob): Promise<string> {
  const res = await fetch(url, {
    method: 'PUT',
    headers: {
      'x-amz-checksum-sha256': checksum_sha256,
      'x-amz-sdk-checksum-algorithm': 'SHA256',
    },
    body,
  })
  if (!res.ok) {
    throw new Error(`загрузка части не удалась (${res.status})`)
  }
  const etag = res.headers.get('ETag')
  if (!etag) throw new Error('S3 не вернул ETag в ответе загрузки')
  return etag
}

/**
 * Upload every part of ``file`` for a presigned ``upload`` and collect ETags.
 * ``onProgress`` fires after each part completes with (done, total) for the file.
 */
export async function uploadFileParts(
  file: File,
  upload: FileUploadOut,
  onProgress?: (doneParts: number, totalParts: number) => void,
): Promise<PartEtagIn[]> {
  const parts: PartEtagIn[] = []
  for (const part of upload.parts) {
    const start = (part.part_number - 1) * upload.part_size
    const end = Math.min(file.size, start + upload.part_size)
    const etag = await uploadPart(part.url, part.checksum_sha256, file.slice(start, end))
    parts.push({ part_number: part.part_number, etag })
    onProgress?.(parts.length, upload.parts.length)
  }
  return parts
}
