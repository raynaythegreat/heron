import { useEffect, useRef, useState } from "react";

export function VideoBackground() {
  const videoRef = useRef<HTMLVideoElement>(null);
  const [isDark, setIsDark] = useState(false);

  useEffect(() => {
    const observer = new MutationObserver(() => {
      setIsDark(document.documentElement.classList.contains("dark"));
    });
    observer.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ["class"],
    });
    setIsDark(document.documentElement.classList.contains("dark"));
    return () => observer.disconnect();
  }, []);

  useEffect(() => {
    if (videoRef.current) {
      videoRef.current.play().catch(() => {});
    }
  }, [isDark]);

  return (
    <div className="fixed inset-0 z-0 overflow-hidden">
      <video
        ref={videoRef}
        key={isDark ? "dark" : "light"}
        autoPlay
        loop
        muted
        playsInline
        className="absolute inset-0 w-full h-full object-cover"
      >
        <source
          src={isDark ? "/DS.mp4" : "/LS.mp4"}
          type="video/mp4"
        />
      </video>
      <div className="absolute inset-0 bg-black/30 dark:bg-black/50" />
    </div>
  );
}
