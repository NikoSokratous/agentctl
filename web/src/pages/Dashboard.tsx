import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Activity, CheckCircle, XCircle, AlertCircle } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import './Dashboard.css'

interface Run {
  run_id: string
  agent: string
  goal: string
  state: string
  created_at: string
  completed_at?: string
}

interface RunsResponse {
  runs: Run[]
  total: number
}

const fetchRuns = async (): Promise<RunsResponse> => {
  const response = await fetch('/v1/runs')
  if (!response.ok) {
    throw new Error('Failed to fetch runs')
  }
  return response.json()
}

const Dashboard: React.FC = () => {
  const navigate = useNavigate()
  const { data, isLoading, error } = useQuery({
    queryKey: ['runs'],
    queryFn: fetchRuns,
    refetchInterval: 5000,
  })

  const getStatusIcon = (state: string) => {
    switch (state) {
      case 'completed':
        return <CheckCircle className="status-icon success" size={20} />
      case 'failed':
        return <XCircle className="status-icon error" size={20} />
      case 'running':
        return <Activity className="status-icon running" size={20} />
      case 'pending':
        return <Activity className="status-icon pending" size={20} />
      default:
        return <AlertCircle className="status-icon" size={20} />
    }
  }

  const getStatusClass = (state: string) => {
    return `status-badge ${state}`
  }

  if (isLoading) {
    return (
      <div className="dashboard">
        <div className="loading">Loading runs...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="dashboard">
        <div className="error">
          Error loading runs: {error instanceof Error ? error.message : 'Unknown error'}
        </div>
      </div>
    )
  }

  return (
    <div className="dashboard">
      <header className="dashboard-header">
        <div>
          <h1>Agent Runs</h1>
          <p className="subtitle">Monitor and manage your autonomous agents</p>
        </div>
        <div className="header-stats">
          <div className="stat-card">
            <div className="stat-value">{data?.total || 0}</div>
            <div className="stat-label">Total Runs</div>
          </div>
          <div className="stat-card">
            <div className="stat-value">
              {data?.runs?.filter(r => r.state === 'running').length || 0}
            </div>
            <div className="stat-label">Active</div>
          </div>
        </div>
      </header>

      <div className="runs-container">
        {!data?.runs || data.runs.length === 0 ? (
          <div className="empty-state">
            <Activity size={48} />
            <h2>No agent runs yet</h2>
            <p>Start an agent run to see it appear here</p>
          </div>
        ) : (
          <div className="runs-grid">
            {data.runs.map((run) => (
              <div
                key={run.run_id}
                className="run-card"
                onClick={() => navigate(`/runs/${run.run_id}`)}
              >
                <div className="run-header">
                  <div className="run-info">
                    {getStatusIcon(run.state)}
                    <div>
                      <h3 className="run-agent">{run.agent}</h3>
                      <p className="run-id">{run.run_id}</p>
                    </div>
                  </div>
                  <span className={getStatusClass(run.state)}>
                    {run.state}
                  </span>
                </div>
                
                <p className="run-goal">{run.goal}</p>
                
                <div className="run-footer">
                  <span className="run-time">
                    {formatDistanceToNow(new Date(run.created_at), { addSuffix: true })}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

export default Dashboard
