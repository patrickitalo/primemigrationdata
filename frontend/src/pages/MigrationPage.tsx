import { useState, useEffect, useRef } from 'react'
import { StartMigration, StopMigration } from '../../wailsjs/go/app/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { models } from '../../wailsjs/go/models'

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

export default function MigrationPage({ config, session, onClose }: Props) {
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

  async function confirmStopMigration() {
    setConfirmStop(false)
    setStopping(true)
    await StopMigration()
  }

  const pct = progress.total > 0 ? Math.round((progress.completed / progress.total) * 100) : 0
  const runningRow = rows.find(r => r.status === 'running')
  const totalNovos = rows.reduce((sum, r) => sum + r.novos, 0)
  const totalSkipped = rows.reduce((sum, r) => sum + r.skipped, 0)
  const totalErros = rows.reduce((sum, r) => sum + r.erros, 0)

  return (
    <div className="h-full flex flex-col" style={{ backgroundColor: '#f8f9fb' }}>
      {/* Header */}
      <header
        className="flex items-center justify-between px-10 flex-shrink-0"
        style={{ backgroundColor: 'white', borderBottom: '1px solid #dfe1e6', height: 64, '--wails-draggable': 'drag' } as React.CSSProperties}
      >
        <div className="flex items-center gap-8">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded flex items-center justify-center" style={{ backgroundColor: '#0052cc' }}>
              <span className="text-white font-bold text-lg leading-none">P</span>
            </div>
            <span className="font-bold text-xl" style={{ color: '#172b4d' }}>Prime Migration</span>
          </div>
          <nav className="flex h-16 items-center">
            {(['Conexão', 'Opções', 'Migrar'] as const).map(label => (
              <span
                key={label}
                className="relative h-full flex items-center px-4 text-sm"
                style={{
                  color: label === 'Migrar' ? '#0052cc' : '#5e6c84',
                  fontWeight: label === 'Migrar' ? 600 : 500,
                  borderBottom: label === 'Migrar' ? '2px solid #0052cc' : '2px solid transparent',
                }}
              >
                {label}
              </span>
            ))}
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <p className="text-xs font-semibold uppercase" style={{ color: '#42526e' }}>{session.Nome}</p>
          {done && (
            <button
              onClick={onClose}
              className="px-4 py-1.5 rounded text-sm font-semibold text-white transition-colors"
              style={{ backgroundColor: '#0052cc' }}
              onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#0747a6')}
              onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#0052cc')}
            >
              Fechar
            </button>
          )}
        </div>
      </header>

      {/* Page body */}
      <div className="flex-1 overflow-y-auto" style={{ backgroundColor: '#f8f9fb' }}>
        <div className="max-w-[1440px] mx-auto py-8 px-10">

          {/* Page title */}
          <div className="flex flex-col md:flex-row md:items-end justify-between gap-3 mb-6">
            <div>
              <h1 className="text-2xl font-bold mb-1" style={{ color: '#191c1e' }}>
                {done ? (doneSuccess ? 'Migração Concluída' : 'Migração Encerrada') : 'Processo de Migração Ativo'}
              </h1>
              <p className="text-sm" style={{ color: '#434654' }}>
                Cliente {config.client_code} — {config.system}
              </p>
            </div>
            <div className="flex items-center gap-2">
              {done ? (
                <span
                  className="px-3 py-1.5 rounded-full text-xs font-semibold flex items-center gap-1.5"
                  style={{
                    backgroundColor: doneSuccess ? '#e3fcef' : '#ffeaea',
                    color: doneSuccess ? '#006644' : '#ba1a1a',
                  }}
                >
                  <span className="w-2 h-2 rounded-full" style={{ backgroundColor: doneSuccess ? '#36b37e' : '#ba1a1a' }} />
                  {doneSuccess ? 'Concluído com sucesso' : 'Encerrado'}
                </span>
              ) : stopping ? (
                <span className="px-3 py-1.5 rounded-full text-xs font-semibold flex items-center gap-1.5" style={{ backgroundColor: '#fff3cd', color: '#856404' }}>
                  <span className="w-2 h-2 rounded-full bg-yellow-500 animate-pulse" />
                  Cancelando...
                </span>
              ) : (
                <span className="px-3 py-1.5 rounded-full text-xs font-semibold flex items-center gap-1.5" style={{ backgroundColor: '#bfd1ff', color: '#485980' }}>
                  <svg className="w-3.5 h-3.5 animate-spin" viewBox="0 0 24 24" fill="none">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  Em andamento
                </span>
              )}
            </div>
          </div>

          {/* Main grid */}
          <div className="grid grid-cols-12 gap-6 items-start">

            {/* Left column */}
            <div className="col-span-8 flex flex-col gap-6">

              {/* Progress card */}
              <div className="bg-white rounded-xl border border-[#dfe1e6] p-8 shadow-sm">
                <div className="flex items-center justify-between mb-6">
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg flex items-center justify-center" style={{ backgroundColor: '#eef2ff' }}>
                      <svg className="w-5 h-5" style={{ color: '#0052cc' }} viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 3C7.58 3 4 4.79 4 7s3.58 4 8 4 8-1.79 8-4-3.58-4-8-4zM4 9v3c0 2.21 3.58 4 8 4s8-1.79 8-4V9c0 2.21-3.58 4-8 4s-8-1.79-8-4zm0 5v3c0 2.21 3.58 4 8 4s8-1.79 8-4v-3c0 2.21-3.58 4-8 4s-8-1.79-8-4z" />
                      </svg>
                    </div>
                    <div>
                      <p className="font-semibold text-base" style={{ color: '#191c1e' }}>
                        {runningRow ? runningRow.name : (done ? 'Finalizado' : 'Aguardando...')}
                      </p>
                      <p className="text-xs mt-0.5" style={{ color: '#434654' }}>
                        Etapa {progress.completed} de {progress.total}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <span className="text-4xl font-bold" style={{ color: '#003d9b' }}>{pct}%</span>
                    <p className="text-xs uppercase tracking-wider mt-0.5" style={{ color: '#737685' }}>Concluído</p>
                  </div>
                </div>
                <div className="w-full h-4 rounded-full overflow-hidden mb-3" style={{ backgroundColor: '#f3f4f6' }}>
                  <div
                    className="h-full rounded-full transition-all duration-700 ease-in-out"
                    style={{
                      width: `${done && doneSuccess ? 100 : pct}%`,
                      backgroundColor: done ? (doneSuccess ? '#36b37e' : '#ba1a1a') : '#003d9b',
                    }}
                  />
                </div>
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium" style={{ color: '#191c1e' }}>
                    {progress.completed} / {progress.total} entidades
                  </p>
                  <div className="flex items-center gap-4 text-xs" style={{ color: '#434654' }}>
                    <span className="flex items-center gap-1">
                      <span className="w-3 h-3 rounded-full" style={{ backgroundColor: '#003d9b' }} />
                      Processado
                    </span>
                    <span className="flex items-center gap-1">
                      <span className="w-3 h-3 rounded-full" style={{ backgroundColor: '#f3f4f6', border: '1px solid #c3c6d6' }} />
                      Restante
                    </span>
                  </div>
                </div>
              </div>

              {/* Entity status table */}
              <div className="bg-white rounded-xl border border-[#dfe1e6] shadow-sm overflow-hidden">
                <div className="px-6 py-4 border-b border-[#ebecf0]">
                  <h3 className="text-xs font-bold uppercase tracking-widest" style={{ color: '#737685' }}>STATUS POR ENTIDADE</h3>
                </div>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr style={{ backgroundColor: '#f8f9fb' }}>
                        <th className="text-left px-6 py-3 text-xs font-semibold uppercase tracking-wider" style={{ color: '#42526e' }}>Entidade</th>
                        <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider" style={{ color: '#42526e' }}>Status</th>
                        <th className="text-right px-4 py-3 text-xs font-semibold uppercase tracking-wider" style={{ color: '#42526e' }}>Origem</th>
                        <th className="text-right px-4 py-3 text-xs font-semibold uppercase tracking-wider" style={{ color: '#42526e' }}>Novos</th>
                        <th className="text-right px-4 py-3 text-xs font-semibold uppercase tracking-wider" style={{ color: '#42526e' }}>Pulados</th>
                        <th className="text-right px-6 py-3 text-xs font-semibold uppercase tracking-wider" style={{ color: '#42526e' }}>Erros</th>
                      </tr>
                    </thead>
                    <tbody>
                      {rows.map((row, i) => (
                        <tr
                          key={row.option}
                          style={{
                            borderTop: i > 0 ? '1px solid #f3f4f6' : 'none',
                            backgroundColor: row.status === 'running' ? 'rgba(0,82,204,0.03)' : 'white',
                          }}
                        >
                          <td className="px-6 py-3 font-medium" style={{ color: '#191c1e' }}>{row.name}</td>
                          <td className="px-4 py-3">
                            <EntityStatusBadge status={row.status} error={row.error} />
                          </td>
                          <td className="px-4 py-3 text-right tabular-nums" style={{ color: '#434654' }}>
                            {row.status !== 'pending' ? row.totalOrigem.toLocaleString() : '—'}
                          </td>
                          <td className="px-4 py-3 text-right tabular-nums font-semibold" style={{ color: '#006644' }}>
                            {(row.status === 'success' || row.status === 'error') ? row.novos.toLocaleString() : '—'}
                          </td>
                          <td className="px-4 py-3 text-right tabular-nums" style={{ color: '#434654' }}>
                            {(row.status === 'success' || row.status === 'error') ? row.skipped.toLocaleString() : '—'}
                          </td>
                          <td className="px-6 py-3 text-right tabular-nums font-semibold" style={{ color: row.erros > 0 ? '#ba1a1a' : '#737685' }}>
                            {(row.status === 'success' || row.status === 'error') ? row.erros.toLocaleString() : '—'}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>

              {/* Log console (dark terminal) */}
              <div className="rounded-xl border overflow-hidden shadow-lg" style={{ borderColor: '#737685' }}>
                {/* Title bar */}
                <div className="flex items-center justify-between px-4 py-2.5" style={{ backgroundColor: '#2d2d2d', borderBottom: '1px solid #3d3d3d' }}>
                  <div className="flex items-center gap-3">
                    <div className="flex gap-1.5">
                      <div className="w-3 h-3 rounded-full" style={{ backgroundColor: '#ff5f56' }} />
                      <div className="w-3 h-3 rounded-full" style={{ backgroundColor: '#ffbd2e' }} />
                      <div className="w-3 h-3 rounded-full" style={{ backgroundColor: '#27c93f' }} />
                    </div>
                    <span className="text-xs font-mono" style={{ color: 'rgba(255,255,255,0.5)' }}>
                      migration_log.bash — {config.client_code}
                    </span>
                  </div>
                  <button
                    className="text-xs transition-colors"
                    style={{ color: 'rgba(255,255,255,0.4)' }}
                    onClick={() => navigator.clipboard?.writeText(logs.join('\n'))}
                    onMouseEnter={e => (e.currentTarget.style.color = 'rgba(255,255,255,0.9)')}
                    onMouseLeave={e => (e.currentTarget.style.color = 'rgba(255,255,255,0.4)')}
                  >
                    Copiar
                  </button>
                </div>
                {/* Log content */}
                <div
                  ref={logRef}
                  className="p-5 font-mono text-sm overflow-y-auto"
                  style={{
                    backgroundColor: '#1e1e1e',
                    height: 320,
                    scrollbarWidth: 'thin',
                    scrollbarColor: '#434654 #1e1e1e',
                  }}
                >
                  {logs.length === 0 ? (
                    <span style={{ color: 'rgba(255,255,255,0.3)' }}>Aguardando logs...</span>
                  ) : (
                    logs.map((line, i) => (
                      <p key={i} className="mb-1 leading-relaxed" style={{ color: logLineColor(line) }}>
                        {line}
                      </p>
                    ))
                  )}
                  {!done && (
                    <span className="animate-pulse" style={{ color: '#b2c5ff' }}>_</span>
                  )}
                </div>
              </div>

            </div>

            {/* Right column */}
            <div className="col-span-4 flex flex-col gap-4">

              {/* Stats mini grid */}
              <div className="grid grid-cols-2 gap-3">
                <StatCard
                  icon={
                    <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor" style={{ color: '#434654' }}>
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 14l-4-4 1.41-1.41L10 13.17l6.59-6.59L18 9l-8 8z" />
                    </svg>
                  }
                  value={totalNovos.toLocaleString()}
                  label="Novos"
                  iconColor="#36b37e"
                />
                <StatCard
                  icon={
                    <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 5h2v6h-2zm0 8h2v2h-2z" />
                    </svg>
                  }
                  value={totalSkipped.toLocaleString()}
                  label="Pulados"
                  iconColor="#434654"
                />
                <StatCard
                  icon={
                    <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                    </svg>
                  }
                  value={totalErros.toLocaleString()}
                  label="Erros"
                  iconColor={totalErros > 0 ? '#ba1a1a' : '#434654'}
                  borderLeft={totalErros > 0}
                />
                <StatCard
                  icon={
                    <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-7 14l-5-5 1.41-1.41L12 14.17l7.59-7.59L21 8l-9 9z" />
                    </svg>
                  }
                  value={`${progress.completed}/${progress.total}`}
                  label="Etapas"
                  iconColor="#0052cc"
                />
              </div>

              {/* Job config card */}
              <div className="rounded-xl border border-[#dfe1e6] p-6" style={{ backgroundColor: '#f3f4f6' }}>
                <h3 className="text-xs font-bold uppercase tracking-widest mb-4" style={{ color: '#434654' }}>Configuração do Job</h3>
                <div className="space-y-3">
                  {[
                    { label: 'Código do Cliente', value: config.client_code },
                    { label: 'Sistema de Origem', value: config.system },
                    { label: 'Host Firebird', value: config.database?.host || '—' },
                    { label: 'Modo', value: config.mode },
                  ].map(({ label, value }) => (
                    <div key={label} className="flex items-center justify-between py-2 border-b border-[#dfe1e6] last:border-0">
                      <span className="text-xs" style={{ color: '#434654' }}>{label}</span>
                      <span className="text-xs font-semibold" style={{ color: '#191c1e' }}>{value}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Action buttons */}
              <div className="flex flex-col gap-2">
                {!done ? (
                  <>
                    <button
                      onClick={() => setConfirmStop(true)}
                      disabled={stopping}
                      className="w-full py-3 px-6 rounded-xl font-semibold flex items-center justify-center gap-2 transition-all shadow-sm disabled:opacity-60 disabled:cursor-not-allowed"
                      style={{ backgroundColor: 'white', border: '1px solid #737685', color: '#191c1e' }}
                      onMouseEnter={e => !stopping && (e.currentTarget.style.backgroundColor = '#edeef0')}
                      onMouseLeave={e => (e.currentTarget.style.backgroundColor = 'white')}
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      {stopping ? 'Cancelando...' : 'Pausar Migração'}
                    </button>
                    <button
                      onClick={() => setConfirmStop(true)}
                      disabled={stopping}
                      className="w-full py-3 px-6 rounded-xl font-semibold flex items-center justify-center gap-2 transition-all disabled:opacity-60 disabled:cursor-not-allowed"
                      style={{ backgroundColor: '#ffdad6', color: '#93000a' }}
                      onMouseEnter={e => !stopping && (e.currentTarget.style.filter = 'brightness(0.95)')}
                      onMouseLeave={e => (e.currentTarget.style.filter = 'none')}
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                      Cancelar Processo
                    </button>
                  </>
                ) : (
                  <button
                    onClick={onClose}
                    className="w-full py-3 px-6 rounded-xl font-bold text-white flex items-center justify-center gap-2 transition-colors"
                    style={{ backgroundColor: '#0052cc' }}
                    onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#0747a6')}
                    onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#0052cc')}
                  >
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                    </svg>
                    Fechar e Voltar
                  </button>
                )}
              </div>

            </div>
          </div>
        </div>
      </div>

      {/* Confirm stop dialog */}
      {confirmStop && (
        <div className="fixed inset-0 flex items-center justify-center z-50" style={{ backgroundColor: 'rgba(0,0,0,0.5)' }}>
          <div className="bg-white rounded-xl border border-[#dfe1e6] p-6 max-w-sm w-full mx-4 shadow-xl">
            <h3 className="font-bold text-lg mb-2" style={{ color: '#172b4d' }}>Cancelar migração?</h3>
            <p className="text-sm mb-5" style={{ color: '#5e6c84' }}>
              A migração será interrompida. Os dados já migrados serão mantidos, mas o processo precisará ser reiniciado.
            </p>
            <div className="flex gap-3">
              <button
                className="flex-1 py-2.5 rounded text-sm font-semibold transition-colors"
                style={{ backgroundColor: '#ffdad6', color: '#93000a' }}
                onMouseEnter={e => (e.currentTarget.style.filter = 'brightness(0.95)')}
                onMouseLeave={e => (e.currentTarget.style.filter = 'none')}
                onClick={confirmStopMigration}
              >
                Cancelar Processo
              </button>
              <button
                className="flex-1 py-2.5 rounded text-sm font-semibold transition-colors"
                style={{ backgroundColor: '#ebecf0', color: '#42526e', border: '1px solid #c1c7d0' }}
                onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#dfe1e6')}
                onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#ebecf0')}
                onClick={() => setConfirmStop(false)}
              >
                Continuar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// ─── Shared components ─────────────────────────────────────────────────────────

function StatCard({
  icon, value, label, iconColor, borderLeft,
}: {
  icon: React.ReactNode
  value: string
  label: string
  iconColor: string
  borderLeft?: boolean
}) {
  return (
    <div
      className="bg-white p-4 rounded-xl border border-[#dfe1e6] flex flex-col"
      style={borderLeft ? { borderLeftWidth: 4, borderLeftColor: 'rgba(186,26,26,0.2)' } : {}}
    >
      <span style={{ color: iconColor }}>{icon}</span>
      <span className="text-xl font-bold mt-1" style={{ color: '#191c1e' }}>{value}</span>
      <span className="text-xs font-semibold uppercase tracking-wider mt-0.5" style={{ color: '#434654' }}>{label}</span>
    </div>
  )
}

function EntityStatusBadge({ status, error }: { status: StatusRow['status']; error: string }) {
  switch (status) {
    case 'pending':
      return (
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold" style={{ backgroundColor: '#edeef0', color: '#434654' }}>
          Aguardando
        </span>
      )
    case 'running':
      return (
        <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold" style={{ backgroundColor: '#bfd1ff', color: '#485980' }}>
          <svg className="w-3 h-3 animate-spin" viewBox="0 0 24 24" fill="none">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
          </svg>
          Migrando
        </span>
      )
    case 'success':
      return (
        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold" style={{ backgroundColor: '#e3fcef', color: '#006644' }}>
          Concluído
        </span>
      )
    case 'error':
      return (
        <span
          className="inline-flex items-center px-2 py-0.5 rounded text-xs font-semibold cursor-help"
          style={{ backgroundColor: '#ffdad6', color: '#93000a' }}
          title={error}
        >
          Erro
        </span>
      )
  }
}

function logLineColor(line: string): string {
  if (line.includes('[ERRO]') || line.toLowerCase().includes('error')) return '#ff6b6b'
  if (line.includes('[AVISO]') || line.toLowerCase().includes('warn')) return '#ffd93d'
  if (line.includes('sucesso') || line.includes('SUCCESS') || line.includes('[OK]')) return '#6bcb77'
  if (line.startsWith('[') && line.includes(']')) return 'rgba(255,255,255,0.5)'
  return 'rgba(255,255,255,0.75)'
}
