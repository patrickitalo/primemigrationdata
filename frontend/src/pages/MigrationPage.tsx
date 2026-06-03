import { useState, useEffect, useRef } from 'react'
import { StartMigration, StopMigration } from '../../wailsjs/go/app/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { models } from '../../wailsjs/go/models'

// MigrationStatus is emitted as a plain event payload (not a Wails-generated model)
interface MigrationStatusEvent {
  Option: string
  Name: string
  Status: string
  TotalOrigem: number
  Novos: number
  Skipped: number
  ErrosReg: number
  Error: string
}

const MAX_LOG_LINES = 500

interface StatusRow {
  option: string
  name: string
  status: 'pending' | 'running' | 'success' | 'error'
  totalOrigem: number
  novos: number
  skipped: number
  erros: number
  error: string
}

interface ProgressEvent { completed: number; total: number }
interface CompleteEvent { success: boolean; error?: string }

interface Props {
  config: models.MigrationConfig
  session: models.UserSession
  onClose: () => void
}

export default function MigrationPage({ config, session: _session, onClose }: Props) {
  const [rows, setRows] = useState<StatusRow[]>(() =>
    config.options.map(o => ({ option: o, name: o, status: 'pending', totalOrigem: 0, novos: 0, skipped: 0, erros: 0, error: '' }))
  )
  const [logs, setLogs] = useState<string[]>([])
  const [progress, setProgress] = useState<ProgressEvent>({ completed: 0, total: config.options.length })
  const [done, setDone] = useState(false)
  const [doneSuccess, setDoneSuccess] = useState(false)
  const [stopping, setStopping] = useState(false)
  const [confirmStop, setConfirmStop] = useState(false)
  const logRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const offs = [
      EventsOn('migration:log', (msg: unknown) => {
        setLogs(prev => {
          const next = [...prev, String(msg)]
          return next.length > MAX_LOG_LINES ? next.slice(next.length - MAX_LOG_LINES) : next
        })
      }),
      EventsOn('migration:status-update', (s: unknown) => {
        const status = s as MigrationStatusEvent
        setRows(prev => prev.map(r =>
          r.option === status.Option
            ? {
                ...r,
                name: status.Name || r.name,
                status: status.Status as StatusRow['status'],
                totalOrigem: status.TotalOrigem ?? r.totalOrigem,
                novos: status.Novos ?? r.novos,
                skipped: status.Skipped ?? r.skipped,
                erros: status.ErrosReg ?? r.erros,
                error: status.Error || '',
              }
            : r
        ))
      }),
      EventsOn('migration:progress', (p: unknown) => {
        setProgress(p as ProgressEvent)
      }),
      EventsOn('migration:complete', (e: unknown) => {
        const ev = e as CompleteEvent
        setDone(true)
        setDoneSuccess(ev.success)
        setStopping(false)
      }),
    ]

    StartMigration(config).catch(err => {
      setLogs(prev => [...prev, `[ERRO] Falha ao iniciar: ${err}`])
      setDone(true)
      setDoneSuccess(false)
    })

    return () => offs.forEach(off => off())
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (logRef.current) {
      logRef.current.scrollTop = logRef.current.scrollHeight
    }
  }, [logs])

  // Resolve option names from status-update events
  useEffect(() => {
    setRows(prev => prev.map(r => {
      if (r.name === r.option) {
        const found = prev.find(x => x.option === r.option && x.name !== x.option)
        return found ? { ...r, name: found.name } : r
      }
      return r
    }))
  }, [])

  function handleStopClick() {
    if (done) { onClose(); return }
    setConfirmStop(true)
  }

  async function confirmStopMigration() {
    setConfirmStop(false)
    setStopping(true)
    await StopMigration()
  }

  const pct = progress.total > 0 ? (progress.completed / progress.total) * 100 : 0

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <header className="flex items-center justify-between px-5 py-3 bg-slate-900 border-b border-slate-800 flex-shrink-0">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-prime-700 flex items-center justify-center">
            <svg className="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 7v10c0 2 1 3 3 3h10c2 0 3-1 3-3V7c0-2-1-3-3-3H7C5 4 4 5 4 7z" />
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6M9 8h6M9 16h4" />
            </svg>
          </div>
          <div>
            <span className="font-display font-bold text-white text-sm">Migração em andamento</span>
            <span className="text-xs text-slate-400 ml-2">{config.system} — cliente {config.client_code}</span>
          </div>
        </div>
        <button
          onClick={handleStopClick}
          disabled={stopping}
          className={done ? 'btn-primary' : 'btn-danger'}
        >
          {done ? 'Fechar' : stopping ? 'Parando...' : 'Parar'}
        </button>
      </header>

      {/* Progress bar */}
      <div className="px-5 py-3 flex-shrink-0 bg-slate-950 border-b border-slate-800">
        <div className="flex items-center justify-between text-xs text-slate-400 mb-2">
          <span>
            {done
              ? doneSuccess ? 'Migração concluída com sucesso' : 'Migração encerrada'
              : stopping ? 'Cancelando...'
              : `Migrando... ${progress.completed} de ${progress.total}`}
          </span>
          <span>{Math.round(pct)}%</span>
        </div>
        <div className="h-2 bg-slate-800 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-500 ${
              done
                ? doneSuccess ? 'bg-green-500' : 'bg-red-500'
                : 'bg-prime-600'
            }`}
            style={{ width: `${done && doneSuccess ? 100 : pct}%` }}
          />
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 flex gap-0 overflow-hidden">
        {/* Status table */}
        <div className="flex-1 overflow-auto border-r border-slate-800">
          <table className="w-full text-sm">
            <thead className="sticky top-0 bg-slate-900 z-10">
              <tr className="text-xs text-slate-400 font-medium">
                <th className="text-left px-4 py-2.5">Entidade</th>
                <th className="text-left px-3 py-2.5">Status</th>
                <th className="text-right px-3 py-2.5">Origem</th>
                <th className="text-right px-3 py-2.5">Novos</th>
                <th className="text-right px-3 py-2.5">Pulados</th>
                <th className="text-right px-4 py-2.5">Erros</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {rows.map(row => (
                <tr key={row.option} className="hover:bg-slate-800/30 transition-colors">
                  <td className="px-4 py-3 text-slate-200 font-medium">{row.name}</td>
                  <td className="px-3 py-3">
                    <StatusBadge status={row.status} error={row.error} />
                  </td>
                  <td className="px-3 py-3 text-right text-slate-400 tabular-nums">
                    {row.status !== 'pending' ? row.totalOrigem : '—'}
                  </td>
                  <td className="px-3 py-3 text-right text-green-400 tabular-nums font-medium">
                    {row.status === 'success' || row.status === 'error' ? row.novos : '—'}
                  </td>
                  <td className="px-3 py-3 text-right text-slate-400 tabular-nums">
                    {row.status === 'success' || row.status === 'error' ? row.skipped : '—'}
                  </td>
                  <td className="px-4 py-3 text-right tabular-nums">
                    <span className={row.erros > 0 ? 'text-red-400 font-medium' : 'text-slate-500'}>
                      {row.status === 'success' || row.status === 'error' ? row.erros : '—'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Log viewer */}
        <div className="w-80 flex flex-col flex-shrink-0">
          <div className="flex items-center justify-between px-4 py-2.5 border-b border-slate-800 flex-shrink-0">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Log</span>
            <button
              className="text-xs text-slate-500 hover:text-slate-300 transition-colors"
              onClick={() => navigator.clipboard?.writeText(logs.join('\n'))}
            >
              Copiar
            </button>
          </div>
          <div
            ref={logRef}
            className="flex-1 overflow-y-auto p-3 font-mono text-xs text-slate-400 space-y-0.5 leading-relaxed"
          >
            {logs.map((line, i) => (
              <div
                key={i}
                className={
                  line.includes('[ERRO]') ? 'text-red-400' :
                  line.includes('[AVISO]') ? 'text-yellow-400' :
                  line.includes('sucesso') ? 'text-green-400' :
                  'text-slate-400'
                }
              >
                {line}
              </div>
            ))}
            {logs.length === 0 && (
              <span className="text-slate-600">Aguardando logs...</span>
            )}
          </div>
        </div>
      </div>

      {/* Confirm stop dialog */}
      {confirmStop && (
        <div className="fixed inset-0 flex items-center justify-center z-50 bg-black/60 backdrop-blur-sm">
          <div className="bg-slate-800 border border-slate-700 rounded-xl p-6 max-w-sm w-full mx-4 shadow-2xl">
            <h3 className="font-display font-bold text-white mb-2">Parar migração?</h3>
            <p className="text-sm text-slate-400 mb-4">
              A migração será interrompida. Os dados já migrados serão mantidos.
            </p>
            <div className="flex gap-3">
              <button className="btn-danger flex-1" onClick={confirmStopMigration}>Parar</button>
              <button className="btn-secondary flex-1" onClick={() => setConfirmStop(false)}>Cancelar</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function StatusBadge({ status, error }: { status: StatusRow['status']; error: string }) {
  switch (status) {
    case 'pending':
      return <span className="badge-info">Aguardando</span>
    case 'running':
      return (
        <span className="badge-info flex items-center gap-1">
          <svg className="w-3 h-3 animate-spin" viewBox="0 0 24 24" fill="none">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          Migrando
        </span>
      )
    case 'success':
      return <span className="badge-success">Concluído</span>
    case 'error':
      return (
        <span className="badge-error" title={error}>Erro</span>
      )
  }
}
