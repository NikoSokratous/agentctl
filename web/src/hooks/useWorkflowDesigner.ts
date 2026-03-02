import { useCallback, useState } from 'react';
import { Node, Edge, Connection, addEdge, applyNodeChanges, applyEdgeChanges } from 'reactflow';

export interface WorkflowNode extends Node {
  data: {
    label: string;
    agent?: string;
    goal?: string;
    condition?: string;
    outputKey?: string;
    timeout?: string;
    retry?: number;
  };
}

export const useWorkflowDesigner = () => {
  const [nodes, setNodes] = useState<WorkflowNode[]>([]);
  const [edges, setEdges] = useState<Edge[]>([]);
  const [isValid, setIsValid] = useState(true);
  const [errors, setErrors] = useState<string[]>([]);

  const onNodesChange = useCallback(
    (changes: any) => setNodes((nds) => applyNodeChanges(changes, nds)),
    []
  );

  const onEdgesChange = useCallback(
    (changes: any) => setEdges((eds) => applyEdgeChanges(changes, eds)),
    []
  );

  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((eds) => addEdge(connection, eds));
      validateWorkflow();
    },
    []
  );

  const addNode = useCallback((type: string, position: { x: number; y: number }) => {
    const newNode: WorkflowNode = {
      id: `node-${Date.now()}`,
      type,
      position,
      data: {
        label: `${type}-${nodes.length + 1}`,
      },
    };
    setNodes((nds) => [...nds, newNode]);
  }, [nodes.length]);

  const updateNode = useCallback((nodeId: string, data: Partial<WorkflowNode['data']>) => {
    setNodes((nds) =>
      nds.map((node) =>
        node.id === nodeId ? { ...node, data: { ...node.data, ...data } } : node
      )
    );
  }, []);

  const deleteNode = useCallback((nodeId: string) => {
    setNodes((nds) => nds.filter((node) => node.id !== nodeId));
    setEdges((eds) => eds.filter((edge) => edge.source !== nodeId && edge.target !== nodeId));
  }, []);

  const validateWorkflow = useCallback(() => {
    const validationErrors: string[] = [];

    // Check for cycles
    const hasCycle = detectCycle(nodes, edges);
    if (hasCycle) {
      validationErrors.push('Workflow contains a cycle');
    }

    // Check for disconnected nodes
    const disconnected = nodes.filter((node) => {
      const hasIncoming = edges.some((edge) => edge.target === node.id);
      const hasOutgoing = edges.some((edge) => edge.source === node.id);
      return !hasIncoming && !hasOutgoing && nodes.length > 1;
    });

    if (disconnected.length > 0) {
      validationErrors.push(`${disconnected.length} disconnected node(s)`);
    }

    // Check for required fields
    nodes.forEach((node) => {
      if (!node.data.agent && node.type === 'agent') {
        validationErrors.push(`Node ${node.data.label} missing agent name`);
      }
      if (!node.data.goal && node.type === 'agent') {
        validationErrors.push(`Node ${node.data.label} missing goal`);
      }
    });

    setErrors(validationErrors);
    setIsValid(validationErrors.length === 0);

    return validationErrors.length === 0;
  }, [nodes, edges]);

  const exportToYAML = useCallback(() => {
    const workflow = {
      name: 'workflow',
      description: 'Generated workflow',
      steps: nodes.map((node) => ({
        name: node.id,
        agent: node.data.agent || 'default',
        goal: node.data.goal || '',
        output_key: node.data.outputKey,
        condition: node.data.condition,
        timeout: node.data.timeout,
        retry: node.data.retry,
        depends_on: edges
          .filter((edge) => edge.target === node.id)
          .map((edge) => edge.source),
      })),
    };

    // Convert to YAML format (simplified)
    let yaml = `name: ${workflow.name}\n`;
    yaml += `description: ${workflow.description}\n\n`;
    yaml += 'steps:\n';

    workflow.steps.forEach((step) => {
      yaml += `  - name: ${step.name}\n`;
      yaml += `    agent: ${step.agent}\n`;
      yaml += `    goal: "${step.goal}"\n`;
      if (step.output_key) yaml += `    output_key: ${step.output_key}\n`;
      if (step.condition) yaml += `    condition: "${step.condition}"\n`;
      if (step.timeout) yaml += `    timeout: ${step.timeout}\n`;
      if (step.retry) yaml += `    retry: ${step.retry}\n`;
      if (step.depends_on && step.depends_on.length > 0) {
        yaml += `    depends_on:\n`;
        step.depends_on.forEach((dep) => {
          yaml += `      - ${dep}\n`;
        });
      }
      yaml += '\n';
    });

    return yaml;
  }, [nodes, edges]);

  const importFromYAML = useCallback((yaml: string) => {
    try {
      // Simple YAML parsing (in production, use a proper parser)
      const lines = yaml.split('\n');
      const newNodes: WorkflowNode[] = [];
      const newEdges: Edge[] = [];
      let currentStep: any = null;
      let stepIndex = 0;

      lines.forEach((line) => {
        if (line.trim().startsWith('- name:')) {
          if (currentStep) {
            const node: WorkflowNode = {
              id: currentStep.name,
              type: 'agent',
              position: { x: 100 + (stepIndex % 3) * 250, y: 100 + Math.floor(stepIndex / 3) * 150 },
              data: {
                label: currentStep.name,
                agent: currentStep.agent,
                goal: currentStep.goal,
                condition: currentStep.condition,
                outputKey: currentStep.output_key,
                timeout: currentStep.timeout,
                retry: currentStep.retry,
              },
            };
            newNodes.push(node);

            if (currentStep.depends_on) {
              currentStep.depends_on.forEach((dep: string) => {
                newEdges.push({
                  id: `${dep}-${currentStep.name}`,
                  source: dep,
                  target: currentStep.name,
                });
              });
            }

            stepIndex++;
          }
          currentStep = { name: line.split(':')[1].trim(), depends_on: [] };
        } else if (currentStep) {
          const [key, ...valueParts] = line.trim().split(':');
          const value = valueParts.join(':').trim().replace(/['"]/g, '');
          
          if (key === 'agent') currentStep.agent = value;
          else if (key === 'goal') currentStep.goal = value;
          else if (key === 'condition') currentStep.condition = value;
          else if (key === 'output_key') currentStep.output_key = value;
          else if (key === 'timeout') currentStep.timeout = value;
          else if (key === 'retry') currentStep.retry = parseInt(value);
          else if (line.trim().startsWith('- ') && currentStep.depends_on) {
            currentStep.depends_on.push(line.trim().substring(2));
          }
        }
      });

      // Add last step
      if (currentStep) {
        const node: WorkflowNode = {
          id: currentStep.name,
          type: 'agent',
          position: { x: 100 + (stepIndex % 3) * 250, y: 100 + Math.floor(stepIndex / 3) * 150 },
          data: {
            label: currentStep.name,
            agent: currentStep.agent,
            goal: currentStep.goal,
            condition: currentStep.condition,
            outputKey: currentStep.output_key,
            timeout: currentStep.timeout,
            retry: currentStep.retry,
          },
        };
        newNodes.push(node);

        if (currentStep.depends_on) {
          currentStep.depends_on.forEach((dep: string) => {
            newEdges.push({
              id: `${dep}-${currentStep.name}`,
              source: dep,
              target: currentStep.name,
            });
          });
        }
      }

      setNodes(newNodes);
      setEdges(newEdges);
      validateWorkflow();
    } catch (error) {
      console.error('Failed to parse YAML:', error);
      throw new Error('Invalid YAML format');
    }
  }, [validateWorkflow]);

  return {
    nodes,
    edges,
    isValid,
    errors,
    onNodesChange,
    onEdgesChange,
    onConnect,
    addNode,
    updateNode,
    deleteNode,
    validateWorkflow,
    exportToYAML,
    importFromYAML,
  };
};

// Cycle detection using DFS
function detectCycle(nodes: WorkflowNode[], edges: Edge[]): boolean {
  const adjacencyList = new Map<string, string[]>();
  
  nodes.forEach((node) => {
    adjacencyList.set(node.id, []);
  });

  edges.forEach((edge) => {
    const neighbors = adjacencyList.get(edge.source) || [];
    neighbors.push(edge.target);
    adjacencyList.set(edge.source, neighbors);
  });

  const visited = new Set<string>();
  const recursionStack = new Set<string>();

  const hasCycleDFS = (nodeId: string): boolean => {
    visited.add(nodeId);
    recursionStack.add(nodeId);

    const neighbors = adjacencyList.get(nodeId) || [];
    for (const neighbor of neighbors) {
      if (!visited.has(neighbor)) {
        if (hasCycleDFS(neighbor)) {
          return true;
        }
      } else if (recursionStack.has(neighbor)) {
        return true;
      }
    }

    recursionStack.delete(nodeId);
    return false;
  };

  for (const node of nodes) {
    if (!visited.has(node.id)) {
      if (hasCycleDFS(node.id)) {
        return true;
      }
    }
  }

  return false;
}
