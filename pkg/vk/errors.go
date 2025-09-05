package vk

import (
	"fmt"
	"github.com/christerso/vulkan-go/pkg/vulkan"
)

// VulkanError wraps a Vulkan result code with additional context
type VulkanError struct {
	Result  vulkan.Result
	Message string
	Context string
}

// Error implements the error interface
func (e *VulkanError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("Vulkan error in %s: %s (%s)", e.Context, e.Message, e.Result.Error())
	}
	return fmt.Sprintf("Vulkan error: %s (%s)", e.Message, e.Result.Error())
}

// Unwrap returns the underlying Vulkan result as an error
func (e *VulkanError) Unwrap() error {
	return e.Result
}

// IsVulkanError checks if an error is a Vulkan error
func IsVulkanError(err error) bool {
	_, ok := err.(*VulkanError)
	return ok
}

// GetVulkanResult extracts the Vulkan result from an error if possible
func GetVulkanResult(err error) (vulkan.Result, bool) {
	if vkErr, ok := err.(*VulkanError); ok {
		return vkErr.Result, true
	}
	return vulkan.SUCCESS, false
}

// NewVulkanError creates a new Vulkan error with context
func NewVulkanError(result vulkan.Result, message, context string) *VulkanError {
	return &VulkanError{
		Result:  result,
		Message: message,
		Context: context,
	}
}

// CheckResult checks a Vulkan result and returns an error if it indicates failure
func CheckResult(result vulkan.Result, operation string) error {
	if result == vulkan.SUCCESS {
		return nil
	}
	
	message := getResultMessage(result)
	return NewVulkanError(result, message, operation)
}

// Must panics if the result indicates an error, otherwise returns the result
func Must(result vulkan.Result, operation string) vulkan.Result {
	if err := CheckResult(result, operation); err != nil {
		panic(err)
	}
	return result
}

// getResultMessage returns a human-readable message for a Vulkan result
func getResultMessage(result vulkan.Result) string {
	switch result {
	case vulkan.SUCCESS:
		return "Operation completed successfully"
	case vulkan.NOT_READY:
		return "A fence or query has not yet completed"
	case vulkan.TIMEOUT:
		return "A wait operation has not completed in the specified time"
	case vulkan.EVENT_SET:
		return "An event is signaled"
	case vulkan.EVENT_RESET:
		return "An event is unsignaled"
	case vulkan.INCOMPLETE:
		return "A return array was too small for the result"
	case vulkan.ERROR_OUT_OF_HOST_MEMORY:
		return "A host memory allocation has failed"
	case vulkan.ERROR_OUT_OF_DEVICE_MEMORY:
		return "A device memory allocation has failed"
	case vulkan.ERROR_INITIALIZATION_FAILED:
		return "Initialization of an object could not be completed"
	case vulkan.ERROR_DEVICE_LOST:
		return "The logical or physical device has been lost"
	case vulkan.ERROR_MEMORY_MAP_FAILED:
		return "Mapping of a memory object has failed"
	case vulkan.ERROR_LAYER_NOT_PRESENT:
		return "A requested layer is not present or could not be loaded"
	case vulkan.ERROR_EXTENSION_NOT_PRESENT:
		return "A requested extension is not supported"
	case vulkan.ERROR_FEATURE_NOT_PRESENT:
		return "A requested feature is not supported"
	case vulkan.ERROR_INCOMPATIBLE_DRIVER:
		return "The requested version of Vulkan is not supported"
	case vulkan.ERROR_TOO_MANY_OBJECTS:
		return "Too many objects of the type have already been created"
	case vulkan.ERROR_FORMAT_NOT_SUPPORTED:
		return "A requested format is not supported on this device"
	case vulkan.ERROR_FRAGMENTED_POOL:
		return "A pool allocation has failed due to fragmentation"
	case vulkan.ERROR_UNKNOWN:
		return "An unknown error has occurred"
	default:
		return fmt.Sprintf("Unknown Vulkan result: %d", int(result))
	}
}

// ErrorHandler provides customizable error handling strategies
type ErrorHandler interface {
	HandleError(err error) error
}

// DefaultErrorHandler is the default error handling strategy
type DefaultErrorHandler struct{}

// HandleError implements ErrorHandler for DefaultErrorHandler
func (h *DefaultErrorHandler) HandleError(err error) error {
	return err // Simply return the error as-is
}

// PanicErrorHandler panics on any error
type PanicErrorHandler struct{}

// HandleError implements ErrorHandler for PanicErrorHandler  
func (h *PanicErrorHandler) HandleError(err error) error {
	if err != nil {
		panic(err)
	}
	return nil
}

// LoggingErrorHandler logs errors before returning them
type LoggingErrorHandler struct {
	Logger func(error)
}

// HandleError implements ErrorHandler for LoggingErrorHandler
func (h *LoggingErrorHandler) HandleError(err error) error {
	if err != nil && h.Logger != nil {
		h.Logger(err)
	}
	return err
}

// Global error handler that can be customized
var GlobalErrorHandler ErrorHandler = &DefaultErrorHandler{}

// SetGlobalErrorHandler sets the global error handling strategy
func SetGlobalErrorHandler(handler ErrorHandler) {
	if handler == nil {
		handler = &DefaultErrorHandler{}
	}
	GlobalErrorHandler = handler
}

// HandleError uses the global error handler to process an error
func HandleError(err error) error {
	return GlobalErrorHandler.HandleError(err)
}

// Validation helpers

// ValidateNotNil checks if a pointer is not nil
func ValidateNotNil(ptr interface{}, name string) error {
	if ptr == nil {
		return fmt.Errorf("parameter %s cannot be nil", name)
	}
	return nil
}

// ValidateSliceNotEmpty checks if a slice is not empty
func ValidateSliceNotEmpty(slice interface{}, name string) error {
	if slice == nil {
		return fmt.Errorf("parameter %s cannot be nil", name)
	}
	// TODO: Add reflection-based length check
	return nil
}

// ValidateStringNotEmpty checks if a string is not empty
func ValidateStringNotEmpty(str string, name string) error {
	if str == "" {
		return fmt.Errorf("parameter %s cannot be empty", name)
	}
	return nil
}

// ValidationError represents a parameter validation error
type ValidationError struct {
	Parameter string
	Message   string
}

// Error implements the error interface for ValidationError
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for parameter %s: %s", e.Parameter, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(parameter, message string) *ValidationError {
	return &ValidationError{
		Parameter: parameter,
		Message:   message,
	}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// Common validation functions

// ValidateInstanceConfig validates instance configuration
func ValidateInstanceConfig(config InstanceConfig) error {
	if err := ValidateStringNotEmpty(config.ApplicationName, "ApplicationName"); err != nil {
		return err
	}
	
	if config.APIVersion.Major == 0 {
		return NewValidationError("APIVersion", "Major version cannot be 0")
	}
	
	return nil
}

// ValidateDeviceConfig validates device configuration  
func ValidateDeviceConfig(config DeviceConfig) error {
	if len(config.QueueCreateInfos) == 0 {
		return NewValidationError("QueueCreateInfos", "At least one queue must be requested")
	}
	
	for i, qci := range config.QueueCreateInfos {
		if qci.QueueCount == 0 {
			return NewValidationError(fmt.Sprintf("QueueCreateInfos[%d].QueueCount", i), "Queue count cannot be 0")
		}
		
		if len(qci.QueuePriorities) != int(qci.QueueCount) {
			return NewValidationError(fmt.Sprintf("QueueCreateInfos[%d].QueuePriorities", i), 
				"Number of priorities must match queue count")
		}
		
		for j, priority := range qci.QueuePriorities {
			if priority < 0.0 || priority > 1.0 {
				return NewValidationError(fmt.Sprintf("QueueCreateInfos[%d].QueuePriorities[%d]", i, j),
					"Queue priority must be between 0.0 and 1.0")
			}
		}
	}
	
	return nil
}

// Error recovery helpers

// WithRecovery wraps a function call with panic recovery
func WithRecovery(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = fmt.Errorf("recovered from panic: %w", e)
			} else {
				err = fmt.Errorf("recovered from panic: %v", r)
			}
		}
	}()
	
	return fn()
}

// Retry executes a function with retry logic
func Retry(attempts int, fn func() error) error {
	var lastErr error
	
	for i := 0; i < attempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			
			// Check if it's a retryable error
			if vkErr, ok := err.(*VulkanError); ok {
				switch vkErr.Result {
				case vulkan.ERROR_DEVICE_LOST:
					// Device lost is not retryable
					return err
				case vulkan.ERROR_OUT_OF_DEVICE_MEMORY,
					 vulkan.ERROR_OUT_OF_HOST_MEMORY:
					// Memory errors might be temporary
					continue
				}
			}
		}
	}
	
	return fmt.Errorf("operation failed after %d attempts: %w", attempts, lastErr)
}