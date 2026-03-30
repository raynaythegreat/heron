import { useMemo } from "react";

interface ChartData {
  type: string;
  data: number[];
  labels?: string[];
  title?: string;
}

interface MarkerParseResult {
  type: "chart" | "diff" | "kanban" | "tts" | "table" | "progress" | "alert" | "text";
  content: string;
  metadata?: Record<string, unknown>;
}

function parseMarkers(text: string): MarkerParseResult[] {
  const results: MarkerParseResult[] = [];
  const markerRegex = /\[heron:(\w+)(?::([^\]]*))?\]([\s\S]*?)(?=\[heron:|$)/gi;
  let lastIndex = 0;
  let match: RegExpExecArray | null;

  while ((match = markerRegex.exec(text)) !== null) {
    if (match.index > lastIndex) {
      const before = text.slice(lastIndex, match.index).trim();
      if (before) {
        results.push({ type: "text", content: before });
      }
    }

    const markerType = match[1];
    const params = match[2] || "";
    const body = match[3]?.trim() || "";

    switch (markerType) {
      case "chart":
        try {
          results.push({
            type: "chart",
            content: "",
            metadata: JSON.parse(params || body),
          });
        } catch {
          results.push({ type: "text", content: match[0] });
        }
        break;
      case "diff":
        results.push({
          type: "diff",
          content: body || params,
        });
        break;
      case "table":
        results.push({
          type: "table",
          content: body || params,
        });
        break;
      case "progress":
        results.push({
          type: "progress",
          content: params,
          metadata: { value: parseInt(params) || 0 },
        });
        break;
      case "alert":
        const alertParts = (body || params).split("|");
        results.push({
          type: "alert",
          content: alertParts[0] || "",
          metadata: { level: alertParts[1] || "info" },
        });
        break;
      default:
        results.push({ type: "text", content: match[0] });
    }

    lastIndex = match.index + match[0].length;
  }

  if (lastIndex < text.length) {
    const remaining = text.slice(lastIndex).trim();
    if (remaining) {
      results.push({ type: "text", content: remaining });
    }
  }

  return results;
}

function DiffRenderer({ content }: { content: string }) {
  const lines = content.split("\n");
  return (
    <div className="rounded-lg border bg-muted/50 p-4 font-mono text-sm overflow-x-auto">
      {lines.map((line, i) => {
        if (line.startsWith("+")) {
          return <div key={i} className="text-green-500 bg-green-500/10">{line}</div>;
        }
        if (line.startsWith("-")) {
          return <div key={i} className="text-red-500 bg-red-500/10">{line}</div>;
        }
        if (line.startsWith("@")) {
          return <div key={i} className="text-cyan-500 bg-cyan-500/10">{line}</div>;
        }
        return <div key={i} className="text-muted-foreground">{line}</div>;
      })}
    </div>
  );
}

function TableRenderer({ content }: { content: string }) {
  const rows = content.trim().split("\n").map(line => line.split("|").map(cell => cell.trim()));
  if (rows.length === 0) return null;

  return (
    <div className="rounded-lg border overflow-hidden">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b bg-muted/50">
            {rows[0].map((header, i) => (
              <th key={i} className="px-4 py-2 text-left font-medium">{header}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.slice(1).map((row, i) => (
            <tr key={i} className="border-b last:border-0">
              {row.map((cell, j) => (
                <td key={j} className="px-4 py-2">{cell}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function ProgressRenderer({ metadata }: { metadata?: Record<string, unknown> }) {
  const value = (metadata?.value as number) || 0;
  return (
    <div className="space-y-2">
      <div className="flex justify-between text-sm">
        <span className="text-muted-foreground">Progress</span>
        <span className="font-medium text-cyan-400">{value}%</span>
      </div>
      <div className="h-2 rounded-full bg-muted overflow-hidden">
        <div
          className="h-full rounded-full bg-cyan-500 transition-all duration-500"
          style={{ width: `${value}%` }}
        />
      </div>
    </div>
  );
}

function AlertRenderer({ content, metadata }: { content: string; metadata?: Record<string, unknown> }) {
  const level = (metadata?.level as string) || "info";
  const styles: Record<string, string> = {
    info: "border-cyan-500 bg-cyan-500/10 text-cyan-400",
    warning: "border-yellow-500 bg-yellow-500/10 text-yellow-400",
    error: "border-red-500 bg-red-500/10 text-red-400",
    success: "border-green-500 bg-green-500/10 text-green-400",
  };
  return (
    <div className={`rounded-lg border p-4 ${styles[level] || styles.info}`}>
      {content}
    </div>
  );
}

export function AgentMarkerRenderer({ content }: { content: string }) {
  const markers = useMemo(() => parseMarkers(content), [content]);

  return (
    <div className="space-y-3">
      {markers.map((marker, i) => {
        switch (marker.type) {
          case "diff":
            return <DiffRenderer key={i} content={marker.content} />;
          case "table":
            return <TableRenderer key={i} content={marker.content} />;
          case "progress":
            return <ProgressRenderer key={i} metadata={marker.metadata} />;
          case "alert":
            return <AlertRenderer key={i} content={marker.content} metadata={marker.metadata} />;
          case "text":
          default:
            return <p key={i} className="text-sm leading-relaxed">{marker.content}</p>;
        }
      })}
    </div>
  );
}

export { parseMarkers };
export type { MarkerParseResult, ChartData };
