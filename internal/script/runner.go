package script

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

// compiledScript 缓存同一脚本源码对应的编译结果（AST Program），
// 避免每次请求都重复 parse/compile，显著降低高并发下的 CPU 开销。
type compiledScript struct {
	program *goja.Program
}

var (
	scriptCache   = make(map[string]*compiledScript)
	scriptCacheMu sync.RWMutex
)

// getProgram 返回 scriptSrc 对应的已编译 Program（命中缓存则直接返回）。
func getProgram(scriptSrc string) (*goja.Program, error) {
	scriptCacheMu.RLock()
	if c, ok := scriptCache[scriptSrc]; ok {
		scriptCacheMu.RUnlock()
		return c.program, nil
	}
	scriptCacheMu.RUnlock()

	prog, err := goja.Compile("script", scriptSrc, false)
	if err != nil {
		return nil, fmt.Errorf("script compile error: %w", err)
	}

	scriptCacheMu.Lock()
	scriptCache[scriptSrc] = &compiledScript{program: prog}
	scriptCacheMu.Unlock()

	return prog, nil
}

// RunMapRequest 执行 JS 脚本中的 mapRequest(input) 函数，将平台标准请求映射为上游格式。
//
// 脚本示例：
//
//	function mapRequest(input) {
//	    return { ...input, model: "vendor-model-name" };
//	}
func RunMapRequest(scriptSrc string, input map[string]interface{}) (map[string]interface{}, error) {
	return runMapFn(scriptSrc, "mapRequest", input)
}

// RunMapResponse 执行 JS 脚本中的 mapResponse(input) 函数，将上游响应映射为平台标准格式。
//
// 脚本示例：
//
//	function mapResponse(output) {
//	    return { url: output.data[0].url, status: 2 };
//	}
func RunMapResponse(scriptSrc string, input map[string]interface{}) (map[string]interface{}, error) {
	return runMapFn(scriptSrc, "mapResponse", input)
}

func runMapFn(scriptSrc, fnName string, input map[string]interface{}) (map[string]interface{}, error) {
	prog, err := getProgram(scriptSrc)
	if err != nil {
		return nil, err
	}

	// 每次请求创建独立的 VM Runtime，保证并发安全（goja Runtime 非线程安全）
	vm := goja.New()
	if _, err := vm.RunProgram(prog); err != nil {
		return nil, fmt.Errorf("script run error: %w", err)
	}

	fn, ok := goja.AssertFunction(vm.Get(fnName))
	if !ok {
		return nil, fmt.Errorf("function %q not found in script", fnName)
	}

	res, err := fn(goja.Undefined(), vm.ToValue(input))
	if err != nil {
		return nil, fmt.Errorf("function %q execution error: %w", fnName, err)
	}

	if goja.IsNull(res) || goja.IsUndefined(res) {
		return input, nil
	}

	exported := res.Export()
	result, ok := exported.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("function %q must return an object, got %T", fnName, exported)
	}
	return result, nil
}
