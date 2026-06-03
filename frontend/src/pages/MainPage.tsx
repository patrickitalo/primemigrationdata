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

const OPTION_DESCRIPTIONS: Record<string, string> = {
  CLIENTES:           'Base cadastral completa de clientes',
  MEDICOS:            'Registros de CRM e especialidades',
  FORNECEDORES:       'Lista de parceiros e contatos comerciais',
  PRODUTOS:           'SKUs, categorias e precificação',
  LOTES:              'Rastreabilidade e datas de validade',
  FORMA_FARMACEUTICA: 'Formas de apresentação dos produtos',
  PRODUCAO_INTERNA:   'Ordens de serviço e fabricação',
  REFAZER:            'ATENÇÃO: Sobrescreve transações existentes',
}

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
  const canStart = Object.keys(errors).length === 0 && form.selectedOptions.length > 0
  const tabErrors: Record<TabKey, boolean> = {
    conexao: !!(errors.clientCode || errors.host || errors.path || errors.user || errors.password || errors.alias || errors.ipServer),
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
    if (!canStart) return
    const cfg = {
      client_code: form.clientCode.trim(),
      system: form.system,
      database: {
        host: form.host, port: form.port, path: form.path,
        user: form.user, password: form.password, conversao: form.conversao,
      },
      options: form.selectedOptions,
      mode: form.mode,
      v_vencido: form.vVencido !== '' ? form.vVencido : undefined,
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

  const selectedNames = form.selectedOptions.map(code => {
    const def = optionDefs.find(o => o.code === code)
    return def?.name ?? code
  })

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <header
        className="flex items-center justify-between px-6 flex-shrink-0"
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
            {tabs.map(t => (
              <button
                key={t.key}
                onClick={() => setActiveTab(t.key)}
                className="relative h-full flex items-center px-4 text-sm transition-colors"
                style={{
                  color: activeTab === t.key ? '#0052cc' : '#5e6c84',
                  fontWeight: activeTab === t.key ? 600 : 500,
                  borderBottom: activeTab === t.key ? '2px solid #0052cc' : '2px solid transparent',
                }}
              >
                {t.label}
                {tabErrors[t.key] && (
                  <span className="ml-1.5 w-1.5 h-1.5 rounded-full inline-block" style={{ backgroundColor: '#de350b' }} />
                )}
              </button>
            ))}
          </nav>
        </div>
        <div className="flex items-center gap-4">
          <p className="text-xs font-semibold uppercase" style={{ color: '#42526e' }}>{session.Nome}</p>
          <button
            onClick={onLogout}
            className="px-4 py-1.5 rounded text-sm font-medium transition-colors"
            style={{ backgroundColor: '#ebecf0', color: '#42526e', border: '1px solid #c1c7d0' }}
            onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#dfe1e6')}
            onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#ebecf0')}
          >
            Sair
          </button>
        </div>
      </header>

      {/* Page body */}
      <div className="flex-1 overflow-y-auto" style={{ backgroundColor: '#f4f5f7' }}>
        <div className="max-w-[1440px] mx-auto py-8 px-6 grid grid-cols-12 gap-8">

          {/* Left: form content */}
          <div className="col-span-8 space-y-6">
            {/* History banner */}
            <div
              className="rounded p-3 flex items-center gap-3"
              style={{
                backgroundColor: historyStatus.connected ? '#e3fcef' : '#fffae6',
                border: `1px solid ${historyStatus.connected ? '#abf5d1' : '#ffe380'}`,
              }}
            >
              <span className="w-2 h-2 rounded-full flex-shrink-0" style={{ backgroundColor: historyStatus.connected ? '#36b37e' : '#ffab00' }} />
              <p className="text-sm font-medium flex-1" style={{ color: historyStatus.connected ? '#006644' : '#974f0c' }}>
                {historyStatus.connected
                  ? 'Histórico central conectado — modo incremental disponível'
                  : (historyStatus.error ? `Histórico offline: ${historyStatus.error}` : 'Histórico central não conectado')}
              </p>
              {!historyStatus.connected && (
                <button
                  onClick={handleReconnect}
                  disabled={reconnecting}
                  className="text-xs font-semibold underline"
                  style={{ color: '#0052cc' }}
                >
                  {reconnecting ? 'Reconectando...' : 'Reconectar'}
                </button>
              )}
            </div>

            {activeTab === 'conexao' && (
              <TabConexao form={form} errors={errors} onField={setField} />
            )}
            {activeTab === 'opcoes' && (
              <TabOpcoes
                form={form}
                optionDefs={optionDefs}
                errors={errors}
                historyStatus={historyStatus}
                onToggleOption={toggleOption}
                onField={setField}
              />
            )}
            {activeTab === 'migrar' && (
              <TabMigrar
                form={form}
                optionDefs={optionDefs}
                errors={errors}
                lastRun={lastRun}
                historyStatus={historyStatus}
                selectedNames={selectedNames}
              />
            )}
          </div>

          {/* Right: sidebar */}
          <aside className="col-span-4 space-y-4">

            {/* ── Conexão tab sidebar ── */}
            {activeTab === 'conexao' && (
              <>
                <div className="bg-white rounded border border-[#dfe1e6] shadow-sm overflow-hidden">
                  <div className="p-6">
                    <h3 className="font-bold mb-2 text-lg" style={{ color: '#172b4d' }}>Configuração de Conexão</h3>
                    <p className="text-sm mb-6" style={{ color: '#5e6c84' }}>Preencha os dados de origem Firebird e destino Pharmacie para avançar.</p>
                    <SidebarBtn onClick={() => setActiveTab('opcoes')} label="Continuar para Opções" />
                  </div>
                  <SummaryStrip form={form} />
                </div>
                <QuickHelp />
              </>
            )}

            {/* ── Opções tab sidebar ── */}
            {activeTab === 'opcoes' && (
              <>
                {form.system === 'FCERTA' && (
                  <section className="bg-white rounded border border-[#dfe1e6] shadow-sm p-6">
                    <h2 className="text-xs font-bold uppercase tracking-widest mb-5" style={{ color: '#737685' }}>OPÇÕES FCERTA</h2>
                    <div className="space-y-4">
                      <Field label="vVencido" hint="Controla o filtro de vencimento aplicado na migração FCERTA">
                        <Sel
                          value={form.vVencido}
                          onChange={e => setField('vVencido', e.target.value)}
                        >
                          <option value="">Não usar</option>
                          <option value="0">0</option>
                          <option value="-1">-1</option>
                        </Sel>
                      </Field>
                      <div className="pt-3 border-t border-[#ebecf0] flex items-start gap-2 text-xs" style={{ color: '#434654' }}>
                        <svg className="w-4 h-4 mt-0.5 flex-shrink-0" style={{ color: '#0052cc' }} viewBox="0 0 24 24" fill="currentColor">
                          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-6h2v6zm0-8h-2V7h2v2z" />
                        </svg>
                        <p>As configurações de data afetam apenas os registros financeiros vinculados às entidades selecionadas.</p>
                      </div>
                    </div>
                  </section>
                )}

                {form.system === 'PRISMA5' && (
                  <section className="bg-white rounded border border-[#dfe1e6] shadow-sm p-6">
                    <h2 className="text-xs font-bold uppercase tracking-widest mb-5" style={{ color: '#737685' }}>OPÇÕES PRISMA5</h2>
                    <Field label="Planilha de Grupos (Excel)">
                      <div className="flex gap-2">
                        <Inp
                          className="flex-1"
                          placeholder="Nenhum arquivo selecionado"
                          readOnly
                          value={form.excelPath ? form.excelPath.split(/[\\/]/).pop() ?? '' : ''}
                        />
                        <button
                          onClick={handleSelectExcel}
                          className="px-4 h-10 rounded text-sm font-semibold flex items-center gap-2 flex-shrink-0 transition-colors"
                          style={{ backgroundColor: '#ebecf0', color: '#42526e', border: '1px solid #c1c7d0' }}
                          onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#dfe1e6')}
                          onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#ebecf0')}
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                          </svg>
                          Selecionar
                        </button>
                      </div>
                    </Field>
                  </section>
                )}

                {/* Blue CTA card */}
                <section
                  className="rounded-xl p-6 shadow-lg relative overflow-hidden"
                  style={{ backgroundColor: '#0052cc', color: 'white' }}
                >
                  {/* Decorative rocket */}
                  <div className="absolute -right-6 -bottom-6 opacity-10 pointer-events-none">
                    <svg className="w-28 h-28" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M13.13 22.19L11.5 18.36C13.07 17.78 14.54 17 15.9 16.09L13.13 22.19M5.64 12.5L1.81 10.87L7.91 8.1C7 9.46 6.22 10.93 5.64 12.5M19.22 4C19.5 4.78 19.75 5.57 19.87 6.38C19.75 5.57 19.5 4.78 19.22 4M21 3L2.97 9.63C1.8 10.1 1.78 11.73 2.95 12.23L5.86 13.44L7.07 18.71C7.36 19.88 8.77 20.27 9.62 19.42L11.53 17.51C13.21 18.31 15.05 18.79 16.96 18.79C18.42 18.79 19.82 18.51 21.1 18.01L21.1 3H21Z" />
                    </svg>
                  </div>
                  <h3 className="text-lg font-bold mb-2">Pronto para iniciar?</h3>
                  <p className="text-sm mb-6 opacity-90">Revise suas escolhas. Uma vez iniciada, a migração não deve ser interrompida para garantir a integridade dos dados.</p>
                  <button
                    onClick={() => setActiveTab('migrar')}
                    className="w-full bg-white font-bold py-3 rounded-lg shadow-md hover:shadow-xl hover:-translate-y-0.5 transition-all flex items-center justify-center gap-2"
                    style={{ color: '#003d9b' }}
                  >
                    Continuar para Migração
                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                    </svg>
                  </button>
                </section>
              </>
            )}

            {/* ── Migrar tab sidebar ── */}
            {activeTab === 'migrar' && (
              <>
                <div className="bg-white rounded border border-[#dfe1e6] shadow-sm overflow-hidden">
                  <div className="p-6">
                    <h3 className="font-bold mb-2 text-lg" style={{ color: '#172b4d' }}>Pronto para iniciar?</h3>
                    <p className="text-sm mb-6" style={{ color: '#5e6c84' }}>Verifique se todas as informações estão corretas antes de migrar.</p>
                    <button
                      onClick={handleStart}
                      disabled={!canStart}
                      className="w-full font-bold py-3 px-4 rounded text-white flex items-center justify-center gap-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                      style={{ backgroundColor: '#0052cc' }}
                      onMouseEnter={e => canStart && (e.currentTarget.style.backgroundColor = '#0747a6')}
                      onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#0052cc')}
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                      </svg>
                      Iniciar Migração
                    </button>
                  </div>
                  <SummaryStrip form={form} />
                </div>
              </>
            )}

          </aside>

        </div>
      </div>

      {/* Saved config modal */}
      {savedConfigModal && (
        <div className="fixed inset-0 flex items-center justify-center z-50" style={{ backgroundColor: 'rgba(0,0,0,0.5)' }}>
          <div className="bg-white rounded border border-[#dfe1e6] p-6 max-w-sm w-full mx-4 shadow-xl">
            <h3 className="font-bold mb-2 text-lg" style={{ color: '#172b4d' }}>Configuração salva encontrada</h3>
            <p className="text-sm mb-5" style={{ color: '#5e6c84' }}>
              Encontramos uma configuração salva para o cliente{' '}
              <strong style={{ color: '#172b4d' }}>{savedConfigModal.codigo_cliente}</strong>{' '}
              ({savedConfigModal.sistema_origem}). Deseja carregá-la?
            </p>
            <div className="flex gap-3">
              <button
                className="flex-1 py-2.5 rounded text-sm font-semibold text-white transition-colors"
                style={{ backgroundColor: '#0052cc' }}
                onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#0747a6')}
                onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#0052cc')}
                onClick={() => loadSavedConfig(savedConfigModal)}
              >
                Carregar
              </button>
              <button
                className="flex-1 py-2.5 rounded text-sm font-semibold transition-colors"
                style={{ backgroundColor: '#ebecf0', color: '#42526e', border: '1px solid #c1c7d0' }}
                onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#dfe1e6')}
                onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#ebecf0')}
                onClick={() => setSavedConfigModal(null)}
              >
                Ignorar
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

// ─── Sidebar helpers ───────────────────────────────────────────────────────────

function SummaryStrip({ form }: { form: FormState }) {
  return (
    <div className="px-6 py-4 border-t space-y-1.5" style={{ backgroundColor: '#f4f5f7', borderColor: '#ebecf0' }}>
      {form.clientCode && (
        <div className="flex items-center justify-between">
          <span className="text-xs" style={{ color: '#5e6c84' }}>Cliente</span>
          <span className="text-xs font-semibold" style={{ color: '#172b4d' }}>{form.clientCode}</span>
        </div>
      )}
      <div className="flex items-center justify-between">
        <span className="text-xs" style={{ color: '#5e6c84' }}>Sistema</span>
        <span className="text-xs font-semibold" style={{ color: '#172b4d' }}>{form.system}</span>
      </div>
      {form.selectedOptions.length > 0 && (
        <div className="flex items-center justify-between">
          <span className="text-xs" style={{ color: '#5e6c84' }}>Entidades</span>
          <span className="text-xs font-semibold" style={{ color: '#172b4d' }}>{form.selectedOptions.length} selecionada(s)</span>
        </div>
      )}
    </div>
  )
}

function QuickHelp() {
  return (
    <div className="rounded p-4" style={{ backgroundColor: 'rgba(222,235,255,0.3)', border: '1px solid rgba(0,82,204,0.2)' }}>
      <h4 className="text-xs font-bold uppercase mb-2" style={{ color: '#0052cc' }}>Ajuda rápida</h4>
      <ul className="space-y-2">
        {['Como configurar o Firebird?', 'Verificar conectividade de rede'].map(text => (
          <li key={text} className="flex items-center gap-2 text-xs" style={{ color: '#5e6c84' }}>
            <span className="w-1 h-1 rounded-full flex-shrink-0" style={{ backgroundColor: '#0052cc' }} />
            {text}
          </li>
        ))}
      </ul>
    </div>
  )
}

function SidebarBtn({ onClick, label }: { onClick: () => void; label: string }) {
  return (
    <button
      onClick={onClick}
      className="w-full font-bold py-3 px-4 rounded text-white flex items-center justify-center gap-2 transition-colors"
      style={{ backgroundColor: '#0052cc' }}
      onMouseEnter={e => (e.currentTarget.style.backgroundColor = '#0747a6')}
      onMouseLeave={e => (e.currentTarget.style.backgroundColor = '#0052cc')}
    >
      <span>{label}</span>
      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7l5 5m0 0l-5 5m5-5H6" />
      </svg>
    </button>
  )
}

// ─── Tab: Conexão ──────────────────────────────────────────────────────────────

interface TabConexaoProps {
  form: FormState
  errors: Record<string, string>
  onField: <K extends keyof FormState>(k: K, v: FormState[K]) => void
}

function TabConexao({ form, errors, onField }: TabConexaoProps) {
  return (
    <div className="space-y-6">
      {/* Identificação */}
      <Section title="IDENTIFICAÇÃO">
        <div className="grid grid-cols-2 gap-6">
          <Field label="Código do Cliente" error={errors.clientCode}>
            <Inp hasError={!!errors.clientCode} placeholder="Ex: 1234" value={form.clientCode}
              onChange={e => onField('clientCode', e.target.value)} />
          </Field>
          <Field label="Sistema de Origem">
            <Sel value={form.system} onChange={e => onField('system', e.target.value)}>
              {SYSTEMS.map(s => <option key={s} value={s}>{s}</option>)}
            </Sel>
          </Field>
        </div>
      </Section>

      {/* Firebird */}
      <Section title="BANCO DE DADOS FIREBIRD (ORIGEM)">
        <div className="space-y-6">
          <div className="grid grid-cols-4 gap-6">
            <Field label="Host" error={errors.host} className="col-span-3">
              <Inp hasError={!!errors.host} placeholder="localhost" value={form.host}
                onChange={e => onField('host', e.target.value)} />
            </Field>
            <Field label="Porta">
              <Inp placeholder="3050" value={form.port}
                onChange={e => onField('port', e.target.value)} />
            </Field>
          </div>
          <Field label="Caminho do Banco" error={errors.path}>
            <Inp hasError={!!errors.path} placeholder="C:\Dados\database.fdb" value={form.path}
              onChange={e => onField('path', e.target.value)} />
          </Field>
          <div className="grid grid-cols-2 gap-6">
            <Field label="Usuário" error={errors.user}>
              <Inp hasError={!!errors.user} placeholder="SYSDBA" value={form.user}
                onChange={e => onField('user', e.target.value)} />
            </Field>
            <Field label="Senha" error={errors.password}>
              <Inp type="password" hasError={!!errors.password} placeholder="••••••••" value={form.password}
                onChange={e => onField('password', e.target.value)} />
            </Field>
          </div>
          <Field label="Conversão de Charset">
            <Sel value={form.conversao} onChange={e => onField('conversao', e.target.value)}>
              <option value="WIN1252">WIN1252</option>
              <option value="UTF8">UTF8</option>
              <option value="ISO8859_1">ISO8859_1</option>
            </Sel>
          </Field>
        </div>
      </Section>

      {/* Pharmacie */}
      <Section title="PHARMACIE (DESTINO)">
        <div className="space-y-6">
          <div className="grid grid-cols-4 gap-6">
            <Field label="Alias" error={errors.alias} className="col-span-3">
              <Inp hasError={!!errors.alias} placeholder="pharmacie" value={form.alias}
                onChange={e => onField('alias', e.target.value)} />
            </Field>
            <Field label="Porta">
              <Inp placeholder="3050" value={form.porta}
                onChange={e => onField('porta', e.target.value)} />
            </Field>
          </div>
          <Field label="IP / Servidor" error={errors.ipServer}>
            <Inp hasError={!!errors.ipServer} placeholder="192.168.1.1" value={form.ipServer}
              onChange={e => onField('ipServer', e.target.value)} />
          </Field>
        </div>
      </Section>
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
}

function TabOpcoes({ form, optionDefs, errors, historyStatus, onToggleOption, onField }: TabOpcoesProps) {
  const incrementalDisabled = !historyStatus.connected

  return (
    <div className="space-y-6">
      {/* Entities */}
      <section className="bg-white rounded-xl border border-[#c3c6d6] p-6 shadow-sm">
        <div className="flex flex-col mb-5">
          <span className="text-xs font-bold uppercase tracking-widest" style={{ color: '#737685' }}>ENTIDADES PARA MIGRAR</span>
          {errors.options && (
            <span className="text-xs font-medium mt-1" style={{ color: '#ba1a1a' }}>{errors.options}</span>
          )}
        </div>
        <div className="grid grid-cols-2 gap-3">
          {optionDefs.map(opt => {
            const checked = form.selectedOptions.includes(opt.code)
            const isRefazer = opt.code.toUpperCase().includes('REFAZER')
            const desc = OPTION_DESCRIPTIONS[opt.code.toUpperCase()] ?? ''
            return (
              <label
                key={opt.code}
                className={`group flex items-center p-4 border rounded-lg cursor-pointer transition-all ${
                  isRefazer ? 'col-span-2 border-dashed' : ''
                }`}
                style={{
                  borderColor: checked ? '#0052cc' : '#c3c6d6',
                  backgroundColor: checked ? 'rgba(0,82,204,0.05)' : isRefazer ? '#f3f4f6' : 'white',
                }}
              >
                <input
                  type="checkbox"
                  className="w-5 h-5 rounded mr-4 flex-shrink-0 cursor-pointer"
                  style={{ accentColor: '#0052cc' }}
                  checked={checked}
                  onChange={() => onToggleOption(opt.code)}
                />
                <div className="flex flex-col min-w-0">
                  <span className="text-sm font-semibold" style={{ color: '#191c1e' }}>{opt.name}</span>
                  {desc && (
                    <span
                      className="text-xs mt-0.5"
                      style={{ color: isRefazer ? '#ba1a1a' : '#434654', fontWeight: isRefazer ? 600 : 400 }}
                    >
                      {desc}
                    </span>
                  )}
                </div>
              </label>
            )
          })}
        </div>
      </section>

      {/* Mode */}
      <section className="bg-white rounded-xl border border-[#c3c6d6] p-6 shadow-sm">
        <span className="text-xs font-bold uppercase tracking-widest block mb-5" style={{ color: '#737685' }}>MODO DE MIGRAÇÃO</span>
        <div className="grid grid-cols-2 gap-4">
          {(['COMPLETA', 'INCREMENTAL'] as const).map(mode => {
            const disabled = mode === 'INCREMENTAL' && incrementalDisabled
            const active = form.mode === mode
            return (
              <label
                key={mode}
                className="relative flex p-5 rounded-xl cursor-pointer transition-all"
                style={{
                  border: active ? '2px solid #0052cc' : '1px solid #c3c6d6',
                  backgroundColor: active ? 'rgba(0,82,204,0.05)' : 'white',
                  opacity: disabled ? 0.45 : 1,
                  cursor: disabled ? 'not-allowed' : 'pointer',
                  outline: active ? '1px solid rgba(0,82,204,0.3)' : 'none',
                  outlineOffset: '2px',
                }}
              >
                <input
                  type="radio"
                  name="mode"
                  value={mode}
                  checked={active}
                  disabled={disabled}
                  onChange={() => onField('mode', mode)}
                  className="hidden"
                />
                <div className="flex flex-col flex-1">
                  <div className="flex items-center justify-between mb-2">
                    <span className="text-base font-semibold" style={{ color: active ? '#0052cc' : '#191c1e' }}>
                      {mode === 'COMPLETA' ? 'Completa' : 'Incremental'}
                    </span>
                    {active ? (
                      <svg className="w-5 h-5" style={{ color: '#0052cc' }} viewBox="0 0 24 24" fill="currentColor">
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 14l-4-4 1.41-1.41L10 13.17l6.59-6.59L18 8l-8 8z" />
                      </svg>
                    ) : (
                      <span className="w-5 h-5 rounded-full border-2 border-[#c3c6d6]" />
                    )}
                  </div>
                  <p className="text-xs leading-relaxed" style={{ color: '#434654' }}>
                    {mode === 'COMPLETA'
                      ? 'Migra todos os registros selecionados, ignorando estados anteriores.'
                      : disabled
                        ? 'Requer histórico central conectado para uso.'
                        : 'Pula registros já migrados. Recomendado para sincronizações rápidas.'}
                  </p>
                </div>
              </label>
            )
          })}
        </div>
      </section>
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
  selectedNames: string[]
}

function TabMigrar({ form, errors, lastRun, historyStatus, selectedNames }: TabMigrarProps) {
  return (
    <div className="space-y-6">
      <Section title="RESUMO DA MIGRAÇÃO">
        <dl className="grid grid-cols-2 gap-x-8 gap-y-4">
          <SummaryItem label="Cliente" value={form.clientCode || '—'} />
          <SummaryItem label="Sistema" value={form.system} />
          <SummaryItem label="Host" value={form.host || '—'} />
          <SummaryItem label="Modo" value={form.mode} />
          <SummaryItem label="Entidades" value={selectedNames.length > 0 ? selectedNames.join(', ') : '—'} className="col-span-2" />
        </dl>
      </Section>

      {historyStatus.connected && (
        <Section title="ÚLTIMO RUN BEM-SUCEDIDO">
          {lastRun ? (
            <dl className="grid grid-cols-2 gap-x-8 gap-y-4">
              <SummaryItem label="Iniciado em" value={lastRun.startedAt} />
              <SummaryItem label="Modo" value={lastRun.mode} />
              <SummaryItem label="Implantador" value={lastRun.implantador || '—'} />
              <SummaryItem label="Status" value={lastRun.status} />
            </dl>
          ) : (
            <p className="text-sm" style={{ color: '#5e6c84' }}>Nenhum run anterior encontrado para este cliente.</p>
          )}
        </Section>
      )}

      {Object.keys(errors).length > 0 && (
        <div
          className="bg-white p-6 rounded border shadow-sm space-y-2"
          style={{ borderColor: '#ffd2cc', borderLeftWidth: 4, borderLeftColor: '#de350b' }}
        >
          <h3 className="text-xs font-bold uppercase tracking-wider mb-3" style={{ color: '#de350b' }}>Campos obrigatórios</h3>
          {Object.values(errors).map((e, i) => (
            <p key={i} className="text-sm flex items-center gap-2" style={{ color: '#de350b' }}>
              <span className="w-1 h-1 rounded-full flex-shrink-0" style={{ backgroundColor: '#de350b' }} />
              {e}
            </p>
          ))}
        </div>
      )}
    </div>
  )
}

// ─── Shared primitives ─────────────────────────────────────────────────────────

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="bg-white p-6 rounded border border-[#dfe1e6] shadow-sm">
      <h2 className="text-xs font-bold uppercase tracking-wider mb-6" style={{ color: '#42526e' }}>{title}</h2>
      {children}
    </section>
  )
}

function Field({
  label, error, hint, className = '', children,
}: {
  label: string; error?: string; hint?: string; className?: string; children: React.ReactNode
}) {
  return (
    <div className={`space-y-1.5 ${className}`}>
      <label className="block text-sm font-semibold" style={{ color: '#42526e' }}>{label}</label>
      {children}
      {hint && !error && <p className="text-[11px]" style={{ color: '#7a869a' }}>{hint}</p>}
      {error && <p className="text-[11px] font-medium" style={{ color: '#de350b' }}>{error}</p>}
    </div>
  )
}

function Inp({
  hasError, className = '', type = 'text', ...props
}: React.InputHTMLAttributes<HTMLInputElement> & { hasError?: boolean }) {
  return (
    <input
      type={type}
      className={`w-full h-10 px-3 border rounded text-sm bg-white text-[#172b4d] placeholder-[#97a0af] focus:outline-none transition-all ${className}`}
      style={{ borderColor: hasError ? '#de350b' : '#c1c7d0' }}
      onFocus={e => { e.currentTarget.style.borderColor = hasError ? '#de350b' : '#0052cc'; e.currentTarget.style.boxShadow = '0 0 0 2px rgba(0,82,204,0.2)' }}
      onBlur={e => { e.currentTarget.style.borderColor = hasError ? '#de350b' : '#c1c7d0'; e.currentTarget.style.boxShadow = 'none' }}
      {...props}
    />
  )
}

function Sel({ children, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <select
      className="w-full h-10 px-3 border border-[#c1c7d0] rounded text-sm text-[#172b4d] focus:outline-none transition-all"
      style={{ backgroundColor: '#f4f5f7' }}
      onFocus={e => { e.currentTarget.style.borderColor = '#0052cc'; e.currentTarget.style.boxShadow = '0 0 0 2px rgba(0,82,204,0.2)' }}
      onBlur={e => { e.currentTarget.style.borderColor = '#c1c7d0'; e.currentTarget.style.boxShadow = 'none' }}
      {...props}
    >
      {children}
    </select>
  )
}

function SummaryItem({ label, value, className = '' }: { label: string; value: string; className?: string }) {
  return (
    <div className={className}>
      <dt className="text-xs mb-0.5" style={{ color: '#5e6c84' }}>{label}</dt>
      <dd className="text-sm font-semibold" style={{ color: '#172b4d' }}>{value}</dd>
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
