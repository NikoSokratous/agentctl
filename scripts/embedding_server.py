#!/usr/bin/env python3
"""
Local Embedding Server for AgentRuntime
Uses sentence-transformers to generate embeddings locally.
"""

import json
import sys
from typing import List

try:
    from sentence_transformers import SentenceTransformer
    HAS_SENTENCE_TRANSFORMERS = True
except ImportError:
    HAS_SENTENCE_TRANSFORMERS = False


def generate_embeddings(model_name: str, texts: List[str]) -> List[List[float]]:
    """Generate embeddings using sentence-transformers."""
    if not HAS_SENTENCE_TRANSFORMERS:
        raise RuntimeError(
            "sentence-transformers not installed. "
            "Install with: pip install sentence-transformers"
        )
    
    model = SentenceTransformer(model_name)
    embeddings = model.encode(texts, convert_to_numpy=True)
    return embeddings.tolist()


def main():
    """Read input from stdin and output embeddings to stdout."""
    try:
        # Read input JSON from stdin
        input_data = json.load(sys.stdin)
        model_name = input_data.get("model", "all-MiniLM-L6-v2")
        texts = input_data.get("texts", [])
        
        if not texts:
            result = {"embeddings": [], "error": ""}
        else:
            embeddings = generate_embeddings(model_name, texts)
            result = {"embeddings": embeddings, "error": ""}
        
        # Output result as JSON
        json.dump(result, sys.stdout)
        sys.stdout.flush()
        
    except Exception as e:
        # Return error in JSON format
        result = {"embeddings": [], "error": str(e)}
        json.dump(result, sys.stdout)
        sys.stdout.flush()
        sys.exit(1)


if __name__ == "__main__":
    main()
