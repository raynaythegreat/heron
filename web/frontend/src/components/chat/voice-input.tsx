import { IconMicrophone, IconMicrophoneOff } from "@tabler/icons-react"
import { useCallback, useEffect, useRef, useState } from "react"

interface VoiceInputProps {
  onTranscript: (text: string) => void
  isListening?: boolean
}

type SpeechRecognitionCtor = new () => {
  continuous: boolean
  interimResults: boolean
  lang: string
  onstart: (() => void) | null
  onend: (() => void) | null
  onerror: (() => void) | null
  onresult:
    | ((ev: {
        resultIndex: number
        results: { length: number; [i: number]: { [0]: { transcript: string } } }
      }) => void)
    | null
  start(): void
  stop(): void
  abort(): void
}

export function VoiceInput({ onTranscript }: VoiceInputProps) {
  const [isListening, setIsListening] = useState(false)
  const [transcript, setTranscript] = useState("")
  const recognitionRef = useRef<InstanceType<SpeechRecognitionCtor> | null>(null)
  const [hasSupport] = useState(() => {
    if (typeof window === "undefined") return false
    return !!(
      (window as unknown as { SpeechRecognition?: unknown }).SpeechRecognition ??
      (window as unknown as { webkitSpeechRecognition?: unknown }).webkitSpeechRecognition
    )
  })

  useEffect(() => {
    const SR =
      (window as unknown as { SpeechRecognition?: SpeechRecognitionCtor }).SpeechRecognition ??
      (window as unknown as { webkitSpeechRecognition?: SpeechRecognitionCtor }).webkitSpeechRecognition

    if (!SR) {
      return
    }

    const recognition = new SR()
    recognition.continuous = true
    recognition.interimResults = true
    recognition.lang = "en-US"

    recognition.onresult = (ev) => {
      let finalTranscript = ""
      let interimTranscript = ""
      for (let i = ev.resultIndex; i < ev.results.length; i++) {
        const result = ev.results[i] as { isFinal?: boolean; 0: { transcript: string } }
        if (result.isFinal) {
          finalTranscript += result[0].transcript
        } else {
          interimTranscript += result[0].transcript
        }
      }
      const full = finalTranscript || interimTranscript
      setTranscript(full)
      if (finalTranscript) {
        onTranscript(finalTranscript)
        setTranscript("")
      }
    }

    recognition.onerror = () => {
      setIsListening(false)
    }

    recognition.onend = () => {
      setIsListening(false)
    }

    recognitionRef.current = recognition

    return () => {
      recognition.abort()
    }
  }, [onTranscript])

  const toggleListening = useCallback(() => {
    if (!recognitionRef.current) return

    if (isListening) {
      recognitionRef.current.stop()
      setIsListening(false)
    } else {
      recognitionRef.current.start()
      setIsListening(true)
    }
  }, [isListening])

  if (!hasSupport) return null

  return (
    <div className="flex items-center gap-2">
      <button
        type="button"
        onClick={toggleListening}
        className={`p-2 rounded-full transition-all ${
          isListening
            ? "bg-red-500 text-white animate-pulse"
            : "bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20"
        }`}
        title={isListening ? "Stop listening" : "Start voice input"}
      >
        {isListening ? (
          <IconMicrophoneOff size={18} />
        ) : (
          <IconMicrophone size={18} />
        )}
      </button>
      {transcript && (
        <span className="text-xs text-muted-foreground animate-pulse">
          {transcript}
        </span>
      )}
    </div>
  )
}
