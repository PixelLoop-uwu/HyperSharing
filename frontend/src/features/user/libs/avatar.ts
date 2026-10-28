export function getAvatarUrl(
  userId: string | number,
  version: string | null,
): string | null {
  if (!version) return null

  const id = encodeURIComponent(String(userId))
  const params = new URLSearchParams({ v: version })

  return `/avatars/${id}?${params.toString()}`
}