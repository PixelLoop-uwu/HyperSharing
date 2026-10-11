const UNITS = ['B', 'KiB', 'MiB', 'GiB', 'TiB'] as const

/** Human-readable byte size (binary units). */
export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), UNITS.length - 1)
  const value = bytes / 1024 ** i
  return `${value.toFixed(value >= 10 || i === 0 ? 0 : 1)} ${UNITS[i]}`
}

/** Short localized timestamp for list items and headers. */
export function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString('ru-RU', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
