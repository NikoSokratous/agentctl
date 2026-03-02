import { useState, useEffect } from 'react';
import { LineChart, Line, BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

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

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];

export default function AnalyticsDashboard() {
  const [timeRange, setTimeRange] = useState('24h');
  const [costData, setCostData] = useState<CostData[]>([]);
  const [slaData, setSLAData] = useState<SLAData[]>([]);
  const [performanceData, setPerformanceData] = useState<PerformanceData[]>([]);
  const [totalCost, setTotalCost] = useState(0);
  const [loading, setLoading] = useState(true);

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
    } catch (error) {
      console.error('Failed to load analytics:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="loading">Loading analytics...</div>;
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
      </div>

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
              {costData.map((entry, index) => (
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
        .analytics-dashboard {
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

        .dashboard-header h1 {
          margin: 0;
        }

        .time-range-selector {
          display: flex;
          gap: 0.5rem;
        }

        .time-range-selector button {
          padding: 0.5rem 1rem;
          border: 1px solid #d1d5db;
          background: white;
          border-radius: 0.375rem;
          cursor: pointer;
          transition: all 0.2s;
        }

        .time-range-selector button:hover {
          background: #f3f4f6;
        }

        .time-range-selector button.active {
          background: #3b82f6;
          color: white;
          border-color: #3b82f6;
        }

        .summary-cards {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
          gap: 1.5rem;
          margin-bottom: 2rem;
        }

        .card {
          background: white;
          border: 1px solid #e5e7eb;
          border-radius: 0.5rem;
          padding: 1.5rem;
          display: flex;
          align-items: center;
          gap: 1rem;
          transition: all 0.2s;
        }

        .card:hover {
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .card-icon {
          font-size: 2rem;
          width: 3rem;
          height: 3rem;
          display: flex;
          align-items: center;
          justify-content: center;
          border-radius: 0.5rem;
        }

        .card-icon.cost {
          background: #fef3c7;
        }

        .card-icon.sla {
          background: #dbeafe;
        }

        .card-icon.perf {
          background: #d1fae5;
        }

        .card-icon.error {
          background: #fee2e2;
        }

        .card-content {
          flex: 1;
        }

        .card-value {
          font-size: 1.75rem;
          font-weight: 600;
          color: #111827;
        }

        .card-label {
          font-size: 0.875rem;
          color: #6b7280;
        }

        .chart-section {
          background: white;
          border: 1px solid #e5e7eb;
          border-radius: 0.5rem;
          padding: 1.5rem;
          margin-bottom: 2rem;
        }

        .chart-section h2 {
          margin: 0 0 1rem 0;
          font-size: 1.25rem;
        }

        .tables-section {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
          gap: 1.5rem;
        }

        .table-container {
          background: white;
          border: 1px solid #e5e7eb;
          border-radius: 0.5rem;
          padding: 1.5rem;
        }

        .table-container h3 {
          margin: 0 0 1rem 0;
        }

        .data-table {
          width: 100%;
          border-collapse: collapse;
        }

        .data-table th,
        .data-table td {
          padding: 0.75rem;
          text-align: left;
          border-bottom: 1px solid #e5e7eb;
        }

        .data-table th {
          background: #f9fafb;
          font-weight: 600;
          font-size: 0.875rem;
        }

        .data-table td {
          font-size: 0.875rem;
        }

        .status {
          padding: 0.25rem 0.75rem;
          border-radius: 1rem;
          font-size: 0.75rem;
          font-weight: 500;
        }

        .status.healthy {
          background: #d1fae5;
          color: #065f46;
        }

        .status.warning {
          background: #fef3c7;
          color: #92400e;
        }

        .loading {
          text-align: center;
          padding: 3rem;
          color: #6b7280;
        }
      `}</style>
    </div>
  );
}
