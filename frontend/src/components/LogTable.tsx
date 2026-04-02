import { useRef } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useLogStore } from '../store/logStore';
import clsx from 'clsx';

export function LogTable() {
  const logs = useLogStore((state) => state.logs);
  const parentRef = useRef<HTMLDivElement>(null);

  const rowVirtualizer = useVirtualizer({
    count: logs.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 40,
    measureElement: (el) => el.getBoundingClientRect().height,
    overscan: 10,
  });

  const getLevelColor = (level: string) => {
    switch (level.toUpperCase()) {
      case 'ERROR': return 'text-[var(--color-core-danger)]';
      case 'WARN': return 'text-[var(--color-core-warning)]';
      case 'INFO': return 'text-[var(--color-core-success)]';
      default: return 'text-[var(--color-core-dim)]';
    }
  };

  return (
    <div className="flex-1 bg-[var(--color-core-panel)] border border-[var(--color-core-border)] rounded-md overflow-hidden flex flex-col m-4 shadow-xl">
      {/* Header */}
      <div className="flex bg-[#010409] border-b border-[var(--color-core-border)] px-4 py-3 font-bold text-xs uppercase tracking-widest text-[var(--color-core-dim)]">
        <div className="w-48 shrink-0">Timestamp</div>
        <div className="w-24 shrink-0">Level</div>
        <div className="w-40 shrink-0">Origem</div>
        <div className="flex-1">Mensagem</div>
      </div>

      {/* Virtualized Body */}
      <div ref={parentRef} className="flex-1 overflow-auto bg-[var(--color-core-panel)] scroll-smooth">
        <div
          style={{
            height: `${rowVirtualizer.getTotalSize()}px`,
            width: '100%',
            position: 'relative',
          }}
        >
          {rowVirtualizer.getVirtualItems().map((virtualRow) => {
            const log = logs[virtualRow.index];
            if (!log) return null;
            
            // Normaliza para ISO UTC e exibe no formato YYYY-MM-DDTHH:MM:SSZ (mesmo dos filtros)
            const tsISO = log.timestamp.includes('T') ? log.timestamp : log.timestamp.replace(' ', 'T') + 'Z';
            const d = new Date(tsISO);
            const displayTime = isNaN(d.getTime()) ? log.timestamp : d.toISOString().slice(0, 19) + 'Z';

            return (
              <div
                key={virtualRow.key}
                data-index={virtualRow.index}
                ref={rowVirtualizer.measureElement}
                style={{
                  position: 'absolute',
                  top: 0,
                  left: 0,
                  width: '100%',
                  transform: `translateY(${virtualRow.start}px)`,
                }}
                className={clsx(
                  "flex items-start px-4 py-2 hover:bg-[#1c2128] border-b border-[var(--color-core-border)] transition-colors text-[13px] font-mono leading-relaxed group",
                  virtualRow.index % 2 === 0 ? "bg-transparent" : "bg-[#0d1117]/40"
                )}
              >
                <div className="w-48 shrink-0 text-[var(--color-core-dim)] overflow-hidden text-ellipsis whitespace-nowrap pt-1">
                  {displayTime}
                </div>
                <div className={clsx("w-24 shrink-0 font-bold pt-1", getLevelColor(log.level))}>
                  [{log.level}]
                </div>
                <div className="w-40 shrink-0 text-[var(--color-core-accent)] overflow-hidden text-ellipsis whitespace-nowrap pt-1 pr-2">
                  {log.source_app}
                </div>
                <div className="flex-1 text-[var(--color-core-text)] whitespace-pre-wrap break-words inline-block pt-1 pb-1">
                  {log.message}
                </div>
              </div>
            );
          })}
        </div>
        {logs.length === 0 && (
          <div className="flex items-center justify-center h-full text-[var(--color-core-dim)]">
            Nenhum log correspondente encontrado.
          </div>
        )}
      </div>
    </div>
  );
}
