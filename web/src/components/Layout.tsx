import React from 'react'
import { Link } from 'react-router-dom'
import { Activity, Settings, GitBranch, Store, BarChart3 } from 'lucide-react'
import './Layout.css'

interface LayoutProps {
  children: React.ReactNode
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  return (
    <div className="layout">
      <nav className="sidebar">
        <div className="sidebar-header">
          <Activity size={32} className="logo-icon" />
          <h1>AgentRuntime</h1>
        </div>
        
        <ul className="nav-menu">
          <li>
            <Link to="/" className="nav-link">
              <Activity size={20} />
              <span>Dashboard</span>
            </Link>
          </li>
          <li>
            <Link to="/workflows/designer" className="nav-link">
              <GitBranch size={20} />
              <span>Workflow Designer</span>
            </Link>
          </li>
          <li>
            <Link to="/workflows/marketplace" className="nav-link">
              <Store size={20} />
              <span>Marketplace</span>
            </Link>
          </li>
          <li>
            <Link to="/analytics" className="nav-link">
              <BarChart3 size={20} />
              <span>Analytics</span>
            </Link>
          </li>
          <li>
            <Link to="/settings" className="nav-link">
              <Settings size={20} />
              <span>Settings</span>
            </Link>
          </li>
        </ul>
        
        <div className="sidebar-footer">
          <div className="version">v1.0.0</div>
        </div>
      </nav>
      
      <main className="main-content">
        {children}
      </main>
    </div>
  )
}

export default Layout
