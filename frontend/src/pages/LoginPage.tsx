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

  return (
    <div className="h-full flex items-center justify-center relative overflow-hidden">
      {/* Background gradient blobs */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -left-40 w-96 h-96 bg-prime-900/30 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 -right-40 w-96 h-96 bg-prime-800/20 rounded-full blur-3xl" />
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[400px] bg-prime-950/40 rounded-full blur-3xl" />
      </div>

      {/* Login card */}
      <div className="relative w-full max-w-sm mx-4">
        <div className="bg-slate-900/80 backdrop-blur border border-slate-700/60 rounded-2xl p-8 shadow-2xl">
          {/* Logo area */}
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-prime-700 mb-4 shadow-lg shadow-prime-900/50">
              <svg className="w-8 h-8 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 7v10c0 2 1 3 3 3h10c2 0 3-1 3-3V7c0-2-1-3-3-3H7C5 4 4 5 4 7z" />
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h6M9 8h6M9 16h4" />
              </svg>
            </div>
            <h1 className="text-xl font-bold font-display text-white tracking-tight">Prime Migration</h1>
            <p className="text-sm text-slate-400 mt-1">Acesse sua conta para continuar</p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="label">Login</label>
              <input
                ref={userRef}
                type="text"
                inputMode="numeric"
                className="input-field"
                placeholder="Código de usuário"
                value={username}
                onChange={e => setUsername(e.target.value)}
                disabled={loading}
                autoComplete="username"
              />
            </div>

            <div>
              <label className="label">Senha</label>
              <input
                type="password"
                className="input-field"
                placeholder="••••••••"
                value={password}
                onChange={e => setPassword(e.target.value)}
                disabled={loading}
                autoComplete="current-password"
              />
            </div>

            <div className="flex items-center gap-2">
              <input
                id="remember"
                type="checkbox"
                className="w-4 h-4 rounded border-slate-600 bg-slate-800 text-prime-600 focus:ring-prime-500 focus:ring-offset-slate-900"
                checked={remember}
                onChange={e => setRemember(e.target.checked)}
              />
              <label htmlFor="remember" className="text-sm text-slate-400 cursor-pointer select-none">
                Lembrar login
              </label>
            </div>

            {error && (
              <div className="flex items-start gap-2 px-3 py-2.5 rounded-lg bg-red-950/60 border border-red-800/50 text-red-300 text-sm">
                <svg className="w-4 h-4 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
                </svg>
                <span>{error}</span>
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full py-2.5 mt-2"
            >
              {loading ? (
                <>
                  <svg className="w-4 h-4 animate-spin" viewBox="0 0 24 24" fill="none">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                  </svg>
                  Entrando...
                </>
              ) : (
                'Entrar'
              )}
            </button>
          </form>
        </div>

        <p className="text-center text-xs text-slate-600 mt-4">
          Prime Software &copy; {new Date().getFullYear()}
        </p>
      </div>
    </div>
  )
}
