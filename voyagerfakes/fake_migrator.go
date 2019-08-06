// Code generated by counterfeiter. DO NOT EDIT.
package voyagerfakes

import (
	"database/sql"
	"sync"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/voyager"
)

type FakeMigrator struct {
	CurrentVersionStub        func(lager.Logger, *sql.DB) (int, error)
	currentVersionMutex       sync.RWMutex
	currentVersionArgsForCall []struct {
		arg1 lager.Logger
		arg2 *sql.DB
	}
	currentVersionReturns struct {
		result1 int
		result2 error
	}
	currentVersionReturnsOnCall map[int]struct {
		result1 int
		result2 error
	}
	MigrateStub        func(lager.Logger, *sql.DB, int) error
	migrateMutex       sync.RWMutex
	migrateArgsForCall []struct {
		arg1 lager.Logger
		arg2 *sql.DB
		arg3 int
	}
	migrateReturns struct {
		result1 error
	}
	migrateReturnsOnCall map[int]struct {
		result1 error
	}
	SupportedVersionStub        func(lager.Logger) int
	supportedVersionMutex       sync.RWMutex
	supportedVersionArgsForCall []struct {
		arg1 lager.Logger
	}
	supportedVersionReturns struct {
		result1 int
	}
	supportedVersionReturnsOnCall map[int]struct {
		result1 int
	}
	UpStub        func(lager.Logger, *sql.DB) error
	upMutex       sync.RWMutex
	upArgsForCall []struct {
		arg1 lager.Logger
		arg2 *sql.DB
	}
	upReturns struct {
		result1 error
	}
	upReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeMigrator) CurrentVersion(arg1 lager.Logger, arg2 *sql.DB) (int, error) {
	fake.currentVersionMutex.Lock()
	ret, specificReturn := fake.currentVersionReturnsOnCall[len(fake.currentVersionArgsForCall)]
	fake.currentVersionArgsForCall = append(fake.currentVersionArgsForCall, struct {
		arg1 lager.Logger
		arg2 *sql.DB
	}{arg1, arg2})
	fake.recordInvocation("CurrentVersion", []interface{}{arg1, arg2})
	fake.currentVersionMutex.Unlock()
	if fake.CurrentVersionStub != nil {
		return fake.CurrentVersionStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.currentVersionReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeMigrator) CurrentVersionCallCount() int {
	fake.currentVersionMutex.RLock()
	defer fake.currentVersionMutex.RUnlock()
	return len(fake.currentVersionArgsForCall)
}

func (fake *FakeMigrator) CurrentVersionCalls(stub func(lager.Logger, *sql.DB) (int, error)) {
	fake.currentVersionMutex.Lock()
	defer fake.currentVersionMutex.Unlock()
	fake.CurrentVersionStub = stub
}

func (fake *FakeMigrator) CurrentVersionArgsForCall(i int) (lager.Logger, *sql.DB) {
	fake.currentVersionMutex.RLock()
	defer fake.currentVersionMutex.RUnlock()
	argsForCall := fake.currentVersionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeMigrator) CurrentVersionReturns(result1 int, result2 error) {
	fake.currentVersionMutex.Lock()
	defer fake.currentVersionMutex.Unlock()
	fake.CurrentVersionStub = nil
	fake.currentVersionReturns = struct {
		result1 int
		result2 error
	}{result1, result2}
}

func (fake *FakeMigrator) CurrentVersionReturnsOnCall(i int, result1 int, result2 error) {
	fake.currentVersionMutex.Lock()
	defer fake.currentVersionMutex.Unlock()
	fake.CurrentVersionStub = nil
	if fake.currentVersionReturnsOnCall == nil {
		fake.currentVersionReturnsOnCall = make(map[int]struct {
			result1 int
			result2 error
		})
	}
	fake.currentVersionReturnsOnCall[i] = struct {
		result1 int
		result2 error
	}{result1, result2}
}

func (fake *FakeMigrator) Migrate(arg1 lager.Logger, arg2 *sql.DB, arg3 int) error {
	fake.migrateMutex.Lock()
	ret, specificReturn := fake.migrateReturnsOnCall[len(fake.migrateArgsForCall)]
	fake.migrateArgsForCall = append(fake.migrateArgsForCall, struct {
		arg1 lager.Logger
		arg2 *sql.DB
		arg3 int
	}{arg1, arg2, arg3})
	fake.recordInvocation("Migrate", []interface{}{arg1, arg2, arg3})
	fake.migrateMutex.Unlock()
	if fake.MigrateStub != nil {
		return fake.MigrateStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.migrateReturns
	return fakeReturns.result1
}

func (fake *FakeMigrator) MigrateCallCount() int {
	fake.migrateMutex.RLock()
	defer fake.migrateMutex.RUnlock()
	return len(fake.migrateArgsForCall)
}

func (fake *FakeMigrator) MigrateCalls(stub func(lager.Logger, *sql.DB, int) error) {
	fake.migrateMutex.Lock()
	defer fake.migrateMutex.Unlock()
	fake.MigrateStub = stub
}

func (fake *FakeMigrator) MigrateArgsForCall(i int) (lager.Logger, *sql.DB, int) {
	fake.migrateMutex.RLock()
	defer fake.migrateMutex.RUnlock()
	argsForCall := fake.migrateArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeMigrator) MigrateReturns(result1 error) {
	fake.migrateMutex.Lock()
	defer fake.migrateMutex.Unlock()
	fake.MigrateStub = nil
	fake.migrateReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeMigrator) MigrateReturnsOnCall(i int, result1 error) {
	fake.migrateMutex.Lock()
	defer fake.migrateMutex.Unlock()
	fake.MigrateStub = nil
	if fake.migrateReturnsOnCall == nil {
		fake.migrateReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.migrateReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeMigrator) SupportedVersion(arg1 lager.Logger) int {
	fake.supportedVersionMutex.Lock()
	ret, specificReturn := fake.supportedVersionReturnsOnCall[len(fake.supportedVersionArgsForCall)]
	fake.supportedVersionArgsForCall = append(fake.supportedVersionArgsForCall, struct {
		arg1 lager.Logger
	}{arg1})
	fake.recordInvocation("SupportedVersion", []interface{}{arg1})
	fake.supportedVersionMutex.Unlock()
	if fake.SupportedVersionStub != nil {
		return fake.SupportedVersionStub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.supportedVersionReturns
	return fakeReturns.result1
}

func (fake *FakeMigrator) SupportedVersionCallCount() int {
	fake.supportedVersionMutex.RLock()
	defer fake.supportedVersionMutex.RUnlock()
	return len(fake.supportedVersionArgsForCall)
}

func (fake *FakeMigrator) SupportedVersionCalls(stub func(lager.Logger) int) {
	fake.supportedVersionMutex.Lock()
	defer fake.supportedVersionMutex.Unlock()
	fake.SupportedVersionStub = stub
}

func (fake *FakeMigrator) SupportedVersionArgsForCall(i int) lager.Logger {
	fake.supportedVersionMutex.RLock()
	defer fake.supportedVersionMutex.RUnlock()
	argsForCall := fake.supportedVersionArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeMigrator) SupportedVersionReturns(result1 int) {
	fake.supportedVersionMutex.Lock()
	defer fake.supportedVersionMutex.Unlock()
	fake.SupportedVersionStub = nil
	fake.supportedVersionReturns = struct {
		result1 int
	}{result1}
}

func (fake *FakeMigrator) SupportedVersionReturnsOnCall(i int, result1 int) {
	fake.supportedVersionMutex.Lock()
	defer fake.supportedVersionMutex.Unlock()
	fake.SupportedVersionStub = nil
	if fake.supportedVersionReturnsOnCall == nil {
		fake.supportedVersionReturnsOnCall = make(map[int]struct {
			result1 int
		})
	}
	fake.supportedVersionReturnsOnCall[i] = struct {
		result1 int
	}{result1}
}

func (fake *FakeMigrator) Up(arg1 lager.Logger, arg2 *sql.DB) error {
	fake.upMutex.Lock()
	ret, specificReturn := fake.upReturnsOnCall[len(fake.upArgsForCall)]
	fake.upArgsForCall = append(fake.upArgsForCall, struct {
		arg1 lager.Logger
		arg2 *sql.DB
	}{arg1, arg2})
	fake.recordInvocation("Up", []interface{}{arg1, arg2})
	fake.upMutex.Unlock()
	if fake.UpStub != nil {
		return fake.UpStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.upReturns
	return fakeReturns.result1
}

func (fake *FakeMigrator) UpCallCount() int {
	fake.upMutex.RLock()
	defer fake.upMutex.RUnlock()
	return len(fake.upArgsForCall)
}

func (fake *FakeMigrator) UpCalls(stub func(lager.Logger, *sql.DB) error) {
	fake.upMutex.Lock()
	defer fake.upMutex.Unlock()
	fake.UpStub = stub
}

func (fake *FakeMigrator) UpArgsForCall(i int) (lager.Logger, *sql.DB) {
	fake.upMutex.RLock()
	defer fake.upMutex.RUnlock()
	argsForCall := fake.upArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeMigrator) UpReturns(result1 error) {
	fake.upMutex.Lock()
	defer fake.upMutex.Unlock()
	fake.UpStub = nil
	fake.upReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeMigrator) UpReturnsOnCall(i int, result1 error) {
	fake.upMutex.Lock()
	defer fake.upMutex.Unlock()
	fake.UpStub = nil
	if fake.upReturnsOnCall == nil {
		fake.upReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.upReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeMigrator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.currentVersionMutex.RLock()
	defer fake.currentVersionMutex.RUnlock()
	fake.migrateMutex.RLock()
	defer fake.migrateMutex.RUnlock()
	fake.supportedVersionMutex.RLock()
	defer fake.supportedVersionMutex.RUnlock()
	fake.upMutex.RLock()
	defer fake.upMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeMigrator) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ voyager.Migrator = new(FakeMigrator)
