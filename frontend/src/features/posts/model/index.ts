import { initDraftPersistence } from './persist'
import { startDigestWatcher } from './digests'

export { useDraftStore } from './draft'
export { useDigests, useFileDigest, type DigestState } from './digests'
export { usePublishStore, type PublishStage } from './publish'
export { buildItems, type BuildResult } from './buildItems'
export { isBlank } from './sections'
export type { DraftData, FileRef, FileTarget, Section } from './types'

/** Call once when the editor mounts (idempotent). Loads the saved draft, then starts hashing. */
export async function initPostEditor(): Promise<void> {
  await initDraftPersistence()
  startDigestWatcher()
}
