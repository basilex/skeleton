export function AuthIllustration() {
  return (
    <div className="relative flex h-full w-full items-center justify-center overflow-hidden bg-muted/30">
      <svg
        viewBox="0 0 400 400"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className="h-3/4 w-3/4 max-w-[320px]"
      >
        <circle cx="200" cy="200" r="160" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.2" />
        <circle cx="200" cy="200" r="120" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.15" />
        <circle cx="200" cy="200" r="80" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.1" />
        <circle cx="200" cy="200" r="40" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.08" />
        <line x1="40" y1="200" x2="360" y2="200" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.15" />
        <line x1="200" y1="40" x2="200" y2="360" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.15" />
        <line x1="87" y1="87" x2="313" y2="313" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.1" />
        <line x1="313" y1="87" x2="87" y2="313" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.1" />
        <rect x="140" y="140" width="120" height="120" rx="4" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.2" />
        <rect x="170" y="170" width="60" height="60" rx="2" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.25" />
        <path d="M200 40 L215 90 L200 80 L185 90 Z" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.2" />
        <path d="M200 360 L215 310 L200 320 L185 310 Z" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.2" />
        <path d="M40 200 L90 215 L80 200 L90 185 Z" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.2" />
        <path d="M360 200 L310 215 L320 200 L310 185 Z" stroke="currentColor" strokeWidth="0.5" strokeOpacity="0.2" />
        <circle cx="200" cy="200" r="3" fill="currentColor" fillOpacity="0.4" />
        <circle cx="200" cy="40" r="2" fill="currentColor" fillOpacity="0.3" />
        <circle cx="200" cy="360" r="2" fill="currentColor" fillOpacity="0.3" />
        <circle cx="40" cy="200" r="2" fill="currentColor" fillOpacity="0.3" />
        <circle cx="360" cy="200" r="2" fill="currentColor" fillOpacity="0.3" />
        <circle cx="87" cy="87" r="2" fill="currentColor" fillOpacity="0.2" />
        <circle cx="313" cy="87" r="2" fill="currentColor" fillOpacity="0.2" />
        <circle cx="87" cy="313" r="2" fill="currentColor" fillOpacity="0.2" />
        <circle cx="313" cy="313" r="2" fill="currentColor" fillOpacity="0.2" />
        <path d="M140 200 Q170 170 200 200 Q230 230 260 200" stroke="currentColor" strokeWidth="1" strokeOpacity="0.25" />
        <path d="M200 140 Q170 170 200 200 Q230 230 200 260" stroke="currentColor" strokeWidth="1" strokeOpacity="0.25" />
        <path
          d="M120 120 Q160 80 240 120 Q320 160 280 240 Q240 320 160 280 Q80 240 120 120Z"
          stroke="currentColor"
          strokeWidth="0.5"
          strokeOpacity="0.12"
          strokeDasharray="6 4"
        />
      </svg>
    </div>
  )
}