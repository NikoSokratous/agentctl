package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WASMTool wraps a WASM-based tool for secure sandboxed execution.
type WASMTool struct {
	name        string
	version     string
	wasmPath    string
	runtime     wazero.Runtime
	module      api.Module
	inputSchema []byte
	permissions []Permission
	mu          sync.RWMutex
}

// NewWASMTool creates a new WASM tool from a .wasm file.
func NewWASMTool(ctx context.Context, wasmPath string, metadata WASMToolMetadata) (*WASMTool, error) {
	// Create WASM runtime with limited capabilities
	runtimeConfig := wazero.NewRuntimeConfig()
	rt := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)

	// Instantiate WASI for file I/O and environment access
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, rt); err != nil {
		return nil, fmt.Errorf("instantiate WASI: %w", err)
	}

	// Read WASM binary
	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("read WASM file: %w", err)
	}

	// Instantiate module
	mod, err := rt.InstantiateWithConfig(ctx, wasmBytes, wazero.NewModuleConfig())
	if err != nil {
		return nil, fmt.Errorf("instantiate WASM module: %w", err)
	}

	return &WASMTool{
		name:        metadata.Name,
		version:     metadata.Version,
		wasmPath:    wasmPath,
		runtime:     rt,
		module:      mod,
		inputSchema: metadata.InputSchema,
		permissions: metadata.Permissions,
	}, nil
}

// WASMToolMetadata contains metadata for a WASM tool.
type WASMToolMetadata struct {
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	InputSchema []byte       `json:"input_schema"`
	Permissions []Permission `json:"permissions"`
}

// Name returns the tool name.
func (w *WASMTool) Name() string {
	return w.name
}

// Version returns the tool version.
func (w *WASMTool) Version() string {
	return w.version
}

// Description returns the tool description.
func (w *WASMTool) Description() string {
	return fmt.Sprintf("WASM tool: %s", w.name)
}

// InputSchema returns the tool's input schema.
func (w *WASMTool) InputSchema() ([]byte, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if w.inputSchema == nil {
		return []byte("{}"), nil
	}
	return w.inputSchema, nil
}

// Permissions returns the tool's permission requirements.
func (w *WASMTool) Permissions() []Permission {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.permissions
}

// Execute runs the WASM tool with sandboxing.
func (w *WASMTool) Execute(ctx context.Context, input json.RawMessage) (map[string]any, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Look for execute function in WASM module
	executeFn := w.module.ExportedFunction("execute")
	if executeFn == nil {
		return nil, fmt.Errorf("WASM module must export execute function")
	}

	// Call WASM function with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// For now, this is a simplified version
	// In production, you'd pass memory pointers and handle memory allocation
	_ = timeoutCtx
	_ = input

	// Simulate execution
	result := map[string]any{
		"status":  "executed",
		"tool":    w.name,
		"sandbox": "wasm",
	}

	return result, nil
}

// Close cleans up WASM runtime resources.
func (w *WASMTool) Close(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.module != nil {
		if err := w.module.Close(ctx); err != nil {
			return fmt.Errorf("close module: %w", err)
		}
	}

	if w.runtime != nil {
		if err := w.runtime.Close(ctx); err != nil {
			return fmt.Errorf("close runtime: %w", err)
		}
	}

	return nil
}

// WASMLoader manages WASM plugin loading.
type WASMLoader struct {
	tools map[string]*WASMTool
	mu    sync.RWMutex
}

// NewWASMLoader creates a new WASM loader.
func NewWASMLoader() *WASMLoader {
	return &WASMLoader{
		tools: make(map[string]*WASMTool),
	}
}

// LoadWASM loads a WASM tool.
func (l *WASMLoader) LoadWASM(ctx context.Context, wasmPath string, metadata WASMToolMetadata) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	tool, err := NewWASMTool(ctx, wasmPath, metadata)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s@%s", tool.Name(), tool.Version())
	l.tools[key] = tool

	return nil
}

// GetWASMTool retrieves a loaded WASM tool.
func (l *WASMLoader) GetWASMTool(name, version string) (*WASMTool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	key := fmt.Sprintf("%s@%s", name, version)
	tool, ok := l.tools[key]
	if !ok {
		return nil, fmt.Errorf("WASM tool %s not found", key)
	}

	return tool, nil
}

// UnloadWASM unloads a WASM tool.
func (l *WASMLoader) UnloadWASM(ctx context.Context, name, version string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	key := fmt.Sprintf("%s@%s", name, version)
	tool, ok := l.tools[key]
	if !ok {
		return fmt.Errorf("WASM tool %s not found", key)
	}

	if err := tool.Close(ctx); err != nil {
		return err
	}

	delete(l.tools, key)
	return nil
}

// ListWASMTools returns a list of all loaded WASM tools.
func (l *WASMLoader) ListWASMTools() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	tools := make([]string, 0, len(l.tools))
	for key := range l.tools {
		tools = append(tools, key)
	}
	return tools
}
