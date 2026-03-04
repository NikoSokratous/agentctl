import { useState, useEffect } from 'react';
import { LineChart, Line, BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import LoadingSpinner from '../components/LoadingSpinner';

interface CostData {
  agent_id: string;
  cost: number;
  requests: number;
}

interface SLAData {
  service_id: string;
  uptime: number;
  avg_latency: number;
  error_rate: number;
}

interface PerformanceData {
  timestamp: string;
  cpu: number;
  memory: number;
  throughput: number;
}

interface DenialLog {
  id: string;
  timestamp: string;
  agent_name: string;
  policy_name: string;
  tool: string;
  deny_reason?: string;
  risk_score: number;
}

interface DenialsStats {
  denied: number;
  allowed: number;
  top_deny_reasons: { reason: string; count: number }[];
}

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];

export default function AnalyticsDashboard() {
  const [timeRange, setTimeRange] = useState('24h');
  const [costData, setCostData] = useState<CostData[]>([]);
  const [slaData, setSLAData] = useState<SLAData[]>([]);
  const [performanceData, setPerformanceData] = useState<PerformanceData[]>([]);
  const [totalCost, setTotalCost] = useState(0);
  const [loading, setLoading] = useState(true);
  const [denialLogs, setDenialLogs] = useState<DenialLog[]>([]);
  const [denialsStats, setDenialsStats] = useState<DenialsStats | null>(null);

  useEffect(() => {
    loadAnalytics();
  }, [timeRange]);

  const loadAnalytics = async () => {
    setLoading(true);
    try {
      // Load cost data
      const costResponse = await fetch(`/api/v1/analytics/costs?range=${timeRange}`);
      const costs = await costResponse.json();
      setCostData(costs.by_agent || []);
      setTotalCost(costs.total || 0);

      // Load SLA data
      const slaResponse = await fetch(`/api/v1/analytics/sla?range=${timeRange}`);
      const sla = await slaResponse.json();
      setSLAData(sla.services || []);

      // Load performance data
      const perfResponse = await fetch(`/api/v1/analytics/performance?range=${timeRange}`);
      const perf = await perfResponse.json();
      setPerformanceData(perf.timeline || []);

      try {
        const denialsResponse = await fetch(`/api/v1/analytics/denials?range=${timeRange}`);
        const denials = await denialsResponse.json();
        setDenialLogs(denials.logs || []);
        setDenialsStats(denials.stats || null);
      } catch { /* optional */ }
    } catch (error) {
      console.error('Failed to load analytics:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="analytics-page"><LoadingSpinner message="Loading analytics..." /></div>;
  }

  return (
    <div className="analytics-dashboard">
      <div className="dashboard-header">
        <h1>Analytics Dashboard</h1>
        <div className="time-range-selector">
          <button
            className={timeRange === '1h' ? 'active' : ''}
            onClick={() => setTimeRange('1h')}
          >
            1 Hour
          </button>
          <button
            className={timeRange === '24h' ? 'active' : ''}
            onClick={() => setTimeRange('24h')}
          >
            24 Hours
          </button>
          <button
            className={timeRange === '7d' ? 'active' : ''}
            onClick={() => setTimeRange('7d')}
          >
            7 Days
          </button>
          <button
            className={timeRange === '30d' ? 'active' : ''}
            onClick={() => setTimeRange('30d')}
          >
            30 Days
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="summary-cards">
        <div className="card">
          <div className="card-icon cost">💰</div>
          <div className="card-content">
            <div className="card-value">${totalCost.toFixed(2)}</div>
            <div className="card-label">Total Cost</div>
          </div>
        </div>

        <div className="card">
          <div className="card-icon sla">📊</div>
          <div className="card-content">
            <div className="card-value">
              {slaData.length > 0
                ? (slaData.reduce((sum, s) => sum + s.uptime, 0) / slaData.length).toFixed(1)
                : '0'}%
            </div>
            <div className="card-label">Average Uptime</div>
          </div>
        </div>

        <div className="card">
          <div className="card-icon perf">⚡</div>
          <div className="card-content">
            <div className="card-value">
              {costData.reduce((sum, c) => sum + c.requests, 0).toLocaleString()}
            </div>
            <div className="card-label">Total Requests</div>
          </div>
        </div>

        <div className="card">
          <div className="card-icon error">⚠️</div>
          <div className="card-content">
            <div className="card-value">
              {slaData.length > 0
                ? (slaData.reduce((sum, s) => sum + s.error_rate, 0) / slaData.length * 100).toFixed(2)
                : '0'}%
            </div>
            <div className="card-label">Error Rate</div>
          </div>
        </div>

        <div className="card">
          <div className="card-icon deny">🛡️</div>
          <div className="card-content">
            <div className="card-value">{denialsStats?.denied ?? denialLogs?.length ?? 0}</div>
            <div className="card-label">Policy Denials</div>
          </div>
        </div>
      </div>

      {/* Policy Denials */}
      {(denialLogs.length > 0 || (denialsStats && denialsStats.denied > 0)) && (
        <div className="chart-section">
          <h2>Policy Denials</h2>
          {denialsStats && (
            <div className="denials-summary">
              <span>Denied: {denialsStats.denied}</span>
              <span>Allowed: {denialsStats.allowed}</span>
              {denialsStats.top_deny_reasons?.length > 0 && (
                <div className="top-reasons">
                  <strong>Top reasons:</strong>
                  {denialsStats.top_deny_reasons.slice(0, 5).map((r, i) => (
                    <span key={i}>{r.reason}: {r.count}</span>
                  ))}
                </div>
              )}
            </div>
          )}
          {denialLogs.length > 0 && (
            <table className="data-table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Agent</th>
                  <th>Tool</th>
                  <th>Reason</th>
                  <th>Risk</th>
                </tr>
              </thead>
              <tbody>
                {denialLogs.slice(0, 20).map((log) => (
                  <tr key={log.id}>
                    <td>{log.timestamp ? new Date(log.timestamp).toLocaleString() : '-'}</td>
                    <td>{log.agent_name}</td>
                    <td>{log.tool}</td>
                    <td>{log.deny_reason || '-'}</td>
                    <td>{(log.risk_score * 100).toFixed(0)}%</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Cost by Agent */}
      <div className="chart-section">
        <h2>Cost by Agent</h2>
        <ResponsiveContainer width="100%" height={300}>
          <PieChart>
            <Pie
              data={costData}
              dataKey="cost"
              nameKey="agent_id"
              cx="50%"
              cy="50%"
              outerRadius={100}
              label={(entry) => `${entry.agent_id}: $${entry.cost.toFixed(2)}`}
            >
              {costData.map((_, index) => (
                <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
              ))}
            </Pie>
            <Tooltip formatter={(value: number) => `$${value.toFixed(2)}`} />
            <Legend />
          </PieChart>
        </ResponsiveContainer>
      </div>

      {/* Performance Over Time */}
      <div className="chart-section">
        <h2>Performance Metrics</h2>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={performanceData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="timestamp" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Line type="monotone" dataKey="cpu" stroke="#8884d8" name="CPU %" />
            <Line type="monotone" dataKey="memory" stroke="#82ca9d" name="Memory %" />
            <Line type="monotone" dataKey="throughput" stroke="#ffc658" name="Throughput" />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* SLA Metrics */}
      <div className="chart-section">
        <h2>SLA Metrics by Service</h2>
        <ResponsiveContainer width="100%" height={300}>
          <BarChart data={slaData}>
            <CartesianGrid strokeDasharray="3 3" />
            <XAxis dataKey="service_id" />
            <YAxis />
            <Tooltip />
            <Legend />
            <Bar dataKey="uptime" fill="#00C49F" name="Uptime %" />
            <Bar dataKey="avg_latency" fill="#FF8042" name="Avg Latency (ms)" />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Detailed Tables */}
      <div className="tables-section">
        <div className="table-container">
          <h3>Cost Breakdown</h3>
          <table className="data-table">
            <thead>
              <tr>
                <th>Agent</th>
                <th>Requests</th>
                <th>Cost</th>
                <th>Cost per Request</th>
              </tr>
            </thead>
            <tbody>
              {costData.map((item) => (
                <tr key={item.agent_id}>
                  <td>{item.agent_id}</td>
                  <td>{item.requests.toLocaleString()}</td>
                  <td>${item.cost.toFixed(2)}</td>
                  <td>${(item.cost / item.requests).toFixed(4)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="table-container">
          <h3>SLA Status</h3>
          <table className="data-table">
            <thead>
              <tr>
                <th>Service</th>
                <th>Uptime</th>
                <th>Avg Latency</th>
                <th>Error Rate</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {slaData.map((item) => (
                <tr key={item.service_id}>
                  <td>{item.service_id}</td>
                  <td>{item.uptime.toFixed(2)}%</td>
                  <td>{item.avg_latency.toFixed(0)}ms</td>
                  <td>{(item.error_rate * 100).toFixed(2)}%</td>
                  <td>
                    <span className={`status ${item.uptime >= 99.9 ? 'healthy' : 'warning'}`}>
                      {item.uptime >= 99.9 ? '✓ Healthy' : '⚠ Warning'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <style>{`
        .analytics-dashboard, .analytics-page {
          padding: 2rem;
          max-width: 1400px;
          margin: 0 auto;
        }

        .dashboard-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 2rem;
        }

        .dashboard-header h1 { margin: 0; color: var(--color-text); }

        .time-range-selector { display: flex; gap: 0.5rem; }

        .time-range-selector button {
          padding: 0.5rem 1rem;
          border: 1px solid var(--color-border);
          background: var(--color-bg-card);
          color: var(--color-text-muted);
          border-radius: 0.375rem;
          cursor: pointer;
          transition: all 0.2s;
        }

        .time-range-selector button:hover {
          background: var(--color-bg-hover);
          color: var(--color-text);
        }

        .time-range-selector button.active {
          background: var(--color-accent);
          color: white;
          border-color: var(--color-accent);
        }

        .summary-cards {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
          gap: 1.25rem;
          margin-bottom: 2rem;
        }

        .card {
          background: var(--color-bg-card);
          border: 1px solid var(--color-border);
          border-radius: 0.75rem;
          padding: 1.25rem;
          display: flex;
          align-items: center;
          gap: 1rem;
          transition: all 0.2s;
        }

        .card:hover { border-color: var(--color-border-strong); }

        .card-icon {
          font-size: 1.5rem;
          width: 2.75rem;
          height: 2.75rem;
          display: flex;
          align-items: center;
          justify-content: center;
          border-radius: 0.5rem;
        }

        .card-icon.cost { background: rgba(245, 158, 11, 0.2); }
        .card-icon.sla { background: rgba(99, 102, 241, 0.2); }
        .card-icon.perf { background: rgba(34, 197, 94, 0.2); }
        .card-icon.error { background: rgba(239, 68, 68, 0.2); }
        .card-icon.deny { background: rgba(99, 102, 241, 0.15); }

        .denials-summary {
          display: flex;
          flex-wrap: wrap;
          gap: 1rem;
          margin-bottom: 1rem;
          font-size: 0.875rem;
        }

        .top-reasons {
          display: flex;
          flex-wrap: wrap;
          gap: 0.5rem;
        }

        .card-content {
          flex: 1;
        }

        .card-value { font-size: 1.5rem; font-weight: 600; color: var(--color-text); }
        .card-label { font-size: 0.875rem; color: var(--color-text-muted); }

        .chart-section {
          background: var(--color-bg-card);
          border: 1px solid var(--color-border);
          border-radius: 0.75rem;
          padding: 1.5rem;
          margin-bottom: 2rem;
        }

        .chart-section h2 { margin: 0 0 1rem 0; font-size: 1.125rem; color: var(--color-text); }

        .tables-section {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(420px, 1fr));
          gap: 1.25rem;
        }

        .table-container {
          background: var(--color-bg-card);
          border: 1px solid var(--color-border);
          border-radius: 0.75rem;
          padding: 1.5rem;
        }

        .table-container h3 { margin: 0 0 1rem 0; color: var(--color-text); }

        .data-table { width: 100%; border-collapse: collapse; }
        .data-table th, .data-table td { padding: 0.75rem; text-align: left; border-bottom: 1px solid var(--color-border); }
        .data-table th { background: var(--color-bg-elevated); font-weight: 600; font-size: 0.8125rem; color: var(--color-text-muted); }
        .data-table td { font-size: 0.875rem; color: var(--color-text); }

        .status {
          padding: 0.25rem 0.75rem;
          border-radius: 1rem;
          font-size: 0.75rem;
          font-weight: 500;
        }

        .status.healthy { background: var(--color-success-muted); color: var(--color-success); }
        .status.warning { background: var(--color-warning-muted); color: var(--color-warning); }
        .loading { text-align: center; padding: 3rem; color: var(--color-text-muted); }
      `}</style>
    </div>
  );
}
