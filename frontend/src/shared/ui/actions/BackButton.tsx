import { useLocation, useNavigate } from "react-router-dom";
import { ArrowLeft } from "lucide-react";

export function BackButton({
  fallback = "/",
  children = "Назад"
}: {
  fallback?: string;
  children?: React.ReactNode;
  className?: string;
}) {
  const navigate = useNavigate();
  const location = useLocation();

  const onClick = () => {
    if (location.key !== "default") navigate(-1);
    else navigate(fallback, { replace: true });
  };

  return (
    <button type="button" onClick={onClick} className={"inline-flex items-center gap-1.5 text-sm text-zinc-400 transition-colors hover:text-zinc-100 focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-zinc-400"}>
      <ArrowLeft className="size-4" />
      {children}
    </button>
  );
}