import Skeleton from '@/shared/ui/feedback/Skeleton'


export default function MePageSkeleton() {
  return (
    <div className="mx-auto w-full max-w-5xl py-10" aria-busy="true">
      <div className="grid gap-8 lg:grid-cols-[17rem_minmax(0,1fr)] lg:items-start">
        <Skeleton className="h-96 w-full rounded-xl" />
        <div className="grid gap-3">
          <Skeleton className="h-7 w-24" />
          <Skeleton className="h-22 w-full rounded-xl" />
          <Skeleton className="h-22 w-full rounded-xl" />
        </div>
      </div>
    </div>
  )
}
