import { useState } from 'react'
import LoginPage from './pages/LoginPage'
import MainPage from './pages/MainPage'
import MigrationPage from './pages/MigrationPage'
import type { models } from '../wailsjs/go/models'

type AppView = 'login' | 'main' | 'migration'

export default function App() {
  const [view, setView] = useState<AppView>('login')
  const [session, setSession] = useState<models.UserSession | null>(null)
  const [migrationConfig, setMigrationConfig] = useState<models.MigrationConfig | null>(null)

  function handleLoginSuccess(s: models.UserSession) {
    setSession(s)
    setView('main')
  }

  function handleStartMigration(cfg: models.MigrationConfig) {
    setMigrationConfig(cfg)
    setView('migration')
  }

  function handleMigrationDone() {
    setView('main')
    setMigrationConfig(null)
  }

  function handleLogout() {
    setSession(null)
    setMigrationConfig(null)
    setView('login')
  }

  return (
    <div className="h-screen flex flex-col overflow-hidden bg-slate-950">
      {view === 'login' && (
        <LoginPage onSuccess={handleLoginSuccess} />
      )}
      {view === 'main' && session && (
        <MainPage
          session={session}
          onLogout={handleLogout}
          onStartMigration={handleStartMigration}
        />
      )}
      {view === 'migration' && migrationConfig && session && (
        <MigrationPage
          config={migrationConfig}
          session={session}
          onClose={handleMigrationDone}
        />
      )}
    </div>
  )
}
