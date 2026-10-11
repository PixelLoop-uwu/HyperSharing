import { FiArchive, FiFile, FiFileText, FiFilm, FiImage, FiMusic } from 'react-icons/fi'

const ARCHIVE_NAME = /\.(zip|rar|7z|tar|gz|tgz|bz2|xz)$/i
const ARCHIVE_TYPE = /zip|x-tar|gzip|x-7z|x-rar|compressed/
const DOCUMENT_NAME = /\.(pdf|docx?|txt|md|rtf|odt)$/i

export default function FileTypeIcon({
  type,
  name,
  className,
}: {
  type: string
  name: string
  className?: string
}) {
  const Icon = type.startsWith('image/')
    ? FiImage
    : type.startsWith('video/')
      ? FiFilm
      : type.startsWith('audio/')
        ? FiMusic
        : ARCHIVE_NAME.test(name) || ARCHIVE_TYPE.test(type)
          ? FiArchive
          : DOCUMENT_NAME.test(name) || type === 'application/pdf' || type.startsWith('text/')
            ? FiFileText
            : FiFile
  return <Icon aria-hidden="true" className={className} />
}
