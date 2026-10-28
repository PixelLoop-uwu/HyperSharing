import { useNavigate } from 'react-router-dom'
import { motion, MotionConfig } from 'framer-motion'
import { LockKeyhole, ArrowRight } from 'lucide-react'
import { PostEditor } from '@/features/posts'
import { useMe } from '@/features/user'
import Card from '@/shared/ui/layout/Card'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'


export default function NewPostPage() {
  const navigate = useNavigate()
  const me = useMe()
  const anonymous = !me.isLoading && !me.data

  return (
    <div className="relative flex-1">
      <PostEditor
       onPublished={(post) => navigate(`/${post.short_key}`)} 
       overlayActive={anonymous}
      />

      {anonymous && (
        <MotionConfig reducedMotion="user">
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ duration: 0.2, ease: 'easeOut' }}
            className="absolute inset-0 z-10 flex items-center justify-center p-4 backdrop-blur-sm"
          >
            <motion.div
              initial={{ opacity: 0, y: 12, scale: 0.97 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              transition={{
                duration: 0.35,
                ease: [0.16, 1, 0.3, 1],
              }}
              className="w-full max-w-sm"
            >
              <Card className="p-6 shadow-2xl shadow-black/20">
                <div className="flex flex-col items-center text-center">
                  <div className="mb-5 flex size-12 items-center justify-center rounded-2xl border border-zinc-800 bg-zinc-900/80 text-zinc-300">
                    <LockKeyhole className="size-5" strokeWidth={1.7} />
                  </div>

                  <div className="mb-3 inline-flex items-center gap-2 rounded-full border border-zinc-800 bg-zinc-900/50 px-3 py-1.5">
                    <span className="size-1.5 rounded-full bg-zinc-500" />
                    <span className="text-xs font-medium text-zinc-400">
                      Требуется авторизация
                    </span>
                  </div>

                  <h2 className="text-xl font-medium tracking-tight text-zinc-100">
                    Войдите, чтобы продолжить
                  </h2>

                  <p className="mt-2 text-sm leading-relaxed text-zinc-400">
                    Создание публикаций доступно только авторизованным пользователям.
                    Войдите в аккаунт, чтобы поделиться своими файлами.
                  </p>

                  <PrimaryButton
                    to="/login"
                    className="mt-6 w-full justify-center"
                  >
                    Войти в аккаунт
                    <ArrowRight className="ml-2 size-4" />
                  </PrimaryButton>
                </div>
              </Card>
            </motion.div>
          </motion.div>
        </MotionConfig>
      )}
    </div>
  )
}
