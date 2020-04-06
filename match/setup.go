package main

import (
	"time"

	"github.com/TempleEight/spec-golang/match/dao"
	"github.com/gorilla/mux"
)

func (e *env) setup(router *mux.Router) {
	// Add user defined code here
	e.hook.BeforeCreate(beforeCreateMatch)
}

func beforeCreateMatch(env *env, req createMatchRequest, input *dao.CreateMatchInput) *HookError {
	input.MatchedOn = time.Now()
	return nil
}
