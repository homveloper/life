package autorouter

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

// HandlerFunc represents the expected handler function signature
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Middleware represents middleware function signature
type Middleware func(http.Handler) http.Handler

// RegistrationOptions configures how handlers are registered
type RegistrationOptions struct {
	Prefix       string       // URL prefix (e.g., "/api/v1/")
	MethodPrefix string       // Method prefix (e.g., "trainer." -> "trainer.Create")
	Middleware   []Middleware // Middleware chain to apply
}

// AutoRouter handles automatic registration of HTTP handlers using reflection
type AutoRouter struct {
	mux     *http.ServeMux
	options RegistrationOptions
}

// NewAutoRouter creates a new auto router
func NewAutoRouter(mux *http.ServeMux, options RegistrationOptions) *AutoRouter {
	return &AutoRouter{
		mux:     mux,
		options: options,
	}
}

// RegisterHandlers automatically registers all methods that match HandlerFunc signature
// from the given handler struct
func (ar *AutoRouter) RegisterHandlers(handler interface{}) error {
	handlerType := reflect.TypeOf(handler)
	handlerValue := reflect.ValueOf(handler)

	// Ensure we're working with a pointer to struct
	if handlerType.Kind() == reflect.Ptr {
		handlerType = handlerType.Elem()
	}

	if handlerType.Kind() != reflect.Struct {
		return fmt.Errorf("handler must be a struct or pointer to struct")
	}

	// Get the number of methods
	numMethods := handlerValue.NumMethod()

	for i := 0; i < numMethods; i++ {
		method := handlerValue.Method(i)
		methodType := handlerValue.Type().Method(i)
		methodName := methodType.Name

		// Skip unexported methods
		if !isExported(methodName) {
			continue
		}

		// Skip methods that start with "Handle" as they're likely already manual handlers
		if strings.HasPrefix(methodName, "Handle") {
			continue
		}

		// Check if method matches HandlerFunc signature
		if !ar.isValidHandlerFunc(method) {
			continue
		}

		// Register the handler
		if err := ar.registerMethod(handler, methodName, method); err != nil {
			return fmt.Errorf("failed to register method %s: %w", methodName, err)
		}
	}

	return nil
}

// RegisterHandlersWithAuth registers handlers with authentication middleware
func (ar *AutoRouter) RegisterHandlersWithAuth(handler interface{}, authMiddleware Middleware) error {
	// Create a copy of options with auth middleware
	optionsWithAuth := ar.options
	optionsWithAuth.Middleware = append([]Middleware{authMiddleware}, ar.options.Middleware...)
	
	// Create temporary router with auth
	tempRouter := &AutoRouter{
		mux:     ar.mux,
		options: optionsWithAuth,
	}
	
	return tempRouter.RegisterHandlers(handler)
}

// RegisterSingleMethod registers a single method with custom path
func (ar *AutoRouter) RegisterSingleMethod(handler interface{}, methodName string, customPath string) error {
	handlerValue := reflect.ValueOf(handler)
	method := handlerValue.MethodByName(methodName)
	
	if !method.IsValid() {
		return fmt.Errorf("method %s not found", methodName)
	}
	
	if !ar.isValidHandlerFunc(method) {
		return fmt.Errorf("method %s does not match handler signature", methodName)
	}
	
	// Use custom path instead of auto-generated one
	handlerFunc := ar.createHandlerFunc(method)
	finalHandler := ar.applyMiddleware(handlerFunc)
	
	fullPath := ar.options.Prefix + customPath
	ar.mux.HandleFunc(fullPath, finalHandler)
	
	fmt.Printf("Auto-registered (custom): %s -> %s\n", fullPath, methodName)
	return nil
}

// isValidHandlerFunc checks if a method matches the HandlerFunc signature
// Expected signature: func(http.ResponseWriter, *http.Request)
func (ar *AutoRouter) isValidHandlerFunc(method reflect.Value) bool {
	methodType := method.Type()
	
	// Must be a function
	if methodType.Kind() != reflect.Func {
		return false
	}
	
	// Must have exactly 2 parameters
	if methodType.NumIn() != 2 {
		return false
	}
	
	// Must have no return values (or only error return)
	if methodType.NumOut() > 1 {
		return false
	}
	
	// If there's a return value, it should be error
	if methodType.NumOut() == 1 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		if !methodType.Out(0).Implements(errorInterface) {
			return false
		}
	}
	
	// First parameter should be http.ResponseWriter
	firstParam := methodType.In(0)
	responseWriterType := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
	if !firstParam.Implements(responseWriterType) {
		return false
	}
	
	// Second parameter should be *http.Request
	secondParam := methodType.In(1)
	requestType := reflect.TypeOf((*http.Request)(nil))
	if secondParam != requestType {
		return false
	}
	
	return true
}

// registerMethod registers a single method as an HTTP handler
func (ar *AutoRouter) registerMethod(handler interface{}, methodName string, method reflect.Value) error {
	// Build the URL path
	urlPath := ar.buildURLPath(methodName)
	
	// Create the handler function
	handlerFunc := ar.createHandlerFunc(method)
	
	// Apply middleware if any
	finalHandler := ar.applyMiddleware(handlerFunc)
	
	// Register with the mux
	ar.mux.HandleFunc(urlPath, finalHandler)
	
	fmt.Printf("Auto-registered: %s -> %s\n", urlPath, methodName)
	return nil
}

// buildURLPath constructs the URL path from method name
func (ar *AutoRouter) buildURLPath(methodName string) string {
	// Start with prefix
	path := ar.options.Prefix
	
	// Add method prefix if specified
	if ar.options.MethodPrefix != "" {
		path += ar.options.MethodPrefix + methodName
	} else {
		path += strings.ToLower(methodName)
	}
	
	return path
}

// createHandlerFunc creates an http.HandlerFunc from a reflect.Value
func (ar *AutoRouter) createHandlerFunc(method reflect.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Call the method with the parameters
		args := []reflect.Value{
			reflect.ValueOf(w),
			reflect.ValueOf(r),
		}
		
		results := method.Call(args)
		
		// Handle error return if present
		if len(results) > 0 && !results[0].IsNil() {
			// Method returned an error
			if err := results[0].Interface().(error); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

// applyMiddleware applies all configured middleware to the handler
func (ar *AutoRouter) applyMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	if len(ar.options.Middleware) == 0 {
		return handler
	}
	
	// Apply middleware in reverse order (last middleware wraps first)
	h := http.Handler(handler)
	for i := len(ar.options.Middleware) - 1; i >= 0; i-- {
		h = ar.options.Middleware[i](h)
	}
	
	return h.ServeHTTP
}

// HandlerInfo provides information about registered handlers
type HandlerInfo struct {
	URLPath    string
	MethodName string
	HasAuth    bool
}

// GetRegisteredHandlers returns information about all registered handlers
// This is useful for debugging and API documentation
func (ar *AutoRouter) GetRegisteredHandlers(handler interface{}) []HandlerInfo {
	var handlers []HandlerInfo
	
	handlerType := reflect.TypeOf(handler)
	handlerValue := reflect.ValueOf(handler)

	if handlerType.Kind() == reflect.Ptr {
		handlerType = handlerType.Elem()
	}

	numMethods := handlerValue.NumMethod()

	for i := 0; i < numMethods; i++ {
		method := handlerValue.Method(i)
		methodType := handlerValue.Type().Method(i)
		methodName := methodType.Name

		if !isExported(methodName) {
			continue
		}

		if strings.HasPrefix(methodName, "Handle") {
			continue
		}

		if !ar.isValidHandlerFunc(method) {
			continue
		}

		handlers = append(handlers, HandlerInfo{
			URLPath:    ar.buildURLPath(methodName),
			MethodName: methodName,
			HasAuth:    len(ar.options.Middleware) > 0,
		})
	}

	return handlers
}

// PrintRegisteredHandlers prints all registered handlers for debugging
func (ar *AutoRouter) PrintRegisteredHandlers(handler interface{}) {
	handlers := ar.GetRegisteredHandlers(handler)
	
	fmt.Println("=== Auto-Registered Handlers ===")
	for _, h := range handlers {
		authStatus := ""
		if h.HasAuth {
			authStatus = " [AUTH]"
		}
		fmt.Printf("  %s -> %s%s\n", h.URLPath, h.MethodName, authStatus)
	}
	fmt.Println("=================================")
}

// isExported reports whether name is an exported Go symbol
func isExported(name string) bool {
	r := rune(name[0])
	return r >= 'A' && r <= 'Z'
}

// QuickRegister is a convenience function for simple handler registration
func QuickRegister(mux *http.ServeMux, prefix string, methodPrefix string, handler interface{}) error {
	router := NewAutoRouter(mux, RegistrationOptions{
		Prefix:       prefix,
		MethodPrefix: methodPrefix,
	})
	return router.RegisterHandlers(handler)
}

// QuickRegisterWithAuth is a convenience function for handler registration with auth
func QuickRegisterWithAuth(mux *http.ServeMux, prefix string, methodPrefix string, handler interface{}, authMiddleware Middleware) error {
	router := NewAutoRouter(mux, RegistrationOptions{
		Prefix:       prefix,
		MethodPrefix: methodPrefix,
	})
	return router.RegisterHandlersWithAuth(handler, authMiddleware)
}