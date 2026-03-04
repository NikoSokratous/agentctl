const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface WorkflowValidationRequest {
  workflow: {
    name: string;
    steps: Array<{
      name: string;
      agent: string;
      goal: string;
      depends_on?: string[];
    }>;
  };
}

export interface WorkflowValidationResponse {
  valid: boolean;
  errors?: string[];
  warnings?: string[];
}

export interface WorkflowTemplate {
  id: string;
  name: string;
  author: string;
  description: string;
  category: string;
  tags: string[];
  template_yaml: string;
  downloads: number;
  rating: number;
  rating_count?: number;
  version?: string;
  published_at: string;
}

export const workflowsAPI = {
  // Validate workflow
  async validate(request: WorkflowValidationRequest): Promise<WorkflowValidationResponse> {
    const response = await fetch(`${API_BASE_URL}/v1/workflows/validate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(request),
    });

    if (!response.ok) {
      throw new Error('Validation request failed');
    }

    return response.json();
  },

  // Preview workflow execution
  async preview(workflowYaml: string): Promise<any> {
    const response = await fetch(`${API_BASE_URL}/v1/workflows/preview`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ workflow_yaml: workflowYaml }),
    });

    if (!response.ok) {
      throw new Error('Preview request failed');
    }

    return response.json();
  },

  // List workflow templates
  async listTemplates(params?: {
    category?: string;
    tags?: string[];
    limit?: number;
  }): Promise<WorkflowTemplate[]> {
    const queryParams = new URLSearchParams();
    if (params?.category) queryParams.append('category', params.category);
    if (params?.tags) params.tags.forEach((tag) => queryParams.append('tag', tag));
    if (params?.limit) queryParams.append('limit', params.limit.toString());

    const response = await fetch(
      `${API_BASE_URL}/v1/workflows/marketplace?${queryParams.toString()}`
    );

    if (!response.ok) {
      throw new Error('Failed to fetch templates');
    }

    return response.json();
  },

  // Get template by ID
  async getTemplate(id: string): Promise<WorkflowTemplate> {
    const response = await fetch(`${API_BASE_URL}/v1/workflows/marketplace/${id}`);

    if (!response.ok) {
      throw new Error('Failed to fetch template');
    }

    return response.json();
  },

  // Install template
  async installTemplate(id: string): Promise<{ yaml: string }> {
    const response = await fetch(`${API_BASE_URL}/v1/workflows/marketplace/${id}/install`, {
      method: 'POST',
    });

    if (!response.ok) {
      throw new Error('Failed to install template');
    }

    return response.json();
  },

  // Publish template
  async publishTemplate(template: {
    name: string;
    description: string;
    category: string;
    tags: string[];
    template_yaml: string;
  }): Promise<WorkflowTemplate> {
    const response = await fetch(`${API_BASE_URL}/v1/workflows/marketplace`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(template),
    });

    if (!response.ok) {
      throw new Error('Failed to publish template');
    }

    return response.json();
  },

  // Rate template
  async rateTemplate(id: string, rating: number): Promise<void> {
    const response = await fetch(`${API_BASE_URL}/v1/workflows/marketplace/${id}/rate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ rating }),
    });

    if (!response.ok) {
      throw new Error('Failed to rate template');
    }
  },
};
