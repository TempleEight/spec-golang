package main

import "github.com/TempleEight/spec-golang/user/dao"

// Hook allows additional code to be executed before and after every datastore interaction
// Hooks are executed in the order they are defined, such that if any hook errors, future hooks are not executed and the request is terminated
type Hook struct {
	beforeCreateHooks []*func(env *env, req createUserRequest, input *dao.CreateUserInput) *HookError
	afterCreateHooks  []*func(env *env, user *dao.User) *HookError
}

// HookError wraps an existing error with HTTP status code
type HookError struct {
	statusCode int
	error      error
}

func (e *HookError) Error() string {
	return e.error.Error()
}

// BeforeCreate adds a new hook to be executed before creating the object in the datastore
func (h *Hook) BeforeCreate(hook func(env *env, req createUserRequest, input *dao.CreateUserInput) *HookError) {
	h.beforeCreateHooks = append(h.beforeCreateHooks, &hook)
}

// AfterCreate adds a new hook to be executed after creating the object in the datastore
func (h *Hook) AfterCreate(hook func(env *env, user *dao.User) *HookError) {
	h.afterCreateHooks = append(h.afterCreateHooks, &hook)
}
