import PostsSection from '@/features/posts/ui/PostsSection'
import { useAuth } from '@/features/user/store'
import ProfileSidebar from '@/features/user/ui/ProfileSidebar'

/**
 * Профиль текущего пользователя в духе страницы юзера на GitHub:
 * слева — аватар и данные аккаунта, справа — список постов.
 */
export default function MePage() {
  const user = useAuth((s) => s.user)
  if (!user) return null // UserOnly гарантирует не-null

  return (
    <div className="mx-auto w-full max-w-6xl py-10">
      <div className="flex flex-col gap-8 lg:flex-row lg:items-start">
        <aside className="lg:sticky lg:top-[calc(var(--header-total-h)+1.5rem)] lg:w-80 lg:shrink-0 xl:w-85">
          <ProfileSidebar user={user} />
        </aside>
        <div className="min-w-0 flex-1">
          <PostsSection />
        </div>
      </div>
    </div>
  )
}
