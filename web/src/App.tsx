import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import RunDetail from './pages/RunDetail'
import WorkflowDesigner from './pages/WorkflowDesigner'
import WorkflowMarketplace from './pages/WorkflowMarketplace'
import AnalyticsDashboard from './pages/AnalyticsDashboard'
import PolicyPlayground from './pages/PolicyPlayground'

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/runs/:runId" element={<RunDetail />} />
        <Route path="/workflows/designer" element={<WorkflowDesigner />} />
        <Route path="/workflows/marketplace" element={<WorkflowMarketplace />} />
        <Route path="/analytics" element={<AnalyticsDashboard />} />
        <Route path="/policy-playground" element={<PolicyPlayground />} />
      </Routes>
    </Layout>
  )
}

export default App
