import { useState, useEffect } from 'react'
import {
  GetEnvDefaults, LoadClientConfig, SaveClientConfig,
  GetOptionNames, CheckHistoryStatus, ReconnectHistory,
  GetLastRunInfo, SelectExcelFile,
} from '../../wailsjs/go/app/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import type { models, app } from '../../wailsjs/go/models'

type EnvDefaults = app.EnvDefaults
type OptionDef = app.OptionDef
type HistoryStatus = app.HistoryStatus
type LastRunInfo = app.LastRunInfo

const SYSTEMS = [
  'FCERTA', 'PRISMA5',
  'FASTERM', 'FARMNET', 'FARMPLUS', 'FASTFARMA',
  'MEGAFARMA', 'SUPERFARMA', 'NEOFARMA',
  'SISTEMA10', 'SISTEMA11', 'SISTEMA12', 'SISTEMA13',
]

type TabKey = 'conexao' | 'opcoes' | 'migrar'

interface FormState {
  clientCode: string
  system: string
  host: string
  port: string
  path: string
  user: string
  password: string
  conversao: string
  alias: string
  ipServer: string
  porta: string
  selectedOptions: string[]
  mode: 'COMPLETA' | 'INCREMENTAL'
  vVencido: string
  excelPath: string
}

const defaultForm: FormState = {
  clientCode: '', system: 'FCERTA',
  host: '', port: '3050', path: '', user: '', password: '', conversao: 'WIN1252',
  alias: '', ipServer: '', porta: '3050',
  selectedOptions: [],
  mode: 'COMPLETA',
  vVencido: '',
  excelPath: '',
}

interface Props {
  session: models.UserSession
  onLogout: () => void
  onStartMigration: (cfg: models.MigrationConfig) => void
}

export default function MainPage({ session, onLogout, onStartMigration }: Props) {
  const [activeTab, setActiveTab] = useState<TabKey>('conexao')
  const [form, setForm] = useState<FormState>(defaultForm)
  const [optionDefs, setOptionDefs] = useState<OptionDef[]>([])
  const [historyStatus, setHistoryStatus] = useState<HistoryStatus>({ connected: false })
  const [lastRun, setLastRun] = useState<LastRunInfo | null>(null)
  const [reconnecting, setReconnecting] = useState(false)
  const [savedConfigModal, setSavedConfigModal] = useState<models.ClientConfig | null>(null)
  const [envDefaults, setEnvDefaults] = useState<EnvDefaults | null>(null)

  const errors = validateForm(form)
  const tabErrors: Record<TabKey, boolean> = {
    conexao: !!(errors.clientCode || errors.host || errors.path || errors.user || errors.password || errors.alias || errors.ipServer || errors.porta),
    opcoes: !!errors.options,
    migrar: false,
  }

  useEffect(() => {
    GetEnvDefaults().then(env => {
      setEnvDefaults(env)
      setForm(f => ({
        ...f,
        host: env.firebird.Host || f.host,
        port: env.firebird.Port || f.port,
        path: env.firebird.Path || f.path,
        user: env.firebird.User || f.user,
        password: env.firebird.Password || f.password,
        conversao: env.firebird.Conversao || f.conversao,
        alias: env.pharmacie.Alias || f.alias,
        ipServer: env.pharmacie.IPServer || f.ipServer,
        porta: env.pharmacie.Porta || f.porta,
      }))
    })
    CheckHistoryStatus().then(setHistoryStatus)
    const off = EventsOn('history:status-changed', (connected: unknown) => {
      setHistoryStatus(s => ({ ...s, connected: !!connected }))
    })
    return () => off()
  }, [])

  useEffect(() => {
    GetOptionNames(form.system).then(setOptionDefs)
    setForm(f => ({ ...f, selectedOptions: [], vVencido: '', excelPath: '' }))
  }, [form.system])

  useEffect(() => {
    if (!form.clientCode.trim() || !form.system) return
    LoadClientConfig(form.clientCode.trim(), form.system)
      .then(saved => { if (saved) setSavedConfigModal(saved) })
      .catch(() => {})
  }, [form.clientCode, form.system])

  useEffect(() => {
    if (!form.clientCode.trim() || !form.system || !historyStatus.connected) return
    GetLastRunInfo(form.clientCode.trim(), form.system).then(setLastRun).catch(() => {})
  }, [form.clientCode, form.system, historyStatus.connected])

  function setField<K extends keyof FormState>(key: K, value: FormState[K]) {
    setForm(f => ({ ...f, [key]: value }))
  }

  function toggleOption(code: string) {
    setForm(f => ({
      ...f,
      selectedOptions: f.selectedOptions.includes(code)
        ? f.selectedOptions.filter(o => o !== code)
        : [...f.selectedOptions, code],
    }))
  }

  async function handleReconnect() {
    setReconnecting(true)
    try { await ReconnectHistory() } catch { /* handled by event */ }
    const status = await CheckHistoryStatus()
    setHistoryStatus(status)
    setReconnecting(false)
  }

  async function handleSelectExcel() {
    try {
      const path = await SelectExcelFile()
      if (path) setField('excelPath', path)
    } catch { /* user cancelled */ }
  }

  function loadSavedConfig(saved: models.ClientConfig) {
    setForm(f => ({
      ...f,
      host: saved.db_host || f.host,
      port: saved.db_port || f.port,
      path: saved.db_path || f.path,
      user: saved.db_user || f.user,
      password: envDefaults?.firebird.Password ? f.password : (saved.db_password || f.password),
      conversao: saved.conversao || f.conversao,
      alias: saved.alias_pharmacie || f.alias,
      ipServer: saved.ipserver_pharmacie || f.ipServer,
      porta: saved.porta_pharmacie || f.porta,
    }))
    setSavedConfigModal(null)
  }

  function handleStart() {
    if (Object.keys(errors).length > 0) return
    const cfg = {
      client_code: form.clientCode.trim(),
      system: form.system,
      database: {
        host: form.host, port: form.port, path: form.path,
        user: form.user, password: form.password, conversao: form.conversao,
      },
      options: form.selectedOptions,
      mode: form.mode,
      v_vencido: form.vVencido || undefined,
      excel_path: form.excelPath,
      alias_pharmacie: form.alias,
      ipserver_pharmacie: form.ipServer,
      porta_pharmacie: form.porta,
    } as unknown as models.MigrationConfig
    SaveClientConfig({
      id: 0,
      codigo_cliente: form.clientCode.trim(),
      sistema_origem: form.system,
      db_host: form.host, db_port: form.port, db_path: form.path,
      db_user: form.user,
      db_password: envDefaults?.firebird.Password ? '' : form.password,
      alias_pharmacie: form.alias,
      ipserver_pharmacie: form.ipServer,
      porta_pharmacie: form.porta,
      conversao: form.conversao,
    } as unknown as models.ClientConfig).catch(() => {})
    onStartMigration(cfg)
  }

  const tabs: { key: TabKey; label: string }[] = [
    { key: 'conexao', label: 'Conexão' },
    { key: 'opcoes', label: 'Opções' },
    { key: 'migrar', label: 'Migrar' },
  ]

  return (
    <div className="h-full flex flex-col">
      {/* Top bar */}
      <header className="flex items-center justify-between px-5 py-3 bg-slate-900 border-b border-slate-800 flex-shrink-0">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-lg bg-prime-700 flex items-center justify-center">
            <svg className="w-4 h-4 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 7v10c0 2 1 3 3 3h10c2 0 3-1 3-3V7c0-2-1-3-3-3H7C5 4 4 5 4 7z" />
              <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6M9 8h6M9 16h4" />
            </svg>
          </div>
          <span className="font-display font-bold text-white text-sm">Prime Migration</span>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs text-slate-400">{session.Nome}</span>
          <button onClick={onLogout} className="btn-secondary text-xs px-3 py-1.5">Sair</button>
        </div>
      </header>

      {/* Tabs */}
      <div className="flex gap-0 px-5 pt-4 flex-shrink-0 border-b border-slate-800 bg-slate-950">
        {tabs.map(t => (
          <button
            key={t.key}
            onClick={() => setActiveTab(t.key)}
            className={`relative px-5 py-2.5 text-sm font-medium transition-colors duration-150 border-b-2 -mb-px ${
              activeTab === t.key
                ? 'border-prime-500 text-prime-400'
                : 'border-transparent text-slate-400 hover:text-slate-200'
            }`}
          >
            {t.label}
            {tabErrors[t.key] && (
              <span className="absolute top-1.5 right-1 w-2 h-2 rounded-full bg-red-500" />
            )}
          </button>
        ))}
      </div>

      {/* Tab content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === 'conexao' && (
          <TabConexao
            form={form}
            errors={errors}
            historyStatus={historyStatus}
            reconnecting={reconnecting}
            onField={setField}
            onReconnect={handleReconnect}
          />
        )}
        {activeTab === 'opcoes' && (
          <TabOpcoes
            form={form}
            optionDefs={optionDefs}
            errors={errors}
            historyStatus={historyStatus}
            onToggleOption={toggleOption}
            onField={setField}
            onSelectExcel={handleSelectExcel}
          />
        )}
        {activeTab === 'migrar' && (
          <TabMigrar
            form={form}
            optionDefs={optionDefs}
            errors={errors}
            lastRun={lastRun}
            historyStatus={historyStatus}
            onStart={handleStart}
          />
        )}
      </div>

      {/* Saved config modal */}
      {savedConfigModal && (
        <div className="fixed inset-0 flex items-center justify-center z-50 bg-black/60 backdrop-blur-sm">
          <div className="bg-slate-800 border border-slate-700 rounded-xl p-6 max-w-sm w-full mx-4 shadow-2xl">
            <h3 className="font-display font-bold text-white mb-2">Configuração salva encontrada</h3>
            <p className="text-sm text-slate-400 mb-4">
              Encontramos uma configuração salva para o cliente <strong className="text-slate-200">{savedConfigModal.codigo_cliente}</strong> ({savedConfigModal.sistema_origem}). Deseja carregá-la?
            </p>
            <div className="flex gap-3">
              <button className="btn-primary flex-1" onClick={() => loadSavedConfig(savedConfigModal)}>Carregar</button>
              <button className="btn-secondary flex-1" onClick={() => setSavedConfigModal(null)}>Ignorar</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// ─── Tab: Conexão ──────────────────────────────────────────────────────────────

interface TabConexaoProps {
  form: FormState
  errors: Record<string, string>
  historyStatus: HistoryStatus
  reconnecting: boolean
  onField: <K extends keyof FormState>(k: K, v: FormState[K]) => void
  onReconnect: () => void
}

function TabConexao({ form, errors, historyStatus, reconnecting, onField, onReconnect }: TabConexaoProps) {
  return (
    <div className="p-5 space-y-5 max-w-2xl">
      {/* History banner */}
      <div className={`flex items-center gap-3 px-4 py-2.5 rounded-lg border text-sm ${
        historyStatus.connected
          ? 'bg-green-950/40 border-green-800/50 text-green-300'
          : 'bg-yellow-950/40 border-yellow-800/50 text-yellow-300'
      }`}>
        <span className={`w-2 h-2 rounded-full flex-shrink-0 ${historyStatus.connected ? 'bg-green-400' : 'bg-yellow-400'}`} />
        <span className="flex-1">
          {historyStatus.connected
            ? 'Histórico central conectado — modo incremental disponível'
            : (historyStatus.error ? `Histórico offline: ${historyStatus.error}` : 'Histórico central não conectado')}
        </span>
        {!historyStatus.connected && (
          <button onClick={onReconnect} disabled={reconnecting} className="text-xs underline opacity-80 hover:opacity-100">
            {reconnecting ? 'Reconectando...' : 'Reconectar'}
          </button>
        )}
      </div>

      {/* Cliente + Sistema */}
      <div className="card space-y-4">
        <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Identificação</h3>
        <div className="grid grid-cols-2 gap-4">
          <FormField label="Código do Cliente" error={errors.clientCode}>
            <input className="input-field" placeholder="Ex: 1234" value={form.clientCode}
              onChange={e => onField('clientCode', e.target.value)} />
          </FormField>
          <FormField label="Sistema de Origem">
            <select className="input-field" value={form.system} onChange={e => onField('system', e.target.value)}>
              {SYSTEMS.map(s => <option key={s} value={s}>{s}</option>)}
            </select>
          </FormField>
        </div>
      </div>

      {/* Banco Firebird */}
      <div className="card space-y-4">
        <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Banco de Dados Firebird (Origem)</h3>
        <div className="grid grid-cols-3 gap-4">
          <FormField label="Host" error={errors.host} className="col-span-2">
            <input className="input-field" placeholder="localhost" value={form.host}
              onChange={e => onField('host', e.target.value)} />
          </FormField>
          <FormField label="Porta">
            <input className="input-field" placeholder="3050" value={form.port}
              onChange={e => onField('port', e.target.value)} />
          </FormField>
        </div>
        <FormField label="Caminho do Banco" error={errors.path}>
          <input className="input-field" placeholder="C:\Dados\database.fdb" value={form.path}
            onChange={e => onField('path', e.target.value)} />
        </FormField>
        <div className="grid grid-cols-2 gap-4">
          <FormField label="Usuário" error={errors.user}>
            <input className="input-field" placeholder="SYSDBA" value={form.user}
              onChange={e => onField('user', e.target.value)} />
          </FormField>
          <FormField label="Senha" error={errors.password}>
            <input type="password" className="input-field" placeholder="••••••••" value={form.password}
              onChange={e => onField('password', e.target.value)} />
          </FormField>
        </div>
        <FormField label="Conversão de Charset">
          <select className="input-field" value={form.conversao} onChange={e => onField('conversao', e.target.value)}>
            <option value="WIN1252">WIN1252</option>
            <option value="UTF8">UTF8</option>
            <option value="ISO8859_1">ISO8859_1</option>
          </select>
        </FormField>
      </div>

      {/* Pharmacie */}
      <div className="card space-y-4">
        <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Pharmacie (Destino)</h3>
        <div className="grid grid-cols-3 gap-4">
          <FormField label="Alias" error={errors.alias} className="col-span-2">
            <input className="input-field" placeholder="pharmacie" value={form.alias}
              onChange={e => onField('alias', e.target.value)} />
          </FormField>
          <FormField label="Porta">
            <input className="input-field" placeholder="3050" value={form.porta}
              onChange={e => onField('porta', e.target.value)} />
          </FormField>
        </div>
        <FormField label="IP / Servidor" error={errors.ipServer}>
          <input className="input-field" placeholder="192.168.1.1" value={form.ipServer}
            onChange={e => onField('ipServer', e.target.value)} />
        </FormField>
      </div>
    </div>
  )
}

// ─── Tab: Opções ───────────────────────────────────────────────────────────────

interface TabOpcoesProps {
  form: FormState
  optionDefs: OptionDef[]
  errors: Record<string, string>
  historyStatus: HistoryStatus
  onToggleOption: (code: string) => void
  onField: <K extends keyof FormState>(k: K, v: FormState[K]) => void
  onSelectExcel: () => void
}

function TabOpcoes({ form, optionDefs, errors, historyStatus, onToggleOption, onField, onSelectExcel }: TabOpcoesProps) {
  const incrementalDisabled = !historyStatus.connected

  return (
    <div className="p-5 space-y-5 max-w-2xl">
      {/* Entities */}
      <div className="card space-y-3">
        <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Entidades para migrar</h3>
        {errors.options && (
          <p className="text-xs text-red-400">{errors.options}</p>
        )}
        <div className="grid grid-cols-2 gap-2">
          {optionDefs.map(opt => (
            <label key={opt.code} className="flex items-center gap-3 px-3 py-2.5 rounded-lg border border-slate-700/50 hover:border-slate-600 cursor-pointer transition-colors">
              <input
                type="checkbox"
                className="w-4 h-4 rounded border-slate-600 bg-slate-800 text-prime-600 focus:ring-prime-500 focus:ring-offset-0"
                checked={form.selectedOptions.includes(opt.code)}
                onChange={() => onToggleOption(opt.code)}
              />
              <span className="text-sm text-slate-300">{opt.name}</span>
            </label>
          ))}
        </div>
      </div>

      {/* Mode */}
      <div className="card space-y-3">
        <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Modo de migração</h3>
        <div className="grid grid-cols-2 gap-3">
          {(['COMPLETA', 'INCREMENTAL'] as const).map(mode => (
            <label
              key={mode}
              className={`flex items-start gap-3 p-3 rounded-lg border cursor-pointer transition-colors ${
                form.mode === mode
                  ? 'border-prime-600 bg-prime-950/40'
                  : 'border-slate-700/50 hover:border-slate-600'
              } ${mode === 'INCREMENTAL' && incrementalDisabled ? 'opacity-40 cursor-not-allowed' : ''}`}
            >
              <input
                type="radio"
                name="mode"
                value={mode}
                checked={form.mode === mode}
                disabled={mode === 'INCREMENTAL' && incrementalDisabled}
                onChange={() => onField('mode', mode)}
                className="mt-0.5 text-prime-600 focus:ring-prime-500 focus:ring-offset-0"
              />
              <div>
                <p className="text-sm font-medium text-slate-200">{mode === 'COMPLETA' ? 'Completa' : 'Incremental'}</p>
                <p className="text-xs text-slate-500 mt-0.5">
                  {mode === 'COMPLETA'
                    ? 'Migra todos os registros'
                    : incrementalDisabled
                      ? 'Requer histórico central conectado'
                      : 'Pula registros já migrados'}
                </p>
              </div>
            </label>
          ))}
        </div>
      </div>

      {/* FCERTA extras */}
      {form.system === 'FCERTA' && (
        <div className="card space-y-3">
          <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Opções FCERTA</h3>
          <FormField label="Data de Vencimento (vVencido)" hint="Ex: 31/12/2023 — deixe em branco para ignorar">
            <input className="input-field" placeholder="dd/mm/aaaa" value={form.vVencido}
              onChange={e => onField('vVencido', e.target.value)} />
          </FormField>
        </div>
      )}

      {/* PRISMA5 extras */}
      {form.system === 'PRISMA5' && (
        <div className="card space-y-3">
          <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Opções PRISMA5</h3>
          <FormField label="Planilha de Grupos (Excel)">
            <div className="flex gap-2">
              <input className="input-field flex-1" placeholder="Nenhum arquivo selecionado" readOnly
                value={form.excelPath ? form.excelPath.split(/[\\/]/).pop() ?? '' : ''} />
              <button className="btn-secondary flex-shrink-0" onClick={onSelectExcel}>
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                </svg>
                Selecionar
              </button>
            </div>
          </FormField>
        </div>
      )}
    </div>
  )
}

// ─── Tab: Migrar ───────────────────────────────────────────────────────────────

interface TabMigrarProps {
  form: FormState
  optionDefs: OptionDef[]
  errors: Record<string, string>
  lastRun: LastRunInfo | null
  historyStatus: HistoryStatus
  onStart: () => void
}

function TabMigrar({ form, optionDefs, errors, lastRun, historyStatus, onStart }: TabMigrarProps) {
  const canStart = Object.keys(errors).length === 0 && form.selectedOptions.length > 0

  const selectedNames = form.selectedOptions.map(code => {
    const def = optionDefs.find(o => o.code === code)
    return def?.name ?? code
  })

  return (
    <div className="p-5 space-y-5 max-w-2xl">
      {/* Summary */}
      <div className="card space-y-3">
        <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Resumo</h3>
        <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
          <SummaryItem label="Cliente" value={form.clientCode || '—'} />
          <SummaryItem label="Sistema" value={form.system} />
          <SummaryItem label="Host" value={form.host || '—'} />
          <SummaryItem label="Modo" value={form.mode} />
          <SummaryItem label="Entidades" value={selectedNames.length > 0 ? selectedNames.join(', ') : '—'} className="col-span-2" />
        </dl>
      </div>

      {/* Last run */}
      {historyStatus.connected && (
        <div className="card space-y-2">
          <h3 className="text-xs font-semibold text-slate-400 uppercase tracking-wider">Último run bem-sucedido</h3>
          {lastRun ? (
            <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
              <SummaryItem label="Iniciado em" value={lastRun.startedAt} />
              <SummaryItem label="Modo" value={lastRun.mode} />
              <SummaryItem label="Implantador" value={lastRun.implantador || '—'} />
              <SummaryItem label="Status" value={lastRun.status} />
            </dl>
          ) : (
            <p className="text-sm text-slate-500">Nenhum run anterior encontrado para este cliente.</p>
          )}
        </div>
      )}

      {/* Validation errors */}
      {Object.keys(errors).length > 0 && (
        <div className="card border-red-800/50 space-y-1.5">
          <h3 className="text-xs font-semibold text-red-400 uppercase tracking-wider">Campos obrigatórios</h3>
          {Object.values(errors).map((e, i) => (
            <p key={i} className="text-sm text-red-300 flex items-center gap-2">
              <span className="w-1 h-1 rounded-full bg-red-400 flex-shrink-0" />
              {e}
            </p>
          ))}
        </div>
      )}

      {/* Start button */}
      <button
        className="btn-primary w-full py-3 text-base"
        disabled={!canStart}
        onClick={onStart}
      >
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        Iniciar Migração
      </button>
    </div>
  )
}

// ─── Shared components ─────────────────────────────────────────────────────────

function FormField({
  label, error, hint, className = '', children,
}: {
  label: string
  error?: string
  hint?: string
  className?: string
  children: React.ReactNode
}) {
  return (
    <div className={className}>
      <label className="label">{label}</label>
      {children}
      {hint && !error && <p className="text-xs text-slate-500 mt-1">{hint}</p>}
      {error && <p className="text-xs text-red-400 mt-1">{error}</p>}
    </div>
  )
}

function SummaryItem({ label, value, className = '' }: { label: string; value: string; className?: string }) {
  return (
    <div className={className}>
      <dt className="text-xs text-slate-500">{label}</dt>
      <dd className="text-slate-200 font-medium">{value}</dd>
    </div>
  )
}

// ─── Validation ────────────────────────────────────────────────────────────────

function validateForm(f: FormState): Record<string, string> {
  const e: Record<string, string> = {}
  if (!f.clientCode.trim()) e.clientCode = 'Código do cliente obrigatório'
  if (!f.host.trim()) e.host = 'Host obrigatório'
  if (!f.path.trim()) e.path = 'Caminho do banco obrigatório'
  if (!f.user.trim()) e.user = 'Usuário obrigatório'
  if (!f.password.trim()) e.password = 'Senha obrigatória'
  if (!f.alias.trim()) e.alias = 'Alias Pharmacie obrigatório'
  if (!f.ipServer.trim()) e.ipServer = 'IP do servidor Pharmacie obrigatório'
  if (f.selectedOptions.length === 0) e.options = 'Selecione pelo menos uma entidade'
  return e
}
