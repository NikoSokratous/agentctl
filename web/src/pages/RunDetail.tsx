import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, Activity } from 'lucide-react'
import { format } from 'date-fns'
import LoadingSpinner from '../components/LoadingSpinner'
import './RunDetail.css'

interface RunMeta {
  run_id: string
  agent: string
  goal: string
  state: string
  created_at: string
  completed_at?: string
  model?: string
  autonomy_level?: string
}

interface Event {
  timestamp: string
  type: string
  agent: string
  data: Record<string, any>
}

const fetchRun = async (runId: string): Promise<RunMeta> => {
  const response = await fetch(`/v1/runs/${runId}`)
  if (!response.ok) {
    throw new Error('Failed to fetch run')
  }
  return response.json()
}

const RunDetail: React.FC = () => {
  const { runId } = useParams<{ runId: string }>()
  const navigate = useNavigate()
  const [events, setEvents] = useState<Event[]>([])
  const [isStreaming, setIsStreaming] = useState(false)

  const { data: run, isLoading, error } = useQuery({
    queryKey: ['run', runId],
    queryFn: () => fetchRun(runId!),
    enabled: !!runId,
  })

  useEffect(() => {
    if (!runId) return

    setIsStreaming(true)
    const eventSource = new EventSource(`/v1/runs/${runId}/stream`)

    eventSource.onmessage = (event) => {
      if (event.data === ': heartbeat') return
      
      try {
        const eventData = JSON.parse(event.data)
        setEvents((prev) => [...prev, eventData])
      } catch (err) {
        console.error('Failed to parse event:', err)
      }
    }

    eventSource.onerror = () => {
      setIsStreaming(false)
      eventSource.close()
    }

    return () => {
      eventSource.close()
      setIsStreaming(false)
    }
  }, [runId])

  if (isLoading) {
    return <div className="run-detail"><LoadingSpinner message="Loading run..." /></div>
  }

  if (error || !run) {
    return (
      <div className="run-detail error">
        Error loading run: {error instanceof Error ? error.message : 'Unknown error'}
      </div>
    )
  }

  return (
    <div className="run-detail">
      <header className="detail-header">
        <button className="back-button" onClick={() => navigate('/')}>
          <ArrowLeft size={20} />
          Back to Dashboard
        </button>
        
        <div className="run-title">
          <h1>{run.agent}</h1>
          <p className="run-id-display">{run.run_id}</p>
        </div>

        <div className="run-metadata">
          <div className="metadata-item">
            <span className="label">Status:</span>
            <span className={`status-badge ${run.state}`}>{run.state}</span>
          </div>
          {run.model && (
            <div className="metadata-item">
              <span className="label">Model:</span>
              <span className="value">{run.model}</span>
            </div>
          )}
          {run.autonomy_level && (
            <div className="metadata-item">
              <span className="label">Autonomy:</span>
              <span className="value">{run.autonomy_level}</span>
            </div>
          )}
        </div>
      </header>

      <section className="goal-section">
        <h2>Goal</h2>
        <div className="goal-box">{run.goal}</div>
      </section>

      <section className="timeline-section">
        <div className="timeline-header">
          <h2>Execution Timeline</h2>
          {isStreaming && (
            <span className="streaming-indicator">
              <Activity size={16} />
              Live
            </span>
          )}
        </div>

        <div className="timeline">
          {events.length === 0 ? (
            <div className="no-events">No events yet</div>
          ) : (
            events.map((event, idx) => (
              <div key={idx} className="timeline-event">
                <div className="event-time">
                  {format(new Date(event.timestamp), 'HH:mm:ss.SSS')}
                </div>
                <div className="event-content">
                  <div className="event-type">{event.type}</div>
                  <pre className="event-data">
                    {JSON.stringify(event.data, null, 2)}
                  </pre>
                </div>
              </div>
            ))
          )}
        </div>
      </section>
    </div>
  )
}

export default RunDetail
