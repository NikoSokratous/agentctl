import React from 'react'

interface LoadingSpinnerProps {
  message?: string
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({ message = 'Loading...' }) => (
  <div className="loading-spinner">
    <div className="spinner" />
    <span>{message}</span>
  </div>
)

export default LoadingSpinner
