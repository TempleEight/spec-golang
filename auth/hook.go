package main
import "github.com/TempleEight/spec-golang/auth/dao"

// Hook allows additional code to be executed before and after every datastore interaction
// Hooks are executed in the order they are defined, such that if any hook errors, future hooks are not executed and the request is terminated
type Hook struct {
	beforeCreateHooks []*func(env *env, req registerAuthRequest, input *dao.CreateAuthInput) *HookError
	beforeReadHooks   []*func(env *env, req loginAuthRequest, input *dao.ReadAuthInput) *HookError

	afterCreateHooks []*func(env *env, auth *dao.Auth, accessToken string) *HookError
	afterReadHooks   []*func(env *env, auth *dao.Auth, accessToken string) *HookError
}

// HookError wraps an existing error with HTTP status code
type HookError struct {
	statusCode int
	error      error
}

func (e *HookError) Error() string {
	return e.error.Error()
}

// BeforeCreate adds a new hook to be executed before creating an object in the data store
func (h *Hook) BeforeCreate(hook func(env *env, req registerAuthRequest, input *dao.CreateAuthInput) *HookError) {
	h.beforeCreateHooks = append(h.beforeCreateHooks, &hook)
}

// BeforeRead adds a new hook to be executed before reading an object in the datastore
func (h *Hook) BeforeRead(hook func(env *env, req loginAuthRequest, input *dao.ReadAuthInput) *HookError) {
	h.beforeReadHooks = append(h.beforeReadHooks, &hook)
}

// AfterCreate adds a new hook to be executed after creating an object in the datastore
func (h *Hook) AfterCreate(hook func(env *env, auth *dao.Auth, accessToken string) *HookError) {
	h.afterCreateHooks = append(h.afterCreateHooks, &hook)
}

// AfterRead adds a new hook to be executed after reading an object in the datastore
func (h *Hook) AfterRead(hook func(env *env, auth *dao.Auth, accessToken string) *HookError) {
	h.afterReadHooks = append(h.afterReadHooks, &hook)
}

