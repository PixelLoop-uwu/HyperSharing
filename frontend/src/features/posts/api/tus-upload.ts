const TUS_VERSION = '1.0.0'
const CHUNK_SIZE = 8 * 1024 * 1024

export interface TusMeta {
  draftId: string
  fileId: string
}

function base64(value: string): string {
  // Идентификаторы draftId и fileId в текущем контракте ASCII.
  return btoa(value)
}

function throwIfAborted(signal?: AbortSignal): void {
  if (signal?.aborted) {
    throw signal.reason ?? new DOMException('Aborted', 'AbortError')
  }
}

export async function uploadFileViaTus(
  file: File,
  meta: TusMeta,
  onProgress?: (uploaded: number, total: number) => void,
  signal?: AbortSignal,
): Promise<void> {
  throwIfAborted(signal)

  const createRes = await fetch('/api/v1/uploads', {
    method: 'POST',
    headers: {
      'Tus-Resumable': TUS_VERSION,
      'Upload-Length': String(file.size),
      'Upload-Metadata':
        `draftId ${base64(meta.draftId)},` +
        `fileId ${base64(meta.fileId)}`,
    },
    signal,
  })

  if (createRes.status !== 201) {
    throw new Error(
      `Не удалось начать загрузку «${file.name}» (${createRes.status})`,
    )
  }

  const location = createRes.headers.get('Location')

  if (!location) {
    throw new Error('Сервер не вернул адрес загрузки')
  }

  /*
   * За reverse proxy Location может содержать адрес внутреннего сервера.
   * Используем только путь, чтобы PATCH шёл через тот же origin,
   * что и приложение, вместе с его авторизационной cookie.
   */
  const uploadURL = new URL(location, window.location.origin)
  const uploadPath = uploadURL.pathname + uploadURL.search

  let offset = 0

  while (offset < file.size) {
    throwIfAborted(signal)

    const end = Math.min(offset + CHUNK_SIZE, file.size)

    const res = await fetch(uploadPath, {
      method: 'PATCH',
      headers: {
        'Tus-Resumable': TUS_VERSION,
        'Upload-Offset': String(offset),
        'Content-Type': 'application/offset+octet-stream',
      },
      body: file.slice(offset, end),
      signal,
    })

    if (res.status !== 204) {
      throw new Error(
        `Загрузка «${file.name}» не удалась (${res.status})`,
      )
    }

    // Проверяем серверное смещение, а не предполагаем, что приняты все байты.
    const serverOffset = res.headers.get('Upload-Offset')

    if (serverOffset === null || Number(serverOffset) !== end) {
      throw new Error(
        `Сервер вернул некорректное смещение для «${file.name}»`,
      )
    }

    offset = end
    onProgress?.(offset, file.size)
  }

  // В том числе сообщает о завершении пустого файла.
  if (file.size === 0) {
    onProgress?.(0, 0)
  }
}