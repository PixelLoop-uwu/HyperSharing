// Public surface of the user feature. Import from here instead of deep paths.

// UI
export { default as AuthShell } from './ui/AuthShell'
export { default as Avatar } from './ui/Avatar'
export { default as MePageSkeleton } from './ui/MePageSkeleton'
export { default as ProfileSidebar } from './ui/ProfileSidebar'
export { default as UserMenu } from './ui/UserMenu'

// Auth & data hooks
export { useMe, useRefreshMe, useLogin, useRegister, useLogout } from './model/hooks'

// Types
export type * from './model/types'
