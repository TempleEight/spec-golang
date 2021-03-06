package main

import "github.com/TempleEight/spec-golang/user/dao"

// Hook allows additional code to be executed before and after every datastore interaction
// Hooks are executed in the order they are defined, such that if any hook errors, future hooks are not executed and the request is terminated
type Hook struct {
	beforeCreateHooks        []*func(env *env, req createUserRequest, input *dao.CreateUserInput) *HookError
	beforeReadHooks          []*func(env *env, input *dao.ReadUserInput) *HookError
	beforeUpdateHooks        []*func(env *env, req updateUserRequest, input *dao.UpdateUserInput) *HookError
	beforeDeleteHooks        []*func(env *env, input *dao.DeleteUserInput) *HookError
	beforeCreatePictureHooks []*func(env *env, req createPictureRequest, input *dao.CreatePictureInput) *HookError
	beforeReadPictureHooks   []*func(env *env, input *dao.ReadPictureInput) *HookError
	beforeUpdatePictureHooks []*func(env *env, req updatePictureRequest, input *dao.UpdatePictureInput) *HookError
	beforeDeletePictureHooks []*func(env *env, input *dao.DeletePictureInput) *HookError

	afterCreateHooks        []*func(env *env, user *dao.User) *HookError
	afterReadHooks          []*func(env *env, user *dao.User) *HookError
	afterUpdateHooks        []*func(env *env, user *dao.User) *HookError
	afterDeleteHooks        []*func(env *env) *HookError
	afterCreatePictureHooks []*func(env *env, picture *dao.Picture) *HookError
	afterReadPictureHooks   []*func(env *env, picture *dao.Picture) *HookError
	afterUpdatePictureHooks []*func(env *env, picture *dao.Picture) *HookError
	afterDeletePictureHooks []*func(env *env) *HookError
}

// HookError wraps an existing error with HTTP status code
type HookError struct {
	statusCode int
	error      error
}

func (e *HookError) Error() string {
	return e.error.Error()
}

// BeforeCreate adds a new hook to be executed before creating an object in the datastore
func (h *Hook) BeforeCreate(hook func(env *env, req createUserRequest, input *dao.CreateUserInput) *HookError) {
	h.beforeCreateHooks = append(h.beforeCreateHooks, &hook)
}

// BeforeRead adds a new hook to be executed before reading an object in the datastore
func (h *Hook) BeforeRead(hook func(env *env, input *dao.ReadUserInput) *HookError) {
	h.beforeReadHooks = append(h.beforeReadHooks, &hook)
}

// BeforeUpdate adds a new hook to be executed before updating an object in the datastore
func (h *Hook) BeforeUpdate(hook func(env *env, req updateUserRequest, input *dao.UpdateUserInput) *HookError) {
	h.beforeUpdateHooks = append(h.beforeUpdateHooks, &hook)
}

// BeforeDelete adds a new hook to be executed before deleting an object in the datastore
func (h *Hook) BeforeDelete(hook func(env *env, input *dao.DeleteUserInput) *HookError) {
	h.beforeDeleteHooks = append(h.beforeDeleteHooks, &hook)
}

// BeforeCreatePicture adds a new hook to be executed before creating an object in the datastore
func (h *Hook) BeforeCreatePicture(hook func(env *env, req createPictureRequest, input *dao.CreatePictureInput) *HookError) {
	h.beforeCreatePictureHooks = append(h.beforeCreatePictureHooks, &hook)
}

// BeforeReadPicture adds a new hook to be executed before reading an object in the datastore
func (h *Hook) BeforeReadPicture(hook func(env *env, input *dao.ReadPictureInput) *HookError) {
	h.beforeReadPictureHooks = append(h.beforeReadPictureHooks, &hook)
}

// BeforeUpdatePicture adds a new hook to be executed before reading an object in the datastore
func (h *Hook) BeforeUpdatePicture(hook func(env *env, req updatePictureRequest, input *dao.UpdatePictureInput) *HookError) {
	h.beforeUpdatePictureHooks = append(h.beforeUpdatePictureHooks, &hook)
}

// BeforeDeletePicture adds a new hook to be executed before deletin an object in the datastore
func (h *Hook) BeforeDeletePicture(hook func(env *env, input *dao.DeletePictureInput) *HookError) {
	h.beforeDeletePictureHooks = append(h.beforeDeletePictureHooks, &hook)
}

// AfterCreate adds a new hook to be executed after creating an object in the datastore
func (h *Hook) AfterCreate(hook func(env *env, user *dao.User) *HookError) {
	h.afterCreateHooks = append(h.afterCreateHooks, &hook)
}

// AfterRead adds a new hook to be executed after reading an object in the datastore
func (h *Hook) AfterRead(hook func(env *env, user *dao.User) *HookError) {
	h.afterReadHooks = append(h.afterReadHooks, &hook)
}

// AfterUpdate adds a new hook to be executed after updating an object in the datastore
func (h *Hook) AfterUpdate(hook func(env *env, user *dao.User) *HookError) {
	h.afterUpdateHooks = append(h.afterUpdateHooks, &hook)
}

// AfterDelete adds a new hook to be executed after deleting an object in the datastore
func (h *Hook) AfterDelete(hook func(env *env) *HookError) {
	h.afterDeleteHooks = append(h.afterDeleteHooks, &hook)
}

// AfterCreatePicture adds a new hook to be executed after creating an object in the datastore
func (h *Hook) AfterCreatePicture(hook func(env *env, user *dao.Picture) *HookError) {
	h.afterCreatePictureHooks = append(h.afterCreatePictureHooks, &hook)
}

// AfterReadPicture adds a new hook to be executed after reading an object in the datastore
func (h *Hook) AfterReadPicture(hook func(env *env, user *dao.Picture) *HookError) {
	h.afterReadPictureHooks = append(h.afterReadPictureHooks, &hook)
}

// AfterUpdatePicture adds a new hook to be executed after reading an object in the datastore
func (h *Hook) AfterUpdatePicture(hook func(env *env, user *dao.Picture) *HookError) {
	h.afterUpdatePictureHooks = append(h.afterUpdatePictureHooks, &hook)
}

// AfterDeletePicture adds a new hook to be executed after deleting an object in the datastore
func (h *Hook) AfterDeletePicture(hook func(env *env) *HookError) {
	h.afterDeletePictureHooks = append(h.afterDeletePictureHooks, &hook)
}
