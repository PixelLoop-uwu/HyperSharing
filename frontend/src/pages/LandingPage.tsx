import About from '@/widgets/About'
import Hero from '@/widgets/Hero'

export default function LandingPage() {
  return (
    <>
      <Hero />
      <div className="backdrop-blur-2xl">
        <About />
      </div>
    </>
  )
}
