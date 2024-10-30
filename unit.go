package gomake

import (
	"strings"
	"sync"
)

var unit_cache = make(map[string]*Unit)
var unit_lock sync.Mutex

// AddUnit defines a new build unit by the specified @name and @handler.
//
// @isDefault determines if this unit should run by default when the build
// program is not given any specific units to run.
func AddUnit(name string, handler UnitFunc, isDefault bool) *Unit {
	unit_lock.Lock()
	defer unit_lock.Unlock()
	unit_cache[strings.ToLower(name)] = &Unit{
		handler:    handler,
		isDefault:  isDefault,
		isFinished: false,
	}
	return unit_cache[strings.ToLower(name)]
}

// UnitFunc defines a unit handler.
type UnitFunc func()

// Unit defines a complete build unit.
// It has a build handler and optionally a set of units it depends on.
type Unit struct {
	lock         sync.Mutex
	handler      UnitFunc
	dependencies []*Unit
	isFinished   bool
	isDefault    bool
}

// IsDefault returns true if this unit should run by default when
// the build program is not given any specific units to run.
func (u *Unit) IsDefault() bool {
	u.lock.Lock()
	v := u.isDefault
	u.lock.Unlock()
	return v
}

// SetDefault determines if this unit should run by default when
// the build program is not given any specific units to run.
func (u *Unit) SetDefault(v bool) {
	u.lock.Lock()
	u.isDefault = v
	u.lock.Unlock()
}

// IsFinished returns true if this unit has run to completion.
func (u *Unit) IsFinished() bool {
	u.lock.Lock()
	v := u.isFinished
	u.lock.Unlock()
	return v
}

// AddDependency adds the given unit as a direct dependency of this unit.
// If this unit is built, the lobrary ensures that the given dependency
// is completed first.
//
// This call fails if adding the unit results in a circular dependency
// somewhere in the dependency graph.
func (u *Unit) AddDependency(dep *Unit) {
	u.lock.Lock()
	if !u.hasDependencyRec(dep) {
		u.dependencies = append(u.dependencies, dep)
	}
	u.lock.Unlock()
}

// HasDependency returns true if this unit (directly or indirectly) depends on @dep.
// This searches the entire dependency graph.
//
// Returns true if @dep is the same as this unit.
func (u *Unit) HasDependency(dep *Unit) bool {
	if dep == u {
		return true
	}

	u.lock.Lock()
	defer u.lock.Unlock()
	return u.hasDependencyRec(dep)
}

func (u *Unit) hasDependencyRec(want *Unit) bool {
	for _, have := range u.dependencies {
		if have == want {
			return true
		}

		if have.HasDependency(want) {
			return true
		}
	}
	return false
}

// Run runs the unit if it isn't already finished.
// This will recusrively run all the unit's dependencies as needed.
func (u *Unit) Run() {
	u.lock.Lock()
	defer u.lock.Unlock()

	for _, dep := range u.dependencies {
		dep.Run()
	}

	if !u.isFinished {
		u.handler()
		u.isFinished = true
	}
}
