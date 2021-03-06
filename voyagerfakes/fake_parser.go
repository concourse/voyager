// Code generated by counterfeiter. DO NOT EDIT.
package voyagerfakes

import (
	"sync"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/voyager"
)

type FakeParser struct {
	ParseFileToMigrationStub        func(lager.Logger, string) (voyager.Migration, error)
	parseFileToMigrationMutex       sync.RWMutex
	parseFileToMigrationArgsForCall []struct {
		arg1 lager.Logger
		arg2 string
	}
	parseFileToMigrationReturns struct {
		result1 voyager.Migration
		result2 error
	}
	parseFileToMigrationReturnsOnCall map[int]struct {
		result1 voyager.Migration
		result2 error
	}
	ParseMigrationFilenameStub        func(lager.Logger, string) (voyager.Migration, error)
	parseMigrationFilenameMutex       sync.RWMutex
	parseMigrationFilenameArgsForCall []struct {
		arg1 lager.Logger
		arg2 string
	}
	parseMigrationFilenameReturns struct {
		result1 voyager.Migration
		result2 error
	}
	parseMigrationFilenameReturnsOnCall map[int]struct {
		result1 voyager.Migration
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeParser) ParseFileToMigration(arg1 lager.Logger, arg2 string) (voyager.Migration, error) {
	fake.parseFileToMigrationMutex.Lock()
	ret, specificReturn := fake.parseFileToMigrationReturnsOnCall[len(fake.parseFileToMigrationArgsForCall)]
	fake.parseFileToMigrationArgsForCall = append(fake.parseFileToMigrationArgsForCall, struct {
		arg1 lager.Logger
		arg2 string
	}{arg1, arg2})
	fake.recordInvocation("ParseFileToMigration", []interface{}{arg1, arg2})
	fake.parseFileToMigrationMutex.Unlock()
	if fake.ParseFileToMigrationStub != nil {
		return fake.ParseFileToMigrationStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.parseFileToMigrationReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeParser) ParseFileToMigrationCallCount() int {
	fake.parseFileToMigrationMutex.RLock()
	defer fake.parseFileToMigrationMutex.RUnlock()
	return len(fake.parseFileToMigrationArgsForCall)
}

func (fake *FakeParser) ParseFileToMigrationCalls(stub func(lager.Logger, string) (voyager.Migration, error)) {
	fake.parseFileToMigrationMutex.Lock()
	defer fake.parseFileToMigrationMutex.Unlock()
	fake.ParseFileToMigrationStub = stub
}

func (fake *FakeParser) ParseFileToMigrationArgsForCall(i int) (lager.Logger, string) {
	fake.parseFileToMigrationMutex.RLock()
	defer fake.parseFileToMigrationMutex.RUnlock()
	argsForCall := fake.parseFileToMigrationArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeParser) ParseFileToMigrationReturns(result1 voyager.Migration, result2 error) {
	fake.parseFileToMigrationMutex.Lock()
	defer fake.parseFileToMigrationMutex.Unlock()
	fake.ParseFileToMigrationStub = nil
	fake.parseFileToMigrationReturns = struct {
		result1 voyager.Migration
		result2 error
	}{result1, result2}
}

func (fake *FakeParser) ParseFileToMigrationReturnsOnCall(i int, result1 voyager.Migration, result2 error) {
	fake.parseFileToMigrationMutex.Lock()
	defer fake.parseFileToMigrationMutex.Unlock()
	fake.ParseFileToMigrationStub = nil
	if fake.parseFileToMigrationReturnsOnCall == nil {
		fake.parseFileToMigrationReturnsOnCall = make(map[int]struct {
			result1 voyager.Migration
			result2 error
		})
	}
	fake.parseFileToMigrationReturnsOnCall[i] = struct {
		result1 voyager.Migration
		result2 error
	}{result1, result2}
}

func (fake *FakeParser) ParseMigrationFilename(arg1 lager.Logger, arg2 string) (voyager.Migration, error) {
	fake.parseMigrationFilenameMutex.Lock()
	ret, specificReturn := fake.parseMigrationFilenameReturnsOnCall[len(fake.parseMigrationFilenameArgsForCall)]
	fake.parseMigrationFilenameArgsForCall = append(fake.parseMigrationFilenameArgsForCall, struct {
		arg1 lager.Logger
		arg2 string
	}{arg1, arg2})
	fake.recordInvocation("ParseMigrationFilename", []interface{}{arg1, arg2})
	fake.parseMigrationFilenameMutex.Unlock()
	if fake.ParseMigrationFilenameStub != nil {
		return fake.ParseMigrationFilenameStub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.parseMigrationFilenameReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeParser) ParseMigrationFilenameCallCount() int {
	fake.parseMigrationFilenameMutex.RLock()
	defer fake.parseMigrationFilenameMutex.RUnlock()
	return len(fake.parseMigrationFilenameArgsForCall)
}

func (fake *FakeParser) ParseMigrationFilenameCalls(stub func(lager.Logger, string) (voyager.Migration, error)) {
	fake.parseMigrationFilenameMutex.Lock()
	defer fake.parseMigrationFilenameMutex.Unlock()
	fake.ParseMigrationFilenameStub = stub
}

func (fake *FakeParser) ParseMigrationFilenameArgsForCall(i int) (lager.Logger, string) {
	fake.parseMigrationFilenameMutex.RLock()
	defer fake.parseMigrationFilenameMutex.RUnlock()
	argsForCall := fake.parseMigrationFilenameArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeParser) ParseMigrationFilenameReturns(result1 voyager.Migration, result2 error) {
	fake.parseMigrationFilenameMutex.Lock()
	defer fake.parseMigrationFilenameMutex.Unlock()
	fake.ParseMigrationFilenameStub = nil
	fake.parseMigrationFilenameReturns = struct {
		result1 voyager.Migration
		result2 error
	}{result1, result2}
}

func (fake *FakeParser) ParseMigrationFilenameReturnsOnCall(i int, result1 voyager.Migration, result2 error) {
	fake.parseMigrationFilenameMutex.Lock()
	defer fake.parseMigrationFilenameMutex.Unlock()
	fake.ParseMigrationFilenameStub = nil
	if fake.parseMigrationFilenameReturnsOnCall == nil {
		fake.parseMigrationFilenameReturnsOnCall = make(map[int]struct {
			result1 voyager.Migration
			result2 error
		})
	}
	fake.parseMigrationFilenameReturnsOnCall[i] = struct {
		result1 voyager.Migration
		result2 error
	}{result1, result2}
}

func (fake *FakeParser) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.parseFileToMigrationMutex.RLock()
	defer fake.parseFileToMigrationMutex.RUnlock()
	fake.parseMigrationFilenameMutex.RLock()
	defer fake.parseMigrationFilenameMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeParser) recordInvocation(key string, args []interface{}) {
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

var _ voyager.Parser = new(FakeParser)
