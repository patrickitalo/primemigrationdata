import { useState, useEffect, useRef } from 'react'
import { Login } from '../../wailsjs/go/app/App'
import type { models } from '../../wailsjs/go/models'

interface Props {
  onSuccess: (session: models.UserSession) => void
}

export default function LoginPage({ onSuccess }: Props) {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [remember, setRemember] = useState(false)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [focused, setFocused] = useState<'username' | 'password' | null>(null)
  const userRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const saved = localStorage.getItem('prime:lastUser')
    if (saved) {
      setUsername(saved)
      setRemember(true)
    }
    userRef.current?.focus()
  }, [])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!username.trim() || !password.trim()) {
      setError('Preencha o login e a senha.')
      return
    }
    setLoading(true)
    setError('')
    try {
      const result = await Login(username.trim(), password)
      if (result.error) {
        setError(result.error)
        return
      }
      if (result.session) {
        if (remember) {
          localStorage.setItem('prime:lastUser', username.trim())
        } else {
          localStorage.removeItem('prime:lastUser')
        }
        onSuccess(result.session)
      }
    } catch (err) {
      setError(String(err))
    } finally {
      setLoading(false)
    }
  }

  const iconColor = (field: 'username' | 'password') =>
    focused === field ? '#003d9b' : '#737685'

  return (
    <div className="h-full flex flex-col" style={{ backgroundColor: '#F4F5F7' }}>
      {/* Nav */}
      <nav
        className="bg-white border-b flex justify-between items-center px-10 py-4 flex-shrink-0"
        style={{ borderColor: '#c3c6d6', '--wails-draggable': 'drag' } as React.CSSProperties}
      >
        <span className="text-2xl font-bold" style={{ color: '#003d9b' }}>Prime Migration</span>
        <div className="flex items-center gap-1.5 text-xs" style={{ color: '#434654' }}>
          <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z" />
          </svg>
          <span>Ambiente Seguro</span>
        </div>
      </nav>

      {/* Main */}
      <main className="flex-1 flex flex-col items-center justify-center px-4">
        <div
          className="w-full max-w-[440px] bg-white flex flex-col gap-8 rounded-lg p-8 shadow-sm"
          style={{ border: '1px solid #DFE1E6' }}
        >
          {/* Logo + heading */}
          <div className="flex flex-col items-center gap-4">
            <div className="w-16 h-16 rounded-lg flex items-center justify-center text-white" style={{ backgroundColor: '#0052cc' }}>
              <svg className="w-9 h-9" viewBox="0 0 24 24" fill="currentColor">
                <path d="M2.5 19h19v2h-19zm7.18-1.73l4.35 1.16 6.87-1.99c.48-.14.74-.64.54-1.09l-.06-.13c-.19-.38-.63-.57-1.03-.45l-2.79.8-1.81-.87.87-1.28A8.17 8.17 0 0019 9h-1.29a6.9 6.9 0 01-3.13 4.7l-1.2.73-2.32-1.11c-.16-.08-.32-.11-.48-.11-.43 0-.83.25-1.02.66l-.03.06c-.22.48-.03 1.05.44 1.29l.98.47.73-.21-.99 3.49zm10.32-8.77c0 1.1-.9 2-2 2s-2-.9-2-2 .9-2 2-2 2 .9 2 2zM5 13c0-3.09 2.03-5.71 5-6.32V5.08C5.22 5.7 3 9.1 3 13c0 1.8.56 3.48 1.53 4.86l1.19-.51A6.94 6.94 0 015 13z" />
              </svg>
            </div>
            <div className="text-center">
              <h1 className="text-2xl font-semibold" style={{ color: '#191c1e' }}>Bem-vindo de volta</h1>
              <p className="text-sm mt-1" style={{ color: '#434654' }}>Acesse o console Prime Migration</p>
            </div>
          </div>

          {/* Form */}
          <form onSubmit={handleSubmit} className="flex flex-col gap-5">
            {/* Username */}
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-semibold tracking-wider uppercase" style={{ color: '#191c1e' }} htmlFor="username">
                Login
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 transition-colors" style={{ color: iconColor('username') }}>
                  <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" />
                  </svg>
                </span>
                <input
                  ref={userRef}
                  id="username"
                  type="text"
                  inputMode="numeric"
                  className="w-full pl-10 pr-4 py-3 bg-white rounded-lg text-base transition-all"
                  style={{
                    border: `1px solid ${focused === 'username' ? '#003d9b' : '#c3c6d6'}`,
                    outline: focused === 'username' ? '2px solid rgba(0,61,155,0.2)' : 'none',
                    outlineOffset: '0px',
                  }}
                  placeholder="Código de usuário"
                  value={username}
                  onChange={e => setUsername(e.target.value)}
                  onFocus={() => setFocused('username')}
                  onBlur={() => setFocused(null)}
                  disabled={loading}
                  autoComplete="username"
                />
              </div>
            </div>

            {/* Password */}
            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-semibold tracking-wider uppercase" style={{ color: '#191c1e' }} htmlFor="password">
                Senha
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 transition-colors" style={{ color: iconColor('password') }}>
                  <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                    <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z" />
                  </svg>
                </span>
                <input
                  id="password"
                  type="password"
                  className="w-full pl-10 pr-4 py-3 bg-white rounded-lg text-base transition-all"
                  style={{
                    border: `1px solid ${focused === 'password' ? '#003d9b' : '#c3c6d6'}`,
                    outline: focused === 'password' ? '2px solid rgba(0,61,155,0.2)' : 'none',
                    outlineOffset: '0px',
                  }}
                  placeholder="••••••••"
                  value={password}
                  onChange={e => setPassword(e.target.value)}
                  onFocus={() => setFocused('password')}
                  onBlur={() => setFocused(null)}
                  disabled={loading}
                  autoComplete="current-password"
                />
              </div>
            </div>

            {/* Remember */}
            <div className="flex items-center gap-2">
              <input
                id="remember"
                type="checkbox"
                className="w-4 h-4 rounded cursor-pointer"
                style={{ accentColor: '#003d9b' }}
                checked={remember}
                onChange={e => setRemember(e.target.checked)}
              />
              <label htmlFor="remember" className="text-sm cursor-pointer select-none" style={{ color: '#434654' }}>
                Lembrar login
              </label>
            </div>

            {/* Error */}
            {error && (
              <div
                className="flex items-start gap-2 px-3 py-2.5 rounded-lg text-sm"
                style={{ backgroundColor: '#ffdad6', border: '1px solid #ffb3ae', color: '#93000a' }}
              >
                <svg className="w-4 h-4 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
                </svg>
                <span>{error}</span>
              </div>
            )}

            {/* Submit */}
            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 px-6 rounded-lg text-xs font-semibold tracking-wider uppercase flex items-center justify-center gap-2 group transition-all active:scale-[0.98] disabled:opacity-70 disabled:cursor-not-allowed"
              style={{ backgroundColor: '#003d9b', color: '#ffffff' }}
              onMouseEnter={e => !loading && ((e.currentTarget.style.backgroundColor = '#0c56d0'))}
              onMouseLeave={e => !loading && ((e.currentTarget.style.backgroundColor = '#003d9b'))}
            >
              {loading ? (
                <>
                  <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  Autenticando...
                </>
              ) : (
                <>
                  Entrar
                  <svg className="w-4 h-4 group-hover:translate-x-1 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                  </svg>
                </>
              )}
            </button>
          </form>

          {/* Trust badge */}
          <div className="flex items-center justify-center gap-2 text-xs opacity-60" style={{ color: '#434654' }}>
            <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm-2 16l-4-4 1.41-1.41L10 14.17l6.59-6.59L18 9l-8 8z" />
            </svg>
            Ambiente Corporativo Seguro
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="flex justify-center items-center py-4 flex-shrink-0">
        <span className="text-xs" style={{ color: '#737685' }}>
          © {new Date().getFullYear()} Prime Software. Todos os direitos reservados.
        </span>
      </footer>

      {/* Background blobs */}
      <div className="fixed inset-0 -z-10 pointer-events-none overflow-hidden">
        <div className="absolute -top-1/4 -left-1/4 w-3/5 h-3/5 rounded-full blur-[120px]" style={{ backgroundColor: '#dae2ff', opacity: 0.4 }} />
        <div className="absolute -bottom-1/4 -right-1/4 w-3/5 h-3/5 rounded-full blur-[120px]" style={{ backgroundColor: '#d8e2ff', opacity: 0.4 }} />
      </div>
    </div>
  )
}
