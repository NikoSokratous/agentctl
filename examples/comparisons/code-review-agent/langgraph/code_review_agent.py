"""
Code Review Agent - LangGraph Implementation

This shows what you need to implement yourself when using LangGraph
for a production-grade code review agent.

Note: This is a simplified version. A production implementation would
require additional code for:
- Policy enforcement (~100 lines)
- Approval workflow (~40 lines)  
- Multi-tenancy (~60 lines)
- Advanced error handling (~50 lines)
- Deployment configuration (~100 lines)

Total production-ready: ~650+ lines
"""

import os
import json
import time
from typing import TypedDict, Annotated, Sequence
from datetime import datetime

from langgraph.graph import StateGraph, END
from langchain_openai import ChatOpenAI
from langchain_core.messages import BaseMessage, HumanMessage, SystemMessage
from langchain.tools import tool
import requests

# ============================================================================
# STATE MANAGEMENT (Manual)
# ============================================================================

class AgentState(TypedDict):
    """
    State that must be manually managed throughout execution.
    AgentRuntime handles this automatically.
    """
    pr_url: str
    severity_threshold: str
    auto_comment: bool
    include_suggestions: bool
    
    # Step outputs - must manually track
    pr_data: dict
    lint_results: dict
    review_comments: str
    breaking_changes: dict
    final_report: str
    comment_result: dict
    
    # Metadata - must manually track
    messages: Sequence[BaseMessage]
    step_count: int
    total_cost: float
    errors: list
    start_time: float


# ============================================================================
# COST TRACKING (Manual Implementation Required)
# ============================================================================

class CostTracker:
    """
    Manual cost tracking - AgentRuntime does this automatically.
    """
    PRICING = {
        "gpt-4": {"input": 0.03 / 1000, "output": 0.06 / 1000},
        "gpt-3.5-turbo": {"input": 0.0015 / 1000, "output": 0.002 / 1000},
    }
    
    def __init__(self):
        self.total_cost = 0.0
        self.per_step_cost = {}
    
    def track_llm_call(self, model: str, input_tokens: int, output_tokens: int, step: str):
        """Calculate and track cost for an LLM call"""
        pricing = self.PRICING.get(model, self.PRICING["gpt-4"])
        cost = (input_tokens * pricing["input"]) + (output_tokens * pricing["output"])
        self.total_cost += cost
        self.per_step_cost[step] = self.per_step_cost.get(step, 0) + cost
        
        # Manual budget check
        if self.total_cost > 10.0:
            raise Exception(f"Cost budget exceeded: ${self.total_cost:.2f}")
        
        return cost
    
    def get_report(self):
        return {
            "total": self.total_cost,
            "by_step": self.per_step_cost
        }


# ============================================================================
# RETRY LOGIC (Manual Implementation Required)
# ============================================================================

def retry_with_backoff(func, max_attempts=3, initial_delay=1):
    """
    Manual retry logic with exponential backoff.
    AgentRuntime has this built-in with configuration.
    """
    for attempt in range(max_attempts):
        try:
            return func()
        except requests.exceptions.HTTPError as e:
            if e.response.status_code == 429:  # Rate limit
                delay = initial_delay * (2 ** attempt)
                print(f"Rate limited. Retrying in {delay}s...")
                time.sleep(delay)
            else:
                raise
        except requests.exceptions.RequestException:
            if attempt < max_attempts - 1:
                delay = initial_delay * (2 ** attempt)
                print(f"Network error. Retrying in {delay}s...")
                time.sleep(delay)
            else:
                raise
    
    raise Exception(f"Max retries ({max_attempts}) exceeded")


# ============================================================================
# APPROVAL GATES (Manual Implementation Required)
# ============================================================================

class ApprovalGate:
    """
    Manual approval workflow - AgentRuntime has this declaratively.
    """
    def __init__(self, webhook_url: str):
        self.webhook_url = webhook_url
        self.pending_approvals = {}
    
    def request_approval(self, action: str, data: dict, timeout: int = 3600):
        """Request human approval for an action"""
        approval_id = f"approval_{int(time.time())}"
        
        # Send approval request
        response = requests.post(self.webhook_url, json={
            "approval_id": approval_id,
            "action": action,
            "data": data,
            "requested_at": datetime.now().isoformat()
        })
        
        if response.status_code != 200:
            raise Exception("Failed to request approval")
        
        # Poll for approval (simplified - production would use webhooks)
        start_time = time.time()
        while time.time() - start_time < timeout:
            status = self.check_approval_status(approval_id)
            if status == "approved":
                return True
            elif status == "rejected":
                raise Exception("Action rejected by approver")
            time.sleep(30)  # Poll every 30 seconds
        
        raise Exception("Approval timeout")
    
    def check_approval_status(self, approval_id: str):
        """Check if approval has been granted"""
        # This would call your approval system API
        # Simplified for example
        return "pending"


# ============================================================================
# GITHUB TOOLS (Manual Implementation)
# ============================================================================

@tool
def fetch_github_pr(pr_url: str) -> dict:
    """Fetch PR details from GitHub"""
    # Extract owner, repo, pr_number from URL
    parts = pr_url.replace("https://github.com/", "").split("/")
    owner, repo, _, pr_number = parts[0], parts[1], parts[2], parts[3]
    
    token = os.getenv("GITHUB_TOKEN")
    headers = {"Authorization": f"Bearer {token}"}
    
    # Fetch PR data
    pr_response = retry_with_backoff(
        lambda: requests.get(
            f"https://api.github.com/repos/{owner}/{repo}/pulls/{pr_number}",
            headers=headers
        )
    )
    pr_data = pr_response.json()
    
    # Fetch files
    files_response = retry_with_backoff(
        lambda: requests.get(
            f"https://api.github.com/repos/{owner}/{repo}/pulls/{pr_number}/files",
            headers=headers
        )
    )
    files = files_response.json()
    
    return {
        "title": pr_data["title"],
        "description": pr_data["body"],
        "author": pr_data["user"]["login"],
        "files": [f["filename"] for f in files],
        "diff": "\n".join([f["patch"] for f in files if "patch" in f]),
        "base_sha": pr_data["base"]["sha"],
        "head_sha": pr_data["head"]["sha"],
        "target_branch": pr_data["base"]["ref"]
    }


# ============================================================================
# AGENT NODES (Manual Implementation)
# ============================================================================

cost_tracker = CostTracker()
llm = ChatOpenAI(model="gpt-4", temperature=0.3)

def fetch_pr_node(state: AgentState) -> AgentState:
    """Step 1: Fetch PR details"""
    print("Step 1: Fetching PR details...")
    
    try:
        pr_data = fetch_github_pr(state["pr_url"])
        state["pr_data"] = pr_data
        state["step_count"] += 1
        
        # Manual cost tracking
        cost = cost_tracker.track_llm_call("gpt-4", 100, 50, "fetch_pr")
        state["total_cost"] += cost
        
    except Exception as e:
        print(f"Error fetching PR: {e}")
        state["errors"].append({"step": "fetch_pr", "error": str(e)})
    
    return state


def static_analysis_node(state: AgentState) -> AgentState:
    """Step 2: Run static analysis"""
    print("Step 2: Running static analysis...")
    
    try:
        # Simplified - production would actually run linters
        messages = [
            SystemMessage(content="You are a code analysis expert."),
            HumanMessage(content=f"""
            Analyze these files for issues:
            Files: {state['pr_data']['files']}
            Severity threshold: {state['severity_threshold']}
            
            Report: linting errors, security issues, code smells
            """)
        ]
        
        response = llm.invoke(messages)
        state["lint_results"] = {"analysis": response.content}
        state["step_count"] += 1
        
        # Track cost
        cost = cost_tracker.track_llm_call("gpt-4", 500, 300, "static_analysis")
        state["total_cost"] += cost
        
    except Exception as e:
        print(f"Error in static analysis: {e}")
        state["errors"].append({"step": "static_analysis", "error": str(e)})
    
    return state


def code_review_node(state: AgentState) -> AgentState:
    """Step 3: Perform code review"""
    print("Step 3: Performing code review...")
    
    try:
        messages = [
            SystemMessage(content="You are an expert code reviewer."),
            HumanMessage(content=f"""
            Review this pull request:
            
            Title: {state['pr_data']['title']}
            Description: {state['pr_data']['description']}
            Files: {state['pr_data']['files']}
            
            Diff:
            {state['pr_data']['diff'][:2000]}  # Truncate for token limits
            
            Provide detailed review focusing on:
            - Code quality and maintainability
            - Potential bugs and edge cases
            - Best practices
            
            Include suggestions: {state['include_suggestions']}
            """)
        ]
        
        response = llm.invoke(messages)
        state["review_comments"] = response.content
        state["step_count"] += 1
        
        # Track cost
        cost = cost_tracker.track_llm_call("gpt-4", 1500, 800, "code_review")
        state["total_cost"] += cost
        
    except Exception as e:
        print(f"Error in code review: {e}")
        state["errors"].append({"step": "code_review", "error": str(e)})
    
    return state


def check_breaking_changes_node(state: AgentState) -> AgentState:
    """Step 4: Check for breaking changes"""
    print("Step 4: Checking for breaking changes...")
    
    # Only run if targeting main/master
    if state["pr_data"]["target_branch"] not in ["main", "master"]:
        state["breaking_changes"] = {"detected": False, "changes": []}
        return state
    
    try:
        messages = [
            SystemMessage(content="You are an API compatibility expert."),
            HumanMessage(content=f"""
            Analyze if this PR introduces breaking changes:
            
            Files: {state['pr_data']['files']}
            Diff: {state['pr_data']['diff'][:1000]}
            
            Check for:
            - Public API modifications
            - Function signature changes
            - Removed exports
            """)
        ]
        
        response = llm.invoke(messages)
        state["breaking_changes"] = {"analysis": response.content}
        state["step_count"] += 1
        
        # Track cost
        cost = cost_tracker.track_llm_call("gpt-4", 800, 400, "check_breaking")
        state["total_cost"] += cost
        
    except Exception as e:
        print(f"Error checking breaking changes: {e}")
        state["errors"].append({"step": "check_breaking", "error": str(e)})
    
    return state


def generate_report_node(state: AgentState) -> AgentState:
    """Step 5: Generate comprehensive report"""
    print("Step 5: Generating report...")
    
    try:
        messages = [
            SystemMessage(content="You are a technical report writer."),
            HumanMessage(content=f"""
            Create a comprehensive code review report in GitHub markdown:
            
            Static Analysis:
            {state['lint_results']}
            
            Code Review:
            {state['review_comments']}
            
            Breaking Changes:
            {state.get('breaking_changes', {})}
            
            Format with sections, emojis, and clear recommendations.
            """)
        ]
        
        response = llm.invoke(messages)
        state["final_report"] = response.content
        state["step_count"] += 1
        
        # Track cost
        cost = cost_tracker.track_llm_call("gpt-4", 1000, 600, "generate_report")
        state["total_cost"] += cost
        
    except Exception as e:
        print(f"Error generating report: {e}")
        state["errors"].append({"step": "generate_report", "error": str(e)})
    
    return state


def post_comment_node(state: AgentState) -> AgentState:
    """Step 6: Post comment to GitHub"""
    print("Step 6: Posting comment...")
    
    # Manual approval check
    if not state["auto_comment"]:
        print("Approval required for posting comment...")
        # In production, this would call ApprovalGate
        # For example purposes, we'll skip
        print("⚠️ Manual approval gate not implemented in this example")
        state["comment_result"] = {"status": "pending_approval"}
        return state
    
    try:
        # Extract owner, repo, pr_number from URL
        parts = state["pr_url"].replace("https://github.com/", "").split("/")
        owner, repo, pr_number = parts[0], parts[1], parts[3]
        
        token = os.getenv("GITHUB_TOKEN")
        headers = {"Authorization": f"Bearer {token}"}
        
        comment_body = f"""{state['final_report']}

---
🤖 Automated review by LangGraph Agent
Cost: ${state['total_cost']:.2f}
Steps: {state['step_count']}
"""
        
        response = retry_with_backoff(
            lambda: requests.post(
                f"https://api.github.com/repos/{owner}/{repo}/issues/{pr_number}/comments",
                headers=headers,
                json={"body": comment_body}
            )
        )
        
        state["comment_result"] = {
            "status": "posted",
            "comment_id": response.json()["id"]
        }
        
    except Exception as e:
        print(f"Error posting comment: {e}")
        state["errors"].append({"step": "post_comment", "error": str(e)})
    
    return state


# ============================================================================
# GRAPH CONSTRUCTION (Manual)
# ============================================================================

def build_code_review_graph():
    """
    Manually construct the workflow graph.
    AgentRuntime does this from YAML configuration.
    """
    workflow = StateGraph(AgentState)
    
    # Add nodes
    workflow.add_node("fetch_pr", fetch_pr_node)
    workflow.add_node("static_analysis", static_analysis_node)
    workflow.add_node("code_review", code_review_node)
    workflow.add_node("check_breaking", check_breaking_changes_node)
    workflow.add_node("generate_report", generate_report_node)
    workflow.add_node("post_comment", post_comment_node)
    
    # Define edges (execution flow)
    workflow.set_entry_point("fetch_pr")
    workflow.add_edge("fetch_pr", "static_analysis")
    workflow.add_edge("static_analysis", "code_review")
    workflow.add_edge("code_review", "check_breaking")
    workflow.add_edge("check_breaking", "generate_report")
    workflow.add_edge("generate_report", "post_comment")
    workflow.add_edge("post_comment", END)
    
    return workflow.compile()


# ============================================================================
# MAIN EXECUTION
# ============================================================================

def run_code_review(
    pr_url: str,
    severity_threshold: str = "medium",
    auto_comment: bool = False,
    include_suggestions: bool = True
):
    """
    Run the code review agent.
    
    Note: This is simplified. Production would need:
    - Better error handling
    - Policy enforcement
    - Comprehensive logging
    - Deployment configuration
    - Multi-tenancy support
    """
    print("=" * 60)
    print("Code Review Agent - LangGraph Implementation")
    print("=" * 60)
    
    # Initialize state
    initial_state = AgentState(
        pr_url=pr_url,
        severity_threshold=severity_threshold,
        auto_comment=auto_comment,
        include_suggestions=include_suggestions,
        pr_data={},
        lint_results={},
        review_comments="",
        breaking_changes={},
        final_report="",
        comment_result={},
        messages=[],
        step_count=0,
        total_cost=0.0,
        errors=[],
        start_time=time.time()
    )
    
    # Build and run graph
    graph = build_code_review_graph()
    
    try:
        result = graph.invoke(initial_state)
        
        # Print results
        duration = time.time() - result["start_time"]
        print("\n" + "=" * 60)
        print("Code Review Complete!")
        print("=" * 60)
        print(f"Duration: {duration:.2f}s")
        print(f"Steps: {result['step_count']}")
        print(f"Total Cost: ${result['total_cost']:.2f}")
        print(f"Errors: {len(result['errors'])}")
        print("\nCost Breakdown:")
        print(json.dumps(cost_tracker.get_report(), indent=2))
        
        if result["errors"]:
            print("\n⚠️ Errors occurred:")
            for error in result["errors"]:
                print(f"  - {error['step']}: {error['error']}")
        
        return result
        
    except Exception as e:
        print(f"\n❌ Fatal error: {e}")
        raise


if __name__ == "__main__":
    import argparse
    
    parser = argparse.ArgumentParser(description="Code Review Agent")
    parser.add_argument("--pr-url", required=True, help="GitHub PR URL")
    parser.add_argument("--severity-threshold", default="medium", help="Severity threshold")
    parser.add_argument("--auto-comment", action="store_true", help="Auto-post comment")
    parser.add_argument("--include-suggestions", action="store_true", default=True, help="Include suggestions")
    
    args = parser.parse_args()
    
    run_code_review(
        pr_url=args.pr_url,
        severity_threshold=args.severity_threshold,
        auto_comment=args.auto_comment,
        include_suggestions=args.include_suggestions
    )
