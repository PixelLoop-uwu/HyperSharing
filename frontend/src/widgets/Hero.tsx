import { motion } from 'framer-motion'
import { FiUpload } from 'react-icons/fi'
import { rise, stagger } from '@/shared/lib/motion'
import PrimaryButton from '@/shared/ui/actions/PrimaryButton'
import Typewriter from '@/shared/ui/content/Typewriter'
import MicroSlats from '@/shared/ui/effects/MicroSlats'
import PostPreview from './PostPreview'

const HEADLINE = 'Все, чем хочется поделиться. В одной ссылке.'

export default function Hero() {
  return (
    <section className="relative -mt-(--header-total-h) flex min-h-svh items-center pb-20 pt-[calc(var(--header-total-h)+3rem)]">
      <div aria-hidden="true" className="absolute inset-y-0 left-1/2 w-screen -translate-x-1/2 overflow-hidden">
        <MicroSlats
          preset="tide"
          color="#3f3f46"
          glintColor="#a1a1aa"
          backgroundColor="#09090b"
          slatWidth={10}
          slatHeight={24}
          gap={3}
          roundness={0.3}
          interactive={true}
          cursorStrength={1}
          cursorSize={40}
          swirl={0}
          trail={1.4}
          lean={0}
          intro
          scale={0.55}
          speed={0.3}
          direction={300}
          chop={1}
          stretch={0.3}
          glint={1.3}
          contrast={1.8}
          perspective={0.8}
          fog={0.35}
          introDuration={1.5}
        />
        <div className="absolute inset-0 bg-linear-to-r from-zinc-950 via-zinc-950/70 to-transparent" />

        <div className="absolute inset-x-0 bottom-0 h-40 bg-linear-to-t from-zinc-950 to-transparent" />
      </div>

      <motion.div
        variants={stagger}
        initial="hidden"
        animate="show"
        className="relative grid w-full items-center gap-14 lg:grid-cols-[minmax(0,1fr)_minmax(0,26rem)] lg:gap-20"
      >
        <div>
          <motion.h1
            variants={rise}
            className="min-h-[2.2em] max-w-160 font-display text-4xl font-semibold leading-[1.08] tracking-tight lg:text-[46px]"
          >
            <span className="sr-only">{HEADLINE}</span>
            <span aria-hidden="true">
              <Typewriter text={HEADLINE} />
            </span>
          </motion.h1>

          <motion.p variants={rise} className="mt-6 max-w-md text-lg leading-relaxed text-zinc-400">
            Текст, документы, архивы, видео и другие файлы. 
            Соберите всё в одном посте и отправьте одну ссылку вместо десятка вложений.
          </motion.p>

          <motion.div variants={rise} className="mt-8 flex flex-wrap items-center gap-x-6 gap-y-4">
            <PrimaryButton to="/posts/new" icon={<FiUpload className="size-4" />} shine>
              Создать пост
            </PrimaryButton>
            <a
              href="#about"
              className="text-sm font-medium text-zinc-400 transition-colors hover:text-zinc-100 focus-visible:text-zinc-100"
            >
              Как это устроено
            </a>
          </motion.div>
        </div>

        <motion.div variants={rise}>
          <PostPreview />
        </motion.div>
      </motion.div>
    </section>
  )
}